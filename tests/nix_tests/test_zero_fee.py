from pathlib import Path

import pytest
from web3 import Web3

from .network import create_snapshots_dir, setup_custom_eidon-chain
from .utils import (
    ADDRS,
    KEYS,
    eth_to_bech32,
    memiavl_config,
    send_transaction,
    wait_for_fn,
)


@pytest.fixture(scope="module")
def custom_eidon-chain(tmp_path_factory):
    yield from setup_custom_eidon-chain(
        tmp_path_factory.mktemp("zero-fee"),
        26900,
        Path(__file__).parent / "configs/zero-fee.jsonnet",
    )


@pytest.fixture(scope="module")
def custom_eidon-chain_rocksdb(tmp_path_factory):
    path = tmp_path_factory.mktemp("zero-fee-rocksdb")
    yield from setup_custom_eidon-chain(
        path,
        26810,
        memiavl_config(path, "zero-fee"),
        post_init=create_snapshots_dir,
        chain_binary="eidond-rocksdb",
    )


@pytest.fixture(scope="module", params=["eidon-chain", "eidon-chain-rocksdb"])
def eidon-chain_cluster(request, custom_eidon-chain, custom_eidon-chain_rocksdb):
    """
    run on eidon-chain and
    eidon-chain built with rocksdb (memIAVL + versionDB)
    """
    provider = request.param
    if provider == "eidon-chain":
        yield custom_eidon-chain
    elif provider == "eidon-chain-rocksdb":
        yield custom_eidon-chain_rocksdb
    else:
        raise NotImplementedError


def test_cosmos_tx(eidon-chain_cluster):
    """
    test basic cosmos transaction works with zero fees
    """
    denom = "aeidon-chain"
    cli = eidon-chain_cluster.cosmos_cli()
    sender = eth_to_bech32(ADDRS["signer1"])
    receiver = eth_to_bech32(ADDRS["signer2"])
    amt = 1000

    old_src_balance = cli.balance(sender, denom)
    old_dst_balance = cli.balance(receiver, denom)

    tx = cli.transfer(
        sender,
        receiver,
        f"{amt}{denom}",
        gas_prices=f"0{denom}",
        generate_only=True,
    )

    tx = cli.sign_tx_json(tx, sender, max_priority_price=0)

    rsp = cli.broadcast_tx_json(tx, broadcast_mode="sync")
    assert rsp["code"] == 0, rsp["raw_log"]

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = cli.balance(receiver, denom)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + amt == new_dst_balance
    new_src_balance = cli.balance(sender, denom)
    # no fees paid, so sender balance should be
    # initial_balance - amount_sent
    assert old_src_balance - amt == new_src_balance


def test_eth_tx(eidon-chain_cluster):
    """
    test basic Ethereum transaction works with zero fees
    """
    w3: Web3 = eidon-chain_cluster.w3

    sender = ADDRS["signer1"]
    receiver = ADDRS["signer2"]
    amt = 1000

    old_src_balance = w3.eth.get_balance(sender)
    old_dst_balance = w3.eth.get_balance(receiver)

    receipt = send_transaction(
        w3,
        {
            "from": sender,
            "to": receiver,
            "value": amt,
            "gasPrice": 0,
        },
        KEYS["signer1"],
    )
    assert receipt.status == 1

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = w3.eth.get_balance(receiver)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + amt == new_dst_balance
    new_src_balance = w3.eth.get_balance(sender)
    # no fees paid, so sender balance should be
    # initial_balance - amount_sent
    assert old_src_balance - amt == new_src_balance
