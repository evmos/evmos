"""
This file contains the definition for the ChangeType class. It is used to parse the section header for a
given type of changes like improvements or bug fixes.
"""

import re
from typing import List

from config import ALLOWED_CHANGE_TYPES
from entry import check_spelling

# Allowed change type pattern, e.g. `### Bug Fixes`
CHANGE_TYPE_PATTERN = re.compile(
    r"^### (?P<type>[a-zA-Z0-9\- ]+)\s*$",
)


class ChangeType:
    """
    This class represents a section header in the changelog.
    """

    def __init__(self, line: str):
        self.line: str = line
        self.fixed: str = line
        self.type: str = ""
        self.problems: List[str] = []

    def parse(self) -> bool:
        """
        Parses a change type entry from a line of text.

        :return: boolean indicating whether the parsing was successful
        """

        problems: List[str] = []
        match = CHANGE_TYPE_PATTERN.match(self.line)
        if not match:
            problems.append(f'Malformed change type: "{self.line}"')
            self.problems = problems
            return False

        self.type = match.group("type")

        type_found, fixed_type, spelling_problems = check_spelling(
            self.type, ALLOWED_CHANGE_TYPES
        )
        if not type_found:
            problems.append(f'"{self.type}" is not a valid change type')
        if spelling_problems:
            problems.extend(spelling_problems)

        self.fixed = f"### {fixed_type}"
        self.problems = problems

        return problems == []
