import os
from pathlib import Path
from shutil import copytree

import pytest
from compile_smart_contracts import (
    HARDHAT_PROJECT_DIR,
    SOLIDITY_SOURCE,
    compile_contracts_in_dir,
    copy_to_contracts_directory,
    find_solidity_contracts,
    is_evmos_repo,
    is_ignored_folder,
)


@pytest.fixture
def setup_example_contracts_files(tmp_path):
    """
    This fixture creates a temporary folder with some Solidity files.
    """

    (tmp_path / "Contract1.sol").touch()
    (tmp_path / "Contract1.json").touch()
    (tmp_path / "Contract2.sol").touch()
    # NOTE: we're not adding the JSON file for Contract2

    (tmp_path / HARDHAT_PROJECT_DIR).mkdir()
    (tmp_path / HARDHAT_PROJECT_DIR / "Contract3.sol").touch()
    (tmp_path / HARDHAT_PROJECT_DIR / "Contract3.json").touch()

    (tmp_path / "precompiles").mkdir()
    (tmp_path / "precompiles" / "Contract4.sol").touch()
    (tmp_path / "precompiles" / "Contract4.json").touch()

    return tmp_path


def test_find_solidity_files(setup_example_contracts_files):
    tmp_path = setup_example_contracts_files
    found_solidity_contracts = find_solidity_contracts(tmp_path)
    assert len(found_solidity_contracts) == 4

    assert found_solidity_contracts[0].filename == "Contract2"
    assert found_solidity_contracts[0].path == tmp_path / "Contract2.sol"
    assert found_solidity_contracts[0].relative_path == Path(".")
    assert found_solidity_contracts[0].compiledJSONPath is None

    assert found_solidity_contracts[1].filename == "Contract1"
    assert found_solidity_contracts[1].path == tmp_path / "Contract1.sol"
    assert found_solidity_contracts[1].relative_path == Path(".")
    assert found_solidity_contracts[1].compiledJSONPath == Path(
        tmp_path / "Contract1.json"
    )

    assert found_solidity_contracts[2].filename == "Contract4"
    assert found_solidity_contracts[2].path == Path(
        tmp_path / "precompiles" / "Contract4.sol"
    )
    assert found_solidity_contracts[2].relative_path == Path("precompiles")
    assert found_solidity_contracts[2].compiledJSONPath == Path(
        tmp_path / "precompiles" / "Contract4.json"
    )

    assert found_solidity_contracts[3].filename == "Contract3"
    assert found_solidity_contracts[3].relative_path == Path(HARDHAT_PROJECT_DIR)
    assert found_solidity_contracts[3].compiledJSONPath == Path(
        tmp_path / HARDHAT_PROJECT_DIR / "Contract3.json"
    )


def test_copy_to_contracts_directory(
    tmp_path,
):
    target = tmp_path
    wd = Path(os.getcwd())
    assert is_evmos_repo(
        wd
    ), "This test should be executed from the top level of the Evmos repo"
    contracts = find_solidity_contracts(wd)

    assert os.listdir(target) == []
    assert copy_to_contracts_directory(target, contracts) is True

    dir_contents_post = os.listdir(target)
    assert len(dir_contents_post) > 0
    assert os.path.exists(
        target / "precompiles" / "staking" / "testdata" / "StakingCaller.sol"
    )


@pytest.fixture
def setup_contracts_directory(tmp_path):
    """
    This fixture creates a dummy hardhat project from the testdata folder.
    It will serve to test the compilation of smart contracts using this
    script's functions.
    """

    testdata_dir = Path(__file__).parent / "testdata"
    copytree(testdata_dir, tmp_path, dirs_exist_ok=True)

    return tmp_path


def test_compile_contracts_in_dir(setup_contracts_directory):
    hardhat_dir = setup_contracts_directory
    target_dir = hardhat_dir / SOLIDITY_SOURCE

    compile_contracts_in_dir(target_dir)
    assert os.path.exists(
        hardhat_dir
        / "artifacts"
        / SOLIDITY_SOURCE
        / "SimpleContract.sol"
        / "SimpleContract.json"
    )


def test_is_ignored_folder():
    assert is_ignored_folder(f"abc/{HARDHAT_PROJECT_DIR}/{SOLIDITY_SOURCE}") is False
    assert (
        is_ignored_folder(f"abc/{HARDHAT_PROJECT_DIR}/{SOLIDITY_SOURCE}/precompiles")
        is True
    )
    assert (
        is_ignored_folder(f"abc/{HARDHAT_PROJECT_DIR}/{SOLIDITY_SOURCE}/01_example")
        is True
    )
    assert is_ignored_folder("abc/node_modules/precompiles") is True
    assert is_ignored_folder("abc/tests/solidity/precompiles") is True
    assert is_ignored_folder("abc/nix_tests/precompiles") is True
