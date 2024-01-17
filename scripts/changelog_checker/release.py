"""
This file contains the definition of the Release class. It is used to parse a release section header in the
changelog.
"""

import re
from typing import List

# Allowed unreleased pattern
UNRELEASED_PATTERN = re.compile(r'^## Unreleased$')

# Allowed release pattern: [vX.Y.Z(-rcN)](LINK) - (YYYY-MM-DD)
RELEASE_PATTERN = re.compile(
    r'^## \[(?P<version>v\d+\.\d+\.\d+(-rc\d+)?)](?P<link>\(.*\))? - \d{4}-\d{2}-\d{2}$',
)


class Release:
    """
    This class represents a release in the changelog.
    """

    def __init__(self, line: str):
        self.line: str = line
        self.link: str = ""
        self.version: str = ""
        self.problems = []

    def parse(self) -> bool:
        """
        This function parses the release section header.

        :return: a boolean value indicating if the parsing was successful.
        """

        problems: List[str] = []

        if UNRELEASED_PATTERN.match(self.line):
            self.version = "Unreleased"
            return True

        release_match = RELEASE_PATTERN.match(self.line)
        if not release_match:
            problems.append(f'Malformed release header: "{self.line}"')
            self.problems = problems
            return False

        self.link = release_match.group("link")
        self.version = release_match.group("version")

        link_problems = check_link(self.link, self.version)
        if link_problems:
            problems.extend(link_problems)

        self.problems = problems
        return problems == []


def check_link(link: str, version: str) -> List[str]:
    """
    This function checks if the link in the release header is correct.

    :param link: the link in the release header.
    :param version: the version in the release header.
    :return: a list of problems found in the link.
    """

    problems: List[str] = []

    if link == "" or link is None:
        problems.append(f'Release link is missing for "{version}"')
        return problems

    link = link[1:-1]
    if not link.startswith("https://github.com/evmos/evmos/releases/tag/"):
        problems.append(f'Release link should point to an Evmos release: "{link}"')

    if version not in link:
        problems.append(f'Release header version "{version}" does not match version in link "{link}"')

    return problems
