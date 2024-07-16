import os
from shutil import copyfile

import pytest  # type: ignore
from check_changelog import Changelog  # type: ignore

# Get the directory of this script
SCRIPT_DIR = os.path.dirname(os.path.realpath(__file__))


@pytest.fixture
def create_tmp_copy():
    tmp_file = os.path.join(SCRIPT_DIR, "testdata", "changelog_tmp.md")
    copyfile(
        os.path.join(SCRIPT_DIR, "testdata", "changelog_fail.md"),
        tmp_file,
    )
    yield tmp_file
    os.remove(tmp_file)


class TestParseChangelog:
    """
    This class collects all tests that are actually parsing dummy changelogs stored in
    markdown files in the testdata directory.
    """

    def test_pass(self):
        expected_result = {
            "Unreleased": {
                "State Machine Breaking": {
                    1922: {"description": "Add `secp256r1` curve precompile."},
                    1949: {"description": "Add `ClaimRewards` custom transaction."},
                    2218: {
                        "description": "Use correct version of proto dependencies to generate swagger."
                    },
                    1687: {"description": "Bump Evmos version to v14."},
                },
                "API Breaking": {
                    2015: {
                        "description": "Rename `inflation` module to `inflation/v1`."
                    },
                    2078: {"description": "Deprecate legacy EIP-712 ante handler."},
                    1851: {
                        "description": "Enable [EIP 3855](https://eips.ethereum.org/EIPS/eip-3855) "
                        + "(`PUSH0` opcode) during upgrade."
                    },
                },
                "Improvements": {
                    1864: {
                        "description": "Add `--base-fee` and `--min-gas-price` flags.",
                    },
                    1912: {"description": "Add Stride outpost interface and ABI."},
                    2104: {
                        "description": "Refactor to use `sdkmath.Int` and `sdkmath.LegacyDec` instead of SDK types."
                    },
                    701: {"description": "Rename Go module to `evmos/evmos`."},
                },
                "Bug Fixes": {
                    1801: {"description": "Fixed the problem `gas_used` is 0."},
                    109: {
                        "description": "Fix hardcoded ERC-20 nonce and `UpdateTokenPairERC20` proposal "
                        + "to support ERC-20s with 0 decimals."
                    },
                },
            },
            "v15.0.0": {
                "API Breaking": {
                    1862: {
                        "description": "Add Authorization Grants to the Vesting extension."
                    },
                    555: {"description": "`v4.0.0` upgrade logic."},
                },
            },
            "v2.0.0": {},
        }

        changelog = Changelog(os.path.join(SCRIPT_DIR, "testdata", "changelog_ok.md"))
        ok = changelog.parse()
        assert changelog.problems == [], "expected no failed entries"
        assert ok is True
        assert changelog.releases == expected_result, "expected different parsed result"

    def test_fail(self):
        changelog = Changelog(os.path.join(SCRIPT_DIR, "testdata", "changelog_fail.md"))
        assert changelog.parse() is False
        assert changelog.problems == [
            'PR link is not matching PR number 1948: "https://github.com/evmos/evmos/pull/1949"',
            "There should be no backslash in front of the # in the PR link",
            '"ABI" should be used instead of "ABi"',
            '"outpost" should be used instead of "Outpost"',
            'PR description should end with a dot: "Fixed the problem `gas_used` is 0"',
            '"Invalid Category" is not a valid change type',
            'Change type "Bug Fixes" is duplicated in Unreleased',
            "PR #1801 is duplicated in the changelog",
            'Release "v15.0.0" is duplicated in the changelog',
            'Change type "API Breaking" is duplicated in v15.0.0',
            "PR #1862 is duplicated in the changelog",
            'Malformed entry: "- malformed entry in changelog"',
        ]

    def test_fix(self, create_tmp_copy):
        changelog = Changelog(create_tmp_copy)
        assert changelog.parse(fix=True) is False
        assert changelog.problems == [
            'PR link is not matching PR number 1948: "https://github.com/evmos/evmos/pull/1949"',
            "There should be no backslash in front of the # in the PR link",
            '"ABI" should be used instead of "ABi"',
            '"outpost" should be used instead of "Outpost"',
            'PR description should end with a dot: "Fixed the problem `gas_used` is 0"',
            '"Invalid Category" is not a valid change type',
            'Change type "Bug Fixes" is duplicated in Unreleased',
            "PR #1801 is duplicated in the changelog",
            'Release "v15.0.0" is duplicated in the changelog',
            'Change type "API Breaking" is duplicated in v15.0.0',
            "PR #1862 is duplicated in the changelog",
            'Malformed entry: "- malformed entry in changelog"',
        ]

        # Here we parse the fixed changelog again and check that the automatic fixes were applied.
        fixed_changelog = Changelog(changelog.filename)
        assert fixed_changelog.parse(fix=False) is False
        assert fixed_changelog.problems == [
            '"Invalid Category" is not a valid change type',
            'Change type "Bug Fixes" is duplicated in Unreleased',
            "PR #1801 is duplicated in the changelog",
            'Release "v15.0.0" is duplicated in the changelog',
            'Change type "API Breaking" is duplicated in v15.0.0',
            "PR #1862 is duplicated in the changelog",
            'Malformed entry: "- malformed entry in changelog"',
        ]

    def test_parse_changelog_nonexistent_file(self):
        with pytest.raises(FileNotFoundError):
            Changelog(os.path.join(SCRIPT_DIR, "testdata", "nonexistent_file.md"))
