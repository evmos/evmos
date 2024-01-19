from change_type import ChangeType


class TestChangeType:
    def test_pass(self):
        change_type = ChangeType("### Bug Fixes")
        assert change_type.parse() is True
        assert change_type.type == "Bug Fixes"
        assert change_type.problems == []

    def test_malformed(self):
        change_type = ChangeType("###Bug Fixes")
        assert change_type.parse() is False
        assert change_type.type == ""
        assert change_type.problems == ['Malformed change type: "###Bug Fixes"']

    def test_spelling(self):
        change_type = ChangeType("### BugFixes")
        assert change_type.parse() is False
        assert change_type.type == "BugFixes"
        assert change_type.problems == [
            '"Bug Fixes" should be used instead of "BugFixes"'
        ]

    def test_invalid_type(self):
        change_type = ChangeType("### Invalid Type")
        assert change_type.parse() is False
        assert change_type.type == "Invalid Type"
        assert change_type.problems == ['"Invalid Type" is not a valid change type']
