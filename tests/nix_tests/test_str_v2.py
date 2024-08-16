import json
import tempfile

import pytest
from web3 import Web3

from .ibc_utils import ATOM_IBC_DENOM, assert_ready, get_balance, prepare_network
from .network import CosmosChain, Evmos
from .utils import (
    ADDRS,
    KEYS,
    WEVMOS_ADDRESS,
    approve_proposal,
    erc20_balance,
    erc20_transfer,
    eth_to_bech32,
    wait_for_ack,
    wait_for_fn,
    wait_for_new_blocks,
)

# uatom from cosmoshub-2 -> cosmoshub-1 IBC representation on the Evmos chain.
ATOM_2_IBC_DENOM_MULTI_HOP = (
    "ibc/D219F3A490310B65BDC312B5A644B0D56FFF1789D894B902A49FBF9D2F560B32"
)
# uatom from cosmoshub-2 -> cosmoshub-1 IBC representation
ATOM_1_IBC_DENOM_ATOM_2 = (
    "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9"
)
# The ERC20 address of ATOM on Evmos
ATOM_1_ERC20_ADDRESS = Web3.toChecksumAddress(
    "0xf36e4C1F926001CEaDa9cA97ea622B25f41e5eB2"
)

UPDATE_PARAMS_PROP = {
    "messages": [
        {
            "@type": "/evmos.erc20.v1.MsgUpdateParams",
            "authority": "evmos10d07y265gmmuvt4z0w9aw880jnsr700jcrztvm",
            "params": {
                "enable_erc20": True,
                "native_precompiles": [],
                "dynamic_precompiles": [],
            },
        }
    ],
    "metadata": "ipfs://CID",
    "deposit": "1aevmos",
    "title": "update erc20 mod params",
    "summary": "update erc20 mod params",
}


@pytest.fixture(scope="module", params=["evmos", "evmos-rocksdb"])
def ibc(request, tmp_path_factory):
    """Prepare the network"""
    name = "str-v2"
    evmos_build = request.param
    path = tmp_path_factory.mktemp(name)
    # specify the custom_scenario
    network = prepare_network(path, name, [evmos_build, "cosmoshub-1", "cosmoshub-2"])
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
    rsp = gaia_cli.ibc_transfer(
        gaia_addr, bech_dst, "5000uatom", "channel-0", 1, fees="10000uatom"
    )
    assert rsp["code"] == 0

    wait_for_ack(evmos_cli, "Evmos")

    w3 = evmos.w3
    active_dynamic_precompiles = evmos_cli.erc20_params()["params"][
        "dynamic_precompiles"
    ]
    new_dest_balance = get_balance(evmos, bech_dst, ATOM_IBC_DENOM)
    erc_dest_balance = erc20_balance(w3, ATOM_1_ERC20_ADDRESS, evmos_addr)

    assert len(active_dynamic_precompiles) == 1
    assert old_dst_balance + 5000 == new_dest_balance
    assert old_dst_balance + 5000 == erc_dest_balance


def test_str_v2_multi_hop(ibc):
    """
    Test Single Token Representation v2 with multi hop Coin.
    It should NOT create an ERC20 precompiled contract and token pair.
    """
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    gaia: CosmosChain = ibc.chains["cosmoshub-1"]
    gaia2: CosmosChain = ibc.chains["cosmoshub-2"]

    evmos_cli = evmos.cosmos_cli()
    evmos_addr = ADDRS["signer2"]
    bech_dst = eth_to_bech32(evmos_addr)

    # The starting balance of the destination address
    evmos_old_balance = get_balance(evmos, bech_dst, ATOM_2_IBC_DENOM_MULTI_HOP)

    # Cosmos hub 1
    gaia_cli = gaia.cosmos_cli()
    gaia_addr = gaia_cli.address("signer2")
    gaia1_old_balance = get_balance(gaia, gaia_addr, ATOM_1_IBC_DENOM_ATOM_2)

    # Cosmos hub 2
    gaia2_cli = gaia2.cosmos_cli()

    rsp = gaia2_cli.ibc_transfer(
        gaia_addr, gaia_addr, "50000uatom", "channel-1", 1, fees="10000uatom"
    )
    assert rsp["code"] == 0

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(gaia, gaia_addr, ATOM_1_IBC_DENOM_ATOM_2)
        return gaia1_old_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)

    new_gaia1_balance = get_balance(gaia, gaia_addr, ATOM_1_IBC_DENOM_ATOM_2)
    assert gaia1_old_balance + 50000 == new_gaia1_balance

    rsp = gaia_cli.ibc_transfer(
        gaia_addr,
        bech_dst,
        f"50000{ATOM_1_IBC_DENOM_ATOM_2}",
        "channel-0",
        1,
        fees="10000uatom",
    )
    assert rsp["code"] == 0

    wait_for_ack(evmos_cli, "Evmos")

    evmos_balance = get_balance(evmos, bech_dst, ATOM_2_IBC_DENOM_MULTI_HOP)
    dynamic_precompiles = evmos_cli.erc20_params()["params"]["dynamic_precompiles"]
    token_pairs = evmos_cli.get_token_pairs()

    # Here it's only one from the previous one we've registered in the first test
    assert evmos_old_balance + 50000 == evmos_balance
    assert len(dynamic_precompiles) == 1
    assert dynamic_precompiles[0] == ATOM_1_ERC20_ADDRESS
    assert len(token_pairs) == 2


