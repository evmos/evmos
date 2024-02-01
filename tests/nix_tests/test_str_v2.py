import pytest
import time
from web3 import Web3

from .ibc_utils import EVMOS_IBC_DENOM, assert_ready, get_balance, prepare_network
from .network import CosmosChain, Evmos
from .utils import (
    ADDRS,
    eth_to_bech32,
    wait_for_fn,
    erc20_balance
)

from .ibc_utils import (
    ATOM_IBC_DENOM
)

# The ERC20 address of ATOM on Evmos
ATOM_ERC20_ADDRESS = Web3.toChecksumAddress("0xf36e4C1F926001CEaDa9cA97ea622B25f41e5eB2")


@pytest.fixture(scope="module", params=["evmos"])
def ibc(request, tmp_path_factory):
    """ Prepare the network """
    name = "str-v2"
    evmos_build = request.param
    path = tmp_path_factory.mktemp(name)
    # specify the custom_scenario
    network = prepare_network(path, name, [evmos_build, "cosmoshub-1"])
    yield from network


def test_str_v2_single_hop(ibc):
    """
    Test Single Token Representation v2 with single hop Coin.
    It should automatically create an ERC20 precompiled contract.
    And register a token pair.
    """
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    gaia: CosmosChain = ibc.chains["cosmoshub-1"]

    evmos_cli = evmos.cosmos_cli()
    evmos_addr = ADDRS["signer2"]
    bech_dst = eth_to_bech32(evmos_addr)

    gaia_cli = gaia.cosmos_cli()
    gaia_addr = gaia_cli.address("signer2")

    old_dst_balance = get_balance(evmos, bech_dst, ATOM_IBC_DENOM)
    print("the bech32 destination", bech_dst)
    print(f"balance on evmos before: {old_dst_balance}")

    rsp = gaia_cli.ibc_transfer(gaia_addr, bech_dst, "5000uatom", "channel-0", 1, fees="10000uatom")
    assert rsp["code"] == 0

    time.sleep(30)

    print("token pairs", evmos_cli.get_token_pairs())
    print("balance on gaia after", gaia_cli.balances(gaia_addr))
    print(evmos_cli.evm_params()["params"]["active_precompiles"])

    # new_dst_balance_erc20_balance = erc20_balance(evmos.w3, ATOM_ERC20_ADDRESS, evmos_addr)
    # print("erc20 balance after", new_dst_balance_erc20_balance)

    new_dest_balance = 0
    def check_balance_after():
        nonlocal new_dest_balance
        new_dest_balance = get_balance(evmos, bech_dst, ATOM_IBC_DENOM)
        print("new balance on evmos after", new_dest_balance)
        assert old_dst_balance < new_dest_balance

    wait_for_fn("balance changed", check_balance_after)

    assert old_dst_balance + 5000 == new_dest_balance

# def test_str_v2_multi_hop(ibc):
#     """
#     Test Single Token Representation v2 with multi hop Coin.
#     It should NOT create an ERC20 precompiled contract and token pair.
#     """
#     assert_ready(ibc)
#
#     cli = ibc.chains["gaia"].cosmos_cli()
#
#
