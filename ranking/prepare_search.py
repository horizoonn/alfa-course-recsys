import argparse
import json
import os
import tempfile
import time
from pathlib import Path

import numpy as np
from scipy.sparse import save_npz

import prepare_text as legacy
import search_index as search
import semantic


def prepare(directory, corpus, timeout, enable_semantic):
    started = time.monotonic()
    directory.mkdir(parents=True, exist_ok=True)
    error = (
        semantic.ensure_model(directory / "model", timeout)
        if enable_semantic
        else "semantic_disabled"
    )
    current = None
    manifest = None
    try:
        manifest = json.loads((directory / "current.json").read_text())
        if (
            manifest["catalog_id"] == corpus["catalog_id"]
            and manifest["pipeline_id"] == search.pipeline_id()
        ):
            current = search.SearchIndex(directory, enable_semantic=False)
            if (
                current.semantic_valid
                and current.manifest.get("semantic_id")
                == search.embedding_id(current.cards)
            ) or error:
                print(
                    f"Search index unchanged; reused in {time.monotonic() - started:.3f}s",
                    flush=True,
                )
                return
    except (OSError, ValueError, KeyError):
        pass
    groups, cards = search.groups_for(corpus)
    if current is None:
        vectorizer = search.vectorizer()
        combined = [" ".join(parts) for parts in zip(*groups)]
        if any(text.strip() for text in combined):
            vectorizer.fit(combined)
            matrix = sum(
                weight * vectorizer.transform(group)
                for weight, group in zip(search.WEIGHTS, groups)
            ).tocsr()
            vocabulary = vectorizer.vocabulary_
            idf = vectorizer.idf_.tolist()
        else:
            from scipy.sparse import csr_matrix

            vocabulary = {"__empty__": 0}
            idf = [1.0]
            matrix = csr_matrix((len(cards), 1), dtype=np.float64)
    else:
        matrix = current.matrix
        vocabulary = current.vectorizer.vocabulary_
        idf = current.vectorizer.idf_.tolist()
    headings = [
        "\n".join(
            field["text"]
            for field in doc["fields"]
            if field["name"] in {"title", "summary"}
        )
        for doc in corpus["documents"]
    ]
    character = search.char_vectorizer()
    character_texts = [search.normalize(card) for card in cards]
    if any(text.strip() for text in character_texts):
        character_matrix = character.fit_transform(character_texts)
        character_vocabulary = {
            term: int(index) for term, index in character.vocabulary_.items()
        }
        character_idf = character.idf_.tolist()
    else:
        from scipy.sparse import csr_matrix

        character_vocabulary = {"__empty__": 0}
        character_idf = [1.0]
        character_matrix = csr_matrix((len(cards), 1), dtype=np.float64)
    vectors = None
    semantic_id = search.embedding_id(cards)
    if manifest and manifest.get("semantic"):
        try:
            previous = directory / manifest["generation"]
            if previous.parent != directory or not previous.name.startswith(
                "generation-"
            ):
                raise ValueError("invalid previous generation")
            old_cards = json.loads((previous / "metadata.json").read_text())["cards"]
            if (
                old_cards == cards
                and manifest.get("semantic_id") in (None, semantic_id)
                and manifest.get("model") == semantic.MODEL
                and manifest.get("model_revision") == semantic.REVISION
                and search.digest_file(previous / "semantic.npy")
                == manifest["files"]["semantic.npy"]
            ):
                vectors = np.load(previous / "semantic.npy", allow_pickle=False)
                if vectors.shape != (len(cards), 384) or not np.isfinite(vectors).all():
                    vectors = None
        except (OSError, ValueError, KeyError):
            pass
    if not error and vectors is None:
        try:
            model = semantic.encoder(directory / "model")
            vectors = np.asarray(
                list(
                    model.embed(
                        ["passage: " + card for card in cards],
                        batch_size=16,
                        parallel=None,
                    )
                ),
                dtype=np.float32,
            ).reshape(len(cards), 384)
        except Exception:
            error = "document_encoding_failed"
    folder = Path(tempfile.mkdtemp(prefix="generation-", dir=directory))
    save_npz(folder / "lexical.npz", matrix)
    save_npz(folder / "character.npz", character_matrix)
    legacy.write_artifact(
        folder / "metadata.json",
        {
            "keys": [document["key"] for document in corpus["documents"]],
            "vocabulary": vocabulary,
            "idf": idf,
            "cards": cards,
            "headings": headings,
            "character_vocabulary": character_vocabulary,
            "character_idf": character_idf,
        },
    )
    files = ["metadata.json", "lexical.npz", "character.npz"]
    if vectors is not None:
        np.save(folder / "semantic.npy", vectors, allow_pickle=False)
        files.append("semantic.npy")
    manifest = {
        "schema_version": search.SCHEMA,
        "catalog_id": corpus["catalog_id"],
        "pipeline_id": search.pipeline_id(),
        "generation": folder.name,
        "semantic": vectors is not None,
        "semantic_error": error or "",
        "semantic_id": semantic_id,
        "model": semantic.MODEL,
        "model_revision": semantic.REVISION,
        "files": {name: search.digest_file(folder / name) for name in files},
    }
    legacy.write_artifact(directory / "current.json", manifest)
    print(
        f"Prepared search index: {len(cards)} materials, E5={vectors is not None}, "
        f"{time.monotonic() - started:.3f}s; reason={error or 'ok'}",
        flush=True,
    )


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("catalog", type=Path)
    parser.add_argument(
        "--output", type=Path, default=Path("data/ranking_text_scores.json")
    )
    parser.add_argument("--search-dir", type=Path, default=Path("data/search"))
    parser.add_argument(
        "--model-timeout",
        type=float,
        default=float(os.environ.get("SEARCH_MODEL_TIMEOUT", "120")),
    )
    args = parser.parse_args()
    corpus = json.loads(args.catalog.read_text())["catalog"]
    try:
        legacy.validate_artifact(json.loads(args.output.read_text()), corpus)
        print("Prepared task scores unchanged; reused", flush=True)
    except (OSError, ValueError, KeyError, TypeError):
        legacy.write_artifact(args.output, legacy.prepare(corpus))
    prepare(
        args.search_dir,
        corpus,
        args.model_timeout,
        os.environ.get("SEARCH_ENABLE_SEMANTIC", "false").lower() == "true",
    )


if __name__ == "__main__":
    main()
