"""
This file contains the logic to check the changelog contents.

It is possible to run this script with the `--fix` flag to automatically
fix a selection of common problems in the changelog.

Usage:
    python3 check_changelog.py <changelog_file> [--fix]

"""

import io
import os
import sys
from typing import Dict, List, Union

from change_type import ChangeType
from config import ALLOWED_DUPLICATES, LEGACY_VERSION
from entry import Entry
from release import Release


def write(file: Union[None, io.TextIOWrapper], line: str):
    """
    This function writes a line to a file.

    :param file: The file to write to.
    :param line: The line to write.
    """

    if file is not None:
        file.write(line)


class Changelog:
    """
    This class represents the contents of the changelog and provides methods to parse it.
    """

    def __init__(self, filename: str):
        self.contents: List[str]
        self.filename: str = filename

        self.problems: List[str] = []
        self.releases: Dict[str, Dict[str, Dict[int, Dict[str, str]]]] = {}

        if not os.path.exists(self.filename):
            raise FileNotFoundError(f'Changelog file "{self.filename}" not found')

        with open(self.filename, "r") as file:
            self.contents = file.read().split("\n")

    def parse(self, fix: bool = False) -> bool:
        """
        This function parses the changelog and checks if the structure is as expected.

        :param fix: An optional parameter specifying if the changelog should be fixed automatically.
        """

        current_release: str = ""
        current_category: str = ""
        f = None
        is_legacy: bool = False
        seen_prs: List[int] = []

        if fix:
            f = open(self.filename, "w")

        try:
            for line in self.contents:
                if is_legacy:
                    if fix:
                        write(f, line + "\n")
                    continue

                # Check for Header 2 (##) to identify releases
                stripped_line = line.strip()
                if stripped_line[:3] == "## ":
                    release = Release(line)
                    release.parse()
                    current_release = release.version
                    if current_release in self.releases:
                        self.problems.append(
                            f'Release "{current_release}" is duplicated in the changelog'
                        )
                    else:
                        self.releases[current_release] = {}
                    self.problems.extend(release.problems)

                    if release <= LEGACY_VERSION:
                        is_legacy = True

                    if fix:
                        write(f, release.fixed + "\n")

                    continue

                # Check for Header 3 (###) to identify change types
                if stripped_line[:4] == "### ":
                    change_type = ChangeType(line)
                    change_type.parse()
                    current_category = change_type.type
                    if current_category in self.releases[current_release]:
                        self.problems.append(
                            f'Change type "{current_category}" is duplicated in {current_release}'
                        )
                    else:
                        self.releases[current_release][current_category] = {}
                    self.problems.extend(change_type.problems)

                    if fix:
                        write(f, change_type.fixed + "\n")

                    continue

                # Check for individual entries
                if stripped_line[:2] != "- ":
                    if fix:
                        write(f, line + "\n")

                    continue

                entry = Entry(line)
                entry.parse()
                self.problems.extend(entry.problems)
                if fix:
                    write(f, entry.fixed + "\n")

                if not current_category:
                    raise ValueError(f'Entry "{line}" is missing a category')

                if entry.pr_number in seen_prs:
                    if (
                        not entry.is_exception
                        and entry.pr_number not in ALLOWED_DUPLICATES
                    ):
                        self.problems.append(
                            f"PR #{entry.pr_number} is duplicated in the changelog"
                        )
                else:
                    seen_prs.append(entry.pr_number)

                self.releases[current_release][current_category][entry.pr_number] = {
                    "description": entry.description
                }
        finally:
            if f is not None:
                f.close()

        return self.problems == []


if __name__ == "__main__":
    changelog = Changelog(sys.argv[1])

    fix_mode = False
    if len(sys.argv) > 2 and sys.argv[2] == "--fix":
        fix_mode = True

    passed = changelog.parse(fix=fix_mode)
    if passed:
        print(" -> Changelog is valid.")
    else:
        print(
            f"Changelog file is not valid - check the following {len(changelog.problems)} problems:\n"
        )
        print("\n".join(changelog.problems))
        sys.exit(1)
