import json
import pytest

from .ibc_utils import (
    BASECRO_IBC_DENOM,
    EVMOS_IBC_DENOM,
    OSMO_IBC_DENOM,
    assert_ready,
    get_balance,
    prepare_network,
    setup_denom_trace,
)
from .network import Evmos
from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    MAX_UINT256,
    check_error,
    debug_trace_tx,
    deploy_contract,
    eth_to_bech32,
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
    "name, args, err_contains, exp_res",
    [
        (
            "empty input args",
            [],
            "improper number of arguments",
            None,
        ),
        (
            "invalid denom trace",
            ["invalid_denom_trace"],
            "invalid denom trace",
            None,
        ),
        (
            "denom trace not found, return empty struct",
            [OSMO_IBC_DENOM],
            None,
            ("", ""),
        ),
        (
            "existing denom trace",
            [BASECRO_IBC_DENOM],
            None,
            ("transfer/channel-0", "basecro"),
        ),
    ],
)
def test_denom_trace(ibc, name, args, err_contains, exp_res):
    """Test ibc precompile denom trace query"""
    assert_ready(ibc)

    # setup: send some funds from chain-main to evmos
    # to register the denom trace (if not registered already)
    setup_denom_trace(ibc)

    # define the query response outside the try-catch block
    # to use it for further validations
    query_res = None
    try:
        # make the query call
        pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
        query_res = pc.functions.denomTrace(*args).call()
    except Exception as err:
        check_error(err, err_contains)

    # check the query response
    assert query_res == exp_res, f"Failed: {name}"


@pytest.mark.parametrize(
    "name, args, err_contains, exp_res",
    [
        (
            "empty input args",
            [],
            "improper number of arguments",
            None,
        ),
        (
            "gets denom traces with pagination",
            [[b"", 0, 3, True, False]],
            None,
            [[("transfer/channel-0", "basecro")], (b"", 1)],
        ),
    ],
)
def test_denom_traces(ibc, name, args, err_contains, exp_res):
    """Test ibc precompile denom traces query"""
    assert_ready(ibc)

    # setup: send some funds from chain-main to evmos
    # to register the denom trace (if not registered already)
    setup_denom_trace(ibc)

    # define the query response outside the try-catch block
    # to use it for further validations
    query_res = None
    try:
        # make the query call
        pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
        query_res = pc.functions.denomTraces(*args).call()
    except Exception as err:
        check_error(err, err_contains)

    # check the query response
    assert query_res == exp_res, f"Failed: {name}"


@pytest.mark.parametrize(
    "name, args, exp_res",
    [
        (
            "trace not found, returns empty string",
            ["transfer/channel-1/uatom"],
            (""),
        ),
        (
            "get the hash of a denom trace",
            ["transfer/channel-0/basecro"],
            (BASECRO_IBC_DENOM.split("/")[1]),
        ),
    ],
)
def test_denom_hash(ibc, name, args, exp_res):
    """Test ibc precompile denom traces query"""
    assert_ready(ibc)

    # setup: send some funds from chain-main to evmos
    # to register the denom trace (if not registered already)
    setup_denom_trace(ibc)

    # define the query response outside the try-catch block
    # to use it for further validations
    query_res = None

    # make the query call
    pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
    query_res = pc.functions.denomHash(*args).call()

    # check the query response
    assert query_res == exp_res, f"Failed: {name}"


