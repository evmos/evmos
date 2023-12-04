import os
import pytest
from check_changelog import parse_changelog

@pytest.fixture
def sample_changelog_path(tmp_path):
    content = """
## Release 1.0.0

### Feature Changes
- [#1](https://github.com/example/repo/pull/1) Add new feature.

### Bug Fixes
- [#2](https://github.com/example/repo/pull/2) Fix a bug.

## Release 0.1.0

### Documentation
- [#3](https://github.com/example/repo/pull/3) Update documentation.

- This is not a valid entry and should be ignored.
"""

    file_path = tmp_path / "test_changelog.md"
    with open(file_path, "w", encoding="utf-8") as file:
        file.write(content)

    return file_path

def test_parse_changelog(sample_changelog_path):
    expected_result = {
        'Release 1.0.0': {
            'Feature Changes': {
                '1': {'link': 'https://github.com/example/repo/pull/1', 'description': 'Add new feature.'},
            },
            'Bug Fixes': {
                '2': {'link': 'https://github.com/example/repo/pull/2', 'description': 'Fix a bug.'},
            },
        },
        'Release 0.1.0': {
            'Documentation': {
                '3': {'link': 'https://github.com/example/repo/pull/3', 'description': 'Update documentation.'},
            },
        },
    }

    result = parse_changelog(sample_changelog_path)
    assert result == expected_result

def test_parse_changelog_nonexistent_file():
    with pytest.raises(FileNotFoundError):
        parse_changelog("nonexistent_file.md")

