from entry import (
    ALLOWED_SPELLINGS,
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

    example = (
            "- (distribution-precompile) [#1949](https://github.com/evmos/evmos/pull/1949) " +
            "Add `ClaimRewards` custom transaction."
    )

    def test_entry_ok(self):
        entry = Entry(self.example)
        assert entry.fixed == self.example
        assert entry.parse() is True
        assert entry.problems == []

    def test_entry_wrong_pr_link_and_missing_dot(self):
        entry = Entry(
            '- (distribution-precompile) [#1949](https://github.com/evmos/evmos/pull/1948) ' +
            'Add `ClaimRewards` custom transaction'
        )
        assert entry.parse() is False
        assert entry.fixed == self.example
        assert entry.problems == [
            'PR link is not matching PR number 1949: "https://github.com/evmos/evmos/pull/1948"',
            'PR description should end with a dot: "Add `ClaimRewards` custom transaction"'
        ]

    def test_malformed_entry(self):
        malformed_example = "- (distribution-precompile) [#194tps://github.com/evmos/evmos/pull/1"
        entry = Entry(malformed_example)
        assert entry.parse() is False
        assert entry.fixed == malformed_example
        assert entry.problems == [
            'Malformed entry: "- (distribution-precompile) [#194tps://github.com/evmos/evmos/pull/1"'
        ]


class TestCheckCategory:
    def test_pass(self):
        fixed, problems = check_category("evm")
        assert fixed == "evm"
        assert problems == []

    def test_invalid_category(self):
        fixed, problems = check_category("invalid")
        assert fixed == "invalid"
        assert problems == ['Invalid change category: "(invalid)"']

    def test_non_lower_category(self):
        fixed, problems = check_category("eVm")
        assert fixed == "evm"
        assert problems == ['Category should be lowercase: "(eVm)"']


class TestCheckLink:
    example = "https://github.com/evmos/evmos/pull/1949"
    def test_pass(self):
        fixed, problems = check_link(self.example, 1949)
        assert fixed == self.example
        assert problems == []

    def test_wrong_base_url(self):
        fixed, problems = check_link("https://github.com/evmds/evmos/pull/1949", 1949)
        assert fixed == self.example
        assert problems == [
            'PR link should point to evmos repository: "https://github.com/evmds/evmos/pull/1949"'
        ]

    def test_wrong_pr_number(self):
        fixed, problems = check_link("https://github.com/evmos/evmos/pull/1948", 1949)
        assert fixed == self.example
        assert problems == [
            'PR link is not matching PR number 1949: "https://github.com/evmos/evmos/pull/1948"'
        ]


class TestCheckDescription:
    def test_pass(self):
        example = "Add `ClaimRewards` custom transaction."
        fixed, problems = check_description(example)
        assert fixed == example
        assert problems == []

    def test_start_with_lowercase(self):
        fixed, problems = check_description("add `ClaimRewards` custom transaction.")
        assert fixed == "Add `ClaimRewards` custom transaction."
        assert problems == [
            'PR description should start with capital letter: "add `ClaimRewards` custom transaction."'
        ]

    def test_end_with_dot(self):
        fixed, problems = check_description("Add `ClaimRewards` custom transaction")
        assert fixed == "Add `ClaimRewards` custom transaction."
        assert problems == [
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
        found, fixed, problems = check_spelling("Fix API.", ALLOWED_SPELLINGS)
        assert found is True
        assert fixed == "Fix API."
        assert problems == []

    def test_spelling(self):
        found, fixed, problems = check_spelling("Fix APi.", ALLOWED_SPELLINGS)
        assert found is True
        assert fixed == "Fix API."
        assert problems == ['"API" should be used instead of "APi"']

    def test_multiple_problems(self):
        found, fixed, problems = check_spelling("Fix Stride Outpost and AbI.", ALLOWED_SPELLINGS)
        assert found is True
        assert fixed == "Fix Stride outpost and ABI."
        assert problems == [
            '"ABI" should be used instead of "AbI"',
            '"outpost" should be used instead of "Outpost"',
        ]

    def test_erc_20(self):
        found, fixed, problems = check_spelling("Add ERC20 contract.", ALLOWED_SPELLINGS)
        assert found is True
        assert fixed == "Add ERC-20 contract."
        assert problems == ['"ERC-20" should be used instead of "ERC20"']
