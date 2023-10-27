import pytest

from .ibc_utils import EVMOS_IBC_DENOM, assert_ready, get_balance, prepare_network
from .utils import (
    ADDRS,
    erc20_balance,
    get_precompile_contract,
    send_transaction,
    wait_for_fn,
    wrap_evmos,
)


@pytest.fixture(scope="module")
def ibc(tmp_path_factory):
    "prepare-network"
    name = "stride-outpost"
    path = tmp_path_factory.mktemp(name)
    network = prepare_network(path, name, ["stride"], True)
    yield from network


# TODO remove this test and replace with the outpost test
def test_ibc_transfer(ibc):
    """
    test transfer IBC precompile.
    """
    assert_ready(ibc)
    evmos = ibc.chains["evmos"]

    cli = evmos.cosmos_cli()
    src_addr = cli.address("signer2")
    sender_addr = ADDRS["signer2"]
    src_denom = "aevmos"
    amt = 1000000000000000000

    wevmos_addr = wrap_evmos(evmos, sender_addr, amt)

    # stride chain is in ibc.orther_chain
    dst_addr = ibc.chains["stride"].cosmos_cli().address("signer2")

    old_src_balance = get_balance(evmos, src_addr, src_denom)
    old_dst_balance = get_balance(ibc.chains["stride"], dst_addr, EVMOS_IBC_DENOM)

    print(cli.balances(src_addr))

    pc = get_precompile_contract(evmos.w3, "IStrideOutpost")
    evmos_gas_price = evmos.w3.eth.gas_price

    # FIXME tx shows that is OK, but no WEVMOS are transferred
    # ?? maybe issue with Stride version
    tx = pc.functions.liquidStake(
        sender_addr,
        wevmos_addr,
        amt,
        dst_addr,
    ).build_transaction({"from": sender_addr, "gasPrice": evmos_gas_price})
    gas_estimation = evmos.w3.eth.estimate_gas(tx)

    receipt = send_transaction(evmos.w3, tx)

    assert receipt.status == 1
    # FIXME gasUsed should be same as estimation
    assert receipt.gasUsed == gas_estimation

    fee = receipt.gasUsed * evmos_gas_price

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        print(ibc.chains["stride"].cosmos_cli().balances(dst_addr))
        print(cli.balances(src_addr))
        wevmos_balance = erc20_balance(evmos.w3, wevmos_addr, sender_addr)
        print(f"WEVMOS balance: {wevmos_balance}")
        new_dst_balance = get_balance(ibc.chains["stride"], dst_addr, EVMOS_IBC_DENOM)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + amt == new_dst_balance
    new_src_balance = get_balance(ibc.chains["evmos"], src_addr, src_denom)
    assert old_src_balance - amt - fee == new_src_balance
