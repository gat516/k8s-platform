import unittest

from collect_ci_benchmark import summarize


def completed_jobs():
    jobs = []
    for variant in ("baseline", "go-cache", "full"):
        for name, start, end in (
            ("test", "00:00:00", "00:01:00"),
            ("web", "00:00:00", "00:00:10"),
            ("scan", "00:00:00", "00:02:00"),
            ("build", "00:02:05", "00:03:00"),
        ):
            jobs.append({
                "name": f"trial ({variant}) / {name}", "conclusion": "success",
                "started_at": f"2026-09-30T{start}Z",
                "completed_at": f"2026-09-30T{end}Z", "steps": [],
            })
    return jobs


class BenchmarkTimingsTest(unittest.TestCase):
    def test_parallel_jobs_are_not_added_to_pipeline_duration(self):
        result = summarize(completed_jobs())["baseline"]
        self.assertEqual(result["pipeline_seconds"], 180)
        self.assertEqual(result["runner_seconds"], 245)
        self.assertEqual(result["build_seconds"], 55)

    def test_incomplete_attempt_is_not_a_measurement(self):
        with self.assertRaises(ValueError):
            summarize(completed_jobs()[:-1])

    def test_skipped_or_failed_checks_cannot_appear_faster(self):
        for conclusion in ("skipped", "failure", "cancelled"):
            jobs = completed_jobs()
            jobs[0]["conclusion"] = conclusion
            with self.subTest(conclusion=conclusion), self.assertRaises(ValueError):
                summarize(jobs)


if __name__ == "__main__":
    unittest.main()
