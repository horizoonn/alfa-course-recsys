import argparse
import hashlib
import json
import math
import os
import re
import tempfile
import time
from functools import lru_cache
from importlib.metadata import version
from pathlib import Path

import numpy as np
from nltk.stem.snowball import RussianStemmer
from sklearn.feature_extraction.text import TfidfVectorizer

ALGORITHM = "tfidf-snowball-v1"
STEMMER = RussianStemmer()
WORDS = re.compile(r"[^\W_]+", re.UNICODE)
STOP_WORDS = frozenset(
    STEMMER.stem(word)
    for word in (
        "как что для или это при его она они вам вас все уже наш свой свое "
        "бизнес предприниматель материалы формат сфера"
    ).split()
)
GROUPS = {
    "title": 0,
    "summary": 0,
    "metadata": 0,
    "body": 1,
    "caption": 2,
    "reference": 2,
}
WEIGHTS = (0.7, 0.3, 0.05)


@lru_cache(maxsize=65536)
def stem(word):
    return STEMMER.stem(word)


def tokens(text):
    return [
        value
        for word in WORDS.findall(text.lower().replace("ё", "е"))
        if len(word) >= 3 and (value := stem(word)) not in STOP_WORDS
    ]


def producer():
    return {
        "scikit_learn": version("scikit-learn"),
        "nltk": version("nltk"),
        "pipeline_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
    }


def validate_artifact(artifact, corpus):
    keys = [doc["key"] for doc in corpus["documents"]]
    if (
        artifact["schema_version"] != 1
        or artifact["algorithm"] != ALGORITHM
        or artifact["catalog_id"] != corpus["catalog_id"]
        or artifact["keys"] != keys
        or artifact["producer"] != producer()
        or set(artifact["scores"]) != set(corpus["queries"])
    ):
        raise ValueError("text scores do not match input or pipeline")
    for values in artifact["scores"].values():
        if len(values) != len(keys) or any(
            type(value) not in (int, float)
            or not math.isfinite(value)
            or not 0 <= value <= 1
            for value in values
        ):
            raise ValueError("invalid text scores")


def prepare(corpus):
    if not re.fullmatch(r"[0-9a-f]{64}", corpus["catalog_id"]):
        raise ValueError("invalid catalog fingerprint")
    documents, queries = corpus["documents"], corpus["queries"]
    keys = [document["key"] for document in documents]
    if len(keys) != len(set(keys)):
        raise ValueError("duplicate document key")
    groups = [[], [], []]
    for document in documents:
        parts = [[], [], []]
        for field in document["fields"]:
            if field["name"] == "promotion":
                continue
            if field["name"] not in GROUPS:
                raise ValueError(f"unknown text field role: {field['name']}")
            parts[GROUPS[field["name"]]].extend(tokens(field["text"]))
        for group, words in zip(groups, parts):
            group.append(" ".join(words))
    combined = [" ".join(parts) for parts in zip(*groups)]
    scores = np.zeros((len(queries), len(keys)), dtype=np.float64)
    if any(text.strip() for text in combined) and queries:
        vectorizer = TfidfVectorizer(
            analyzer=str.split, sublinear_tf=True, smooth_idf=True, norm="l2"
        )
        vectorizer.fit(combined)
        query_vectors = vectorizer.transform(
            [" ".join(tokens(text)) for text in queries.values()]
        )
        for weight, group in zip(WEIGHTS, groups):
            scores += weight * (query_vectors @ vectorizer.transform(group).T).toarray()
    artifact = {
        "schema_version": 1,
        "algorithm": ALGORITHM,
        "catalog_id": corpus["catalog_id"],
        "keys": keys,
        "scores": {
            task: row.tolist()
            for task, row in zip(queries, np.round(np.clip(scores, 0, 1), 9))
        },
        "producer": producer(),
    }
    validate_artifact(artifact, corpus)
    return artifact


def write_artifact(path, artifact):
    path.parent.mkdir(parents=True, exist_ok=True)
    descriptor, temporary = tempfile.mkstemp(prefix=".text-scores-", dir=path.parent)
    try:
        with os.fdopen(descriptor, "w") as output:
            json.dump(
                artifact,
                output,
                ensure_ascii=False,
                separators=(",", ":"),
                allow_nan=False,
            )
            output.flush()
            os.fsync(output.fileno())
        os.replace(temporary, path)
    finally:
        Path(temporary).unlink(missing_ok=True)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("catalog", type=Path)
    parser.add_argument(
        "--output", type=Path, default=Path("data/ranking_text_scores.json")
    )
    parser.add_argument("--force", action="store_true")
    args = parser.parse_args()
    started = time.monotonic()
    corpus = json.loads(args.catalog.read_text())["catalog"]
    if not args.force and args.output.is_file():
        try:
            validate_artifact(json.loads(args.output.read_text()), corpus)
        except (ValueError, KeyError, TypeError):
            pass
        else:
            print(
                f"Text scores unchanged; reused {args.output} in {time.monotonic() - started:.3f}s"
            )
            return
    artifact = prepare(corpus)
    write_artifact(args.output, artifact)
    print(
        f"Prepared {len(artifact['keys'])} materials x {len(artifact['scores'])} tasks "
        f"in {time.monotonic() - started:.3f}s; {args.output.stat().st_size} bytes"
    )


if __name__ == "__main__":
    main()
