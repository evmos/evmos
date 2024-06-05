from hashlib import sha256
import json
import pytest

from .ibc_utils import (
    BASECRO_IBC_DENOM,
    EVMOS_IBC_DENOM,
    OSMO_IBC_DENOM,
    assert_ready,
    get_balance,
    prepare_network,
)
from .network import Evmos
from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    MAX_UINT256,
    debug_trace_tx,
    deploy_contract,
    get_precompile_contract,
    send_transaction,
    wait_for_fn,
)


@pytest.fixture(scope="module", params=["evmos"])
def ibc(request, tmp_path_factory):
    """
    Prepares the network.
    """
    name = "ibc-precompile"
    evmos_build = request.param
    path = tmp_path_factory.mktemp(name)
    network = prepare_network(path, name, [evmos_build, "chainmain"])
    yield from network


@pytest.mark.parametrize(
    "name, args, exp_err, err_contains, exp_spend_limit",
    [
        (
            "channel does not exist",
            [
                ADDRS["signer2"],
                [["transfer", "channel-1", [["aevmos", int(1e18)]], []]],
            ],
            True,
            "channel not found",
            0,
        ),
        (
            "MaxInt256 allocation",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", MAX_UINT256]], []]],
            ],
            False,
            "",
            MAX_UINT256,
        ),
        (
            "create authorization with specific spend limit",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            False,
            "",
            int(1e18),
        ),
    ],
)
def test_approve(ibc, name, args, exp_err, err_contains, exp_spend_limit):
    """Test precompile approvals"""
    assert_ready(ibc)

    gas_limit = 200_000
    pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price

    # signer1 creates authorization for signer2
    approve_tx = pc.functions.approve(*args).build_transaction(
        {
            "from": ADDRS["signer1"],
            "gasPrice": evmos_gas_price,
            "gas": gas_limit,
        }
    )

    tx_receipt = send_transaction(ibc.chains["evmos"].w3, approve_tx, KEYS["signer1"])
    if exp_err:
        assert tx_receipt.status == 0, f"Failed: {name}"
        # get the corresponding error message from the trace
        trace = debug_trace_tx(ibc.chains["evmos"], tx_receipt.transactionHash.hex())
        assert err_contains in trace["error"]
        return

    assert tx_receipt.status == 1, f"Failed: {name}"

    # check the IBCTransferAuthorization event was emitted
    auth_event = pc.events.IBCTransferAuthorization().processReceipt(tx_receipt)[0]
    assert auth_event.address == "0x0000000000000000000000000000000000000802"
    assert auth_event.event == "IBCTransferAuthorization"
    assert auth_event.args.grantee == args[0]
    assert auth_event.args.granter == ADDRS["signer1"]
    assert len(auth_event.args.allocations) == 1
    assert auth_event.args.allocations[0] == (
        "transfer",
        "channel-0",
        [("aevmos", exp_spend_limit)],
        [],
    )

    # check the authorization was created
    cli = ibc.chains["evmos"].cosmos_cli()
    granter = cli.address("signer1")
    grantee = cli.address("signer2")
    res = cli.authz_grants(granter, grantee)
    assert len(res["grants"]) == 1, f"Failed: {name}"
    assert (
        res["grants"][0]["authorization"]["type"]
        == "/ibc.applications.transfer.v1.TransferAuthorization"
    )
    assert (
        int(
            res["grants"][0]["authorization"]["value"]["allocations"][0]["spend_limit"][
                0
            ]["amount"]
        )
        == exp_spend_limit
    ), f"Failed: {name}"


