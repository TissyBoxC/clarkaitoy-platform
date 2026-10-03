#!/usr/bin/env python3
"""Guards the release workflow's failure isolation contract."""

from __future__ import annotations

import sys
from pathlib import Path


REQUIRED_FRAGMENTS = (
    "continue-on-error: true",
    "steps.upload_release.outcome == 'success'",
    "Report download publication status",
    "GitHub Release and container image publication will continue",
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

    sftp_index = workflow.index("Configure SFTP")
    images_index = workflow.index("Build and publish platform images")
    release_index = workflow.index("Publish GitHub release")
    if not sftp_index < images_index < release_index:
        print(
            "release workflow order must keep SFTP before images and GitHub release",
            file=sys.stderr,
        )
        return 1

    print("release workflow failure-isolation test passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
