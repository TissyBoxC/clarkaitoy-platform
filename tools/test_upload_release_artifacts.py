import json
import tempfile
import unittest
from pathlib import Path
from unittest import mock

from tools import upload_release_artifacts


class UploadReleaseArtifactsTest(unittest.TestCase):
    def test_first_publish_starts_from_empty_index(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            local_index = root / "index.json"
            manifest = {
                "version": "0.9.0",
                "channel": "stable",
                "artifacts": [
                    {
                        "platform": "android",
                        "kind": "apk",
                        "file_name": "sprout-parent-app-v0.9.0.apk",
                        "sha256": "a" * 64,
                    }
                ],
            }

            with mock.patch.object(
                upload_release_artifacts,
                "run_sftp",
                side_effect=RuntimeError("remote index: No such file"),
            ):
                upload_release_artifacts.download_optional_index(
                    host="example.com",
                    port=2022,
                    user="sprout-release",
                    private_key=root / "id_ed25519",
                    known_hosts=root / "known_hosts",
                    remote_path="releases/index.json",
                    local_path=local_index,
                )

            upload_release_artifacts.write_index_candidate(
                local_index_path=local_index,
                manifest=manifest,
                public_base_url="https://download.example.com",
                title="测试发布",
                published_at="2026-10-03T00:00:00Z",
            )
            payload = json.loads(local_index.read_text(encoding="utf-8"))

        self.assertEqual(payload["releases"][0]["version"], "0.9.0")

    def test_authentication_failure_is_not_treated_as_missing_index(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            with mock.patch.object(
                upload_release_artifacts,
                "run_sftp",
                side_effect=RuntimeError("Permission denied (publickey)"),
            ):
                with self.assertRaises(RuntimeError):
                    upload_release_artifacts.download_optional_index(
                        host="example.com",
                        port=2022,
                        user="sprout-release",
                        private_key=root / "id_ed25519",
                        known_hosts=root / "known_hosts",
                        remote_path="releases/index.json",
                        local_path=root / "index.json",
                    )


if __name__ == "__main__":
    unittest.main()