@pytest.mark.parametrize(
    "name, args, exp_err, err_contains",
    [
        (
            "not a correct grantee address",
            ["0xinvalid_addr"],
            True,
            "invalid grantee address",
        ),
        (
            "authorization does not exist",
            [ADDRS["community"]],
            True,
            "does not exist",
        ),
        (
            "deletes authorization grant",
            [ADDRS["validator"]],
            False,
            "",
        ),
    ],
)
def test_revoke(ibc, name, args, exp_err, err_contains):
    """Test precompile approval revocation"""
    assert_ready(ibc)

    gas_limit = 200_000
    pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price

    # setup: create authorization to revoke
    # signer1 creates authorization for validator address
    approve_tx = pc.functions.approve(
        ADDRS["validator"],
        [["transfer", "channel-0", [["aevmos", MAX_UINT256]], []]],
    ).build_transaction(
        {
            "from": ADDRS["signer1"],
            "gasPrice": evmos_gas_price,
            "gas": gas_limit,
        }
    )
    tx_receipt = send_transaction(ibc.chains["evmos"].w3, approve_tx, KEYS["signer1"])
    assert tx_receipt.status == 1

    # signer1 revokes authorization
    try:
        revoke_tx = pc.functions.revoke(*args).build_transaction(
            {
                "from": ADDRS["signer1"],
                "gasPrice": evmos_gas_price,
                "gas": gas_limit,
            }
        )

        tx_receipt = send_transaction(
            ibc.chains["evmos"].w3, revoke_tx, KEYS["signer1"]
        )
    except Exception as err:
        if exp_err is True:
            # error expected, continue with next test
            return
        else:
            print(f"Unexpected {err=}, {type(err)=}")
            raise

    if exp_err:
        assert tx_receipt.status == 0, f"Failed: {name}"
        # get the corresponding error message from the trace
        trace = debug_trace_tx(ibc.chains["evmos"], tx_receipt.transactionHash.hex())
        assert err_contains in trace["error"]
        return

    assert tx_receipt.status == 1, f"Failed: {name}"

    # check the IBCTransferAuthorization event was emitted
    auth_event = pc.events.IBCTransferAuthorization().processReceipt(tx_receipt)[0]
    assert auth_event.address == "0x0000000000000000000000000000000000000802"
    assert auth_event.event == "IBCTransferAuthorization"
    assert auth_event.args.grantee == args[0]
    assert auth_event.args.granter == ADDRS["signer1"]
    assert len(auth_event.args.allocations) == 0

    # check the authorization to the validator was revoked
    cli = ibc.chains["evmos"].cosmos_cli()
    granter = cli.address("signer1")
    grantee = cli.address("validator")
    res = cli.authz_grants(granter, grantee)
    assert res["pagination"] == {}, f"Failed: {name}"


