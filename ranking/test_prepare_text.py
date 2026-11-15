import copy
import tempfile
import unittest
from pathlib import Path

from prepare_text import prepare, tokens, validate_artifact, write_artifact


def corpus(documents):
    return {
        "catalog_id": "a" * 64,
        "documents": documents,
        "queries": {
            "accounting": "Налоги налоговая отчётность НДФЛ НДС",
            "clients": "Лидогенерация клиенты",
        },
    }


def document(key, role, text):
    return {"key": key, "fields": [{"name": role, "text": text}]}


class TextPreparationTest(unittest.TestCase):
    def test_snowball_and_abbreviations(self):
        self.assertEqual(tokens("налоги налогов налогами"), ["налог"] * 3)
        self.assertEqual(tokens("клиент клиентами клиентов"), ["клиент"] * 3)
        self.assertEqual(tokens("НДС НДФЛ SEO"), ["ндс", "ндфл", "seo"])
        self.assertEqual(tokens("учёт учет"), tokens("учет учет"))
        self.assertEqual(tokens("как для предпринимателя бизнесом"), [])

    def test_full_body_roles_and_all_documents(self):
        data = corpus(
            [
                document("body", "body", "Налоги налоговая отчётность НДФЛ НДС"),
                document(
                    "promotion", "promotion", "Налоги налоговая отчётность НДФЛ НДС"
                ),
                document("caption", "caption", "Налоги налоговая отчётность НДФЛ НДС"),
                document(
                    "reference", "reference", "Налоги налоговая отчётность НДФЛ НДС"
                ),
                document(
                    "metadata", "metadata", "Налоги налоговая отчётность НДФЛ НДС"
                ),
                document("empty", "title", ""),
                document("unrelated", "body", "Производственный календарь"),
            ]
        )
        artifact = prepare(data)
        values = dict(zip(artifact["keys"], artifact["scores"]["accounting"]))
        self.assertEqual(artifact["keys"], [doc["key"] for doc in data["documents"]])
        self.assertAlmostEqual(values["body"], 0.3)
        self.assertAlmostEqual(values["metadata"], 0.7)
        self.assertEqual(values["promotion"], 0)
        self.assertLessEqual(values["caption"], 0.05)
        self.assertLessEqual(values["reference"], 0.05)
        self.assertEqual(values["unrelated"], 0)
        self.assertEqual(values["empty"], 0)

    def test_empty_and_stopword_only_catalogs(self):
        for docs in [[], [document("empty", "body", "как для это")]]:
            artifact = prepare(corpus(docs))
            self.assertTrue(
                all(value == 0 for row in artifact["scores"].values() for value in row)
            )
        with self.assertRaises(ValueError):
            prepare(corpus([document("x", "unknown", "Налог")]))
        with self.assertRaises(ValueError):
            prepare(corpus([document("x", "body", ""), document("x", "body", "")]))

    def test_cache_validation_and_atomic_private_output(self):
        data = corpus([document("tax", "title", "Налоги")])
        artifact = prepare(data)
        validate_artifact(artifact, data)
        for mutate in [
            lambda a: a.update(catalog_id="b" * 64),
            lambda a: a["scores"]["accounting"].append(0),
            lambda a: a["scores"]["accounting"].__setitem__(0, float("nan")),
            lambda a: a["producer"].update(pipeline_sha256="changed"),
        ]:
            broken = copy.deepcopy(artifact)
            mutate(broken)
            with self.assertRaises(ValueError):
                validate_artifact(broken, data)
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "scores.json"
            path.write_text("old")
            path.chmod(0o644)
            write_artifact(path, artifact)
            self.assertEqual(path.stat().st_mode & 0o777, 0o600)
            self.assertEqual(len(list(Path(directory).iterdir())), 1)


if __name__ == "__main__":
    unittest.main()
