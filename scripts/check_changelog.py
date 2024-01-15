import os
import re
import sys
from typing import Dict, List, Tuple

# Allowed entry pattern: `- (module) [#PR](link) - description`
ENTRY_PATTERN = re.compile(
    r'^-\s+\([a-zA-Z0-9\-]+\) \[\\?#(?P<pr>\d+)]\((?P<link>.+)\) (?P<desc>.+)$',
)

# Allowed release pattern: vX.Y.Z(-rcN) (YYYY-MM-DD)
RELEASE_PATTERN = re.compile(
    r'^## (Unreleased|\[(?P<version>v\d+\.\d+\.\d+(-rc\d+)?)] - \d{4}-\d{2}-\d{2})$',
)

ALLOWED_CATEGORIES = [
    'API Breaking',
    'Bug Fixes',
    'Improvements',
    'State Machine Breaking',
]


class Changelog:
    """
    This class represents the contents of the changelog and provides methods to parse it.
    """

    def __init__(self, filename: str):
        self.contents: List[str]
        self.filename: str = filename

        self.failed_entries: List[str] = []
        # TODO: extract releases type
        self.releases: Dict[str, Dict[str, Dict[int, Dict[str, str]]]] = {}

        with open(self.filename, 'r') as file:
            self.contents = file.read()

    def parse(self) -> bool:
        """
        This function parses the changelog and checks if the structure is as expected.
        """

        current_release = None
        current_category = None

        for line in self.contents.split('\n'):
            # Check for Header 2 (##) to identify releases
            stripped_line = line.strip()
            if stripped_line[:3] == '## ':
                release_match = RELEASE_PATTERN.match(line)
                if not release_match:
                    raise ValueError("Header 2 should be used for releases - invalid release pattern: " + line)

                version_match = release_match.group("version")
                current_release = version_match if version_match is not None else "Unreleased"
                self.releases[current_release] = {}
                continue

            # Check for Header 3 (###) to identify categories
            category_match = re.match(r'^###\s+(.+)$', line)
            if category_match:
                current_category = category_match.group(1)
                if current_category not in ALLOWED_CATEGORIES:
                    self.failed_entries.append(f'Invalid change category in {current_release}: "{current_category}"')
                self.releases[current_release][current_category] = {}
                continue

            # Check for individual entries
            if stripped_line[:2] != '- ':
                continue

            entry_match = ENTRY_PATTERN.match(line)
            if not entry_match:
                self.failed_entries.append(f'Invalid entry in {current_release} - {current_category}: "{line}"')
                continue

            pr_number = entry_match.group("pr")
            pr_link = entry_match.group("link")
            pr_description = entry_match.group("desc")

            if pr_number not in pr_link:
                self.failed_entries.append(
                    f'PR link is not matching PR number in {current_release} - {current_category}: "{line}"'
                )

            if not pr_description[0].isupper():
                self.failed_entries.append(
                    f'PR description should start with capital letter in {current_release} - {current_category}: "{line}"'
                )

            if pr_description[-1] != '.':
                self.failed_entries.append(
                    f'PR description should end with a dot in {current_release} - {current_category}: "{line}"'
                )

            self.releases[current_release][current_category][int(pr_number)] = {
                "description": pr_description
            }

        return self.failed_entries == []


class Release:
    """
    This class represents a release in the changelog.
    """

    def __init__(self, contents: List[str]):
        self.contents = contents
        self.version = None
        self.date = None
        self.entries: List[Entry] = []


class Entry:
    """
    This class represents an individual changelog entry that is describing the changes on one specific PR.
    """

    def __init__(self, line: str):
        self.line: str = line
        self.pr_number: int = 0
        self.category: str = ""
        self.description: str = ""

    def parse(self) -> tuple[bool, str]:
        """
        Parses a changelog entry from a line of text.

        :param line: the line of text to parse
        :return: a tuple indicating whether the parsing was successful and an error message in case of failure
        """

        match = ENTRY_PATTERN.match(self.line)
        if not match:
            return False, f'Invalid entry: "{self.line}"'

        self.pr_number = int(match.group("pr"))
        self.category = match.group("category")
        self.description = match.group("desc")

        return True, ""


if __name__ == "__main__":
    changelog_file_path = sys.argv[1]
    if not os.path.exists(changelog_file_path):
        print('Changelog file not found')
        sys.exit(1)

    changelog = Changelog(sys.argv[1])
    failed = changelog.parse()
    if failed:
        print(f'Changelog file is not valid - check the following {len(changelog.failed_entries)} problems:\n')
        print('\n'.join(changelog.failed_entries))
        sys.exit(1)
