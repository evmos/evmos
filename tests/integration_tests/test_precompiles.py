import pytest
import os
import subprocess

from .ibc_utils import assert_ready, prepare_network


@pytest.fixture(scope="module", params=[False])
def ibc(request, tmp_path_factory):
    "prepare-network"
    incentivized = request.param
    name = "ibc-precompile"
    path = tmp_path_factory.mktemp(name)
    network = prepare_network(path, name, incentivized)
    yield from network


def test_precompiles(ibc):
    """
    test precompiles transactions.
    """
    assert_ready(ibc)
    abspath = os.path.abspath(__file__)
    dname = os.path.dirname(abspath)
    os.chdir(f"{dname}/hardhat")
    proc = subprocess.Popen(
        ["npm", "run", "test-evmos"],
        preexec_fn=os.setsid,
    )
    # check process exit code is OK
    code = proc.wait()
    assert code == 0
