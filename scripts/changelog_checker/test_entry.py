from entry import (
    check_category,
    check_description,
    check_link,
    check_spelling,
    check_whitespace,
    Entry,
)


class TestEntry:
    """
    This class collects all tests that are checking individual changelog entries.
    """

    def test_entry_ok(self):
        entry = Entry(
            "- (distribution-precompile) [#1949](https://github.com/evmos/evmos/pull/1949) " +
            "Add `ClaimRewards` custom transaction."
        )
        assert entry.parse() is True

    def test_entry_wrong_pr_link_and_missing_dot(self):
        entry = Entry(
            "- (distribution-precompile) [#1949](https://github.com/evmos/evmos/pull/1948) " +
            "Add `ClaimRewards` custom transaction"
        )
        assert entry.parse() is False
        assert entry.problems == [
            'PR link is not matching PR number 1949: "https://github.com/evmos/evmos/pull/1948"',
            'PR description should end with a dot: "Add `ClaimRewards` custom transaction"'
        ]

    def test_malformed_entry(self):
        entry = Entry(
            "- (distribution-precompile) [#194tps://github.com/evmos/evmos/pull/1"
        )
        assert entry.parse() is False
        assert entry.problems == [
            'Malformed entry: "- (distribution-precompile) [#194tps://github.com/evmos/evmos/pull/1"'
        ]


class TestCheckCategory:
    def test_pass(self):
        assert check_category("evm") == []

    def test_invalid_category(self):
        assert check_category("invalid") == ['Invalid change category: "(invalid)"']

    def test_non_lower_category(self):
        assert check_category("eVm") == ['Category should be lowercase: "(eVm)"']


class TestCheckLink:
    def test_pass(self):
        assert check_link("https://github.com/evmos/evmos/pull/1949", 1949) == []

    def test_wrong_base_url(self):
        assert check_link("https://github.com/evmds/evmos/pull/1949", 1949) == [
            'PR link should point to evmos repository: "https://github.com/evmds/evmos/pull/1949"'
        ]

    def test_wrong_pr_number(self):
        assert check_link("https://github.com/evmos/evmos/pull/1948", 1949) == [
            'PR link is not matching PR number 1949: "https://github.com/evmos/evmos/pull/1948"'
        ]


class TestCheckDescription:
    def test_pass(self):
        assert check_description("Add `ClaimRewards` custom transaction.") == []

    def test_start_with_lowercase(self):
        assert check_description("add `ClaimRewards` custom transaction.") == [
            'PR description should start with capital letter: "add `ClaimRewards` custom transaction."'
        ]

    def test_end_with_dot(self):
        assert check_description("Add `ClaimRewards` custom transaction") == [
            'PR description should end with a dot: "Add `ClaimRewards` custom transaction"'
        ]


class TestCheckWhitespace:
    def test_missing_whitespace(self):
        assert check_whitespace(["", " ", " "]) == [
            'There should be exactly one space between the leading dash and the category'
        ]

    def test_multiple_spaces(self):
        assert check_whitespace([" ", " ", "  "]) == [
            'There should be exactly one space between the PR link and the description'
        ]


class TestCheckSpelling:
    def test_pass(self):
        assert check_spelling("Fix API.") == []

    def test_spelling(self):
        assert check_spelling("Fix APi.") == [
            '"API" should be used instead of "APi"'
        ]

    def test_erc_20(self):
        assert check_spelling("Add ERC20 contract.") == [
            '"ERC-20" should be used instead of "ERC20"'
        ]
