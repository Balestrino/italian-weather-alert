#!/usr/bin/env python3
"""Synthetic Git tests for changelog recording and enforcement."""

from __future__ import annotations

from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest


SOURCE = Path(__file__).resolve().parent.parent


class ChangelogTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        (self.root / "scripts").mkdir()
        (self.root / ".githooks").mkdir()
        shutil.copy2(SOURCE / "scripts/changelog.py", self.root / "scripts/changelog.py")
        shutil.copy2(SOURCE / ".githooks/pre-commit", self.root / ".githooks/pre-commit")
        (self.root / "CHANGELOG.md").write_text("# Changelog\n\n## Unreleased\n\n")
        (self.root / "obsolete.txt").write_text("Fixture\n")
        self.git("init", "-q")
        self.git("config", "user.name", "Fixture")
        self.git("config", "user.email", "fixture@example.test")
        self.git("add", ".")
        self.git("commit", "-qm", "baseline")
        self.base = self.git("rev-parse", "HEAD").stdout.strip()

    def git(self, *args: str, check: bool = True) -> subprocess.CompletedProcess[str]:
        return subprocess.run(("git", *args), cwd=self.root, capture_output=True, text=True, check=check)

    def changelog(self, *args: str, check: bool = True) -> subprocess.CompletedProcess[str]:
        return subprocess.run((sys.executable, "scripts/changelog.py", *args), cwd=self.root, capture_output=True, text=True, check=check)

    def test_explicit_entry_is_staged_once_and_passes_check(self) -> None:
        (self.root / "plan.md").write_text("Synthetic plan\n")
        self.git("add", "plan.md")
        self.changelog("record", "--kind", "Planned", "--summary", "Define a fixture plan")
        self.changelog("record", "--kind", "Planned", "--summary", "Define a fixture plan")
        self.assertEqual((self.root / "CHANGELOG.md").read_text().count("Define a fixture plan"), 1)
        self.git("commit", "-qm", "plan: fixture")
        self.changelog("check", "--base", self.base)

    def test_hook_inserts_entry_and_ci_detects_bypass(self) -> None:
        self.git("config", "core.hooksPath", ".githooks")
        (self.root / "fix.go").write_text("package fixture\n")
        self.git("add", "fix.go")
        self.git("commit", "-qm", "fix: fixture")
        self.assertIn("[Changed] Update fix.go", (self.root / "CHANGELOG.md").read_text())
        self.changelog("check", "--base", self.base)
        clean = self.git("rev-parse", "HEAD").stdout.strip()
        (self.root / "plan.md").write_text("Synthetic plan\n")
        self.git("add", "plan.md")
        self.git("-c", "core.hooksPath=/dev/null", "commit", "-qm", "plan: bypass hook")
        failed = self.changelog("check", "--base", clean, check=False)
        self.assertNotEqual(failed.returncode, 0)
        self.assertIn("without a dated CHANGELOG.md entry", failed.stderr)

    def test_unstaged_changelog_edit_is_not_staged(self) -> None:
        with (self.root / "CHANGELOG.md").open("a") as changelog:
            changelog.write("Draft text\n")
        (self.root / "fix.go").write_text("package fixture\n")
        self.git("add", "fix.go")
        failed = self.changelog("record", check=False)
        self.assertNotEqual(failed.returncode, 0)
        self.assertIn("Draft text", (self.root / "CHANGELOG.md").read_text())
        self.assertEqual(self.git("diff", "--cached", "--name-only").stdout.strip(), "fix.go")

    def test_deletion_gets_an_automatic_entry(self) -> None:
        self.git("config", "core.hooksPath", ".githooks")
        self.git("rm", "obsolete.txt")
        self.git("commit", "-qm", "remove: obsolete fixture")
        self.assertIn("Update obsolete.txt", (self.root / "CHANGELOG.md").read_text())
        self.changelog("check", "--base", self.base)

    def test_staged_changelog_without_entry_is_rejected(self) -> None:
        with (self.root / "CHANGELOG.md").open("a") as changelog:
            changelog.write("Draft text\n")
        (self.root / "fix.go").write_text("package fixture\n")
        self.git("add", "CHANGELOG.md", "fix.go")
        failed = self.changelog("record", check=False)
        self.assertNotEqual(failed.returncode, 0)
        self.assertIn("without a new dated entry", failed.stderr)

    def test_direct_merge_without_changelog_is_rejected(self) -> None:
        main = self.git("symbolic-ref", "--short", "HEAD").stdout.strip()
        self.git("checkout", "-qb", "feature")
        (self.root / "fix.go").write_text("package fixture\n")
        self.git("add", "fix.go")
        self.git("commit", "-qm", "fix: without changelog")
        self.git("checkout", "-q", main)
        self.git("merge", "--no-ff", "-m", "Merge feature", "feature")
        merge = self.git("rev-parse", "--short=12", "HEAD").stdout.strip()
        failed = self.changelog("check", "--base", self.base, check=False)
        self.assertNotEqual(failed.returncode, 0)
        self.assertIn(merge, failed.stderr)


if __name__ == "__main__":
    unittest.main()
