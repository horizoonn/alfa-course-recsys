import hashlib
import json
import inspect
import difflib
import re
import threading
import time
from importlib.metadata import version
from pathlib import Path

import numpy as np
from scipy.sparse import load_npz
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.preprocessing import normalize as normalize_matrix

import semantic
from intent import parse_intent, CONTEXTS

SCHEMA = 2
ALGORITHM = "query-search-v3"
from text import GROUPS, WEIGHTS, WORDS, SHORT, STOP, stem, tokens, normalize


def pipeline_id():
    digest = hashlib.sha256()
    digest.update(
        json.dumps(
            {
                "schema": SCHEMA,
                "groups": GROUPS,
                "weights": WEIGHTS,
                "short": sorted(SHORT),
                "stop": sorted(STOP),
                "words": WORDS.pattern,
            },
            sort_keys=True,
        ).encode()
    )
    for function in (stem, tokens, normalize, groups_for, vectorizer, char_vectorizer):
        digest.update(inspect.getsource(function).encode())
    for name in ["scikit-learn", "nltk", "fastembed"]:
        digest.update(version(name).encode())
    return digest.hexdigest()


def embedding_id(cards):
    return hashlib.sha256(
        json.dumps(
            {
                "cards": cards,
                "model": semantic.MODEL,
                "revision": semantic.REVISION,
                "sha256": semantic.MODEL_SHA256,
                "prefix": "passage: ",
                "pooling": "mean",
                "normalization": True,
            },
            ensure_ascii=False,
            sort_keys=True,
        ).encode()
    ).hexdigest()


def groups_for(corpus):
    groups = [[], [], []]
    cards = []
    for document in corpus["documents"]:
        parts = [[], [], []]
        fields = document["fields"]
        for field in fields:
            if field["name"] == "promotion":
                continue
            if field["name"] not in GROUPS:
                raise ValueError("unknown text field role")
            parts[GROUPS[field["name"]]].extend(tokens(field["text"]))
        for group, words in zip(groups, parts):
            group.append(" ".join(words))
        cards.append(
            "\n".join(
                field["text"]
                for role in ("title", "summary", "metadata")
                for field in fields
                if field["name"] == role
            )
        )
    return groups, cards


def vectorizer(vocabulary=None):
    return TfidfVectorizer(
        analyzer=str.split,
        sublinear_tf=True,
        smooth_idf=True,
        norm="l2",
        vocabulary=vocabulary,
    )


def char_vectorizer(vocabulary=None):
    return TfidfVectorizer(
        analyzer="char_wb",
        ngram_range=(3, 5),
        sublinear_tf=True,
        norm="l2",
        max_features=50000,
        vocabulary=vocabulary,
    )


def digest_file(path):
    digest = hashlib.sha256()
    with Path(path).open("rb") as source:
        for chunk in iter(lambda: source.read(1 << 20), b""):
            digest.update(chunk)
    return digest.hexdigest()


