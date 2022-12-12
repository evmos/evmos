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
  - `--add-order [POSITION]`: Adds a HTML command specifying the desired position in the sub-folder order
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
            "Script has to be called with a filename and optional flags to be control the execution"
        )

    file = sys.argv[1]
    optlist, _ = getopt.gnu_getopt(sys.argv[2:], "ho:", ["header", "order"])

    _ADD_ORDER = False
    _ADJUST_HEADER = False
    _POSITION: int = ...
    for key, value in optlist:
        if "--order" == key:
            _ADD_ORDER = True
            _POSITION = value
        elif "--header" == key:
            _ADJUST_HEADER = True

    if _ADJUST_HEADER:
        format_header_in_file(file)
    if _ADD_ORDER:
        add_order(file, _POSITION)

    return


def format_header_in_file(file: str) -> None:
    """
    format_header_in_file will adjust the formatting in the file at the given path to
    match the Evmos and Ethermint docs.

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
    header string. Also, the module prefix "x/" is removed from the string.

    :param header: String which contains a markdown header 1
    :return: adjusted string
    """

    if header[:2] != "# ":
        raise ValueError(
            f"Expected markdown header 1 (e.g. '# Example')\nGot: '{header}'"
        )

    formatted_header = re.sub("`*(x/)*", "", header)

    return formatted_header


def add_order(file: str, position: int) -> None:
    """
    add_order adds lines to the beginning of the markdown file at the given path, which
    specify the position in the sub-folder order.

    :param file: path to the markdown file to be adjusted.
    :param position: integer value of the desired position
    """

    if not os.path.exists(file):
        raise FileNotFoundError(f"File '{file}' not found.")

    filename, extension = os.path.splitext(file)
    tmp_file = f"{filename}_tmp{extension}"
    added_string = f"<!--\norder: {position}\n-->"

    with open(file, "r") as f_read:
        with open(tmp_file, "w") as f_write:
            f_write.write(added_string + "\n")

            for line in f_read:
                f_write.write(line)

    os.remove(file)
    os.rename(tmp_file, file)


if __name__ == '__main__':
    main()