@pytest.mark.parametrize(
    "name, auth_args, args, err_contains, exp_res",
    [
        (
            "empty input args",
            [
                ADDRS["community"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [],
            "improper number of arguments",
            None,
        ),
        (
            "authorization does not exist - returns empty array",
            [
                ADDRS["community"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [
                ADDRS["community"],
                ADDRS["signer1"],
            ],
            None,
            [],
        ),
        (
            "existing authorization with one denom",
            [
                ADDRS["community"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], []]],
            ],
            [
                ADDRS["community"],
                ADDRS["validator"],
            ],
            None,
            [("transfer", "channel-0", [("aevmos", 1000000000000000000)], [])],
        ),
        (
            "existing authorization with a multiple coin denomination",
            [
                ADDRS["community"],
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
                ADDRS["community"],
                ADDRS["validator"],
            ],
            None,
            [
                (
                    "transfer",
                    "channel-0",
                    [("aevmos", 1000000000000000000), ("uatom", 1000000000000000000)],
                    [],
                )
            ],
        ),
    ],
)
def test_query_allowance(ibc, name, auth_args, args, err_contains, exp_res):
    """Test precompile increase allowance"""
    assert_ready(ibc)

    gas_limit = 200_000
    pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price

    # setup: create authorization to revoke
    # validator address creates authorization for the provided auth_args
    approve_tx = pc.functions.approve(*auth_args).build_transaction(
        {
            "from": ADDRS["validator"],
            "gasPrice": evmos_gas_price,
            "gas": gas_limit,
        }
    )
    tx_receipt = send_transaction(ibc.chains["evmos"].w3, approve_tx, KEYS["validator"])
    assert tx_receipt.status == 1

    allowance_res = None
    try:
        # signer1 creates authorization for signer2
        allowance_res = pc.functions.allowance(*args).call()
    except Exception as err:
        check_error(err, err_contains)

    assert allowance_res == exp_res, f"Failed: {name}"


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
            "invocation failed due to no matching argument types",
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
        check_error(err, err_contains)
        return

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
            "improper number of arguments",
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
        check_error(err, err_contains)
        return

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
            "improper number of arguments",
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
        check_error(err, err_contains)
        return

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
    "name, auth_coins, args, exp_err, err_contains, transfer_amt, exp_spend_limit",
    [
        ("empty input args", None, [], True, "improper number of arguments", 0, None),
        (
            "channel does not exist",
            None,
            [
                "transfer",
                "channel-1",
                "aevmos",
                int(1e18),
                "cro1apdh4yc2lnpephevc6lmpvkyv6s5cjh652n6e4",  # signer2 in chain-main
            ],
            True,
            "channel not found",
            int(1e18),
            None,
        ),
        (
            "non authorized denom",
            [["aevmos", int(1e18)]],
            [
                "transfer",
                "channel-0",
                "uatom",
                int(1e18),
                "cro1apdh4yc2lnpephevc6lmpvkyv6s5cjh652n6e4",
            ],
            True,
            "requested amount is more than spend limit",
            int(1e18),
            None,
        ),
        (
            "allowance is less than transfer amount",
            [["aevmos", int(1e18)]],
            [
                "transfer",
                "channel-0",
                "aevmos",
                int(2e18),
                "cro1apdh4yc2lnpephevc6lmpvkyv6s5cjh652n6e4",
            ],
            True,
            "requested amount is more than spend limit",
            int(2e18),
            None,
        ),
        (
            "transfer 1 Evmos from chainA to chainB and spend the entire allowance",
            [["aevmos", int(1e18)]],
            [
                "transfer",
                "channel-0",
                "aevmos",
                int(1e18),
                "cro1apdh4yc2lnpephevc6lmpvkyv6s5cjh652n6e4",
            ],
            False,
            None,
            int(1e18),
            None,
        ),
        (
            "transfer 1 Evmos from chainA to chainB and don't change the unlimited spending limit",
            [["aevmos", MAX_UINT256]],
            [
                "transfer",
                "channel-0",
                "aevmos",
                int(1e18),
                "cro1apdh4yc2lnpephevc6lmpvkyv6s5cjh652n6e4",
            ],
            False,
            None,
            int(1e18),
            json.loads(f"""[{{"denom": "aevmos", "amount": "{MAX_UINT256}"}}]"""),
        ),
        (
            "transfer 1 Evmos from chainA to chainB and only change 1 spend limit",
            [["aevmos", int(1e18)], ["uatom", int(1e18)]],
            [
                "transfer",
                "channel-0",
                "aevmos",
                int(1e18),
                "cro1apdh4yc2lnpephevc6lmpvkyv6s5cjh652n6e4",
            ],
            False,
            None,
            int(1e18),
            json.loads(f"""[{{"denom": "uatom", "amount": "{int(1e18)}"}}]"""),
        ),
    ],
)
def test_ibc_transfer_with_authorization(
    ibc, name, auth_coins, args, exp_err, err_contains, transfer_amt, exp_spend_limit
):
    """Test ibc transfer with authorization (using a smart contract)"""
    assert_ready(ibc)

    pc = get_precompile_contract(ibc.chains["evmos"].w3, "ICS20I")
    gas_limit = 200_000
    src_denom = "aevmos"
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price
    src_address = ibc.chains["evmos"].cosmos_cli().address("signer2")
    dst_address = ibc.chains["chainmain"].cosmos_cli().address("signer2")

    # test setup:
    # deploy contract that calls the ics-20 precompile
    w3 = ibc.chains["evmos"].w3
    eth_contract, tx_receipt = deploy_contract(w3, CONTRACTS["ICS20FromContract"])
    assert tx_receipt.status == 1

    counter = eth_contract.functions.counter().call()
    assert counter == 0

    # create the authorization for the deployed contract
    # based on the specific coins for each test case
    if auth_coins is not None:
        approve_tx = pc.functions.approve(
            eth_contract.address, [["transfer", "channel-0", auth_coins, []]]
        ).build_transaction(
            {
                "from": ADDRS["signer2"],
                "gasPrice": evmos_gas_price,
                "gas": gas_limit,
            }
        )
        receipt = send_transaction(ibc.chains["evmos"].w3, approve_tx, KEYS["signer2"])
        assert receipt.status == 1, f"Failed: {name}"

        def check_allowance_set():
            new_allowance = pc.functions.allowance(
                eth_contract.address, ADDRS["signer2"]
            ).call()
            return new_allowance != []

        wait_for_fn("allowance has changed", check_allowance_set)

    # get the balances previous to the transfer to validate them after the tx
    src_amount_evmos_prev = get_balance(ibc.chains["evmos"], src_address, src_denom)
    dst_balance_prev = get_balance(
        ibc.chains["chainmain"], dst_address, EVMOS_IBC_DENOM
    )
    try:
        # Calling the actual transfer function on the custom contract
        transfer_tx = eth_contract.functions.transferFromEOA(*args).build_transaction(
            {
                "from": ADDRS["signer2"],
                "gasPrice": evmos_gas_price,
                "gas": gas_limit,
            }
        )
        receipt = send_transaction(ibc.chains["evmos"].w3, transfer_tx, KEYS["signer2"])
    except Exception as err:
        check_error(err, err_contains)
        return

    if exp_err:
        assert receipt.status == 0, f"Failed: {name}"
        # get the corresponding error message from the trace
        trace = debug_trace_tx(ibc.chains["evmos"], receipt.transactionHash.hex())
        # stringify the tx trace to look for the expected error message
        trace_str = json.dumps(trace, separators=(",", ":"))
        assert err_contains in trace_str
        return

    assert receipt.status == 1, debug_trace_tx(
        ibc.chains["evmos"], receipt.transactionHash.hex()
    )
    fees = receipt.gasUsed * evmos_gas_price

    # check ibc-transfer event was emitted
    transfer_event = pc.events.IBCTransfer().processReceipt(receipt)[0]
    assert transfer_event.address == "0x0000000000000000000000000000000000000802"
    assert transfer_event.event == "IBCTransfer"
    assert transfer_event.args.sender == ADDRS["signer2"]
    # TODO check if we want to keep the keccak256 hash bytes or smth better
    # assert transfer_event.args.receiver == dst_addr
    assert transfer_event.args.sourcePort == "transfer"
    assert transfer_event.args.sourceChannel == "channel-0"
    assert transfer_event.args.denom == "aevmos"
    assert transfer_event.args.amount == transfer_amt
    assert transfer_event.args.memo == ""

    # check the authorization was updated
    cli = ibc.chains["evmos"].cosmos_cli()
    granter = cli.address("signer2")
    grantee = eth_to_bech32(eth_contract.address)
    res = cli.authz_grants(granter, grantee)

    if exp_spend_limit is None:
        assert "grants" not in res
    else:
        assert len(res["grants"]) == 1, f"Failed: {name}"
        assert (
            res["grants"][0]["authorization"]["type"]
            == "/ibc.applications.transfer.v1.TransferAuthorization"
        )
        assert (
            res["grants"][0]["authorization"]["value"]["allocations"][0]["spend_limit"]
            == exp_spend_limit
        ), f"Failed: {name}"

    # check balances were updated
    # contract balance should be 0
    final_contract_balance = eth_contract.functions.balanceOfContract().call()
    assert final_contract_balance == 0

    # signer2 (src) balance should be reduced by the fees paid
    src_amount_evmos_final = get_balance(ibc.chains["evmos"], src_address, src_denom)

    assert src_amount_evmos_final == src_amount_evmos_prev - fees - transfer_amt

    # dst_address should have received the IBC coins
    dst_balance_final = 0

    def check_balance_change():
        nonlocal dst_balance_final
        dst_balance_final = get_balance(
            ibc.chains["chainmain"], dst_address, EVMOS_IBC_DENOM
        )
        return dst_balance_final > dst_balance_prev

    wait_for_fn("balance change", check_balance_change)

    assert dst_balance_final - dst_balance_prev == transfer_amt

    # check counter of contract has the corresponding value
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


def revoke_all_grants(cli, w3, granter):
    """
    Helper function to revoke all grants
    """
    for grantee in ["signer1", "signer2", "community", "validator"]:
        if grantee == granter:
            continue
        res = cli.authz_grants(cli.address(granter), cli.address(grantee))
        if "grants" not in res or len(res["grants"]) == 0:
            continue
        pc = get_precompile_contract(w3, "ICS20I")
        revoke_tx = pc.functions.revoke(ADDRS[grantee]).build_transaction(
            {
                "from": ADDRS[granter],
                "gasPrice": w3.eth.gas_price,
                "gas": 200_000,
            }
        )
        send_transaction(w3, revoke_tx, KEYS[granter])
