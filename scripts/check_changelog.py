import os
import re
import sys
from typing import Dict, List, Tuple

# Allowed entry pattern: `- (module) [#PR](link) - description`
ENTRY_PATTERN = re.compile(
    r'^-\s+\([a-zA-Z0-9\-]+\) \[\\?#(?P<pr>\d+)\]\((?P<link>.+)\) (?P<desc>.+)$',
)

# Allowed release pattern: vX.Y.Z(-rcN) (YYYY-MM-DD)
RELEASE_PATTERN = re.compile(
    r'^## (Unreleased|\[(?P<version>v\d+\.\d+\.\d+(-rc\d+)?)\] - \d{4}-\d{2}-\d{2})$',
)

ALLOWED_CATEGORIES = [
    'API Breaking',
    'Bug Fixes',
    'Improvements',
    'State Machine Breaking',
]


def parse_changelog(file_path) -> Tuple[Dict[str, Dict[str, Dict[int, str]]], List[str]]:
    """
    This function parses the changelog and checks if the structure is as expected.
    """

    with open(file_path, 'r') as file:
        content = file.read()

    releases = {}
    failed_entries = []
    current_release = None
    current_category = None

    for line in content.split('\n'):
        # Check for Header 2 (##) to identify releases
        stripped_line = line.strip()
        if stripped_line [:3] == '## ':
            release_match = RELEASE_PATTERN.match(line)
            if not release_match:
                raise ValueError("Header 2 should be used for releases - invalid release pattern: " + line)

            current_release = release_match.group("version") if release_match.group("version") is not None else "Unreleased"
            releases[current_release] = {}
            continue

        # Check for Header 3 (###) to identify categories
        category_match = re.match(r'^###\s+(.+)$', line)
        if category_match:
            current_category = category_match.group(1)
            if current_category not in ALLOWED_CATEGORIES:
                failed_entries.append(f'Invalid change category in {current_release}: "{current_category}"')
            releases[current_release][current_category] = {}
            continue

        # Check for individual entries
        if stripped_line[:2] != '- ':
            continue

        entry_match = ENTRY_PATTERN.match(line)
        if not entry_match:
            failed_entries.append(f'Invalid entry in {current_release} - {current_category}: "{line}"')
            continue

        pr_number = entry_match.group("pr")
        pr_link = entry_match.group("link")
        pr_description = entry_match.group("desc")

        if pr_number not in pr_link:
            failed_entries.append(f'PR link is not matching PR number in {current_release} - {current_category}: "{line}"')

        if not pr_description[0].isupper():
            failed_entries.append(f'PR description should start with capital letter in {current_release} - {current_category}: "{line}"')

        if pr_description[-1] != '.':
            failed_entries.append(f'PR description should end with a dot in {current_release} - {current_category}: "{line}"')

        releases[current_release][current_category][int(pr_number)] = {
            'description': pr_description
        }

    return releases, failed_entries


if __name__ == "__main__":
    changelog_file_path = sys.argv[1]
    if not os.path.exists(changelog_file_path):
        print('Changelog file not found')
        sys.exit(1)

    _, fails = parse_changelog(changelog_file_path)
    if len(fails) > 0:
        print(f'Changelog file is not valid - check the following {len(fails)} problems:\n')
        print('\n'.join(fails))
        sys.exit(1)
