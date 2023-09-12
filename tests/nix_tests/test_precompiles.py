import re

import pytest

from .ibc_utils import EVMOS_IBC_DENOM, assert_ready, get_balance, prepare_network
from .utils import ADDRS, get_precompile_contract, wait_for_fn


@pytest.fixture(scope="module", params=[False])
def ibc(request, tmp_path_factory):
    "prepare-network"
    incentivized = request.param
    name = "ibc-precompile"
    path = tmp_path_factory.mktemp(name)
    network = prepare_network(path, name, "chainmain", incentivized)
    yield from network


def test_ibc_transfer(ibc):
    """
    test transfer IBC precompile.
    """
    assert_ready(ibc)

    dst_addr = ibc.other_chain.cosmos_cli().address("signer2")
    amt = 1000000

    cli = ibc.evmos.cosmos_cli()
    src_addr = cli.address("signer2")
    src_denom = "aevmos"

    old_src_balance = get_balance(ibc.evmos, src_addr, src_denom)
    old_dst_balance = get_balance(ibc.other_chain, dst_addr, EVMOS_IBC_DENOM)

    pc = get_precompile_contract(ibc.evmos.w3, "ICS20I")
    evmos_gas_price = ibc.evmos.w3.eth.gas_price

    tx_hash = pc.functions.transfer(
        "transfer",
        "channel-0",
        src_denom,
        amt,
        ADDRS["signer2"],
        dst_addr,
        [1, 10000000000],
        0,
        "",
    ).transact({"from": ADDRS["signer2"], "gasPrice": evmos_gas_price})

    receipt = ibc.evmos.w3.eth.wait_for_transaction_receipt(tx_hash)

    assert receipt.status == 1
    # check gas used
    assert receipt.gasUsed == 127581

    fee = receipt.gasUsed * evmos_gas_price

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(ibc.other_chain, dst_addr, EVMOS_IBC_DENOM)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + amt == new_dst_balance
    new_src_balance = get_balance(ibc.evmos, src_addr, src_denom)
    assert old_src_balance - amt - fee == new_src_balance


def test_ibc_transfer_invalid_packet(ibc):
    """
    test transfer IBC precompile invalid packet error.
    NOTE: it is important for this error message to not change
    because it is already stored on mainnet.
    Changing this error message is a state breaking change
    """
    assert_ready(ibc)

    # IMPORTANT: THIS ERROR MSG SHOULD NEVER CHANGE OR WILL BE A STATE BREAKING CHANGE ON MAINNET
    exp_err = "constructed packet failed basic validation: packet timeout height and packet timeout timestamp cannot both be 0: invalid packet"  # noqa: E501
    w3 = ibc.evmos.w3

    dst_addr = ibc.other_chain.cosmos_cli().address("signer2")
    amt = 1000000

    cli = ibc.evmos.cosmos_cli()
    src_addr = cli.address("signer2")
    src_denom = "aevmos"

    old_src_balance = get_balance(ibc.evmos, src_addr, src_denom)

    pc = get_precompile_contract(w3, "ICS20I")
    evmos_gas_price = w3.eth.gas_price

    try:
        pc.functions.transfer(
            "transfer",
            "channel-0",
            src_denom,
            amt,
            ADDRS["signer2"],
            dst_addr,
            [0, 0],
            0,
            "",
        ).transact({"from": ADDRS["signer2"], "gasPrice": evmos_gas_price})
    except Exception as error:
        assert error.args[0]["message"] == f"rpc error: code = Unknown desc = {exp_err}"

        new_src_balance = get_balance(ibc.evmos, src_addr, src_denom)
        assert old_src_balance == new_src_balance


def test_ibc_transfer_timeout(ibc):
    """
    test transfer IBC precompile timeout packet error.
    NOTE: it is important for this error message to not change
    because it is already stored on mainnet.
    Changing this error message is a state breaking change
    """
    assert_ready(ibc)

    # IMPORTANT: THIS ERROR MSG SHOULD NEVER CHANGE OR WILL BE A STATE BREAKING CHANGE ON MAINNET
    exp_err = r"rpc error\: code = Unknown desc = receiving chain block timestamp \>\= packet timeout timestamp \(\d{4}\-\d{2}\-\d{2} \d{2}\:\d{2}\:\d{2}\.\d{6,9} \+0000 UTC \>\= \d{4}\-\d{2}\-\d{2} \d{2}\:\d{2}\:\d{2}\.\d{6,9} \+0000 UTC\)\: packet timeout"  # noqa: E501
    w3 = ibc.evmos.w3

    dst_addr = ibc.other_chain.cosmos_cli().address("signer2")
    amt = 1000000

    cli = ibc.evmos.cosmos_cli()
    src_addr = cli.address("signer2")
    src_denom = "aevmos"

    old_src_balance = get_balance(ibc.evmos, src_addr, src_denom)

    pc = get_precompile_contract(w3, "ICS20I")
    evmos_gas_price = w3.eth.gas_price

    try:
        pc.functions.transfer(
            "transfer",
            "channel-0",
            src_denom,
            amt,
            ADDRS["signer2"],
            dst_addr,
            [0, 0],
            1000,
            "",
        ).transact({"from": ADDRS["signer2"], "gasPrice": evmos_gas_price})
    except Exception as error:
        assert re.search(exp_err, error.args[0]["message"]) is not None

        new_src_balance = get_balance(ibc.evmos, src_addr, src_denom)
        assert old_src_balance == new_src_balance
