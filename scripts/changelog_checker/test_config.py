from config import get_allowed_categories


class TestGetAllowedCategories:
    def test_pass(self):
        allowed_categories = get_allowed_categories()
        assert (
            "app" in allowed_categories
        ), "expected pre-configured value to be in allowed categories"
        assert (
            "evm" in allowed_categories
        ), "expected module to be in allowed categories"
        assert (
            "osmosis-outpost" in allowed_categories
        ), "expected outpost to be in allowed categories"
        assert (
            "distribution-precompile" in allowed_categories
        ), "expected precompile to be in allowed categories"
