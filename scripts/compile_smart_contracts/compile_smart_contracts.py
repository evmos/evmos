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
REPO_PATH = Path(__file__).parent.parent.parent


# This is the main target directory inside of the contracts folder.
CONTRACTS_TARGET = REPO_PATH / "contracts" / "contracts"


# This list contains all files that should be ignored when scanning the
# repository for Solidity files.
IGNORED_FILES: List[str] = [
    # Ignored because it uses a different OpenZeppelin contracts version to
    # compile
    "ERC20Minter_OpenZeppelinV5.sol",
]


# This list contains all folders that should be ignored when scanning the
# repository for Solidity files.
IGNORED_FOLDERS: List[str] = [
    "nix_tests",
    "node_modules",
    "scripts",
    "tests/solidity",
    # Ignored because the files are already in the correct folder
    "contracts/contracts",
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
    compiledJSONPath: Path | None


def find_solidity_contracts(path: Path) -> List[Contract]:
    """
    Finds all Solidity files in the given Path.
    It also checks if the compiled JSON file is present (in the same directory)
    which is the indicator if the compilation result should be copied
    back to the source directory.
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
                compiledJSONPath = Path(root) / f"{filename}.json"
                if not os.path.exists(compiledJSONPath):
                    print(
                        "failed to find compiled JSON file for contract: ",
                        file
                    )
                    compiledJSONPath = None

                solidity_files.append(
                    Contract(
                        filename=filename,
                        path=Path(os.path.join(root, file)),
                        relative_path=relative_path,
                        compiledJSONPath=compiledJSONPath,
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
        # if sub dir already exists this is skipped when using exist_ok=True
        sub_dir.mkdir(parents=True, exist_ok=True)
        copy(contract.path, sub_dir)

        print(
            f"copying {contract.path} to contracts directory -- " +
            f"relative path: {contract.relative_path}",
        )

    return True


def is_evmos_repo(path: Path) -> bool:
    """
    This function checks if the given path is the root of the Evmos repository,
    where this script is designed to be executed.
    """

    print("Path: ", path)
    contents = os.listdir(path)

    if "go.mod" not in contents:
        return False

    with open(path / "go.mod", "r") as go_mod:
        while True:
            line = go_mod.readline()
            if not line:
                break

            if "module github.com/evmos/evmos" in line:
                return True

    return False


def compile_contracts_in_dir(target_dir: Path):
    """
    This function compiles the Solidity contracts in the target directory
    with Hardhat.
    """

    cur_dir = os.getcwd()

    # Change to the root directory of the hardhat setup to compile.
    os.chdir(target_dir.parent)
    if not os.path.exists("hardhat.config.js"):
        raise ValueError(
            "compilation can only work in a HardHat setup"
        )

    install_failed = os.system("npm install")
    if install_failed:
        raise ValueError("Failed to install npm packages.")

    compilation_failed = os.system("npx hardhat compile")
    if compilation_failed:
        raise ValueError("Failed to compile Solidity contracts.")

    os.chdir(cur_dir)


def copy_compiled_contracts_back_to_source(
    contracts: List[Contract],
    compiled_dir: Path,
):
    """
    This function checks if the given contracts have
    been compiled in the compilation target directory
    and copies those back, that have a corresponding JSON
    file found originally.
    """

    for contract in contracts:
        if contract.compiledJSONPath is None:
            continue

        compiledPath = CONTRACTS_TARGET / \
            contract.relative_path / f"{contract.filename}.json"
        if not os.path.exists(compiledPath):
            print(f"compiled JSON data not found for {contract.filename}")
            continue

        copy(compiledPath, contract.compiledJSONPath)


def clean_up_hardhat_project(hardhat_dir: Path):
    os.removedirs(hardhat_dir / "node_modules")
    os.removedirs(hardhat_dir / "artefacts")
    os.removedirs(hardhat_dir / "cache")


if __name__ == "__main__":
    if not is_evmos_repo(REPO_PATH):
        raise ValueError(
            "This script should only be executed in the evmos repository. " +
            f"Current path: {REPO_PATH}"
        )

    found_contracts = find_solidity_contracts(REPO_PATH)
    if not copy_to_contracts_directory(CONTRACTS_TARGET, found_contracts):
        raise ValueError("Failed to copy contracts to target directory.")
    print("Successfully copied contracts to target directory.")

    compile_contracts_in_dir(CONTRACTS_TARGET)

    copy_compiled_contracts_back_to_source(found_contracts, CONTRACTS_TARGET)

    # clean_up_hardhat_project(CONTRACTS_TARGET.parent)
