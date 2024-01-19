"""
This file contains the definition of the Release class. It is used to parse a release section header in the
changelog.
"""

import re
from typing import List, Tuple

# Allowed unreleased pattern
UNRELEASED_PATTERN = re.compile(r"^## Unreleased$")

# Unreleased version
UNRELEASED_VERSION = "Unreleased"

# Allowed release pattern: [vX.Y.Z(-rcN)](LINK) - (YYYY-MM-DD)
RELEASE_PATTERN = re.compile(
    r"^## \[(?P<version>v\d+\.\d+\.\d+(-rc\d+)?)](?P<link>\(.*\))? - (?P<date>\d{4}-\d{2}-\d{2})$",
)


class Release:
    """
    This class represents a release in the changelog.
    """

    def __init__(self, line: str):
        self.line: str = line
        self.fixed: str = ""
        self.link: str = ""
        self.version: str = ""
        self.problems: List[str] = []

    def parse(self) -> bool:
        """
        This function parses the release section header.

        :return: a boolean value indicating if the parsing was successful.
        """

        problems: List[str] = []

        if UNRELEASED_PATTERN.match(self.line):
            self.fixed = self.line
            self.version = UNRELEASED_VERSION
            return True

        release_match = RELEASE_PATTERN.match(self.line)
        if not release_match:
            problems.append(f'Malformed release header: "{self.line}"')
            self.problems = problems
            return False

        date = release_match.group("date")
        self.link = release_match.group("link")
        self.version = release_match.group("version")

        fixed_link, link_problems = check_link(self.link, self.version)
        if link_problems:
            problems.extend(link_problems)

        fixed = f"## [{self.version}]{fixed_link} - {date}"
        self.fixed = fixed
        self.problems = problems

        return problems == []

    def __le__(self, other: int):
        if self.version == UNRELEASED_VERSION:
            return False

        version_match = re.match(
            r"^v(?P<major>\d+)\.(\d+)\.(\d+)(-rc\d+)?$", self.version
        )
        if not version_match:
            raise ValueError(f'Invalid version "{self.version}"')

        major = int(version_match.group("major"))
        return major <= other


def check_link(link: str, version: str) -> Tuple[str, List[str]]:
    """
    This function checks if the link in the release header is correct.

    :param link: the link in the release header.
    :param version: the version in the release header.
    :return: a tuple containing the fixed link and a list of problems, which is empty if there are none.
    """

    base_url: str = "https://github.com/evmos/evmos/releases/tag/"
    problems: List[str] = []
    # NOTE: the fixed link is the same for all problems
    fixed: str = f"({base_url}{version})"

    if link == "" or link is None:
        problems.append(f'Release link is missing for "{version}"')
        return fixed, problems

    link = link[1:-1]
    if not link.startswith(base_url):
        problems.append(f'Release link should point to an Evmos release: "{link}"')

    if version not in link:
        problems.append(
            f'Release header version "{version}" does not match version in link "{link}"'
        )

    return fixed, problems
