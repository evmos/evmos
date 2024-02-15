import pytest
import time
from web3 import Web3

from .ibc_utils import (
    EVMOS_IBC_DENOM,
    assert_ready,
    get_balance,
    prepare_network,
    get_balances,
)
from .utils import (
    wait_for_cosmos_tx_receipt,
)
from .network import CosmosChain, Evmos
from .utils import ADDRS, eth_to_bech32, wait_for_fn, erc20_balance


@pytest.fixture(scope="module", params=["evmos"])
def ibc(request, tmp_path_factory):
    """Prepare the network"""
    name = "str-v2-token_factory"
    evmos_build = request.param
    path = tmp_path_factory.mktemp(name)
    # specify the custom_scenario
    network = prepare_network(path, name, [evmos_build, "osmosis-1"])
    yield from network


def test_str_v2_token_factory(ibc):
    """
    Test Single Token Representation v2 with single hop Coin.
    It should automatically create an ERC20 precompiled contract.
    And register a token pair.
    """
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    osmosis: CosmosChain = ibc.chains["cosmoshub-1"]

    evmos_cli = evmos.cosmos_cli()
    evmos_addr = ADDRS["signer2"]
    bech_dst = eth_to_bech32(evmos_addr)

    osmosis_cli = osmosis.cosmos_cli()
    osmosis_addr = osmosis_cli.address("signer2")

    # create a token factory coin
    token_factory_denom = create_token_factory_coin("utest", osmosis_addr, osmosis_cli)

    # NOTE: Sleep some time because wait_for_fn doesn't work for some reason ?
    time.sleep(30)

    balance = get_balances(osmosis_cli, osmosis_addr)
    print("balances in Osmosis", balance)


def create_token_factory_coin(denom, creator_addr, osmosis_cli):
    full_denom = f"factory/{creator_addr}/{denom}"
    rsp = osmosis_cli.token_factory_create_denom(denom, creator_addr)
    assert rsp["code"] == 0

    # check for tx receipt to confirm tx was successful
    receipt = wait_for_cosmos_tx_receipt(osmosis_cli, rsp["txhash"])
    assert receipt["tx_result"]["code"] == 0
    print("Created token factory token", full_denom)

    rsp = osmosis_cli.token_factory_mint_denom(1000, full_denom, creator_addr)
    assert rsp["code"] == 0

    # check for tx receipt to confirm tx was successful
    receipt = wait_for_cosmos_tx_receipt(osmosis_cli, rsp["txhash"])
    assert receipt["tx_result"]["code"] == 0
    print("Minted token factory token", full_denom)

    return full_denom