@pytest.mark.parametrize(
    "name, auth_args, args, exp_err, err_contains, exp_spend_limit",
    [
        (
            "empty input args",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [],
            True,
            "invalid number of arguments",
            0,
        ),
        (
            "authorization does not exist",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [
                ADDRS["signer1"],
                "transfer",
                "channel-1",
                "aevmos",
                int(1e18),
            ],
            True,
            "does not exist",
            0,
        ),
        (
            "allocation for specified denom does not exist",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [
                ADDRS["signer2"],
                "transfer",
                "channel-0",
                "atom",
                int(1e18),
            ],
            True,
            "no matching allocation found",
            0,
        ),
        (
            "the new spend limit overflows the maxUint256",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [
                ADDRS["signer2"],
                "transfer",
                "channel-0",
                "aevmos",
                MAX_UINT256 - 1,
            ],
            True,
            "integer overflow when increasing allowance",
            0,
        ),
        (
            "increase allowance by 1 EVMOS",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [
                ADDRS["signer2"],
                "transfer",
                "channel-0",
                "aevmos",
                int(1e18),
            ],
            False,
            "",
            json.loads("""[{"denom": "aevmos", "amount": "2000000000000000000"}]"""),
        ),
        (
            "increase allowance by 1 Atom for allocation with a multiple coin denomination",
            [
                ADDRS["signer2"],
                [
                    [
                        "transfer",
                        "channel-0",
                        [["aevmos", int(1e18)], ["uatom", int(1e18)]],
                        [],
                    ]
                ],
            ],
            [
                ADDRS["signer2"],
                "transfer",
                "channel-0",
                "uatom",
                int(1e18),
            ],
            False,
            "",
            json.loads(
                """[{"denom": "aevmos", "amount": "1000000000000000000"},{"denom": "uatom", "amount": "2000000000000000000"}]"""
            ),
        ),
    ],
)
def test_increase_allowance(
    ibc, name, auth_args, args, exp_err, err_contains, exp_spend_limit
):
    """Test precompile increase allowance"""
    assert_ready(ibc)

    gas_limit = 200_000
    pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price

    # setup: create authorization to revoke
    # validator address creates authorization for signer2 address
    # for 1 EVMOS
    approve_tx = pc.functions.approve(*auth_args).build_transaction(
        {
            "from": ADDRS["validator"],
            "gasPrice": evmos_gas_price,
            "gas": gas_limit,
        }
    )
    tx_receipt = send_transaction(ibc.chains["evmos"].w3, approve_tx, KEYS["validator"])
    assert tx_receipt.status == 1

    try:
        # signer1 creates authorization for signer2
        incr_allowance_tx = pc.functions.increaseAllowance(*args).build_transaction(
            {
                "from": ADDRS["validator"],
                "gasPrice": evmos_gas_price,
                "gas": gas_limit,
            }
        )
        tx_receipt = send_transaction(
            ibc.chains["evmos"].w3, incr_allowance_tx, KEYS["validator"]
        )
    except Exception as err:
        if exp_err is True:
            # error expected, continue with next test
            return
        else:
            print(f"Unexpected {err=}, {type(err)=}")
            raise

    if exp_err:
        assert tx_receipt.status == 0, f"Failed: {name}"
        # get the corresponding error message from the trace
        trace = debug_trace_tx(ibc.chains["evmos"], tx_receipt.transactionHash.hex())
        assert err_contains in trace["error"]
        return

    assert tx_receipt.status == 1, f"Failed: {name}"

    # check the IBCTransferAuthorization event was emitted
    auth_event = pc.events.IBCTransferAuthorization().processReceipt(tx_receipt)[0]
    assert auth_event.address == "0x0000000000000000000000000000000000000802"
    assert auth_event.event == "IBCTransferAuthorization"
    assert auth_event.args.grantee == args[0]
    assert auth_event.args.granter == ADDRS["validator"]
    assert len(auth_event.args.allocations) == 1

    # check the authorization was created
    cli = ibc.chains["evmos"].cosmos_cli()
    granter = cli.address("validator")
    grantee = cli.address("signer2")
    res = cli.authz_grants(granter, grantee)
    assert len(res["grants"]) == 1, f"Failed: {name}"
    assert (
        res["grants"][0]["authorization"]["type"]
        == "/ibc.applications.transfer.v1.TransferAuthorization"
    )
    assert (
        res["grants"][0]["authorization"]["value"]["allocations"][0]["spend_limit"]
        == exp_spend_limit
    ), f"Failed: {name}"


