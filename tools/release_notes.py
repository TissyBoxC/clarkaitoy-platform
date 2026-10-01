#!/usr/bin/env python3
"""Generate Chinese release notes from conventional commits."""

from __future__ import annotations

import argparse
import datetime as dt
import os
import re
import subprocess
from pathlib import Path

SECTION_ORDER = (
    "feat",
    "fix",
    "perf",
    "refactor",
    "security",
    "build",
    "ci",
    "test",
    "docs",
    "chore",
    "other",
)

SECTION_TITLES = {
    "feat": "新功能",
    "fix": "问题修复",
    "perf": "性能优化",
    "refactor": "代码重构",
    "security": "安全与隐私",
    "build": "构建与依赖",
    "ci": "持续集成",
    "test": "测试",
    "docs": "文档",
    "chore": "维护",
    "other": "其他变更",
}

SCOPE_TITLES = {
    "admin_web": "后台管理端",
    "brand": "品牌与命名",
    "ci": "持续集成",
    "contracts": "共享契约",
    "deploy": "本地部署",
    "device_capabilities": "设备能力",
    "device_identity": "设备身份",
    "device_platform": "设备与家庭业务服务",
    "error_code": "固件错误码",
    "firmware": "固件",
    "httpapi": "HTTP 响应契约",
    "observability": "可观测性与日志",
    "parent_app": "家长控制端",
    "release": "版本发布",
    "repo": "仓库维护",
    "sprout": "芽系列扩展",
    "sub2api": "AI 网关",
    "version_info": "版本信息",
    "voice_gateway": "实时语音服务",
    "workspace": "工作区",
}

DESCRIPTION_TITLES = {
    "add": "新增",
    "align": "对齐",
    "define": "定义",
    "enforce": "强制校验",
    "expose": "开放",
    "keep": "保留",
    "rename": "重命名",
    "repair": "修复",
    "reserve": "预留",
    "share": "共享",
    "sync": "同步",
    "update": "更新",
    "validate": "校验",
    "verify": "验证",
}

DESCRIPTION_PHRASES = {
    "add api error and auth refresh foundation": "新增错误处理与登录续期基础",
    "add local service stack": "新增本地服务编排",
    "add shared generated client types": "新增共享生成客户端类型",
    "add validated request labels": "新增经过校验的请求标签",
    "add versioned chinese release workflow": "新增按版本触发的中文发布流程",
    "define cross-service domain schemas": "定义跨服务领域结构",
    "expose secured runtime contract": "开放受保护的运行时契约",
    "ignore python caches and use python3": "忽略 Python 缓存并使用 python3",
    "reserve remote text modules": "预留远程文本模块",
    "preserve concrete release note descriptions": "保留具体的发行说明内容",
    "share versioned response envelope": "共享带版本响应信封",
    "validate versioned base contracts": "校验带版本的基础契约",
    "verify contracts apps and services": "验证契约、应用与服务",
}

CONVENTIONAL_COMMIT = re.compile(
    r"^(?P<type>[A-Za-z]+)"
    r"(?:\((?P<scope>[^)]+)\))?"
    r"(?P<breaking>!)?: "
    r"(?P<description>.+)$"
)
RELEASE_TAG = re.compile(r"^v(?P<version>\d+\.\d+\.\d+)$")


def run_git(*arguments: str) -> str:
    result = subprocess.run(
        ["git", *arguments],
        check=True,
        capture_output=True,
        text=True,
        encoding="utf-8",
    )
    return result.stdout.strip()


def resolve_previous_tag(current_ref: str, current_tag: str | None) -> str | None:
    if current_tag is None:
        return None

    tags = run_git(
        "tag",
        "--merged",
        current_ref,
        "--sort=-v:refname",
        "--format=%(refname:short)",
    ).splitlines()
    for tag in tags:
        if not RELEASE_TAG.fullmatch(tag):
            continue
        if tag == current_tag:
            continue
        return tag
    return None


