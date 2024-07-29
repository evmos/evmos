import json

import pytest

from .ibc_utils import (
    BASECRO_IBC_DENOM,
    EVMOS_IBC_DENOM,
    OSMO_IBC_DENOM,
    assert_ready,
    get_balance,
    prepare_network,
    setup_denom_trace,
)
from .network import CosmosChain, Evmos
from .utils import (
    ACCOUNTS,
    ADDRS,
    CONTRACTS,
    KEYS,
    MAX_UINT256,
    approve_proposal,
    check_error,
    debug_trace_tx,
    decode_bech32,
    deploy_contract,
    eth_to_bech32,
    get_precompile_contract,
    send_transaction,
    w3_wait_for_new_blocks,
    wait_for_ack,
    wait_for_fn,
    wait_for_new_blocks,
)


@pytest.fixture(scope="module", params=["evmos"])
def ibc(request, tmp_path_factory):
    """
    Prepares the network.

    NOTE: The tests on this file cover only cases of native ERC20 contracts.
    For tests with IBC coins, checkout the test_str_v2.py
    and test_str_v2_token_factory.py files
    """
    name = "ibc-precompile"  # use the ibc-precompile.jsonnet config
    evmos_build = request.param
    path = tmp_path_factory.mktemp(name)
    network = prepare_network(path, name, [evmos_build, "chainmain"])
    yield from network


@pytest.mark.parametrize(
    "name, convert_amt, transfer_amt",
    [
        (
            "should convert erc20 ibc voucher to original erc20",
            10,
            10,
        ),
        (
            "should convert all available balance "
            + "of erc20 coin to original erc20 token",
            10,
            5,
        ),
        (
            "send native ERC-20 to chainmain, "
            + "when sending back IBC coins should "
            + "convert full balance back to erc20 token",
            0,
            5,
        ),
    ],
)
def test_ibc_callbacks(ibc, name, convert_amt, transfer_amt):
    """Test ibc precompile denom trace query"""
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    chainmain: CosmosChain = ibc.chains["chainmain"]

    w3 = evmos.w3
    evmos_cli = evmos.cosmos_cli()
    evmos_addr = ADDRS["signer2"]
    bech32_evmos_addr = evmos_cli.address("signer2")
    dst_addr = chainmain.cosmos_cli().address("signer2")

    # deploy erc20 contract
    contract, _ = deploy_contract(
        w3,
        CONTRACTS["TestERC20A"],
    )
    w3_wait_for_new_blocks(w3, 1)

    # Check token pairs before IBC transfer,
    # should only exist the WEVMOS pair
    pairs = evmos_cli.get_token_pairs()
    pairs_count_before = len(pairs)

    # register token pair
    rsp = evmos_cli.gov_legacy_proposal("signer2", "register-erc20")
    assert rsp["code"] == 0, rsp["raw_log"]
    wait_for_new_blocks(evmos_cli, 1)

    props = evmos_cli.query_proposals()
    props_count = len(props)
    assert props_count >= 1

    approve_proposal(evmos, props[props_count - 1].id)

    pairs = evmos_cli.get_token_pairs()
    assert len(pairs) == pairs_count_before + 1

    # check erc20 balance
    initial_amt = 100000000000000000000000000
    erc20_balance = contract.functions.balanceOf(evmos_addr).call()
    assert erc20_balance == initial_amt

    # convert to IBC voucher
    ibc_voucher_denom = f"erc20/${contract.address}"
    evmos_cli.convert_erc20(contract.address, convert_amt, "signer2")
    wait_for_new_blocks(evmos_cli, 1)

    # check erc20 balance & IBC voucher balance
    erc20_balance = contract.functions.balanceOf(evmos_addr).call()
    assert erc20_balance == initial_amt - convert_amt

    ibc_voucher_balance = get_balance(evmos, bech32_evmos_addr, ibc_voucher_denom)
    assert ibc_voucher_balance == convert_amt

    # send erc20 via IBC
    rsp = evmos_cli.ibc_transfer(
        bech32_evmos_addr,
        dst_addr,
        f"{transfer_amt}{ibc_voucher_denom}",
        "channel-0",
        1,
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    wait_for_new_blocks(evmos_cli, 1)

    # get the IBC denom on the destination chain
    erc20_ibc_denom = None

    # check balance on destination chain
    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        nonlocal erc20_ibc_denom
        erc20_ibc_denom = chainmain.cosmos_cli().denom_hash(
            f"transfer/channel-0/erc20/${contract.address}"
        )
        print(erc20_ibc_denom)
        if erc20_ibc_denom is None:
            return False
        new_dst_balance = get_balance(chainmain, dst_addr, erc20_ibc_denom)
        return new_dst_balance > 0

    wait_for_fn("balance change", check_balance_change)
    assert new_dst_balance == transfer_amt

    # send back erc20 IBC voucher to origin
    rsp = chainmain.cosmos_cli().ibc_transfer(
        dst_addr,
        bech32_evmos_addr,
        f"{transfer_amt}{erc20_ibc_denom}",
        "channel-0",
        1,
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    wait_for_ack(chainmain.cosmos_cli(), "chainmain")

    # check balance on source and destination chains
    chain_main_balance = get_balance(chainmain, dst_addr, erc20_ibc_denom)
    assert chain_main_balance == 0

    # check erc20 and IBC voucher balances
    # IBC coin balance should be zero
    # all balance should be in ERC20
    erc20_balance = contract.functions.balanceOf(evmos_addr).call()
    assert erc20_balance == initial_amt

    ibc_voucher_balance = get_balance(evmos, bech32_evmos_addr, ibc_voucher_denom)
    assert ibc_voucher_balance == 0
