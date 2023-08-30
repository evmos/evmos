import pytest

from .ibc_utils import (
    RATIO,
    assert_ready,
    get_balance,
    hermes_transfer,
    prepare_network,
)
from .utils import (
    ADDRS,
    eth_to_bech32,
    parse_events,
    parse_events_rpc,
    wait_for_fn,
)


@pytest.fixture(scope="module", params=[True, False])
def ibc(request, tmp_path_factory):
    "prepare-network"
    incentivized = request.param
    name = "ibc"
    path = tmp_path_factory.mktemp(name)
    network = prepare_network(path, name, incentivized)
    yield from network


def get_balances(chain, addr):
    return chain.cosmos_cli().balances(addr)


def test_ibc_transfer_with_hermes(ibc):
    """
    test ibc transfer tokens with hermes cli
    """
    src_amount = hermes_transfer(ibc)
    dst_amount = src_amount * RATIO  # the decimal places difference
    dst_denom = "aevmos"
    dst_addr = eth_to_bech32(ADDRS["signer2"])
    old_dst_balance = get_balance(ibc.evmos, dst_addr, dst_denom)

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(ibc.evmos, dst_addr, dst_denom)
        return new_dst_balance != old_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + dst_amount == new_dst_balance

    # assert that the relayer transactions do enables the dynamic fee extension option.
    cli = ibc.evmos.cosmos_cli()
    criteria = "message.action=/ibc.core.channel.v1.MsgChannelOpenInit"
    tx = cli.tx_search(criteria)["txs"][0]
    events = parse_events_rpc(tx["events"])
    fee = int(events["tx"]["fee"].removesuffix("basetcro"))
    gas = int(tx["gas_wanted"])
    # the effective fee is decided by the max_priority_fee (base fee is zero)
    # rather than the normal gas price
    assert fee == gas * 1000000


def test_ibc_incentivized_transfer(ibc):
    if not ibc.incentivized:
        # this test case only works for incentivized channel.
        return
    src_chain = ibc.evmos.cosmos_cli()
    dst_chain = ibc.chainmain.cosmos_cli()
    receiver = dst_chain.address("signer2")
    sender = src_chain.address("signer2")
    relayer = src_chain.address("signer1")
    original_amount = src_chain.balance(relayer, denom="ibcfee")
    original_amount_sender = src_chain.balance(sender, denom="ibcfee")

    rsp = src_chain.ibc_transfer(
        sender,
        receiver,
        "1000aevmos",
        "channel-0",
        1,
        "100000000basecro",
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    evt = parse_events(rsp["logs"])["send_packet"]
    print("packet event", evt)
    packet_seq = int(evt["packet_sequence"])

    rsp = src_chain.pay_packet_fee(
        "transfer",
        "channel-0",
        packet_seq,
        recv_fee="10ibcfee",
        ack_fee="10ibcfee",
        timeout_fee="10ibcfee",
        from_=sender,
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    # fee is locked
    assert src_chain.balance(sender, denom="ibcfee") == original_amount_sender - 30

    # wait for relayer receive the fee
    def check_fee():
        amount = src_chain.balance(relayer, denom="ibcfee")
        if amount > original_amount:
            assert amount == original_amount + 20
            return True
        else:
            return False

    wait_for_fn("wait for relayer to receive the fee", check_fee)

    # timeout fee is refunded
    assert src_chain.balance(sender, denom="ibcfee") == original_amount_sender - 20


def test_evmos_transfer_tokens(ibc):
    """
    test sending aevmos from evmos to crypto-org-chain using cli transfer_tokens.
    depends on `test_ibc` to send the original coins.
    """
    assert_ready(ibc)
    dst_addr = ibc.chainmain.cosmos_cli().address("signer2")
    dst_amount = 2
    dst_denom = "basecro"
    cli = ibc.evmos.cosmos_cli()
    src_amount = dst_amount * RATIO  # the decimal places difference
    src_addr = cli.address("signer2")
    src_denom = "aevmos"

    # case 1: use evmos cli
    old_src_balance = get_balance(ibc.evmos, src_addr, src_denom)
    old_dst_balance = get_balance(ibc.chainmain, dst_addr, dst_denom)
    rsp = cli.transfer_tokens(
        src_addr,
        dst_addr,
        f"{src_amount}{src_denom}",
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(ibc.chainmain, dst_addr, dst_denom)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + dst_amount == new_dst_balance
    new_src_balance = get_balance(ibc.evmos, src_addr, src_denom)
    assert old_src_balance - src_amount == new_src_balance


def test_evmos_transfer_tokens_acknowledgement_error(ibc):
    """
    test sending aevmos from evmos to crypto-org-chain using cli transfer_tokens
    with invalid receiver for acknowledgement error.
    depends on `test_ibc` to send the original coins.
    """
    assert_ready(ibc)
    dst_addr = "invalid_address"
    dst_amount = 2
    cli = ibc.evmos.cosmos_cli()
    src_amount = dst_amount * RATIO  # the decimal places difference
    src_addr = cli.address("signer2")
    src_denom = "aevmos"

    old_src_balance = get_balance(ibc.evmos, src_addr, src_denom)
    rsp = cli.transfer_tokens(
        src_addr,
        dst_addr,
        f"{src_amount}{src_denom}",
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    new_src_balance = 0

    def check_balance_change():
        nonlocal new_src_balance
        new_src_balance = get_balance(ibc.evmos, src_addr, src_denom)
        return old_src_balance == new_src_balance

    wait_for_fn("balance no change", check_balance_change)
    new_src_balance = get_balance(ibc.evmos, src_addr, src_denom)
