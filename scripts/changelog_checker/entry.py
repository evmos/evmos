"""
This file contains the definition for the Entry class. It is used to parse the individual entries, that
relate to the changes in a specific pull request.
"""

import re
from typing import List


# Allowed entry pattern: `- (module) [#PR](link) - description`
ENTRY_PATTERN = re.compile(
    r'^-\s+\([a-zA-Z0-9\-]+\) \[\\?#(?P<pr>\d+)]\((?P<link>.+)\) (?P<desc>.+)$',
)


class Entry:
    """
    This class represents an individual changelog entry that is describing the changes on one specific PR.
    """

    def __init__(self, line: str):
        self.line: str = line
        self.category: str = ""
        self.description: str = ""
        self.link: str = ""
        self.pr_number: int = 0
        self.problems: List[str] = []

    def parse(self) -> tuple[bool, str]:
        """
        Parses a changelog entry from a line of text.

        :return: a tuple indicating whether the parsing was successful and an error message in case of failure
        """

        match = ENTRY_PATTERN.match(self.line)
        if not match:
            self.problems.append(f'Invalid entry: "{self.line}"')

        self.pr_number = int(match.group("pr"))
        self.category = match.group("category")
        self.link = match.group("link")
        self.description = match.group("desc")

        if self.pr_number not in self.link:
            self.problems.append(
                f'PR link is not matching PR number: "{self.line}"'
            )

        if not self.description[0].isupper():
            self.problems.append(
                f'PR description should start with capital letter: "{self.line}"'
            )

        if self.description[-1] != '.':
            self.problems.append(
                f'PR description should end with a dot: "{self.line}"'
            )

        return True, ""


if __name__ == "__main__":
    print("This is a library file and should not be executed directly.")
