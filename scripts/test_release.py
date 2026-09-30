import json
import unittest
from unittest.mock import patch, call

import release


class ReleaseTests(unittest.TestCase):
    @patch("release.health", return_value="previous")
    @patch("release.kubectl", return_value=json.dumps({"metadata":{"annotations":{"deployment.kubernetes.io/revision":"8"}}}))
    def test_capture_pins_previous_revision(self, kubectl, health):
        self.assertEqual(release.capture("default", "app", "https://example.test/health"), {"previous_revision":"8", "previous_version":"previous"})

    @patch("release.kubectl", return_value="")
    def test_first_install_has_no_rollback_target(self, kubectl):
        self.assertEqual(release.capture("default", "app", "https://example.test/health"), {"previous_revision":"", "previous_version":""})

    @patch("release.verify")
    @patch("release.kubectl", return_value="")
    def test_rollback_waits_and_verifies_exact_previous_version(self, kubectl, verify):
        release.rollback("book", "reader", "8", "previous", "https://example.test/health")
        self.assertEqual(kubectl.call_args_list, [call("book", "rollout", "undo", "deployment/reader", "--to-revision=8"),call("book", "rollout", "status", "deployment/reader", "--timeout=120s")])
        verify.assert_called_once_with("https://example.test/health", "previous")

    @patch("release.verify", side_effect=RuntimeError("unhealthy"))
    @patch("release.kubectl", return_value="")
    def test_rollback_is_failed_if_service_remains_unhealthy(self, kubectl, verify):
        with self.assertRaises(RuntimeError):
            release.rollback("book", "reader", "8", "previous", "https://example.test/health")

    @patch("release.health", return_value="old-version")
    def test_old_healthy_version_does_not_pass_new_release_check(self, health):
        with self.assertRaises(RuntimeError):
            release.verify("https://example.test/health", "new-version", timeout=0)

    @patch("release.health", return_value="new-version")
    def test_expected_version_passes(self, health):
        self.assertGreaterEqual(release.verify("https://example.test/health", "new-version", timeout=0), 0)


if __name__ == "__main__":
    unittest.main()
