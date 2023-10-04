import json
import tempfile

from web3 import Web3

from .utils import (
    ACCOUNTS,
    ADDRS,
    CONTRACTS,
    DEFAULT_DENOM,
    deploy_contract,
    eth_to_bech32,
    get_fees_from_tx_result,
    wait_for_new_blocks,
)


def test_send_funds_to_distr_mod(evmos):
    """
    test transfer funds to distribution module account should be forbidden
    """
    cli = evmos.cosmos_cli()
    sender = eth_to_bech32(ADDRS["signer1"])
    amt = 1000

    mod_accs = cli.query_module_accounts()

    for acc in mod_accs:
        if acc["name"] != "distribution":
            continue
        receiver = acc["base_account"]["address"]

    assert receiver is not None

    old_src_balance = cli.balance(sender, DEFAULT_DENOM)

    tx = cli.transfer(
        sender,
        receiver,
        f"{amt}{DEFAULT_DENOM}",
        gas_prices=f"{cli.query_base_fee() + 100000}{DEFAULT_DENOM}",
        generate_only=True,
    )

    tx = cli.sign_tx_json(tx, sender, max_priority_price=0)

    rsp = cli.broadcast_tx_json(tx, broadcast_mode="sync")
    assert rsp["code"] == 0, rsp["raw_log"]
    txhash = rsp["txhash"]

    wait_for_new_blocks(cli, 2)
    receipt = cli.tx_search_rpc(f"tx.hash='{txhash}'")[0]

    assert receipt["tx_result"]["code"] == 4
    assert (
        f"{receiver} is not allowed to receive funds: unauthorized"
        in receipt["tx_result"]["log"]
    )
    fees = get_fees_from_tx_result(receipt["tx_result"])

    # only fees should be deducted from sender balance
    new_src_balance = cli.balance(sender, DEFAULT_DENOM)
    assert old_src_balance - fees == new_src_balance


def test_authz_nested_msg(evmos):
    """
    test sending MsgEthereumTx nested in a MsgExec should be forbidden
    """
    w3: Web3 = evmos.w3
    cli = evmos.cosmos_cli()

    sender_acc = ACCOUNTS["signer1"]
    sender_bech32_addr = eth_to_bech32(sender_acc.address)

    contract, _ = deploy_contract(w3, CONTRACTS["Greeter"])

    tx = contract.functions.setGreeting("world").build_transaction(
        {
            "from": sender_acc.address,
            "nonce": w3.eth.get_transaction_count(sender_acc.address),
            "gas": 999_999_999_999,
        }
    )

    tx_call = sender_acc.sign_transaction(tx)

    # save the eth tx to nest inside a MsgExec to a json file
    with tempfile.NamedTemporaryFile("w") as tx_file:
        json.dump(cli.build_evm_tx(tx_call.rawTransaction.hex()), tx_file)
        tx_file.flush()

        # create the tx with the MsgExec with the eth tx generated previously
        # as nested message
        tx = cli.authz_exec(tx_file.name, sender_bech32_addr)
        tx = cli.sign_tx_json(tx, sender_bech32_addr, max_priority_price=0)

        rsp = cli.broadcast_tx_json(tx, broadcast_mode="sync")

        assert rsp["code"] == 4, rsp["raw_log"]
        assert (
            "found disabled msg type: /ethermint.evm.v1.MsgEthereumTx" in rsp["raw_log"]
        )


def test_create_vesting_acc(evmos):
    """
    test create vesting account with negative/zero amounts should be forbidden
    """
    test_cases = [
        {
            "name": "fail - create vesting account with negative amount",
            "funder": eth_to_bech32(ADDRS["validator"]),
            "address": eth_to_bech32(ADDRS["signer1"]),
            "lockup": {
                "start_time": 1625204910,
                "periods": [
                    {
                        "length_seconds": 2419200,
                        "coins": "10000000000aevmos",
                    }
                ],
            },
            "vesting": {
                "start_time": 1625204910,
                "periods": [
                    {
                        "length_seconds": 2419200,
                        "coins": "10000000000aevmos",
                    },
                    {
                        "length_seconds": 2419200,
                        "coins": "10000000000aevmos",
                    },
                    {
                        "length_seconds": 2419200,
                        "coins": "-10000000000aevmos",
                    },
                ],
            },
        },
        {
            "name": "fail - create vesting account with zero amount",
            "funder": eth_to_bech32(ADDRS["validator"]),
            "address": eth_to_bech32(ADDRS["signer2"]),
            "lockup": {
                "start_time": 1625204910,
                "periods": [
                    {
                        "length_seconds": 2419200,
                        "coins": "10000000000aevmos",
                    }
                ],
            },
            "vesting": {
                "start_time": 1625204910,
                "periods": [
                    {
                        "length_seconds": 2419200,
                        "coins": "10000000000aevmos",
                    },
                    {
                        "length_seconds": 2419200,
                        "coins": "0aevmos",
                    },
                ],
            },
        },
    ]

    for tc in test_cases:
        print("\nCase: {}".format(tc["name"]))
        # create the vesting account

        # funder funds the vesting account and defines the
        # vesting and lockup schedules
