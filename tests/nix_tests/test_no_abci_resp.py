from pathlib import Path

import pytest

from .network import create_snapshots_dir, setup_custom_eidon-chain
from .utils import memiavl_config, wait_for_block


@pytest.fixture(scope="module")
def custom_eidon-chain(tmp_path_factory):
    path = tmp_path_factory.mktemp("no-abci-resp")
    yield from setup_custom_eidon-chain(
        path,
        26260,
        Path(__file__).parent / "configs/discard-abci-resp.jsonnet",
    )


@pytest.fixture(scope="module")
def custom_eidon-chain_rocksdb(tmp_path_factory):
    path = tmp_path_factory.mktemp("no-abci-resp-rocksdb")
    yield from setup_custom_eidon-chain(
        path,
        26810,
        memiavl_config(path, "discard-abci-resp"),
        post_init=create_snapshots_dir,
        chain_binary="eidond-rocksdb",
    )


@pytest.fixture(scope="module", params=["eidon-chain", "eidon-chain-rocksdb"])
def eidon-chain_cluster(request, custom_eidon-chain, custom_eidon-chain_rocksdb):
    """
    run on eidon-chain and
    eidon-chain built with rocksdb (memIAVL + versionDB)
    """
    provider = request.param
    if provider == "eidon-chain":
        yield custom_eidon-chain

    elif provider == "eidon-chain-rocksdb":
        yield custom_eidon-chain_rocksdb

    else:
        raise NotImplementedError


def test_gas_eth_tx(eidon-chain_cluster):
    """
    When node does not persist ABCI responses
    eth_gasPrice should return an error instead of crashing
    """
    wait_for_block(eidon-chain_cluster.cosmos_cli(), 3)
    try:
        eidon-chain_cluster.w3.eth.gas_price  # pylint: disable=pointless-statement
        raise Exception(  # pylint: disable=broad-exception-raised
            "This query should have failed"
        )
    except Exception as error:
        assert "block result not found" in error.args[0]["message"]
