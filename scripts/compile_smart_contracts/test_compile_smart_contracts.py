from compile_smart_contracts import *
import pytest


@pytest.fixture
def setup_contracts_folder(tmp_path):
    """
    This fixture creates a temporary folder with some Solidity files.
    """

    (tmp_path / "Contract1.sol").touch()
    (tmp_path / "Contract2.sol").touch()

    (tmp_path / "contracts").mkdir()
    (tmp_path / "contracts" / "Contract3.sol").touch()

    (tmp_path / "precompiles").mkdir()
    (tmp_path / "precompiles" / "Contract4.sol").touch()

    return tmp_path


def test_find_solidity_files(setup_contracts_folder):
    tmp_path = setup_contracts_folder
    found_solidity_files = find_solidity_files(tmp_path)
    assert len(found_solidity_files) == 4
