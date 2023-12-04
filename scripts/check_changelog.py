import os
import re
import sys


def parse_changelog(file_path) -> bool:
    with open(file_path, 'r') as file:
        content = file.read()

    releases = {}
    current_release = None
    current_category = None

    for line in content.split('\n'):
        # Check for Header 2 (##) to identify releases
        release_match = re.match(r'^##\s+(.+)$', line)
        if release_match:
            current_release = release_match.group(1)
            releases[current_release] = {}
            continue

        # Check for Header 3 (###) to identify categories
        category_match = re.match(r'^###\s+(.+)$', line)
        if category_match:
            current_category = category_match.group(1)
            releases[current_release][current_category] = {}
            continue

        # Check for individual entries
        entry_match = re.match(r'^-\s+\[#(\d+)\]\((.+)\)\s+(.+)$', line)
        if entry_match and current_category:
            pr_number = entry_match.group(1)
            pr_link = entry_match.group(2)
            pr_description = entry_match.group(3)
            releases[current_release][current_category][pr_number] = {
                'link': pr_link,
                'description': pr_description
            }

    return releases


if __name__ == "__main__":
    changelog_file_path = sys.argv[1]
    if not os.path.exists(changelog_file_path):
        print('Changelog file not found')
        sys.exit(1)

    ok = parse_changelog(changelog_file_path)
    if not ok:
        print('Changelog file is not valid')
        sys.exit(1)
