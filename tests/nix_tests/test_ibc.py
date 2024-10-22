import pytest

from .ibc_utils import (
    EVMOS_IBC_DENOM,
    assert_ready,
    get_balance,
    hermes_transfer,
    prepare_network,
)
from .utils import parse_events_rpc, wait_for_fn


@pytest.fixture(scope="module", params=["eidon-chain", "eidon-chain-rocksdb"])
def ibc(request, tmp_path_factory):
    """
    prepare IBC network with an eidon-chain chain
    (default build or with memIAVL + versionDB)
    and a chainmain (crypto.org) chain
    """
    name = "ibc"
    eidon-chain_build = request.param
    path = tmp_path_factory.mktemp(name)
    network = prepare_network(path, name, [eidon-chain_build, "chainmain"])
    yield from network


def get_balances(chain, addr):
    return chain.cosmos_cli().balances(addr)


def test_ibc_transfer_with_hermes(ibc):
    """
    test ibc transfer tokens with hermes cli
    """
    amt = hermes_transfer(ibc)
    # ibc denom of the basecro sent
    dst_denom = "ibc/6411AE2ADA1E73DB59DB151A8988F9B7D5E7E233D8414DB6817F8F1A01611F86"
    dst_addr = ibc.chains["eidon-chain"].cosmos_cli().address("signer2")
    old_dst_balance = get_balance(ibc.chains["eidon-chain"], dst_addr, dst_denom)
    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(ibc.chains["eidon-chain"], dst_addr, dst_denom)
        return new_dst_balance != old_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + amt == new_dst_balance

    # assert that the relayer transactions do enables the
    # dynamic fee extension option.
    cli = ibc.chains["eidon-chain"].cosmos_cli()
    criteria = "message.action='/ibc.core.channel.v1.MsgChannelOpenInit'"
    tx = cli.tx_search(criteria)["txs"][0]
    events = parse_events_rpc(tx["events"])
    fee = int(events["tx"]["fee"].removesuffix("aeidon-chain"))
    gas = int(tx["gas_wanted"])
    # the effective fee is decided by the max_priority_fee (base fee is zero)
    # rather than the normal gas price
    assert fee == gas * 1000000


def test_eidon-chain_ibc_transfer(ibc):
    """
    test sending aeidon-chain from eidon-chain to crypto-org-chain using cli.
    """
    assert_ready(ibc)
    dst_addr = ibc.chains["chainmain"].cosmos_cli().address("signer2")
    amt = 1000000

    cli = ibc.chains["eidon-chain"].cosmos_cli()
    src_addr = cli.address("signer2")
    src_denom = "aeidon-chain"

    # case 1: use eidon-chain cli
    old_src_balance = get_balance(ibc.chains["eidon-chain"], src_addr, src_denom)
    old_dst_balance = get_balance(ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM)

    rsp = cli.ibc_transfer(
        src_addr,
        dst_addr,
        f"{amt}{src_denom}",
        "channel-0",
        1,
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(
            ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM
        )
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + amt == new_dst_balance
    new_src_balance = get_balance(ibc.chains["eidon-chain"], src_addr, src_denom)
    assert old_src_balance - amt == new_src_balance


def test_eidon-chain_ibc_transfer_acknowledgement_error(ibc):
    """
    test sending aeidon-chain from eidon-chain to crypto-org-chain using cli
    transfer_tokens with invalid receiver for acknowledgement error.
    """
    assert_ready(ibc)
    dst_addr = "invalid_address"
    amt = 1000000

    cli = ibc.chains["eidon-chain"].cosmos_cli()
    src_addr = cli.address("signer2")
    src_denom = "aeidon-chain"

    old_src_balance = get_balance(ibc.chains["eidon-chain"], src_addr, src_denom)
    rsp = cli.ibc_transfer(
        src_addr,
        dst_addr,
        f"{amt}{src_denom}",
        "channel-0",
        1,
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    new_src_balance = 0

    def check_balance_change():
        nonlocal new_src_balance
        new_src_balance = get_balance(ibc.chains["eidon-chain"], src_addr, src_denom)
        return old_src_balance == new_src_balance

    wait_for_fn("balance no change", check_balance_change)
    new_src_balance = get_balance(ibc.chains["eidon-chain"], src_addr, src_denom)
