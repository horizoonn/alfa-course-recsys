import unittest

from evaluate import ranking_metrics


class MetricsTest(unittest.TestCase):
    def test_common_pool_penalizes_missing_results(self):
        labels = {"a": 2, "b": 2, "c": 1}
        self.assertLess(ranking_metrics(["a"], labels)["ndcg_at_3"], 1)
        self.assertEqual(ranking_metrics(["a", "b", "c"], labels)["ndcg_at_3"], 1)

    def test_unjudged_is_not_negative(self):
        self.assertEqual(
            ranking_metrics(["unknown"], {"a": 2}), {"status": "unjudged_predictions"}
        )

    def test_invalid_labels_and_duplicate_predictions(self):
        for labels in [{"a": True}, {"a": -1}, {"a": 3}]:
            with self.assertRaises(ValueError):
                ranking_metrics(["a"], labels)
        with self.assertRaises(ValueError):
            ranking_metrics(["a", "a"], {"a": 2})


if __name__ == "__main__":
    unittest.main()