class SearchIndex:
    def __init__(self, directory, enable_semantic=True):
        self.directory = Path(directory)
        self.manifest = json.loads((self.directory / "current.json").read_text())
        manifest = self.manifest
        if (
            manifest["schema_version"] != SCHEMA
            or manifest["pipeline_id"] != pipeline_id()
        ):
            raise ValueError("search index version mismatch; run preparation")
        generation = manifest["generation"]
        if not re.fullmatch(r"generation-[a-z0-9_]+", generation):
            raise ValueError("invalid search generation")
        folder = self.directory / generation
        if not {"metadata.json", "lexical.npz", "character.npz"}.issubset(
            manifest["files"]
        ):
            raise ValueError("search index files mismatch")
        self.semantic_valid = bool(
            manifest["semantic"] and "semantic.npy" in manifest["files"]
        )
        for name, checksum in manifest["files"].items():
            if name not in {
                "metadata.json",
                "lexical.npz",
                "character.npz",
                "semantic.npy",
            }:
                raise ValueError("unknown index file")
            try:
                valid = digest_file(folder / name) == checksum
            except OSError:
                valid = False
            if name == "semantic.npy":
                self.semantic_valid = self.semantic_valid and valid
            elif not valid:
                raise ValueError("search index checksum mismatch")
        metadata = json.loads((folder / "metadata.json").read_text())
        self.keys = metadata["keys"]
        if len(self.keys) != len(set(self.keys)):
            raise ValueError("duplicate search keys")
        self.vectorizer = vectorizer(metadata["vocabulary"])
        self.vectorizer.idf_ = np.asarray(metadata["idf"], dtype=np.float64)
        self.matrix = load_npz(folder / "lexical.npz")
        if (
            self.matrix.shape != (len(self.keys), len(metadata["vocabulary"]))
            or not np.isfinite(self.matrix.data).all()
        ):
            raise ValueError("invalid lexical matrix")
        self.related_matrix = normalize_matrix(self.matrix)
        self.cards = metadata["cards"]
        if len(self.cards) != len(self.keys):
            raise ValueError("invalid card texts")
        self.character = char_vectorizer(metadata["character_vocabulary"])
        self.character.idf_ = np.asarray(metadata["character_idf"], dtype=np.float64)
        self.character_matrix = load_npz(folder / "character.npz")
        if (
            self.character_matrix.shape
            != (len(self.keys), len(metadata["character_vocabulary"]))
            or not np.isfinite(self.character_matrix.data).all()
        ):
            raise ValueError("invalid character matrix")
        self.headings = metadata["headings"]
        if len(self.headings) != len(self.keys):
            raise ValueError("invalid card headings")
        self.card_terms = [set(tokens(card)) for card in self.headings]
        self.title_terms = [set(tokens(card.split("\n", 1)[0])) for card in self.cards]
        self.title_matrix = self.vectorizer.transform([" ".join(sorted(terms)) for terms in self.title_terms])
        self.spelling = {}
        for word in self.vectorizer.vocabulary_:
            if 4 <= len(word) <= 20:
                self.spelling.setdefault((word[0], len(word)), []).append(word)
        self.semantic = None
        self.encoder = None
        self.semantic_error = manifest.get("semantic_error", "model_not_prepared")
        self.lock = threading.Lock()
        self.semantic_valid = self.semantic_valid and manifest.get(
            "semantic_id"
        ) == embedding_id(self.cards)
        if not enable_semantic:
            self.semantic_error = "semantic_disabled"
        elif manifest["semantic"]:
            try:
                if not self.semantic_valid:
                    raise ValueError("invalid semantic index")
                vectors = np.load(folder / "semantic.npy", allow_pickle=False)
                if (
                    vectors.shape != (len(self.keys), 384)
                    or not np.isfinite(vectors).all()
                ):
                    raise ValueError("invalid semantic matrix")
                self.encoder = semantic.encoder(self.directory / "model")
                self.semantic = vectors
                self.semantic_error = ""
            except Exception:
                self.semantic_error = "semantic_load_failed"

    def search(self, query, mode):
        started = time.perf_counter()
        corrections = {}
        attempts = 0
        for word in dict.fromkeys(WORDS.findall(normalize(query))):
            term = stem(word)
            if (
                term in self.vectorizer.vocabulary_
                or term in STOP
                or not 4 <= len(term) <= 20
            ):
                continue
            if attempts >= 6:
                break
            attempts += 1
            candidates = [
                candidate
                for length in range(len(term) - 1, len(term) + 2)
                for candidate in self.spelling.get((term[0], length), ())
            ]
            closest = difflib.get_close_matches(term, candidates, n=2, cutoff=0.82)
            if closest and (
                len(closest) == 1
                or difflib.SequenceMatcher(None, term, closest[0]).ratio()
                - difflib.SequenceMatcher(None, term, closest[1]).ratio()
                >= 0.025
            ):
                corrections[word] = closest[0]
        corrected = WORDS.sub(
            lambda match: corrections.get(match.group(), match.group()),
            normalize(query),
        )
        intent = parse_intent(corrected)
        result = {
            "catalog_id": self.manifest["catalog_id"],
            "index_id": self.manifest["generation"],
            "algorithm": ALGORITHM + "/" + mode,
            "mode": mode,
            "state": "ok",
            "message": "",
            "matches": [],
            "limitations": [],
            "free_only": intent.free_only,
            "exclude_kinds": intent.excluded_kinds,
            "allowed_kinds": intent.allowed_kinds,
            "query_view": {**intent.diagnostic(), "corrections": corrections},
            "elapsed_ms": 0.0,
        }
        if intent.clarification:
            result.update(state="unclear", message=intent.clarification)
            return result
        query_words = sorted(set(tokens(intent.positive)))
        features = sorted(
            {
                self.vectorizer.vocabulary_[word]
                for word in query_words
                if word in self.vectorizer.vocabulary_
            }
        )
        coverage = np.asarray((self.matrix[:, features] > 0).sum(axis=1)).ravel() / max(
            1, len(query_words)
        )
        query_vector = self.vectorizer.transform([" ".join(tokens(intent.search_text))])
        word = np.clip((self.matrix @ query_vector.T).toarray().ravel(), 0, 1)
        character = np.clip(
            (
                self.character_matrix
                @ self.character.transform([normalize(intent.positive)]).T
            )
            .toarray()
            .ravel(),
            0,
            1,
        )
        generic = set(
            tokens(
                "настроить увеличить повысить открыть зарегистрировать найти начать сделать вести получить работа составить такой такое шаблон образец статья курс первый открыл посещаемость мало много больше меньше"
            )
        )
        subjects = set(query_words) - generic
        focus = set(subjects)
        if intent.goals:
            roots = [root for goal in intent.goals for root in (*goal.subjects, *goal.actions)]
            focus = {
                term for term in focus
                if not any(term.startswith(root) for root in roots)
                and not any(re.search(pattern, term) for pattern in CONTEXTS.values())
            }
        focus = {term for term in focus if term in self.vectorizer.vocabulary_}
        if intent.goals and (len(focus) < 2 or not any(focus <= title for title in self.title_terms)):
            focus = set()
        focus_features = [self.vectorizer.vocabulary_[term] for term in focus]
        focus_coverage = np.asarray((self.matrix[:, focus_features] > 0).sum(axis=1)).ravel() / max(1, len(focus))
        heading_focus = np.array([len(focus & terms) / max(1, len(focus)) for terms in self.card_terms])
        subject_fit = np.array([bool(subjects & terms) for terms in self.card_terms])
        goal_fit = np.ones(len(self.keys))
        matched_goals = [[] for _ in self.keys]
        if intent.goals:
            for i, terms in enumerate(self.card_terms):
                subject_only = False
                for goal in intent.goals:
                    subject = any(
                        term.startswith(root)
                        for term in terms
                        for root in goal.subjects
                    )
                    if goal.name == "registration":
                        subject = bool({"ип", "ооо"} & self.title_terms[i])
                    subject_only = subject_only or subject
                    action = any(
                        term.startswith(root) for term in terms for root in goal.actions
                    )
                    if subject and action:
                        matched_goals[i].append(goal.name)
                goal_fit[i] = 1.0 if matched_goals[i] else 0.5 if subject_only else 0.15
        if focus:
            goal_fit = np.maximum(goal_fit, heading_focus)
        lexical = np.maximum(0.85 * word + 0.15 * character, 0.5 * character) * goal_fit
        if focus:
            focused_query = self.vectorizer.transform([" ".join(sorted(focus))])
            focused_word = np.clip((self.matrix @ focused_query.T).toarray().ravel(), 0, 1)
            title_score = np.clip((self.title_matrix @ focused_query.T).toarray().ravel(), 0, 1)
            lexical = (0.25 * lexical + 0.45 * focused_word + 0.3 * title_score) * (0.4 + 0.6 * heading_focus)
        lexical_support = (
            (word >= 0.06)
            & ((coverage >= 0.6) | (goal_fit == 1.0 if intent.goals else False))
        ) | ((character >= 0.24) & subject_fit)
        if intent.goals:
            lexical_support &= goal_fit >= 0.5
        else:
            lexical_support &= subject_fit
        if focus:
            required_coverage = 1.0 if len(focus) <= 4 else 0.75
            lexical_support &= focus_coverage >= required_coverage
        semantic_scores = None
        if mode in {"e5", "rrf"}:
            if self.encoder is not None:
                try:
                    with self.lock:
                        embedding = np.asarray(
                            list(
                                self.encoder.embed(
                                    ["query: " + intent.search_text],
                                    batch_size=1,
                                    parallel=None,
                                )
                            ),
                            dtype=np.float32,
                        )[0]
                    if embedding.shape != (384,) or not np.isfinite(embedding).all():
                        raise ValueError("invalid query embedding")
                    semantic_scores = np.clip(self.semantic @ embedding, 0, 1)
                except Exception:
                    semantic_scores = None
            if semantic_scores is None:
                result["limitations"].append(
                    {
                        "code": "SEMANTIC_UNAVAILABLE",
                        "message": "Семантический поиск недоступен. Использовано текстовое совпадение.",
                    }
                )
                mode = "tfidf"
                result.update(mode=mode, algorithm=ALGORITHM + "/tfidf")
        admitted = lexical_support.copy()
        if semantic_scores is not None:
            semantic_support = (semantic_scores >= 0.78) & subject_fit
            if intent.goals:
                semantic_support |= (semantic_scores >= 0.78) & (goal_fit >= 0.5)
                semantic_support &= goal_fit >= 0.5
            peak = np.sort(semantic_scores)
            if len(peak) > 1 and peak[-1] >= 0.92 and peak[-1] - peak[-2] >= 0.025:
                semantic_support |= semantic_scores == peak[-1]
            if focus:
                semantic_support &= focus_coverage >= required_coverage
            admitted |= semantic_support
        for i, key in enumerate(self.keys):
            kind = key.split(":", 1)[0]
            if kind in intent.excluded_kinds or (
                intent.allowed_kinds and kind not in intent.allowed_kinds
            ):
                admitted[i] = False
            if any(
                any(term.startswith(excluded) for term in self.card_terms[i])
                for excluded in intent.excluded_terms
            ):
                admitted[i] = False
        if intent.goals and not focus:
            goal_title_fit = np.array([
                any(any(term.startswith(root) for term in title for root in goal.subjects) for goal in intent.goals)
                for title in self.title_terms
            ])
            lexical *= np.where(goal_title_fit, 1.0, 0.35)
        if mode == "tfidf":
            scores = lexical.copy()
        elif mode == "e5":
            scores = semantic_scores.copy() * goal_fit
        else:
            scores = np.zeros(len(self.keys), dtype=np.float64)
            for values in (lexical, semantic_scores):
                order = sorted(
                    np.flatnonzero(admitted & (values > 0)),
                    key=lambda i: (-float(values[i]), self.keys[i]),
                )[:50]
                for rank, index in enumerate(order, 1):
                    scores[index] += 30.5 / (60 + rank)
        if mode == "rrf":
            scores *= goal_fit
        for i, key in enumerate(self.keys):
            kind = key.split(":", 1)[0]
            if intent.requested_kind == kind:
                scores[i] *= 1.15
            if kind == "term" and not intent.definition:
                scores[i] *= 0.65
        if intent.definition and focus:
            admitted &= np.array([bool(focus & title) for title in self.title_terms])
        scores = np.clip(scores, 0, 1)
        selected = np.flatnonzero(admitted & (scores > 0))
        order = sorted(selected, key=lambda i: (-float(scores[i]), self.keys[i]))
        result["matches"] = [
            {
                "key": self.keys[i],
                "score": float(scores[i]),
                "rank": rank,
                "lexical": float(lexical[i]),
                "word": float(word[i]),
                "character": float(character[i]),
                "semantic": None
                if semantic_scores is None
                else float(semantic_scores[i]),
                "goals": matched_goals[i],
            }
            for rank, i in enumerate(order, 1)
        ]
        result["elapsed_ms"] = (time.perf_counter() - started) * 1000
        return result

    def related(self, key, query=""):
        if key not in self.keys:
            raise KeyError(key)
        source = self.keys.index(key)
        generic = set(tokens(
            "как что такое увеличить повысить составить сделать начать открыть "
            "узнать работать использовать бизнес работа"
        ))

        def specificity(term):
            feature = self.vectorizer.vocabulary_.get(term)
            return -float(self.vectorizer.idf_[feature]) if feature is not None else 0

        anchors = sorted(
            self.title_terms[source] - generic,
            key=lambda term: (specificity(term), term),
        )[:5]
        features = [
            self.vectorizer.vocabulary_[term]
            for term in anchors
            if term in self.vectorizer.vocabulary_
        ]
        intent = parse_intent(query)
        result = {
            "catalog_id": self.manifest["catalog_id"],
            "algorithm": "related-content-v1",
            "source_key": key,
            "anchors": anchors,
            "matches": [],
            "free_only": intent.free_only,
            "exclude_kinds": intent.excluded_kinds,
            "allowed_kinds": intent.allowed_kinds,
        }
        if not features or (query.strip() and intent.clarification):
            return result
        matrix = self.related_matrix
        similarity = np.clip((matrix @ matrix[source].T).toarray().ravel(), 0, 1)
        title_query = self.vectorizer.transform([" ".join(anchors)])
        subject = np.clip((matrix @ title_query.T).toarray().ravel(), 0, 1)
        weights = self.vectorizer.idf_[features]
        covered = (
            np.asarray((self.matrix[:, features] > 0) @ weights).ravel()
            / float(weights.sum())
        )
        scores = 0.55 * similarity + 0.45 * subject
        order = sorted(
            range(len(self.keys)),
            key=lambda index: (-float(scores[index]), self.keys[index]),
        )
        source_kind = key.split(":", 1)[0]
        for index in order:
            kind = self.keys[index].split(":", 1)[0]
            if index == source or scores[index] < 0.12 or covered[index] < 0.45:
                continue
            union = self.title_terms[source] | self.title_terms[index]
            overlap = len(self.title_terms[source] & self.title_terms[index]) / max(1, len(union))
            same_level = source_kind == kind or kind not in {"course", "article"}
            if (
                (overlap >= 0.8 and same_level)
                or kind in intent.excluded_kinds
                or (intent.allowed_kinds and kind not in intent.allowed_kinds)
            ):
                continue
            if any(
                any(term.startswith(excluded) for term in self.card_terms[index])
                for excluded in intent.excluded_terms
            ):
                continue
            result["matches"].append({"key": self.keys[index], "score": float(scores[index])})
            if len(result["matches"]) == 40:
                break
        return result
