#!/usr/bin/env python3
"""
This file contains utility functions to adjust the Cosmos SDK and IBC-Go spec files
that are downloaded during the creation of the docs (e.g. using `make docs-serve`).
The downloads are specified in the `pre.sh` script in the docs folder.

Usage:
  $ ./format_cosmos_specs.py FILENAME [FLAGS...]

The different downloaded files need different adjustments. The following functionality
can be executed by passing the corresponding flags:

  - `--header`: Formats the markdown header 1 to fit the rest of the Evmos and Ethermint docs
  - `--order [POSITION]`: Adds a HTML command specifying the desired position in the sub-folder order
  - `--title [TITLE]`: Specifies the shown page title in the generated docs
  - `--parent [PARENT]`: Specifies the parent tile in the dropdown menu of the generated docs
"""

import getopt
import os
import re
import sys


def main():
    """
    Main function to execute the formatting of the Cosmos-SDK and IBC-Go specs.
    """

    if len(sys.argv) < 2:
        raise ValueError(
            "Script has to be called with a filename and optional flags to control the execution"
        )

    file = sys.argv[1]
    optlist, _ = getopt.gnu_getopt(
        sys.argv[2:], "ho:t:p:", ["header", "order=", "title=", "parent="]
    )

    # Initialize variables
    add_parent = False
    add_position = False
    add_title = False
    adjust_header = False
    position = 1
    title = ""
    parent = ""

    # Parse input arguments
    for key, value in optlist:
        if key in ("-o", "--order"):
            add_position = True
            position = value
        elif "--header" == key:
            adjust_header = True
        elif "--title" == key:
            add_title = True
            title = value
        elif "--parent" == key:
            add_parent = True
            parent = value

    if adjust_header:
        format_header_in_file(file)
    if add_position or add_title or add_parent:
        add_metadata(file, position, title, parent)

    return


def format_header_in_file(file: str) -> None:
    """
    format_header_in_file will adjust the formatting in the file at the given path to
    match the Evmos and Ethermint docs.
    Additionally, it will remove all lines before the markdown header 1, because the
    docusaurus commands (like sidebar_position) are not interpreted in our setup.

    :param file: Path to a markdown file
    """

    if not os.path.exists(file):
        raise FileNotFoundError(f"File '{file}' not found.")

    filename, extension = os.path.splitext(file)
    tmp_file = f"{filename}_tmp{extension}"
    _write = False

    with open(file, "r") as f_read:
        with open(tmp_file, "w") as f_write:
            for line in f_read:
                if line.strip()[:2] == "# ":
                    f_write.write(format_header(line))
                    _write = True  # only include lines after the heading (to remove the "sidebar_position" lines)
                elif _write:
                    f_write.write(line)

    os.remove(file)
    os.rename(tmp_file, file)


def format_header(header: str) -> str:
    """
    format_header removes any formatting other than the header 1 setting from the given
    header string. Also, the module prefix "x/" is removed from the string and the string
    converted to lower case.

    :param header: String which contains a markdown header 1
    :return: adjusted string
    """

    split_header = header.split()
    if split_header[0] != "#":
        raise ValueError(
            f"Expected markdown header 1 (e.g. '# Example')\nGot: '{header}'"
        )

    formatted_header = re.sub(r"x/", "", split_header[1])
    formatted_header = formatted_header.lower()
    if "`" not in formatted_header:
        formatted_header = f"`{formatted_header}`"

    return f"# {formatted_header}\n"


def add_metadata(file: str, position: int, title: str, parent: str) -> None:
    """
    add_metadata adds lines to the beginning of the markdown file at the given path, which
    specify the position in the sub-folder order and define a title and parent title
    (which are used for the 'Modules' dropdown on the Evmos docs).

    :param file: path to the markdown file to be adjusted.
    :param position: integer value of the desired position
    """

    if not os.path.exists(file):
        raise FileNotFoundError(f"File '{file}' not found.")

    filename, extension = os.path.splitext(file)
    tmp_file = f"{filename}_tmp{extension}"

    added_string = f"<!--\norder: {position}\n"
    if title != "":
        added_string += f'title: "{title}"\n'
    if parent != "":
        added_string += f'parent:\n  title: "{parent}"\n'
    added_string += "-->\n\n"

    with open(file, "r") as f_read:
        with open(tmp_file, "w") as f_write:
            f_write.write(added_string)

            for line in f_read:
                f_write.write(line)

    os.remove(file)
    os.rename(tmp_file, file)


if __name__ == "__main__":
    main()
