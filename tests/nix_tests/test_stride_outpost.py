import pytest

from .ibc_utils import EVMOS_IBC_DENOM, assert_ready, get_balance, prepare_network
from .utils import (
    ADDRS,
    KEYS,
    erc20_balance,
    get_precompile_contract,
    register_host_zone,
    send_transaction,
    wait_for_fn,
    wrap_evmos,
)


@pytest.fixture(scope="module", params=["evmos"])
def ibc(request, tmp_path_factory):
    "prepare-network"
    name = "stride-outpost"
    evmos_build = request.param
    path = tmp_path_factory.mktemp(name)
    # specify the custom_scenario
    # to patch evmos to use channel-0 for Stride outpost
    # and allow to register WEVMOS token
    network = prepare_network(path, name, [evmos_build, "stride"])
    yield from network


def test_liquid_stake(ibc):
    """
    test liquidStaking precompile function.
    """
    assert_ready(ibc)

    cli = ibc.chains["evmos"].cosmos_cli()
    src_addr = cli.address("signer2")
    sender_addr = ADDRS["signer2"]
    src_denom = "aevmos"
    st_token = "staevmos"
    amt = 1000000000000000000

    # ATM is not required to have WEVMOS balance to be able to
    # liquid stake. It will deduct 'aevmos' coins instead of WEVMOS.
    # We still need to register the token pair tho.
    # This is done in this function
    wevmos_addr = wrap_evmos(ibc.chains["evmos"], sender_addr, amt)

    dst_addr = ibc.chains["stride"].cosmos_cli().address("signer2")

    # need to register evmos chain as host zone in stride
    register_host_zone(
        ibc.chains["stride"],
        dst_addr,
        "connection-0",
        src_denom,
        "evmos",
        EVMOS_IBC_DENOM,
        "channel-0",
        1000000,
    )

    old_src_balance = get_balance(ibc.chains["evmos"], src_addr, src_denom)
    old_dst_balance = get_balance(ibc.chains["stride"], dst_addr, st_token)

    pc = get_precompile_contract(ibc.chains["evmos"].w3, "IStrideOutpost")
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price

    liquid_stake_params = {
        "channelID": "channel-0",
        "sender": sender_addr,
        "receiver": sender_addr,
        "strideForwarder": dst_addr,
        "token": wevmos_addr,
        "amount": amt,
    }
    tx = pc.functions.liquidStake(liquid_stake_params).build_transaction(
        {"from": sender_addr, "gasPrice": evmos_gas_price}
    )
    gas_estimation = ibc.chains["evmos"].w3.eth.estimate_gas(tx)

    receipt = send_transaction(ibc.chains["evmos"].w3, tx, KEYS["signer2"])
    assert receipt.status == 1

    # FIXME gasUsed should be same as estimation
    # ATM gas estimation is always higher than gas used
    # in precompiles.
    # Possible fix here https://github.com/evmos/evmos/pull/1943
    # assert receipt.gasUsed == gas_estimation
    print(f"gas estimation {gas_estimation}")
    print(f"gas used: {receipt.gasUsed}")

    fee = receipt.gasUsed * evmos_gas_price
    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(ibc.chains["stride"], dst_addr, st_token)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + amt == new_dst_balance
    new_src_balance = get_balance(ibc.chains["evmos"], src_addr, src_denom)
    wevmos_balance = erc20_balance(ibc.chains["evmos"].w3, wevmos_addr, sender_addr)
    # FIXME For better UX, the 'amt' should be deducted
    # from WEVMOS balance instead of the native coin 'aevmos'
    assert old_src_balance - amt - fee == new_src_balance
    assert wevmos_balance == amt