def read_commits(previous_tag: str | None, current_ref: str) -> list[dict[str, str]]:
    revision_range = f"{previous_tag}..{current_ref}" if previous_tag else current_ref
    output = run_git(
        "log",
        "--no-merges",
        "--format=%H%n%an%n%s%n%b%n%x1e",
        revision_range,
    )

    commits: list[dict[str, str]] = []
    for raw_record in output.split("\x1e"):
        raw_record = raw_record.strip()
        if not raw_record:
            continue
        fields = raw_record.splitlines()
        if len(fields) < 3:
            continue

        sha, author, subject = fields[:3]
        body = "\n".join(fields[3:])
        match = CONVENTIONAL_COMMIT.match(subject)
        if match:
            commit_type = match.group("type").lower()
            scope = match.group("scope") or ""
            description = match.group("description")
            is_breaking = bool(match.group("breaking")) or "BREAKING CHANGE:" in body
        else:
            commit_type = "other"
            scope = ""
            description = subject
            is_breaking = "BREAKING CHANGE:" in body

        if commit_type not in SECTION_TITLES:
            commit_type = "other"

        commits.append(
            {
                "sha": sha,
                "author": author,
                "scope": scope,
                "type": commit_type,
                "description": description,
                "is_breaking": str(is_breaking).lower(),
            }
        )
    return commits


def render_release_notes(
    version: str,
    repository: str,
    previous_tag: str | None,
    commits: list[dict[str, str]],
) -> str:
    tag = f"v{version}"
    if previous_tag:
        comparison = f"`{previous_tag}` 至 `{tag}`"
    else:
        comparison = f"首个发行版本 `{tag}`"

    lines = [
        f"# {repository} {tag}",
        "",
        f"发布日期：{dt.date.today().isoformat()}",
        "",
        f"比较范围：{comparison}",
        "",
        "本版本相对上一发行版本的改动如下。",
        "",
        "## 版本摘要",
        "",
    ]

    counts = {section: 0 for section in SECTION_ORDER}
    for commit in commits:
        counts[commit["type"]] += 1
    for section in SECTION_ORDER:
        if counts[section]:
            lines.append(f"- {SECTION_TITLES[section]}：{counts[section]} 项")
    if not commits:
        lines.append("- 本版本没有代码提交。")

    breaking_changes = [commit for commit in commits if commit["is_breaking"] == "true"]
    if breaking_changes:
        lines.extend(["", "## 破坏性变更", ""])
        lines.extend(render_commit_lines(breaking_changes, repository))

    for section in SECTION_ORDER:
        section_commits = [commit for commit in commits if commit["type"] == section]
        if not section_commits:
            continue
        lines.extend(["", f"## {SECTION_TITLES[section]}", ""])
        lines.extend(render_commit_lines(section_commits, repository))

    return "\n".join(lines).rstrip() + "\n"


def render_commit_lines(commits: list[dict[str, str]], repository: str) -> list[str]:
    lines: list[str] = []
    for commit in commits:
        scope = SCOPE_TITLES.get(commit["scope"], commit["scope"] or "项目")
        short_sha = commit["sha"][:7]
        if repository:
            commit_reference = (
                f"[{short_sha}](https://github.com/{repository}/commit/{commit['sha']})"
            )
        else:
            commit_reference = f"`{short_sha}`"
        lines.append(
            f"- **{scope}**：{localize_description(commit['description'])}"
            f"（提交 {commit_reference}，作者：{commit['author']}）"
        )
    return lines


def localize_description(description: str) -> str:
    phrase = DESCRIPTION_PHRASES.get(description.lower())
    if phrase is not None:
        return phrase

    words = description.split(maxsplit=1)
    if not words:
        return description

    # Preserve the original wording when a commit is already readable in
    # Chinese, so release notes stay specific instead of becoming generic.
    if contains_cjk(description):
        return description.rstrip("。.!！")

    action = DESCRIPTION_TITLES.get(words[0].lower())
    if action is None:
        return description
    return action


def contains_cjk(text: str) -> bool:
    return any("\u4e00" <= character <= "\u9fff" for character in text)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--version", required=True)
    parser.add_argument("--current-ref", default="HEAD")
    parser.add_argument("--current-tag")
    parser.add_argument("--repository", default=os.environ.get("GITHUB_REPOSITORY", ""))
    parser.add_argument("--output", default="release-notes.md")
    arguments = parser.parse_args()

    if not re.fullmatch(r"\d+\.\d+\.\d+", arguments.version):
        raise SystemExit(f"invalid release version: {arguments.version}")

    previous_tag = resolve_previous_tag(arguments.current_ref, arguments.current_tag)
    commits = read_commits(previous_tag, arguments.current_ref)
    notes = render_release_notes(
        arguments.version,
        arguments.repository,
        previous_tag,
        commits,
    )
    Path(arguments.output).write_text(notes, encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
