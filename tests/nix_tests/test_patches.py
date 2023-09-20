from .utils import ADDRS, eth_to_bech32, get_fees_from_tx_result, wait_for_new_blocks


def test_send_funds_to_distr_mod(evmos):
    """
    test transfer funds to distribution module account should be forbidden
    """
    denom = "aevmos"
    cli = evmos.cosmos_cli()
    sender = eth_to_bech32(ADDRS["signer1"])
    amt = 1000

    mod_accs = cli.query_module_accounts()

    for acc in mod_accs:
        if acc["name"] != "distribution":
            continue
        receiver = acc["base_account"]["address"]

    assert receiver is not None

    old_src_balance = cli.balance(sender, denom)

    tx = cli.transfer(
        sender,
        receiver,
        f"{amt}{denom}",
        gas_prices=f"{cli.query_base_fee() + 100000}{denom}",
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
    new_src_balance = cli.balance(sender, denom)
    assert old_src_balance - fees == new_src_balance


def test_authz_nested_msg(evmos):
    """
    test sending MsgEthereumTx nested in a MsgExec should be forbidden
    """
    