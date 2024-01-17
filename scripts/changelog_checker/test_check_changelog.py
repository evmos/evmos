import os
import pytest
from check_changelog import Changelog, Entry

# Get the directory of this script
SCRIPT_DIR = os.path.dirname(os.path.realpath(__file__))


class TestParseChangelog:
    """
    This class collects all tests that are actually parsing dummy changelogs stored in
    markdown files in the testdata directory.
    """

    def test_pass(self):
        expected_result = {
            'Unreleased': {
                'State Machine Breaking': {
                    1922: {'description': 'Add `secp256r1` curve precompile.'},
                    1949: {'description': 'Add `ClaimRewards` custom transaction.'},
                },
                'API Breaking': {
                    2015: {'description': 'Rename `inflation` module to `inflation/v1`.'},
                    2078: {'description': 'Deprecate legacy EIP712 ante handler.'},
                },
                'Improvements': {
                    1864: {
                        'description':
                            'Add `--base-fee` and `--min-gas-price` flags.',
                    },
                    1912: {'description': 'Add Stride Outpost interface and ABI.'},
                },
                'Bug Fixes': {
                    1801: {'description': 'Fixed the problem gas_used is 0.'},
                },
            },
            'v15.0.0': {
                'API Breaking': {
                    1862: {'description': 'Add Authorization Grants to the Vesting extension.'},
                },
            },
        }

        changelog = Changelog(os.path.join(SCRIPT_DIR, "testdata", "changelog_ok.md"))
        ok = changelog.parse()
        assert ok
        assert changelog.failed_entries == [], "expected no failed entries"
        assert changelog.releases == expected_result, "expected different parsed result"

    # TODO: uncomment when done with refactor
    # def test_parse_changelog_invalid_pr_link(self):
    #     expected_result = {
    #         'Unreleased': {
    #             'State Machine Breaking': {
    #                 1948: {'description': 'Add `ClaimRewards` custom transaction.'},
    #             },
    #         },
    #     }
    #     changelog = Changelog(os.path.join(SCRIPT_DIR, "testdata", "changelog_invalid_entry_pr_not_in_link.md"))
    #     ok = changelog.parse()
    #     assert not ok
    #     assert changelog.failed_entries == [
    #         'PR link is not matching PR number in Unreleased - State Machine Breaking: "- (distribution-precompile)' +
    #         ' [#1948](https://github.com/evmos/evmos/pull/1949) Add `ClaimRewards` custom transaction."',
    #     ]
    #     assert changelog.releases == expected_result
    #
    # def test_parse_changelog_malformed_description(self):
    #     expected_result = {
    #        'Unreleased': {
    #            'State Machine Breaking': {
    #                1949: {'description': 'add `ClaimRewards` custom transaction'},
    #            },
    #        },
    #     }
    #     releases, failed_entries = parse_changelog(
    #         os.path.join(SCRIPT_DIR, "testdata", "changelog_invalid_entry_misformatted_description.md"),
    #     )
    #     assert failed_entries == [
    #         'PR description should start with capital letter in Unreleased - State Machine Breaking: "- (distribution' +
    #         '-precompile) [#1949](https://github.com/evmos/evmos/pull/1949) add `ClaimRewards` custom transaction"',
    #         'PR description should end with a dot in Unreleased - State Machine Breaking: "- (distribution' +
    #         '-precompile) [#1949](https://github.com/evmos/evmos/pull/1949) add `ClaimRewards` custom transaction"',
    #     ]
    #     assert releases == expected_result
    #
    # def test_parse_changelog_invalid_category(self):
    #     expected_result = {
    #        'Unreleased': {
    #            'Invalid Category': {
    #                1949: {'description': 'Add `ClaimRewards` custom transaction.'},
    #            },
    #        },
    #     }
    #     releases, failed_entries = parse_changelog(
    #         os.path.join(SCRIPT_DIR, "testdata", "changelog_invalid_category.md")
    #     )
    #     assert failed_entries == ["Invalid change category in Unreleased: \"Invalid Category\""]
    #     assert releases == expected_result
    #
    # def test_parse_changelog_invalid_header(self):
    #     with pytest.raises(ValueError):
    #         parse_changelog(os.path.join(SCRIPT_DIR, "testdata", "changelog_invalid_version.md"))
    #
    # def test_parse_changelog_invalid_date(self):
    #     with pytest.raises(ValueError):
    #         parse_changelog(os.path.join(SCRIPT_DIR, "testdata", "changelog_invalid_date.md"))
    #
    # def test_parse_changelog_nonexistent_file(self):
    #     with pytest.raises(FileNotFoundError):
    #         parse_changelog("nonexistent_file.md")


class TestEntry:
    """
    This class collects all tests that are checking individual changelog entries.
    """

    def test_entry_ok(self):
        entry = Entry(
            "- (distribution-precompile) [#1949](https://github.com/evmos/evmos/pull/1949)" +
            "Add `ClaimRewards` custom transaction."
        )
        ok, reason = entry.parse()
        assert ok
        assert reason == ""

    def test_entry_wrong_pr_link(self):
        entry = Entry(
            "- (distribution-precompile) [#1949](https://github.com/evmos/evmos/pull/1948)" +
            "Add `ClaimRewards` custom transaction."
        )
        ok, reason = entry.parse()
        assert not ok
        assert reason == "PR link is not matching PR number"
