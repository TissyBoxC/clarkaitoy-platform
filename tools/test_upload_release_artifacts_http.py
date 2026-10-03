#!/usr/bin/env python3
"""Tests the HTTPS release uploader's request contract."""

from __future__ import annotations

import io
import json
import tempfile
import unittest
from pathlib import Path
from unittest import mock

from tools import upload_release_artifacts_http


class FakeResponse:
    def __init__(self, payload: dict[str, object]) -> None:
        self.payload = json.dumps(payload).encode("utf-8")

    def read(self) -> bytes:
        return self.payload

    def __enter__(self) -> "FakeResponse":
        return self

    def __exit__(self, *args: object) -> None:
        return None


class UploadReleaseArtifactsHTTPTest(unittest.TestCase):
    def test_multipart_body_contains_structured_release_metadata(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            source = Path(temporary_directory) / "resource.bin"
            source.write_bytes(b"release bytes")

            body, boundary = upload_release_artifacts_http.multipart_body(
                {
                    "version": "0.12.4",
                    "channel": "stable",
                    "platform": "any",
                    "kind": "resource",
                    "filename": "resource.bin",
                    "overwrite": "true",
                },
                source,
            )

            self.assertIn(b"release bytes", body)
            self.assertIn(b'name="version"', body)
            self.assertIn(b"0.12.4", body)
            self.assertIn(b'filename="resource.bin"', body)
            self.assertTrue(boundary)

    def test_main_uploads_every_artifact_then_refreshes_index(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            source = root / "dist/releases/0.12.4/stable/any/resource/resource.bin"
            source.parent.mkdir(parents=True)
            source.write_bytes(b"release bytes")
            manifest_path = root / "manifest.json"
            manifest_path.write_text(
                json.dumps(
                    {
                        "version": "0.12.4",
                        "channel": "stable",
                        "artifacts": [
                            {
                                "platform": "any",
                                "kind": "resource",
                                "file_name": "resource.bin",
                                "relative_path": "any/resource/resource.bin",
                            }
                        ],
                    }
                ),
                encoding="utf-8",
            )
            arguments = [
                "upload_release_artifacts_http.py",
                "--api-base-url",
                "https://api.example.test",
                "--token",
                "t" * 32,
                "--dist",
                str(root / "dist"),
                "--version",
                "0.12.4",
                "--manifest",
                str(manifest_path),
            ]
            requests: list[object] = []

            def fake_urlopen(request: object, timeout: int) -> FakeResponse:
                requests.append((request, timeout))
                return FakeResponse({"data": {}})

            with mock.patch.object(upload_release_artifacts_http.sys, "argv", arguments):
                with mock.patch.object(
                    upload_release_artifacts_http.urllib.request,
                    "urlopen",
                    fake_urlopen,
                ):
                    upload_release_artifacts_http.main()

            self.assertEqual(len(requests), 2)
            upload_request = requests[0][0]
            index_request = requests[1][0]
            self.assertTrue(
                upload_request.full_url.endswith("/internal/v1/release-files")
            )
            self.assertTrue(
                index_request.full_url.endswith("/internal/v1/release-index/refresh")
            )
            self.assertEqual(
                index_request.get_header("Authorization"),
                "Bearer " + "t" * 32,
            )


if __name__ == "__main__":
    unittest.main()
