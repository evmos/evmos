import os
import re
import sys
from typing import Dict, List

from entry import Entry
from change_type import ChangeType

# Allowed release pattern: vX.Y.Z(-rcN) (YYYY-MM-DD)
RELEASE_PATTERN = re.compile(
    r'^## (Unreleased|\[(?P<version>v\d+\.\d+\.\d+(-rc\d+)?)] - \d{4}-\d{2}-\d{2})$',
)


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
                # TODO: extract release type
                release_match = RELEASE_PATTERN.match(line)
                if not release_match:
                    raise ValueError("Header 2 should be used for releases - invalid release pattern: " + line)

                version_match = release_match.group("version")
                current_release = version_match if version_match is not None else "Unreleased"
                self.releases[current_release] = {}
                continue

            # Check for Header 3 (###) to identify change types
            if stripped_line[:4] == '### ':
                change_type = ChangeType(line)
                current_category = change_type.type
                change_type.parse()
                self.failed_entries.extend(change_type.problems)

            # Check for individual entries
            if stripped_line[:2] != '- ':
                continue

            # TODO: order by extending the types by entries and then process afterwards within each release to have sorted output.
            entry = Entry(line)
            entry.parse()
            self.failed_entries.extend(entry.problems)

            self.releases[current_release][current_category][entry.pr_number] = {
                "description": entry.description
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


if __name__ == "__main__":
    changelog_file_path = sys.argv[1]
    if not os.path.exists(changelog_file_path):
        print('Changelog file not found')
        sys.exit(1)

    changelog = Changelog(sys.argv[1])
    passed = changelog.parse()
    if not passed:
        print(f'Changelog file is not valid - check the following {len(changelog.failed_entries)} problems:\n')
        print('\n'.join(changelog.failed_entries))
        sys.exit(1)
