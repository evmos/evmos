import pytest

from .ibc_utils import EVMOS_IBC_DENOM, assert_ready, get_balance, prepare_network
from .utils import ADDRS, get_precompile_contract, register_ibc_coin, wait_for_fn


@pytest.fixture(scope="module")
def ibc(tmp_path_factory):
    "prepare-network"
    name = "stride-outpost"
    path = tmp_path_factory.mktemp(name)
    network = prepare_network(path, name, ["stride"])
    yield from network


# TODO remove this test and replace with the outpost test
def test_ibc_transfer(ibc):
    """
    test transfer IBC precompile.
    """
    assert_ready(ibc)
    cli = ibc.chains["evmos"].cosmos_cli()
    proposal_id = register_ibc_coin(cli)

    # stride chain is in ibc.orther_chain
    dst_addr = ibc.chains["stride"].cosmos_cli().address("signer2")
    amt = 1000000

    src_addr = cli.address("signer2")
    src_denom = "aevmos"
    src_token_addr = cli.query_token(src_denom)["token"]["address"]

    old_src_balance = get_balance(ibc.chains["evmos"], src_addr, src_denom)
    old_dst_balance = get_balance(ibc.chains["stride"], dst_addr, EVMOS_IBC_DENOM)

    pc = get_precompile_contract(ibc.chains["evmos"].w3, "IStrideOutpost")
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price

    tx_hash = pc.functions.liquidStake(
        src_addr,
        src_token_addr,
        amt,
        "",
    ).transact({"from": ADDRS["signer2"], "gasPrice": evmos_gas_price})

    receipt = ibc.chains["evmos"].w3.eth.wait_for_transaction_receipt(tx_hash)

    assert receipt.status == 1
    # check gas used
    assert receipt.gasUsed == 133680

    fee = receipt.gasUsed * evmos_gas_price

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(ibc.chains["stride"], dst_addr, EVMOS_IBC_DENOM)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + amt == new_dst_balance
    new_src_balance = get_balance(ibc.chains["evmos"], src_addr, src_denom)
    assert old_src_balance - amt - fee == new_src_balance
