import pytest
from web3 import Web3

from .ibc_utils import  assert_ready, get_balance, prepare_network
from .network import CosmosChain, Evmos
from .utils import (
    ADDRS,
    KEYS,
    WEVMOS_ADDRESS,
    erc20_balance,
    eth_to_bech32,
    wait_for_ack,
    wait_for_fn,
)

@pytest.fixture(scope="module", params=["evmos"])
def ibc(request, tmp_path_factory):
    """Prepare the network"""
    name = "callbacks"
    evmos_build = request.param
    path = tmp_path_factory.mktemp(name)
    # specify the custom_scenario
    network = prepare_network(path, name,  ["cosmoshub-1", evmos_build])
    yield from network

def test_ibc_transfer_callback(ibc):
    """ """
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    gaia: CosmosChain = ibc.chains["cosmoshub-1"]

    evmos_cli = evmos.cosmos_cli()
    evmos_addr = ADDRS["signer2"]
    bech_dst = eth_to_bech32(evmos_addr)

    gaia_cli = gaia.cosmos_cli()
    gaia_addr = gaia_cli.address("signer2")

    rsp = evmos_cli.ibc_transfer(
        evmos_addr, gaia_addr, "5000aevmos", "channel-0", 1, fees="10000aevmos"
    )
    assert rsp["code"] == 0

    wait_for_ack(evmos_cli, "Evmos")
