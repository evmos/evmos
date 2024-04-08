import os
import pytest
from compile_smart_contracts import *
from pathlib import Path


@pytest.fixture
def setup_example_contracts_files(tmp_path):
    """
    This fixture creates a temporary folder with some Solidity files.
    """

    (tmp_path / "Contract1.sol").touch()
    (tmp_path / "Contract1.json").touch()
    (tmp_path / "Contract2.sol").touch()
    # NOTE: we're not adding the JSON file for Contract2

    (tmp_path / "contracts").mkdir()
    (tmp_path / "contracts" / "Contract3.sol").touch()
    (tmp_path / "contracts" / "Contract3.json").touch()

    (tmp_path / "precompiles").mkdir()
    (tmp_path / "precompiles" / "Contract4.sol").touch()
    (tmp_path / "precompiles" / "Contract4.json").touch()

    return tmp_path


def test_find_solidity_files(setup_example_contracts_files):
    tmp_path = setup_example_contracts_files
    found_solidity_contracts = find_solidity_contracts(tmp_path)
    assert len(found_solidity_contracts) == 3

    assert found_solidity_contracts[0].filename == "Contract1.sol"
    assert found_solidity_contracts[0].path == tmp_path
    assert found_solidity_contracts[0].compiledJSONPath == Path(
        tmp_path / "Contract1.json"
    )

    assert found_solidity_contracts[1].filename == "Contract4.sol"
    assert found_solidity_contracts[1].path == Path(tmp_path / "precompiles")
    assert found_solidity_contracts[1].compiledJSONPath == Path(
        tmp_path / "precompiles" / "Contract4.json"
    )

    assert found_solidity_contracts[2].filename == "Contract3.sol"
    assert found_solidity_contracts[2].path == Path(tmp_path / "contracts")
    assert found_solidity_contracts[2].compiledJSONPath == Path(
        tmp_path / "contracts" / "Contract3.json"
    )


@pytest.fixture
def setup_contracts_directory(tmp_path):
    """
    This fixture creates the target contracts folder,
    where any found smart contracts should be copied to
    in order to be compiled with Hardhat.
    """

    # TODO: This could actually be removed if nothing is added
    # except using tmp_path
    return tmp_path


def test_copy_to_contracts_directory(
    setup_contracts_directory
):
    target = setup_contracts_directory
    contracts = find_solidity_contracts(os.getcwd())

    assert len(os.listdir(target)) == 0
    assert copy_to_contracts_directory(
        target,
        contracts
    ) is True
    # TODO: have better assertion
    assert len(os.listdir(target)) > 0
