#!/usr/bin/env python3
"""Generate a non-enumerable release index from uploaded artifact files."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import tempfile
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

VERSION_PATTERN = re.compile(r"^[0-9]+\.[0-9]+\.[0-9]+$")
CHANNEL_PATTERN = re.compile(r"^[a-z][a-z0-9._-]*$")
PLATFORM_PATTERN = re.compile(r"^[a-z][a-z0-9._-]*$")
KIND_PATTERN = re.compile(r"^[a-z][a-z0-9._-]*$")
FILE_PATTERN = re.compile(r"^[A-Za-z0-9._+-]+$")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Write a release manifest and regenerate the global index."
    )
    parser.add_argument("--root", required=True, type=Path)
    parser.add_argument("--version", required=True)
    parser.add_argument("--channel", default="stable")
    parser.add_argument("--public-base-url", required=True)
    parser.add_argument("--title", default="")
    parser.add_argument("--published-at", default="")
    return parser.parse_args()


def validate_token(value: str, pattern: re.Pattern[str], label: str) -> str:
    if not pattern.fullmatch(value):
        raise ValueError(f"invalid {label}: {value!r}")
    return value


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as input_file:
        for chunk in iter(lambda: input_file.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def public_url(base_url: str, relative_path: str) -> str:
    return f"{base_url.rstrip('/')}/{relative_path.lstrip('/')}"


def atomic_json_write(path: Path, payload: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    serialized = json.dumps(payload, ensure_ascii=False, indent=2, sort_keys=True) + "\n"
    file_descriptor, temporary_name = tempfile.mkstemp(
        dir=path.parent, prefix=f".{path.name}.", suffix=".tmp"
    )
    try:
        with os.fdopen(file_descriptor, "w", encoding="utf-8") as output_file:
            output_file.write(serialized)
            output_file.flush()
            os.fsync(output_file.fileno())
        os.replace(temporary_name, path)
    except BaseException:
        try:
            os.unlink(temporary_name)
        except FileNotFoundError:
            pass
        raise


def collect_release(root: Path, version: str, channel: str, base_url: str) -> dict[str, Any]:
    release_root = root / version / channel
    if not release_root.is_dir():
        raise FileNotFoundError(f"release directory does not exist: {release_root}")

    artifacts: list[dict[str, Any]] = []
    for path in sorted(release_root.rglob("*")):
        if not path.is_file() or path.name == "manifest.json":
            continue
        relative_to_release = path.relative_to(release_root)
        if len(relative_to_release.parts) < 3:
            continue
        platform, kind, filename = relative_to_release.parts[-3:]
        validate_token(platform, PLATFORM_PATTERN, "platform")
        validate_token(kind, KIND_PATTERN, "kind")
        validate_token(filename, FILE_PATTERN, "filename")
        relative_path = path.relative_to(root).as_posix()
        source_relative_path = path.relative_to(release_root).as_posix()
        artifacts.append(
            {
                "file_name": filename,
                "kind": kind,
                "platform": platform,
                "relative_path": source_relative_path,
                "sha256": sha256_file(path),
                "size_bytes": path.stat().st_size,
                "url": public_url(base_url, relative_path),
            }
        )

    return {
        "artifacts": artifacts,
        "channel": channel,
        "version": version,
    }


def load_existing_index(index_path: Path) -> dict[str, Any]:
    if not index_path.is_file():
        return {"releases": [], "schema_version": 1}
    try:
        payload = json.loads(index_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {"releases": [], "schema_version": 1}
    if not isinstance(payload, dict) or not isinstance(payload.get("releases"), list):
        return {"releases": [], "schema_version": 1}
    return payload


def merge_release_index(
    index: dict[str, Any],
    manifest: dict[str, Any],
    base_url: str,
    title: str,
    published_at: str,
) -> dict[str, Any]:
    releases = [
        release
        for release in index["releases"]
        if not (
            isinstance(release, dict)
            and release.get("version") == manifest["version"]
            and release.get("channel") == manifest["channel"]
        )
    ]
    entry = {
        "artifact_count": len(manifest["artifacts"]),
        "channel": manifest["channel"],
        "manifest_url": public_url(
            base_url, f"{manifest['version']}/{manifest['channel']}/manifest.json"
        ),
        "published_at": published_at,
        "title": title,
        "version": manifest["version"],
    }
    releases.append(entry)
    releases.sort(key=lambda item: (item.get("published_at", ""), item.get("version", "")), reverse=True)
    return {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "releases": releases,
        "schema_version": 1,
    }


def update_index(
    index_path: Path,
    manifest: dict[str, Any],
    base_url: str,
    title: str,
    published_at: str,
) -> None:
    merged = merge_release_index(
        load_existing_index(index_path), manifest, base_url, title, published_at
    )
    atomic_json_write(index_path, merged)


def main() -> None:
    args = parse_args()
    version = validate_token(args.version, VERSION_PATTERN, "version")
    channel = validate_token(args.channel, CHANNEL_PATTERN, "channel")
    root = args.root.resolve()
    root.mkdir(parents=True, exist_ok=True)
    manifest = collect_release(root, version, channel, args.public_base_url)
    manifest_path = root / version / channel / "manifest.json"
    atomic_json_write(manifest_path, manifest)
    published_at = args.published_at or datetime.now(timezone.utc).isoformat()
    update_index(root / "index.json", manifest, args.public_base_url, args.title, published_at)
    print(json.dumps(manifest, ensure_ascii=False, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
