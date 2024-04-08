from compile_smart_contracts import *
import pytest


@pytest.fixture
def setup_contracts_folder(tmp_path):
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


def test_find_solidity_files(setup_contracts_folder):
    tmp_path = setup_contracts_folder
    found_solidity_contracts = find_solidity_contracts(tmp_path)
    assert len(found_solidity_contracts) == 3

    assert found_solidity_contracts[0].filename == "Contract1.sol"
    assert found_solidity_contracts[0].path == tmp_path
    assert found_solidity_contracts[0].compiledJSONPath == tmp_path / "Contract1.json"
    assert found_solidity_contracts[1].filename == "Contract2"
    assert found_solidity_contracts[2].filename == "Contract3"
    assert found_solidity_contracts[3].filename == "Contract4"
