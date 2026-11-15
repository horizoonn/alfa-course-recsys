import argparse
import json
import math
import os
from pathlib import Path


def ranking_metrics(keys, judgments):
    if any(
        type(value) is not int or value not in (0, 1, 2) for value in judgments.values()
    ):
        raise ValueError("relevance judgments must be integers in 0..2")
    if len(keys) != len(set(keys)):
        raise ValueError("duplicate prediction")
    if any(key not in judgments for key in keys):
        return {"status": "unjudged_predictions"}
    if not any(judgments.values()):
        return {"status": "no_judged_relevant_materials"}

    def dcg(values):
        return sum(
            (2**value - 1) / math.log2(rank + 2) for rank, value in enumerate(values)
        )

    return {
        "ndcg_at_3": dcg([judgments[key] for key in keys[:3]])
        / dcg(sorted(judgments.values(), reverse=True)[:3]),
        "returned": len(keys[:3]),
        "useful": sum(judgments[key] > 0 for key in keys[:3]),
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("catalog", type=Path)
    parser.add_argument("--labels", type=Path)
    parser.add_argument(
        "--output", type=Path, default=Path("data/ranking_predictions.json")
    )
    args = parser.parse_args()
    data = json.loads(args.catalog.read_text())
    if not data["cases"]:
        parser.error(
            "export recommendations without --documents-only before evaluation"
        )
    labels = json.loads(args.labels.read_text()) if args.labels else {}
    rows = []
    for case in data["cases"]:
        variants = {
            "baseline": case["baseline"],
            "content_rules": case["recommendations"],
        }
        judgments = labels.get(case["id"], {})
        rows.append(
            {
                "id": case["id"],
                "predictions": variants,
                "metrics": {
                    name: ranking_metrics(keys, judgments)
                    for name, keys in variants.items()
                }
                if judgments
                else {},
            }
        )
    result = {
        "catalog_id": data["catalog"]["catalog_id"],
        "labels_supplied": bool(labels),
        "cases": rows,
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    descriptor = os.open(args.output, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    os.fchmod(descriptor, 0o600)
    with os.fdopen(descriptor, "w") as output:
        json.dump(result, output, ensure_ascii=False, indent=2)
    print(f"Compared {len(rows)} profiles; labels supplied: {bool(labels)}")


if __name__ == "__main__":
    main()
