"""
This file contains the script to compile all Solidity smart contracts
in this repository.

Usage:
    python3 compile_smart_contracts.py

"""

from shutil import copy
import os
import re
from dataclasses import dataclass
from pathlib import Path
from typing import List

# The path to the main level of the Evmos repository.
EVMOS_REPO = Path(__file__).parent.parent.parent


# This is the main target directory inside of the contracts folder.
MAIN_CONTRACTS_TARGET = EVMOS_REPO / "contracts" / "contracts"


# This list contains all files that should be ignored when scanning the repository
# for Solidity files.
IGNORED_FILES: List[str] = [
    # Ignored because it uses a different OpenZeppelin contracts version to compile
    "ERC20Minter_OpenZeppelinV5.sol",
]


# This list contains all folders that should be ignored when scanning the repository
# for Solidity files.
IGNORED_FOLDERS: List[str] = [
    "contracts/contracts", # Ignored because the files are already in the correct folder
    "nix_tests",
    "node_modules",
    "scripts",
    "tests/solidity",
]


@dataclass
class Contract:
    """
    Dataclass to store the name and path of a Solidity contract
    as well as the path to where the compiled JSON data is stored.
    """

    filename: str
    path: Path
    relative_path: Path
    # TODO: Maybe this can also be removed again
    compiledJSONPath: Path


def find_solidity_contracts(path: Path) -> List[Contract]:
    """
    Finds all Solidity files in the given Path.
    """

    solidity_files: List[Contract] = []

    for root, _, files in os.walk(path):
        if is_ignored_folder(root):
            continue

        relative_path = Path(root).relative_to(path)

        for file in files:
            if file in IGNORED_FILES:
                print(f"Ignoring file: {file}")
                continue

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
                        path=Path(os.path.join(root, file)),
                        relative_path=relative_path,
                        compiledJSONPath=Path(root) / f"{filename}.json"
                    )
                )

    return solidity_files


def is_ignored_folder(path: str) -> bool:
    """
    Check if the folder is in the list of ignored folders.
    """

    return any([re.search(folder, path) for folder in IGNORED_FOLDERS])


def copy_to_contracts_directory(
    target_dir: Path,
    contracts: List[Contract]
) -> bool:
    """
    This function copies the list of Contracts found in the repository
    to the target directory.

    In the context of the fully-functional tool, this is necessary to compile
    them with Hardhat, which relies on the Solidity files to be nested within
    the `contracts` directory.
    """

    if not os.path.isdir(target_dir) or not os.path.exists(target_dir):
        return False

    for contract in contracts:
        sub_dir = target_dir / contract.relative_path
        sub_dir.mkdir(parents=True, exist_ok=True) # if sub dir already exists this is skipped when using exist_ok=True
        copy(contract.path, sub_dir)

        print(f"copying {contract.path} to contracts directory  relative path: {contract.relative_path}")

    return True

def is_evmos_repo(path: Path) -> bool:
    """
    This function checks if the given path is the root of the Evmos repository,
    where this script is designed to be executed.
    """

    print("Path: ", path)
    contents = os.listdir(path)

    if not "go.mod" in contents:
        return False

    with open(path / "go.mod", "r") as go_mod:
        for line in go_mod.readlines():
            if "module github.com/evmos/evmos" in line:
                return True

    return False


if __name__ == "__main__":
    dir_to_execute = Path(__file__).parent.parent.parent
    if not is_evmos_repo(dir_to_execute):
        raise ValueError("This script should only be executed in the evmos repository.")

    found_contracts = find_solidity_contracts(dir_to_execute)
    copy_to_contracts_directory(dir_to_execute / "contracts", found_contracts)
