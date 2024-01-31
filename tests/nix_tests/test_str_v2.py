import pytest
import time

from .ibc_utils import EVMOS_IBC_DENOM, assert_ready, get_balance, prepare_network
from .network import CosmosChain, Evmos
from .utils import (
    ADDRS,
    eth_to_bech32,
    wait_for_fn
)

from .ibc_utils import (
    EVMOS_IBC_DENOM,
    OSMO_IBC_DENOM,
)

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

    print("token pairs", evmos_cli.get_token_pairs())

    gaia_cli = gaia.cosmos_cli()
    gaia_addr = gaia_cli.address("signer2")

    print("old balance of gaia", gaia_cli.balances(gaia_addr))
    bech_dst = eth_to_bech32(evmos_addr)

    old_dst_balance = evmos_cli.balances(bech_dst)
    rsp = gaia_cli.ibc_transfer(gaia_addr, bech_dst, "200uatom", "channel-0", 1, fees="10000uatom")
    assert rsp["code"] == 0

    time.sleep(20)
    # def check_balance_change():
    #     new_dst_balance = evmos_cli.balances(bech_dst)
    #     return old_dst_balance != new_dst_balance
    #
    # wait_for_fn("balance change", check_balance_change, timeout=25)

    print("token pairs", evmos_cli.get_token_pairs())
    print(f"balance on evmos after: {evmos_cli.balances(bech_dst)}")
    print("balance on gaia after", gaia_cli.balances(gaia_addr))
    print(evmos_cli.evm_params())

    # assert old_dst_balance + 200 == new_dst_balance

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
