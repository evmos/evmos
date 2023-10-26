import pytest

from .ibc_utils import EVMOS_IBC_DENOM, assert_ready, get_balance, prepare_network
from .utils import ADDRS, get_precompile_contract, wait_for_fn


@pytest.fixture(scope="module")
def ibc(tmp_path_factory):
    """
    Prepares the network.
    """
    name = "osmosis-outpost"
    path = tmp_path_factory.mktemp(name)
    # Setup the IBC connections
    # evmos     (channel-0) <> (channel-0)  gaia
    # evmos     (channel-1) <> (channel-0)  osmosis
    # osmosis   (channel-1) <> (channel-1)  gaia
    network = prepare_network(path, name, ["gaia", "osmosis"])
    yield from network


# TODO remove this test and replace with the outpost test
def test_ibc_transfer(ibc):
    """
    test transfer IBC precompile.
    """
    assert_ready(ibc)

    dst_addr = ibc.chains["osmosis"].cosmos_cli().address("signer2")
    amt = 1000000

    cli = ibc.chains["evmos"].cosmos_cli()
    src_addr = cli.address("signer2")
    src_denom = "aevmos"

    old_src_balance = get_balance(ibc.chains["evmos"], src_addr, src_denom)
    old_dst_balance = get_balance(ibc.chains["osmosis"], dst_addr, EVMOS_IBC_DENOM)

    pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price

    tx_hash = pc.functions.transfer(
        "transfer",
        "channel-1",  # Connection with Osmosis is on channel-1
        src_denom,
        amt,
        ADDRS["signer2"],
        dst_addr,
        [1, 10000000000],
        0,
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
        new_dst_balance = get_balance(ibc.chains["osmosis"], dst_addr, EVMOS_IBC_DENOM)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + amt == new_dst_balance
    new_src_balance = get_balance(ibc.chains["evmos"], src_addr, src_denom)
    assert old_src_balance - amt - fee == new_src_balance
