import json
import tempfile

from web3 import Web3

from .utils import (
    ACCOUNTS,
    ADDRS,
    CONTRACTS,
    KEYS,
    SCALE_FACTOR_6DEC,
    decode_bech32,
    deploy_contract,
    eth_to_bech32,
    get_fees_from_tx_result,
    get_scaling_factor,
    send_transaction,
    wait_for_cosmos_tx_receipt,
    wait_for_new_blocks,
)


def test_send_funds_to_distr_mod(evmos_cluster):
    """
    This tests the transfer of funds to the distribution module account,
    which should be forbidden, since this is a blocked address.
    """
    cli = evmos_cluster.cosmos_cli()
    sender = eth_to_bech32(ADDRS["signer1"])
    amt = 1000
    fee_denom = cli.evm_denom()
    scaling_factor = get_scaling_factor(cli)

    mod_accs = cli.query_module_accounts()

    for acc in mod_accs:
        if acc["value"]["name"] != "distribution":
            continue
        receiver = acc["value"]["address"]

    assert receiver is not None

    old_src_balance = cli.balance(sender, fee_denom)

    tx = cli.transfer(
        sender,
        receiver,
        f"{amt}{fee_denom}",
        gas_prices=f"{cli.query_base_fee() + 100000/scaling_factor:.12f}{fee_denom}",
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
    fees = get_fees_from_tx_result(receipt["tx_result"], denom=fee_denom)

    # only fees should be deducted from sender balance
    new_src_balance = cli.balance(sender, fee_denom)
    assert old_src_balance - fees == new_src_balance


def test_send_funds_to_distr_mod_eth_tx(evmos_cluster):
    """
    This tests the transfer of funds to the distribution module account,
    via ethereum tx
    which should be forbidden, since this is a blocked address.
    """
    cli = evmos_cluster.cosmos_cli()
    w3 = evmos_cluster.w3

    sender = ADDRS["signer1"]
    mod_accs = cli.query_module_accounts()
    evm_denom = cli.evm_denom()
    old_src_balance = cli.balance(eth_to_bech32(sender), evm_denom)

    for acc in mod_accs:
        if acc["value"]["name"] != "distribution":
            continue
        receiver = decode_bech32(acc["value"]["address"])

    assert receiver is not None

    receipt = send_transaction(
        w3,
        {
            "from": sender,
            "to": receiver,
            "value": int(1e18),
        },
        KEYS["signer1"],
    )
    assert receipt.status == 0  # failed status expected

    wait_for_new_blocks(cli, 2)
    # only fees should be deducted from sender balance
    fees = receipt["gasUsed"] * receipt["effectiveGasPrice"]
    assert fees > 0

    # check if evm has 6 dec,
    # actual fees will have 6 dec
    # instead of 18
    scaling_factor = get_scaling_factor(cli)
    fees = int(fees / scaling_factor)

    new_src_balance = cli.balance(eth_to_bech32(sender), evm_denom)
    assert old_src_balance - fees == new_src_balance


def test_authz_nested_msg(evmos_cluster):
    """
    test sending MsgEthereumTx nested in a MsgExec should be forbidden
    """
    w3: Web3 = evmos_cluster.w3
    cli = evmos_cluster.cosmos_cli()

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
        json.dump(cli.build_evm_tx(tx, tx_call), tx_file)
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


def test_create_invalid_vesting_acc(evmos_cluster):
    """
    test create vesting account with account address != signer address
    """
    cli = evmos_cluster.cosmos_cli()
    # create the vesting account
    tx = cli.create_vesting_acc(
        eth_to_bech32(ADDRS["validator"]),
        "evmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqzqpgshrm7",
    )
    try:
        tx = cli.sign_tx_json(tx, eth_to_bech32(ADDRS["signer1"]), max_priority_price=0)
        raise Exception(  # pylint: disable=broad-exception-raised
            "This command should have failed"
        )
    except Exception as error:
        assert "tx intended signer does not match the given signer" in error.args[0]


def test_vesting_acc_schedule(evmos_cluster):
    """
    test vesting account with negative/zero amounts should be forbidden
    """
    test_cases = [
        {
            "name": "fail - vesting account with negative amount",
            "funder": eth_to_bech32(ADDRS["validator"]),
            "address": eth_to_bech32(ADDRS["signer1"]),
            "exp_err": "invalid decimal coin expression: -10000000000aevmos",
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
            "name": "fail - vesting account with zero amount",
            "funder": eth_to_bech32(ADDRS["validator"]),
            "address": eth_to_bech32(ADDRS["signer2"]),
            "exp_err": "invalid request",
            "lockup": {
                "start_time": 1625204910,
                "periods": [
                    {
                        "length_seconds": 2419200,
                        "coins": "0aevmos",
                    }
                ],
            },
            "vesting": {
                "start_time": 1625204910,
                "periods": [
                    {
                        "length_seconds": 2419200,
                        "coins": "0aevmos",
                    },
                ],
            },
        },
    ]

    cli = evmos_cluster.cosmos_cli()
    for tc in test_cases:
        print("\nCase: {}".format(tc["name"]))
        # create the vesting account
        tx = cli.create_vesting_acc(
            tc["funder"],
            tc["address"],
        )
        tx = cli.sign_tx_json(tx, tc["address"], max_priority_price=0)

        rsp = cli.broadcast_tx_json(tx, broadcast_mode="sync")
        # assert tx returns OK code
        assert rsp["code"] == 0

        # wait tx to be committed
        wait_for_new_blocks(cli, 2)

        # funder funds the vesting account and defines the
        # vesting and lockup schedules
        with tempfile.NamedTemporaryFile("w") as lockup_file:
            json.dump(tc["lockup"], lockup_file)
            lockup_file.flush()

            with tempfile.NamedTemporaryFile("w") as vesting_file:
                json.dump(tc["vesting"], vesting_file)
                vesting_file.flush()

                # expect an error
                try:
                    tx = cli.fund_vesting_acc(
                        tc["address"],
                        tc["funder"],
                        lockup_file.name,
                        vesting_file.name,
                    )
                    raise Exception(  # pylint: disable=broad-exception-raised
                        "This command should have failed"
                    )
                except Exception as error:
                    assert tc["exp_err"] in error.args[0]


def test_unvested_token_delegation(evmos_cluster):
    """
    test vesting account cannot delegate unvested tokens
    """
    cli = evmos_cluster.cosmos_cli()
    funder = eth_to_bech32(ADDRS["signer1"])
    # add a new key that will be the vesting account
    acc = cli.create_account("vesting_acc")
    address = acc["address"]

    fee_denom = cli.evm_denom()
    scaling_factor = get_scaling_factor(cli)

    # transfer some funds to pay for tx fees
    # when creating the vesting account
    tx = cli.transfer(
        funder,
        address,
        f"{int(7000000000000000/scaling_factor)}{fee_denom}",
        gas_prices=f"{cli.query_base_fee() + 100000/scaling_factor:.12f}{fee_denom}",
        generate_only=True,
    )

    tx = cli.sign_tx_json(tx, funder, max_priority_price=0)

    rsp = cli.broadcast_tx_json(tx, broadcast_mode="sync")
    assert rsp["code"] == 0, rsp["raw_log"]

    wait_for_new_blocks(cli, 2)

    # create the vesting account
    tx = cli.create_vesting_acc(funder, address)
    tx = cli.sign_tx_json(tx, address, max_priority_price=0)
    rsp = cli.broadcast_tx_json(tx, broadcast_mode="sync")

    # assert tx returns OK code
    assert rsp["code"] == 0, rsp["raw_log"]

    # wait tx to be committed
    # get tx receipt and check if tx was succesful
    receipt = wait_for_cosmos_tx_receipt(cli, rsp["txhash"])
    assert receipt["tx_result"]["code"] == 0, receipt["log"]

    # fund vesting account
    with tempfile.NamedTemporaryFile("w") as lockup_file:
        json.dump(
            {
                "start_time": 1625204910,
                "periods": [
                    {
                        "length_seconds": 1675184400,
                        "coins": "10000000000000000000aevmos",
                    }
                ],
            },
            lockup_file,
        )
        lockup_file.flush()

        with tempfile.NamedTemporaryFile("w") as vesting_file:
            json.dump(
                {
                    "start_time": 1625204910,
                    "periods": [
                        {
                            "length_seconds": 1675184400,
                            "coins": "3000000000000000000aevmos",
                        },
                        {
                            "length_seconds": 2419200,
                            "coins": "3000000000000000000aevmos",
                        },
                        {
                            "length_seconds": 2419200,
                            "coins": "4000000000000000000aevmos",
                        },
                    ],
                },
                vesting_file,
            )
            vesting_file.flush()

            tx = cli.fund_vesting_acc(
                address,
                funder,
                lockup_file.name,
                vesting_file.name,
            )
            tx = cli.sign_tx_json(tx, funder, max_priority_price=0)
            rsp = cli.broadcast_tx_json(tx, broadcast_mode="sync")
            # assert tx returns OK code
            assert rsp["code"] == 0, rsp["raw_log"]

            # wait tx to be committed
            # get tx receipt and check if tx was succesful
            receipt = wait_for_cosmos_tx_receipt(cli, rsp["txhash"])
            assert receipt["tx_result"]["code"] == 0, receipt["log"]

    # check vesting balances
    # vested should be zero at this point
    balances = cli.vesting_balance_http(address)
    assert balances["vested"] == []
    assert len(balances["locked"]) == 1
    assert balances["locked"] == balances["unvested"]

    # try to delegate more than the allowed tokens
    del_amt = "7000000000000000000aevmos"
    validator_addr = cli.validators()[0]["operator_address"]
    tx = cli.delegate_amount(
        validator_addr,
        del_amt,
        address,
    )
    tx = cli.sign_tx_json(tx, address, max_priority_price=0)
    rsp = cli.broadcast_tx_json(tx, broadcast_mode="sync")
    # get tx receipt to check if tx failed as expected
    receipt = wait_for_cosmos_tx_receipt(cli, rsp["txhash"])
    # assert tx fails with corresponding error message
    assert receipt["tx_result"]["code"] == 2
    assert "insufficient vested coins" in receipt["tx_result"]["log"]
