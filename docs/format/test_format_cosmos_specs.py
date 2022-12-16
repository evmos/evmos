"""
test_format_cosmos_specs contains the unit-testing suite for
the format_cosmos_specs package.
"""

import os

import format_cosmos_specs as fcs
import pytest

# ------------------
# globals
#
FILE = "test.md"


# ------------------
# format_header tests
#
def test_format_header_should_return_string():
    header = "# `x/auth`\n"
    assert fcs.format_header(header) == "# `auth`\n"


def test_format_header_no_header_1_should_raise_ValueError():
    header = "## x/auth`\n"

    with pytest.raises(ValueError, match="Expected markdown header 1"):
        fcs.format_header(header)


@pytest.fixture()
def header_file_setup():
    with open(FILE, "w") as f:
        f.write("\n\n# `x/auth`\n\n")

    yield

    os.remove(FILE)


def test_format_header_in_file(header_file_setup):
    fcs.format_header_in_file(FILE)

    with open(FILE, "r") as f:
        file_contents = f.read()
    assert file_contents == "# `auth`\n\n"


# ------------------
# add_metadata tests
#
contents = "Test\n"


@pytest.fixture()
def order_file_setup():
    with open(FILE, "w") as f:
        f.write(contents)

    yield

    os.remove(FILE)


def test_add_order_should_add_to_file_contents(order_file_setup):
    fcs.add_metadata(FILE, 1, "", "")

    with open(FILE, "r") as f:
        file_contents = f.read()
    assert file_contents == "<!--\norder: 1\n-->\n\n" + contents


def test_add_title_should_add_to_file_contents(order_file_setup):
    title = "Test-Title"
    fcs.add_metadata(FILE, 1, title, "")

    with open(FILE, "r") as f:
        file_contents = f.read()
    assert file_contents == f'<!--\norder: 1\ntitle: "{title}"\n-->\n\n' + contents


def test_add_parent_should_add_to_file_contents(order_file_setup):
    parent = "Test-Parent"
    fcs.add_metadata(FILE, 1, "", parent)

    with open(FILE, "r") as f:
        file_contents = f.read()
    assert (
        file_contents
        == f'<!--\norder: 1\nparent:\n  title: "{parent}"\n-->\n\n' + contents
    )


def test_add_all_options_should_add_to_file_contents(order_file_setup):
    title = "Test-Title"
    parent = "Test-Parent"
    fcs.add_metadata(FILE, 1, title, parent)

    with open(FILE, "r") as f:
        file_contents = f.read()
    assert (
        file_contents
        == f'<!--\norder: 1\ntitle: "{title}"\nparent:\n  title: "{parent}"\n-->\n\n'
        + contents
    )
