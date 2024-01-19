from release import Release


class TestRelease:
    def test_pass(self):
        release = Release(
            "## [v15.0.2](https://github.com/evmos/evmos/releases/tag/v15.0.2) - 2021-08-02"
        )
        assert release.parse() is True
        assert release.version == "v15.0.2"
        assert release.problems == []

        # Check version comparisons
        assert (release <= 15) is True
        assert (release <= 14) is False

    def test_pass_unreleased(self):
        release = Release("## Unreleased")
        assert release.parse() is True
        assert release.version == "Unreleased"
        assert release.problems == []

    def test_malformed(self):
        release = Release("## `v15.0.2])")
        assert release.parse() is False
        assert release.version == ""
        assert release.problems == ['Malformed release header: "## `v15.0.2])"']

    def test_missing_link(self):
        release = Release("## [v15.0.2] - 2021-08-02")
        assert release.parse() is False
        assert release.version == "v15.0.2"
        assert release.problems == ['Release link is missing for "v15.0.2"']

    def test_wrong_version_in_link(self):
        release = Release(
            "## [v15.0.2](https://github.com/evmos/evmos/releases/tag/v16.0.0) - 2021-08-02"
        )
        assert release.parse() is False
        assert release.version == "v15.0.2"
        assert release.problems == [
            'Release header version "v15.0.2" does not match version in link '
            + '"https://github.com/evmos/evmos/releases/tag/v16.0.0"'
        ]

    def test_wrong_base_url(self):
        release = Release(
            "## [v15.0.2](https://github.com/evmos/evmds/releases/tag/v15.0.2) - 2021-08-02"
        )
        assert release.parse() is False
        assert release.version == "v15.0.2"
        assert release.problems == [
            'Release link should point to an Evmos release: "https://github.com/evmos/evmds/releases/tag/v15.0.2"'
        ]
