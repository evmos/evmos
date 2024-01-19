import re

from entry import (  # type: ignore
    ALLOWED_SPELLINGS,
    Entry,
    check_category,
    check_description,
    check_link,
    check_spelling,
    check_whitespace,
    get_match,
)


class TestEntry:
    """
    This class collects all tests that are checking individual changelog entries.
    """

    example = (
        "- (distribution-precompile) [#1949](https://github.com/evmos/evmos/pull/1949) "
        + "Add `ClaimRewards` custom transaction."
    )

    def test_pass(self):
        entry = Entry(self.example)
        ok = entry.parse()
        assert entry.problems == []
        assert ok is True
        assert entry.fixed == self.example

    def test_pass_includes_link(self):
        example = (
            "- (evm) [#1851](https://github.com/evmos/evmos/pull/1851) "
            + "Enable [EIP 3855](https://eips.ethereum.org/EIPS/eip-3855) (`PUSH0` opcode) during upgrade."
        )
        entry = Entry(example)
        ok = entry.parse()
        assert entry.link == "https://github.com/evmos/evmos/pull/1851"
        assert entry.description == (
            "Enable [EIP 3855](https://eips.ethereum.org/EIPS/eip-3855) (`PUSH0` opcode) during upgrade."
        )
        assert entry.problems == []
        assert ok is True
        assert entry.fixed == example

    def test_fail_has_backslash_in_link(self):
        example = r"- (evm) [\#1851](https://github.com/evmos/evmos/pull/1851) Test."
        entry = Entry(example)
        ok = entry.parse()
        assert entry.problems == [
            "There should be no backslash in front of the # in the PR link"
        ]
        assert ok is False
        assert entry.fixed == example.replace(r"\#", "#")

    def test_entry_wrong_pr_link_and_missing_dot(self):
        entry = Entry(
            "- (distribution-precompile) [#1949](https://github.com/evmos/evmos/pull/1948) "
            + "Add `ClaimRewards` custom transaction"
        )
        assert entry.parse() is False
        assert entry.fixed == self.example
        assert entry.problems == [
            'PR link is not matching PR number 1949: "https://github.com/evmos/evmos/pull/1948"',
            'PR description should end with a dot: "Add `ClaimRewards` custom transaction"',
        ]

    def test_malformed_entry(self):
        malformed_example = (
            "- (distribution-precompile) [#194tps://github.com/evmos/evmos/pull/1"
        )
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

    def test_start_with_codeblock(self):
        fixed, problems = check_description(
            "```\nAdd `ClaimRewards` custom transaction."
        )
        assert fixed == "```\nAdd `ClaimRewards` custom transaction."
        assert problems == []


class TestCheckWhitespace:
    def test_missing_whitespace(self):
        assert check_whitespace(["", " ", "", " "]) == [
            "There should be exactly one space between the leading dash and the category"
        ]

    def test_multiple_spaces(self):
        assert check_whitespace([" ", " ", "", "  "]) == [
            "There should be exactly one space between the PR link and the description"
        ]

    def test_space_in_link(self):
        assert check_whitespace([" ", " ", " ", " "]) == [
            "There should be no whitespace inside of the markdown link"
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
        found, fixed, problems = check_spelling(
            "Fix Stride Outpost and AbI.", ALLOWED_SPELLINGS
        )
        assert found is True
        assert fixed == "Fix Stride outpost and ABI."
        assert problems == [
            '"ABI" should be used instead of "AbI"',
            '"outpost" should be used instead of "Outpost"',
        ]

    def test_pass_codeblocks(self):
        found, fixed, problems = check_spelling("Fix `in evm code`.", ALLOWED_SPELLINGS)
        assert found is False
        assert fixed == "Fix `in evm code`."
        assert problems == []

    def test_fail_in_word(self):
        found, fixed, problems = check_spelling("FixAbI in word.", ALLOWED_SPELLINGS)
        assert found is False
        assert fixed == "FixAbI in word."
        assert problems == []

    def test_erc_20(self):
        found, fixed, problems = check_spelling(
            "Add ERC20 contract.", ALLOWED_SPELLINGS
        )
        assert found is True
        assert fixed == "Add ERC-20 contract."
        assert problems == ['"ERC-20" should be used instead of "ERC20"']


class TestGetMatch:
    def test_pass(self):
        assert get_match(re.compile("abi", re.IGNORECASE), "Fix ABI.") == "ABI"

    def test_fail_codeblocks(self):
        assert get_match(re.compile("abi", re.IGNORECASE), "Fix `in AbI code`.") == ""

    def test_fail_in_word(self):
        assert get_match(re.compile("abi", re.IGNORECASE), "FixAbI in word.") == ""

    def test_fail_in_link(self):
        assert (
            get_match(
                re.compile("abi", re.IGNORECASE),
                "Fix [abcdef](https://example/aBi.com).",
            )
            == ""
        )