@pytest.mark.parametrize(
    "name, auth_args, args, exp_err, err_contains, exp_spend_limit",
    [
        (
            "empty input args",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [],
            True,
            "invalid number of arguments",
            0,
        ),
        (
            "authorization does not exist",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [
                ADDRS["signer1"],
                "transfer",
                "channel-1",
                "aevmos",
                int(1e18),
            ],
            True,
            "does not exist",
            0,
        ),
        (
            "allocation for specified denom does not exist",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [
                ADDRS["signer2"],
                "transfer",
                "channel-0",
                "atom",
                int(1e18),
            ],
            True,
            "no matching allocation found",
            0,
        ),
        (
            "the new spend limit is negative",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [
                ADDRS["signer2"],
                "transfer",
                "channel-0",
                "aevmos",
                int(2e18),
            ],
            True,
            "negative amount when decreasing allowance",
            0,
        ),
        (
            "decrease allowance by 0.5 EVMOS",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [
                ADDRS["signer2"],
                "transfer",
                "channel-0",
                "aevmos",
                int(5e17),
            ],
            False,
            "",
            json.loads("""[{"denom": "aevmos", "amount": "500000000000000000"}]"""),
        ),
        (
            "decrease allowance by 0.5 Atom for allocation with a multiple coin denomination",
            [
                ADDRS["signer2"],
                [
                    [
                        "transfer",
                        "channel-0",
                        [["aevmos", int(1e18)], ["uatom", int(1e18)]],
                        [],
                    ]
                ],
            ],
            [
                ADDRS["signer2"],
                "transfer",
                "channel-0",
                "uatom",
                int(5e17),
            ],
            False,
            "",
            json.loads(
                """[{"denom": "aevmos", "amount": "1000000000000000000"},{"denom": "uatom", "amount": "500000000000000000"}]"""
            ),
        ),
    ],
)
def test_decrease_allowance(
    ibc, auth_args, name, args, exp_err, err_contains, exp_spend_limit
):
    """Test precompile decrease allowance"""
    assert_ready(ibc)

    gas_limit = 200_000
    pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price

    # setup: create authorization to revoke
    # community address creates authorization for signer2 address
    # for 1 EVMOS
    approve_tx = pc.functions.approve(*auth_args).build_transaction(
        {
            "from": ADDRS["community"],
            "gasPrice": evmos_gas_price,
            "gas": gas_limit,
        }
    )
    tx_receipt = send_transaction(ibc.chains["evmos"].w3, approve_tx, KEYS["community"])
    assert tx_receipt.status == 1

    try:
        # signer1 creates authorization for signer2
        decr_allowance_tx = pc.functions.decreaseAllowance(*args).build_transaction(
            {
                "from": ADDRS["community"],
                "gasPrice": evmos_gas_price,
                "gas": gas_limit,
            }
        )
        tx_receipt = send_transaction(
            ibc.chains["evmos"].w3, decr_allowance_tx, KEYS["community"]
        )
    except Exception as err:
        if exp_err is True:
            # error expected, continue with next test
            return
        else:
            print(f"Unexpected {err=}, {type(err)=}")
            raise

    if exp_err:
        assert tx_receipt.status == 0, f"Failed: {name}"
        # get the corresponding error message from the trace
        trace = debug_trace_tx(ibc.chains["evmos"], tx_receipt.transactionHash.hex())
        assert err_contains in trace["error"]
        return

    assert tx_receipt.status == 1, f"Failed: {name}"

    # check the IBCTransferAuthorization event was emitted
    auth_event = pc.events.IBCTransferAuthorization().processReceipt(tx_receipt)[0]
    assert auth_event.address == "0x0000000000000000000000000000000000000802"
    assert auth_event.event == "IBCTransferAuthorization"
    assert auth_event.args.grantee == args[0]
    assert auth_event.args.granter == ADDRS["community"]
    assert len(auth_event.args.allocations) == 1

    # check the authorization was created
    cli = ibc.chains["evmos"].cosmos_cli()
    granter = cli.address("community")
    grantee = cli.address("signer2")
    res = cli.authz_grants(granter, grantee)
    assert len(res["grants"]) == 1, f"Failed: {name}"
    assert (
        res["grants"][0]["authorization"]["type"]
        == "/ibc.applications.transfer.v1.TransferAuthorization"
    )
    assert (
        res["grants"][0]["authorization"]["value"]["allocations"][0]["spend_limit"]
        == exp_spend_limit
    ), f"Failed: {name}"


