#!/usr/bin/env python3
"""Guards the release workflow's required publication order."""

from __future__ import annotations

import sys
from pathlib import Path


REQUIRED_FRAGMENTS = (
    "Publish release artifacts",
    "SPROUT_RELEASE_UPLOAD_TOKEN",
    "upload_release_artifacts_http.py",
    "steps.upload_release.outcome == 'success'",
    "Verify published download service",
)


def main() -> int:
    workflow_path = Path(sys.argv[1]) if len(sys.argv) > 1 else Path(
        ".github/workflows/release.yml"
    )
    workflow = workflow_path.read_text(encoding="utf-8")
    missing = [fragment for fragment in REQUIRED_FRAGMENTS if fragment not in workflow]
    if missing:
        for fragment in missing:
            print(f"missing workflow fragment: {fragment}", file=sys.stderr)
        return 1

    upload_index = workflow.index("Publish release artifacts")
    images_index = workflow.index("Build and publish platform images")
    release_index = workflow.index("Publish GitHub release")
    if not upload_index < images_index < release_index:
        print(
            "release workflow order must keep artifact upload before images and GitHub release",
            file=sys.stderr,
        )
        return 1

    print("release workflow publication-order test passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
