"""
This file contains the definition for the Entry class. It is used to parse the individual entries, that
relate to the changes in a specific pull request.

The expected structure of an entry is: `- (category) [#PR](link) description`
"""

import re
from typing import Dict, List, Tuple

from config import ALLOWED_CATEGORIES, ALLOWED_SPELLINGS, KNOWN_EXCEPTIONS

# Allowed entry pattern: `- (category) [#PR](link) - description`
ENTRY_PATTERN = re.compile(
    r"^-(?P<ws1>\s*)\((?P<category>[a-zA-Z0-9\-]+)\)"
    + r"(?P<ws2>\s*)\[(?P<bs>\\)?#(?P<pr>\d+)](?P<ws3>\s*)\((?P<link>[^)]*)\)"
    + r"(?P<ws4>\s*)(?P<desc>.+)$",
)


class Entry:
    """
    This class represents an individual changelog entry that is describing the changes on one specific PR.
    """

    def __init__(self, line: str):
        self.line: str = line
        self.fixed: str = line
        self.backslash: bool = False
        self.category: str = ""
        self.description: str = ""
        self.is_exception: bool = False
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
            if self.line in KNOWN_EXCEPTIONS:
                self.is_exception = True

                return True

            problems.append(f'Malformed entry: "{self.line}"')
            self.problems = problems
            return False

        self.pr_number = int(match.group("pr"))
        self.category = match.group("category")
        self.backslash = True if match.group("bs") else False
        self.link = match.group("link")
        self.description = match.group("desc")
        self.whitespaces = [
            match.group("ws1"),
            match.group("ws2"),
            match.group("ws3"),
            match.group("ws4"),
        ]

        if self.backslash:
            problems.append(
                "There should be no backslash in front of the # in the PR link"
            )

        ws_problems = check_whitespace(self.whitespaces)
        if ws_problems:
            problems.extend(ws_problems)

        fixed_cat, cat_problems = check_category(self.category)
        if cat_problems:
            problems.extend(cat_problems)

        fixed_link, link_problems = check_link(self.link, self.pr_number)
        if link_problems:
            problems.extend(link_problems)

        fixed_desc, description_problems = check_description(self.description)
        if description_problems:
            problems.extend(description_problems)

        self.fixed = f"- ({fixed_cat}) [#{self.pr_number}]({fixed_link}) {fixed_desc}"
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
        problems.append(
            "There should be exactly one space between the leading dash and the category"
        )

    if whitespaces[1] != " ":
        problems.append(
            "There should be exactly one space between the category and PR link"
        )

    if whitespaces[2] != "":
        problems.append("There should be no whitespace inside of the markdown link")

    if whitespaces[3] != " ":
        problems.append(
            "There should be exactly one space between the PR link and the description"
        )

    return problems


def check_category(category: str) -> Tuple[str, List[str]]:
    """
    Check if the category is valid.

    :param category: the category to check
    :return: a tuple containing the fixed category and a list of problems, which is empty if there are none
    """

    problems: List[str] = []
    fixed: str = category

    if not category.islower():
        problems.append(f'Category should be lowercase: "({category})"')
        fixed = category.lower()

    if category.lower() not in ALLOWED_CATEGORIES:
        problems.append(f'Invalid change category: "({category})"')

    return fixed, problems


def check_link(link: str, pr_number: int) -> Tuple[str, List[str]]:
    """
    Check if the link is valid.

    :param link: the link to check
    :param pr_number: the PR number to match in the link
    :return: a tuple containing the fixed link and a list of problems, which is empty if there are none
    """

    problems: List[str] = []
    fixed = link

    if not link.startswith("https://github.com/evmos/evmos/pull/"):
        fixed = f"https://github.com/evmos/evmos/pull/{pr_number}"
        problems.append(f'PR link should point to evmos repository: "{link}"')

    if str(pr_number) not in link:
        fixed = f"https://github.com/evmos/evmos/pull/{pr_number}"
        problems.append(f'PR link is not matching PR number {pr_number}: "{link}"')

    return fixed, problems


def check_description(description: str) -> Tuple[str, List[str]]:
    """
    Check if the description is valid.

    :param description: the description to check
    :return: a tuple containing the fixed description and a list of problems, which is empty if there are none
    """

    problems: List[str] = []
    fixed: str = description

    if re.search(r"\w", description[0]) and not description[0].isupper():
        fixed = description[0].upper() + description[1:]
        problems.append(
            f'PR description should start with capital letter: "{description}"'
        )

    if description[-1] != ".":
        problems.append(f'PR description should end with a dot: "{description}"')
        fixed += "."

    _, fixed, abbreviation_problems = check_spelling(fixed, ALLOWED_SPELLINGS)
    if abbreviation_problems:
        problems.extend(abbreviation_problems)

    return fixed, problems


def check_spelling(
    description: str, expected_spellings: Dict[str, re.Pattern]
) -> Tuple[bool, str, List[str]]:
    """
    Checks some common spelling requirements.
    Any matches that occur inside of code blocks, are part of a link or inside a word are ignored.

    :param expected_spellings: a dictionary of expected spellings and the matching patterns
    :param description: the description to check
    :return: a tuple containing a boolean value indicating whether a matching pattern was found and
             a list of problems with the match
    """

    problems: List[str] = []
    found: bool = False
    fixed: str = description

    for spelling, pattern in expected_spellings.items():
        match = get_match(pattern, description)
        if match:
            if match != spelling:
                problems.append(f'"{spelling}" should be used instead of "{match}"')
                fixed = pattern.sub(spelling, fixed)
            found = True

    return found, fixed, problems


def get_match(pattern: re.Pattern, text: str) -> str:
    """
    Returns the first match of the pattern in the text.
    Matching patterns inside of code blocks, inside of links or inside of words are ignored.

    :param pattern: the pattern to match
    :param text: the text to match against
    :return: the first match of the pattern in the text
    """

    codeblocks_pattern = re.compile(
        r"`[^`]*(" + pattern.pattern + r")[^`]*`", pattern.flags
    )
    match = codeblocks_pattern.search(text)
    if match:
        return ""

    isolated_word_pattern = re.compile(
        r"(^|\s)(" + pattern.pattern + r")(?=$|[\s.])", pattern.flags
    )
    match = isolated_word_pattern.search(text)
    if match:
        return match.group(2)

    return ""


if __name__ == "__main__":
    print("This is a library file and should not be executed directly.")