@pytest.mark.parametrize(
    "name, args, exp_err, err_contains, exp_res",
    [
        (
            "empty input args",
            [],
            True,
            "improper number of arguments",
            None,
        ),
        (
            "invalid denom trace",
            ["invalid_denom_trace"],
            True,
            "invalid denom trace",
            None,
        ),
        (
            "denom trace not found, return empty struct",
            [OSMO_IBC_DENOM],
            False,
            None,
            ("", ""),
        ),
        (
            "existing denom trace",
            [BASECRO_IBC_DENOM],
            False,
            None,
            ("transfer/channel-0", "basecro"),
        ),
    ],
)
def test_denom_trace(ibc, name, args, exp_err, err_contains, exp_res):
    """Test ibc precompile denom trace query"""
    assert_ready(ibc)

    # setup: send some funds from chain-main to evmos
    # to register the denom trace (if not registered already)
    res = ibc.chains["evmos"].cosmos_cli().denom_traces()

    if len(res["denom_traces"]) == 0:
        amt = 100
        src_denom = "basecro"
        dst_addr = ibc.chains["evmos"].cosmos_cli().address("signer2")
        src_addr = ibc.chains["chainmain"].cosmos_cli().address("signer2")
        rsp = (
            ibc.chains["chainmain"]
            .cosmos_cli()
            .ibc_transfer(
                src_addr,
                dst_addr,
                f"{amt}{src_denom}",
                "channel-0",
                1,
                fees="10000000000basecro",
            )
        )
        assert rsp["code"] == 0, rsp["raw_log"]

        # wait for the ack and registering the denom trace
        def check_denom_trace_change():
            res = ibc.chains["evmos"].cosmos_cli().denom_traces()
            return len(res["denom_traces"]) > 0

        wait_for_fn("denom trace registration", check_denom_trace_change)

    # define the query response outside the try-catch block
    # to use it for further validations
    query_res = None
    try:
        # make the query call
        pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
        query_res = pc.functions.denomTrace(*args).call()
    except Exception as err:
        if exp_err is True:
            # stringify error in case it is a struct
            err_msg = json.dumps(err.args[0], separators=(",", ":"))
            assert err_contains in err_msg
            # error expected, continue with next test
            return
        else:
            print(f"Unexpected {err=}, {type(err)=}")
            raise

    # check the query response
    assert query_res == exp_res, f"Failed: {name}"


def test_ibc_transfer_from_contract(ibc):
    """Test ibc transfer from contract"""
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    w3 = evmos.w3

    dst_addr = ibc.chains["chainmain"].cosmos_cli().address("signer2")
    amt = 1000000000000000000
    src_denom = "aevmos"
    gas_limit = 200_000

    pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price

    src_adr = ibc.chains["evmos"].cosmos_cli().address("signer2")

    # Deployment of contracts and initial checks
    eth_contract, tx_receipt = deploy_contract(w3, CONTRACTS["ICS20FromContract"])
    assert tx_receipt.status == 1

    counter = eth_contract.functions.counter().call()
    assert counter == 0

    # Approve the contract to spend the src_denom
    approve_tx = pc.functions.approve(
        eth_contract.address, [["transfer", "channel-0", [[src_denom, amt]], []]]
    ).build_transaction(
        {
            "from": ADDRS["signer2"],
            "gasPrice": evmos_gas_price,
            "gas": gas_limit,
        }
    )
    tx_receipt = send_transaction(ibc.chains["evmos"].w3, approve_tx, KEYS["signer2"])
    assert tx_receipt.status == 1

    def check_allowance_set():
        new_allowance = pc.functions.allowance(
            eth_contract.address, ADDRS["signer2"]
        ).call()
        return new_allowance != []

    wait_for_fn("allowance has changed", check_allowance_set)

    src_amount_evmos_prev = get_balance(ibc.chains["evmos"], src_adr, src_denom)
    # Deposit into the contract
    deposit_tx = eth_contract.functions.deposit().build_transaction(
        {
            "from": ADDRS["signer2"],
            "value": amt,
            "gas": gas_limit,
            "gasPrice": evmos_gas_price,
        }
    )
    deposit_receipt = send_transaction(
        ibc.chains["evmos"].w3, deposit_tx, KEYS["signer2"]
    )
    assert deposit_receipt.status == 1
    fees = deposit_receipt.gasUsed * evmos_gas_price

    def check_contract_balance():
        new_contract_balance = eth_contract.functions.balanceOfContract().call()
        return new_contract_balance > 0

    wait_for_fn("contract balance change", check_contract_balance)

    # Calling the actual transfer function on the custom contract
    send_tx = eth_contract.functions.transfer(
        "transfer", "channel-0", src_denom, amt, dst_addr
    ).build_transaction(
        {
            "from": ADDRS["signer2"],
            "gasPrice": evmos_gas_price,
            "gas": gas_limit,
        }
    )
    receipt = send_transaction(ibc.chains["evmos"].w3, send_tx, KEYS["signer2"])
    assert receipt.status == 1
    fees += receipt.gasUsed * evmos_gas_price

    final_dest_balance = 0

    def check_dest_balance():
        nonlocal final_dest_balance
        final_dest_balance = get_balance(
            ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM
        )
        return final_dest_balance > 0

    # check balance of destination
    wait_for_fn("destination balance change", check_dest_balance)
    assert final_dest_balance == amt

    # check balance of contract
    final_contract_balance = eth_contract.functions.balanceOfContract().call()
    assert final_contract_balance == 0

    src_amount_evmos = get_balance(ibc.chains["evmos"], src_adr, src_denom)
    assert src_amount_evmos == src_amount_evmos_prev - amt - fees

    # check counter of contract
    counter_after = eth_contract.functions.counter().call()
    assert counter_after == 0


