"""
This file contains the script to compile all Solidity smart contracts
in this repository.

Usage:
    python3 compile_smart_contracts.py

"""

import os
import re
from dataclasses import dataclass
from pathlib import Path
from typing import List


# This list contains all files that should be ignored when scanning the repository
# for Solidity files.
IGNORED_FILES: List[str] = [
    "ERC20Minter_OpenZeppelinV5.sol", # Ignored because it uses a different OpenZeppelin contracts version to compile
]


# This list contains all folders that should be ignored when scanning the repository
# for Solidity files.
IGNORED_FOLDERS: List[str] = [
    "scripts",
]


@dataclass
class Contract:
    """
    Dataclass to store the name and path of a Solidity contract
    as well as the path to where the compiled JSON data is stored.
    """

    filename: str
    path: Path
    compiledJSONPath: Path


def find_solidity_contracts(path: Path) -> List[Contract]:
    """
    Finds all Solidity files in the given Path.
    """

    solidity_files: List[Contract] = []

    for root, _, files in os.walk(path):
        if is_ignored_folder(root):
            continue

        for file in files:
            if re.search(r"(?!\.dbg)\.sol$", file):
                filename = os.path.splitext(file)[0]
                compiledJSONPath = os.path.join(root, f"{filename}.json")
                if not os.path.exists(compiledJSONPath):
                    # TODO: collect failed compilations
                    print("failed to find compiled JSON file for contract: ", file)
                    continue

                solidity_files.append(
                    Contract(
                        filename=file,
                        path=Path(root),
                        compiledJSONPath=Path(root) / f"{file}.json"
                    )
                )

    return solidity_files


def is_ignored_folder(path: str) -> bool:
    """
    Check if the folder is in the list of ignored folders.
    """

    return any([re.search(folder, path) for folder in IGNORED_FOLDERS])
