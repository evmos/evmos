"""
This file contains the script to compile all Solidity smart contracts
in this repository.

Usage:
    python3 compile_smart_contracts.py

"""

import os
import re
from pathlib import Path
from typing import List


def find_solidity_files(path: Path) -> List[Path]:
    """
    Finds all Solidity files in the given Path.
    """

    solidity_files: List[Path] = []

    for root, _, files in os.walk(path):
        print("root: ", root)
        for file in files:
            print(file)
            if re.search(r"(?!\.dbg)\.sol$", file):
                solidity_files.append(Path(root) / file)

    return solidity_files