def test_ibc_transfer_from_eoa_through_contract(ibc):
    """Test ibc transfer from EOA through a Smart Contract call"""
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    w3 = evmos.w3

    amt = 1000000000000000000
    src_denom = "aevmos"
    gas_limit = 200_000
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price

    dst_addr = ibc.chains["chainmain"].cosmos_cli().address("signer2")
    src_adr = ibc.chains["evmos"].cosmos_cli().address("signer2")

    # Deployment of contracts and initial checks
    eth_contract, tx_receipt = deploy_contract(w3, CONTRACTS["ICS20FromContract"])
    assert tx_receipt.status == 1

    counter = eth_contract.functions.counter().call()
    assert counter == 0

    pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
    # Approve the contract to spend the src_denom
    approve_tx = pc.functions.approve(
        eth_contract.address, [["transfer", "channel-0", [[src_denom, amt]], []]]
    ).build_transaction(
        {
            "from": ADDRS["signer2"],
            "gasPrice": evmos_gas_price,
            "gas": gas_limit,
        }
    )
    tx_receipt = send_transaction(ibc.chains["evmos"].w3, approve_tx, KEYS["signer2"])
    assert tx_receipt.status == 1

    def check_allowance_set():
        new_allowance = pc.functions.allowance(
            eth_contract.address, ADDRS["signer2"]
        ).call()
        return new_allowance != []

    wait_for_fn("allowance has changed", check_allowance_set)

    src_starting_balance = get_balance(ibc.chains["evmos"], src_adr, "aevmos")
    dest_starting_balance = get_balance(
        ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM
    )
    # Calling the actual transfer function on the custom contract
    send_tx = eth_contract.functions.transferFromEOA(
        "transfer", "channel-0", src_denom, amt, dst_addr
    ).build_transaction(
        {"from": ADDRS["signer2"], "gasPrice": evmos_gas_price, "gas": gas_limit}
    )
    receipt = send_transaction(ibc.chains["evmos"].w3, send_tx, KEYS["signer2"])
    assert receipt.status == 1
    fees = receipt.gasUsed * evmos_gas_price

    final_dest_balance = dest_starting_balance

    def check_dest_balance():
        nonlocal final_dest_balance
        final_dest_balance = get_balance(
            ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM
        )
        return final_dest_balance > dest_starting_balance

    wait_for_fn("destination balance change", check_dest_balance)
    assert final_dest_balance == dest_starting_balance + amt

    src_final_amount_evmos = get_balance(ibc.chains["evmos"], src_adr, src_denom)
    assert src_final_amount_evmos == src_starting_balance - amt - fees

    counter_after = eth_contract.functions.counter().call()
    assert counter_after == 0
