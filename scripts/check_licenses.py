import os
import re
import sys
from typing import Dict, List

FILTER: re.Pattern = re.compile(r"^((?!(_test|\.pb|\.pb\.gw)\.go$).)*\.(go|proto)$")
EXEMPT_FILES: List[str] = [
    r"x/revenue/v1/",  # All files in this folder
    r"x/claims/genesis\.go$",
    r"x/erc20/keeper/proposals\.go$",
    r"x/erc20/types/utils\.go$",
]

LGPL3_LICENSE = [
    "// Copyright Tharsis Labs Ltd.(Evmos)\n",
    "// SPDX-License-Identifier:LGPL-3.0-only\n"
]

ENCL_LICENSE = [
    "// Copyright Tharsis Labs Ltd.(Evmos)\n",
    "// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)\n"
]


def check_licenses_in_path(
    path: str, name_filter: re.Pattern = re.compile(".*"), add: bool = False
) -> Dict[str, int]:
    """
    Iterate over all files in the current directory and its subdirectories
    and check if the appropriate licenses are contained in the files.
    """

    files_with_wrong_license = []
    n_files = 0
    n_files_with = 0
    n_files_generated = 0
    n_files_lic_removed = 0
    n_files_geth = 0
    n_files_lgpl3 = 0
    n_files_encl = 0

    for root, _, files in os.walk(path):
        for file in files:
            full_path = os.path.join(root, file)
            if not name_filter.search(full_path):
                continue

            n_files += 1

            lgpl3 = check_if_in_exempt_files(full_path)
            checked_license = LGPL3_LICENSE if lgpl3 else ENCL_LICENSE

            found = check_license_in_file(os.path.join(root, file), checked_license)
            if found is True:
                n_files_with += 1
                if lgpl3:
                    n_files_lgpl3 += 1
                else:
                    n_files_encl += 1
            elif found == "generated":
                n_files_with += 1
                n_files_generated += 1
                continue
            elif found == "geth":
                n_files_with += 1
                n_files_geth += 1
                continue
            else:
                files_with_wrong_license.append(full_path)

    print(f"\n{n_files_with}/{n_files} contain a license comment")
    print(f" -> {n_files_generated} are generated files")
    print(f" -> {n_files_geth} have a geth license")
    print(f" -> {n_files_lgpl3} have the LGPL3 license")
    print(f" -> {n_files_encl} have the ENCL license")
    if len(files_with_wrong_license) > 0:
        print("---------------------------")
        print(
            f""" -> {len(files_with_wrong_license)} files have the wrong license or are missing a license altogether!
    Please check the output above."""
        )

    return {
        "total": n_files,
        "with_license": n_files_with,
        "generated": n_files_generated,
        "geth": n_files_geth,
        "license_removed": n_files_lic_removed,
        "lgpl3": n_files_lgpl3,
        "encl": n_files_encl,
        "wrong_license": len(files_with_wrong_license),
    }


def check_if_in_exempt_files(file: str) -> bool:
    """
    Check if the file is in the exempt files list.
    """

    for exempt_file in EXEMPT_FILES:
        if re.search(exempt_file, file):
            return True
    return False


def check_license_in_file(file: str, checked_license: List[str]) -> bool | str:
    """
    Check if the file has the license.
    """

    with open(file, "r") as f:
        lines = f.readlines()

        if "generated" in lines[0].lower() or "do not edit" in lines[0].lower():
            return "generated"
        elif "ethereum" in lines[0].lower():
            return "geth"

        for expected_line, line in zip(checked_license, lines[: len(checked_license)]):
            if line != expected_line:
                print(" - ", file)
                return False

        return True


if __name__ == "__main__":
    result = check_licenses_in_path(sys.argv[1], FILTER)
    if result["wrong_license"] > 0:
        sys.exit(1)
