#!/usr/bin/env python3
"""Record staged changes and verify that change commits update CHANGELOG.md."""

from __future__ import annotations

import argparse
import datetime as dt
import os
from pathlib import Path
import re
import subprocess
import sys


ROOT = Path(__file__).resolve().parent.parent
CHANGELOG = ROOT / "CHANGELOG.md"
KINDS = ("Added", "Changed", "Fixed", "Planned", "Removed", "Security")
ENTRY = re.compile(r"^\+\- \d{4}-\d{2}-\d{2} \[(?:" + "|".join(KINDS) + r")\] .+", re.MULTILINE)


def git(*args: str) -> str:
    result = subprocess.run(
        ("git", *args), cwd=ROOT, check=True, capture_output=True, text=True
    )
    return result.stdout


def staged_paths() -> list[str]:
    return [
        path
        for path in git("diff", "--cached", "--name-only", "-z").split("\0")
        if path
    ]


def has_added_entry(patch: str) -> bool:
    return ENTRY.search(patch) is not None


def proposed_summary(paths: list[str]) -> str:
    plans = sorted(
        {path.split("/")[2].replace("-", " ") for path in paths if path.startswith("openspec/changes/") and len(path.split("/")) > 2}
    )
    packages = sorted({path.split("/")[1] for path in paths if path.startswith("internal/") and len(path.split("/")) > 1})
    if plans and packages:
        return f"Update the {', '.join(plans)} plan and {', '.join(packages)} implementation"
    if plans:
        return f"Update the {', '.join(plans)} plan"
    if packages:
        return f"Update the {', '.join(packages)} implementation"
    named = ", ".join(paths[:3])
    return f"Update {named}" + (" and related files" if len(paths) > 3 else "")


def record(summary: str | None, kind: str | None) -> int:
    paths = [path for path in staged_paths() if path != "CHANGELOG.md"]
    if not paths:
        return 0
    staged_patch = git("diff", "--cached", "--unified=0", "--", "CHANGELOG.md")
    if has_added_entry(staged_patch):
        return 0
    if "CHANGELOG.md" in staged_paths():
        print("CHANGELOG.md is staged without a new dated entry", file=sys.stderr)
        return 1
    if git("diff", "--name-only", "--", "CHANGELOG.md").strip():
        print("Stage or discard the existing CHANGELOG.md edit before recording", file=sys.stderr)
        return 1
    text = CHANGELOG.read_text(encoding="utf-8")
    anchor = "## Unreleased\n\n"
    if text.count(anchor) != 1:
        print("CHANGELOG.md must contain one '## Unreleased' heading", file=sys.stderr)
        return 1
    chosen = (summary or os.environ.get("IWA_CHANGELOG_SUMMARY") or proposed_summary(paths)).strip()
    if not chosen or "\n" in chosen or "\r" in chosen:
        print("The changelog summary must be one nonempty line", file=sys.stderr)
        return 1
    chosen_kind = kind or os.environ.get("IWA_CHANGELOG_KIND") or (
        "Planned" if all(path.startswith("openspec/") for path in paths) else "Changed"
    )
    if chosen_kind not in KINDS:
        print(f"Invalid changelog kind: {chosen_kind}", file=sys.stderr)
        return 1
    line = f"- {dt.datetime.now(dt.timezone.utc).date().isoformat()} [{chosen_kind}] {chosen}\n"
    CHANGELOG.write_text(text.replace(anchor, anchor + line, 1), encoding="utf-8")
    git("add", "--", "CHANGELOG.md")
    print(f"Recorded and staged changelog entry: {line.strip()}")
    return 0


def check(base: str) -> int:
    if not base or set(base) == {"0"}:
        base = "HEAD^"
    try:
        ancestor = git("merge-base", base, "HEAD").strip()
    except subprocess.CalledProcessError:
        print(f"Cannot compare changelog to base {base}", file=sys.stderr)
        return 1
    commits = git("rev-list", "--reverse", f"{ancestor}..HEAD").splitlines()
    missing = []
    for commit in commits:
        paths = [
            path for path in git("diff", "--name-only", "-z", f"{commit}^1", commit).split("\0") if path
        ]
        if not any(path != "CHANGELOG.md" for path in paths):
            continue
        patch = git("diff", "--unified=0", f"{commit}^1", commit, "--", "CHANGELOG.md")
        if not has_added_entry(patch):
            missing.append(commit[:12])
    if missing:
        print("Change commits without a dated CHANGELOG.md entry: " + ", ".join(missing), file=sys.stderr)
        return 1
    print(f"Checked {len(commits)} commit(s): changelog entries present")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    recorder = commands.add_parser("record", help="add and stage an Unreleased entry for staged changes")
    recorder.add_argument("--summary", help="single-line explanation; defaults to a path-based summary")
    recorder.add_argument("--kind", choices=KINDS)
    checker = commands.add_parser("check", help="require an entry in each change commit since a base")
    checker.add_argument("--base", required=True, help="base revision, such as a PR base SHA")
    args = parser.parse_args()
    return record(args.summary, args.kind) if args.command == "record" else check(args.base)


if __name__ == "__main__":
    sys.exit(main())
