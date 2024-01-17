"""
This file contains the definition for the Entry class. It is used to parse the individual entries, that
relate to the changes in a specific pull request.

The expected structure of an entry is: `- (category) [#PR](link) description`
"""

import re
from typing import Dict, List, Tuple

# List of allowed categories for an individual changelog entry.
ALLOWED_CATEGORIES = [
    "ante",
    "api",
    "ci",
    "distribution-precompile",
    "erc20-precompile",
    "evm",
    "ibc-precompile",
    "inflation",
    "osmosis-outpost",
    "p256-precompile",
    "staking-precompile",
    "stride-outpost",
    "testnet",
    "tests",
    "vesting",
]

# A dictionary of allowed spellings for some common patterns in changelog entries.
ALLOWED_SPELLINGS = {
    "ABI": re.compile("abi", re.IGNORECASE),
    "API": re.compile("api", re.IGNORECASE),
    "CI": re.compile("ci", re.IGNORECASE),
    "CLI": re.compile("cli", re.IGNORECASE),
    "ERC-20": re.compile("erc-*20", re.IGNORECASE),
    "EVM": re.compile("evm", re.IGNORECASE),
    "IBC": re.compile("ibc", re.IGNORECASE),
    "outpost": re.compile("outpost", re.IGNORECASE),
    "PR": re.compile("(pr)(\s|$)", re.IGNORECASE),
    "precompile": re.compile("precompile", re.IGNORECASE),
    "SDK": re.compile("sdk", re.IGNORECASE),
    "WERC-20": re.compile("werc-*20", re.IGNORECASE),
}

# Allowed entry pattern: `- (category) [#PR](link) - description`
ENTRY_PATTERN = re.compile(
    r'^-(?P<ws1>\s*)\((?P<category>[a-zA-Z0-9\-]+)\)' +
    r'(?P<ws2>\s*)\[\\?#(?P<pr>\d+)]\((?P<link>.+)\)' +
    r'(?P<ws3>\s*)(?P<desc>.+)$',
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
        self.whitespaces: List[str] = []

    def parse(self) -> bool:
        """
        Parses a changelog entry from a line of text.

        :return: a tuple indicating whether the parsing was successful and an error message in case of failure
        """

        problems: List[str] = []
        match = ENTRY_PATTERN.match(self.line)
        if not match:
            problems.append(f'Malformed entry: "{self.line}"')
            self.problems = problems
            return False

        self.pr_number = int(match.group("pr"))
        self.category = match.group("category")
        self.link = match.group("link")
        self.description = match.group("desc")
        self.whitespaces = [
            match.group("ws1"),
            match.group("ws2"),
            match.group("ws3"),
        ]

        ws_problems = check_whitespace(self.whitespaces)
        if ws_problems:
            problems.extend(ws_problems)

        cat_problems = check_category(self.category)
        if cat_problems:
            problems.extend(cat_problems)

        link_problems = check_link(self.link, self.pr_number)
        if link_problems:
            problems.extend(link_problems)

        description_problems = check_description(self.description)
        if description_problems:
            problems.extend(description_problems)

        self.problems = problems
        return problems == []


def check_whitespace(whitespaces: List[str]) -> List[str]:
    """
    Check if the whitespaces are valid.

    :param whitespaces: the whitespaces to check
    :return: a list of problems, empty if there are none
    """

    problems: List[str] = []

    if whitespaces[0] != " ":
        problems.append(f'There should be exactly one space between the leading dash and the category')

    if whitespaces[1] != " ":
        problems.append(f'There should be exactly one space between the category and PR link')

    if whitespaces[2] != " ":
        problems.append(f'There should be exactly one space between the PR link and the description')

    return problems


def check_category(category: str) -> List[str]:
    """
    Check if the category is valid.

    :param category: the category to check
    :return: a list of problems, empty if there are none
    """

    problems: List[str] = []

    if not category.islower():
        problems.append(f'Category should be lowercase: "({category})"')

    if category.lower() not in ALLOWED_CATEGORIES:
        problems.append(f'Invalid change category: "({category})"')

    return problems


def check_link(link: str, pr_number: int) -> List[str]:
    """
    Check if the link is valid.

    :param link: the link to check
    :param pr_number: the PR number to match in the link
    :return: a list of problems, empty if there are none
    """

    problems: List[str] = []

    if not link.startswith("https://github.com/evmos/evmos/pull/"):
        problems.append(f'PR link should point to evmos repository: "{link}"')

    if str(pr_number) not in link:
        problems.append(f'PR link is not matching PR number {pr_number}: "{link}"')

    return problems


def check_description(description: str) -> List[str]:
    """
    Check if the description is valid.

    :param description: the description to check
    :return: a list of problems, empty if there are none
    """

    problems: List[str] = []

    if not description[0].isupper():
        problems.append(
            f'PR description should start with capital letter: "{description}"'
        )

    if description[-1] != '.':
        problems.append(
            f'PR description should end with a dot: "{description}"'
        )

    _, abbreviation_problems = check_spelling(description, ALLOWED_SPELLINGS)
    if abbreviation_problems:
        problems.extend(abbreviation_problems)

    return problems


def check_spelling(description: str, expected_spellings: Dict[str, re.Pattern]) -> Tuple[bool, List[str]]:
    """
    Checks some common spelling requirements.

    :param expected_spellings: a dictionary of expected spellings and the matching patterns
    :param description: the description to check
    :return: a tuple indicating whether a matching pattern was found and a list of problems with the match
    """

    problems: List[str] = []
    found: bool = False

    for spelling, pattern in expected_spellings.items():
        match = pattern.search(description)
        if match:
            if match.group(0) != spelling:
                problems.append(
                    f'"{spelling}" should be used instead of "{match.group(0)}"'
                )
            found = True

    return found, problems


if __name__ == "__main__":
    print("This is a library file and should not be executed directly.")
