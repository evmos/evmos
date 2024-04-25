import pytest
import json
import time
from web3 import Web3

from .ibc_utils import  assert_ready, get_balance, prepare_network
from .network import CosmosChain, Evmos
from .utils import (
    ADDRS,
    CONTRACTS,
    deploy_contract,
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
    network = prepare_network(path, name,  [evmos_build, "cosmoshub-1"])
    yield from network

def test_ibc_transfer_callback(ibc):
    """ """
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    gaia: CosmosChain = ibc.chains["cosmoshub-1"]

    w3 = evmos.w3
    eth_contract, tx_receipt = deploy_contract(w3, CONTRACTS["PacketActorCounter"])
    print("the address", eth_contract.address)
    print("the counter", eth_contract.functions.counter().call())

    memo = {"src_callback": { "address": f'{eth_contract.address}',}}

    evmos_cli = evmos.cosmos_cli()
    evmos_addr = ADDRS["signer1"]
    bech_src = eth_to_bech32(evmos_addr)

    gaia_cli = gaia.cosmos_cli()
    gaia_addr = gaia_cli.address("signer2")

    rsp = evmos_cli.ibc_transfer(
        bech_src, gaia_addr, "5000aevmos", "channel-0", 1, fees="10000000000000000aevmos", memo=json.dumps(memo)
    )
    assert rsp["code"] == 0
    wait_for_ack(evmos_cli, "Evmos")

    time.sleep(15)

    print("the counter after", eth_contract.functions.counter().call())