def test_wevmos_precompile_transfer(ibc):
    """
    Test the ERC20 transfer from one signer to another using the now
    registered ERC20 precompiled contract for WEVMOS.
    """
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    signer1 = ADDRS["signer1"]
    signer2 = ADDRS["signer2"]
    bech_dst = eth_to_bech32(signer2)
    src_denom = "aevmos"

    w3 = evmos.w3
    evmos_balance = get_balance(evmos, bech_dst, src_denom)
    signer2_balance = erc20_balance(w3, WEVMOS_ADDRESS, signer2)

    assert evmos_balance == signer2_balance

    receipt = erc20_transfer(
        w3, WEVMOS_ADDRESS, signer1, signer2, 1000000, KEYS["signer1"]
    )
    assert receipt.status == 1

    signer_2_balance_after = erc20_balance(w3, WEVMOS_ADDRESS, signer2)
    assert signer_2_balance_after == signer2_balance + 1000000

    evmos_balance_after = get_balance(evmos, bech_dst, src_denom)
    assert evmos_balance_after == evmos_balance + 1000000


def test_toggle_erc20_precompile(ibc):
    """
    Test Enabling/Disabling an ERC20 precompile
    for an IBC coin (single hop).
    It should automatically update the code hash and code
    to empty when disabled, and should add the code hash when enabling.
    NOTE: This test relies on the IBC coins transferred in a previous test
    """
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]

    # this is the code hash of the ERC20 contract deployed previous to the STRv2 upgrade
    erc20_code_hash = (
        "0x7b477c761b4d0469f03f27ba58d0a7eacbfdd62b69b82c6c683ae5f81c67fe80"
    )
    # this is the code hash resulting from the keccak256(nil)
    empty_code_hash = (
        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    )
    evmos_cli = evmos.cosmos_cli()
    w3 = evmos.w3

    # assert that there's code and code hash
    # on the erc20 contract address
    contract_bech32_addr = eth_to_bech32(ATOM_1_ERC20_ADDRESS)
    acc = evmos_cli.account(contract_bech32_addr)
    assert acc["code_hash"] == erc20_code_hash

    code = w3.eth.get_code(ATOM_1_ERC20_ADDRESS)
    assert len(code) > 0

    # get the initial params to use them later
    initial_params = evmos_cli.erc20_params()

    # update params via gov proposal to disable all the erc20 precompile
    update_erc20_params(evmos)

    # check that code and code hash were updated
    acc = evmos_cli.account(contract_bech32_addr)
    assert acc["code_hash"] == empty_code_hash

    code = w3.eth.get_code(ATOM_1_ERC20_ADDRESS)
    assert len(code) == 0

    # enable back the erc20 precompiles
    update_erc20_params(
        evmos,
        initial_params["params"]["native_precompiles"],
        initial_params["params"]["dynamic_precompiles"],
    )

    # check that code and code hash were restored
    acc = evmos_cli.account(contract_bech32_addr)
    assert acc["code_hash"] == erc20_code_hash

    code = w3.eth.get_code(ATOM_1_ERC20_ADDRESS)
    assert len(code) > 0


def update_erc20_params(evmos: Evmos, native_precomiles=[], dynamic_precompiles=[]):
    cli = evmos.cosmos_cli()
    with tempfile.NamedTemporaryFile("w") as fp:
        UPDATE_PARAMS_PROP["messages"][0]["params"][
            "native_precompiles"
        ] = native_precomiles
        UPDATE_PARAMS_PROP["messages"][0]["params"][
            "dynamic_precompiles"
        ] = dynamic_precompiles
        json.dump(UPDATE_PARAMS_PROP, fp)
        fp.flush()
        rsp = cli.gov_proposal("signer2", fp.name)
        assert rsp["code"] == 0, rsp["raw_log"]
        txhash = rsp["txhash"]

        wait_for_new_blocks(cli, 2)
        receipt = cli.tx_search_rpc(f"tx.hash='{txhash}'")[0]
        assert receipt["tx_result"]["code"] == 0, receipt["tx_result"]["log"]

    res = cli.query_proposals()
    props = res["proposals"]
    props_count = len(props)
    assert props_count >= 1

    approve_proposal(evmos, props[props_count - 1]["id"])
    wait_for_new_blocks(cli, 2)
