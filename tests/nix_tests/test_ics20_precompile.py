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
    ACCOUNTS,
    ADDRS,
    CONTRACTS,
    KEYS,
    MAX_UINT256,
    check_error,
    debug_trace_tx,
    decode_bech32,
    deploy_contract,
    eth_to_bech32,
    get_precompile_contract,
    send_transaction,
    wait_for_fn,
    wait_for_new_blocks,
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
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
            ],
            [],
            "improper number of arguments",
            None,
        ),
        (
            "authorization does not exist - returns empty array",
            [
                ADDRS["community"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
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
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
            ],
            [
                ADDRS["community"],
                ADDRS["validator"],
            ],
            None,
            [("transfer", "channel-0", [("aevmos", 1000000000000000000)], [], [])],
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
                [["transfer", "channel-1", [["aevmos", int(1e18)]], [], []]],
            ],
            True,
            "channel not found",
            0,
        ),
        (
            "MaxInt256 allocation",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", MAX_UINT256]], [], []]],
            ],
            False,
            "",
            MAX_UINT256,
        ),
        (
            "create authorization with specific spend limit",
            [
                ADDRS["signer2"],
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
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
        [["transfer", "channel-0", [["aevmos", MAX_UINT256]], [], []]],
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
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
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
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
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
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
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
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
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
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
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
                """[{"denom": "aevmos", "amount": "1000000000000000000"},"""
                """{"denom": "uatom", "amount": "2000000000000000000"}]"""
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
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
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
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
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
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
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
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
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
                [["transfer", "channel-0", [["aevmos", int(1e18)]], [], []]],
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
                """[{"denom": "aevmos", "amount": "1000000000000000000"},"""
                """{"denom": "uatom", "amount": "500000000000000000"}]"""
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
            eth_contract.address, [["transfer", "channel-0", auth_coins, [], []]]
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
        eth_contract.address, [["transfer", "channel-0", [[src_denom, amt]], [], []]]
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


@pytest.mark.parametrize(
    "name, src_addr, other_addr, transfer_before, transfer_after",
    [
        (
            "IBC transfer with internal transfer before and after the precompile call",
            ADDRS["signer2"],
            None,
            True,
            True,
        ),
        (
            "IBC transfer with internal transfer before the precompile call",
            ADDRS["signer2"],
            None,
            True,
            False,
        ),
        (
            "IBC transfer with internal transfer after the precompile call",
            ADDRS["signer2"],
            None,
            False,
            True,
        ),
        (
            "IBC transfer with internal transfer to the escrow addr "
            + "before and after the precompile call",
            ADDRS["signer2"],
            "ESCROW_ADDR",
            True,
            True,
        ),
        (
            "IBC transfer with internal transfer to the escrow addr before the precompile call",
            ADDRS["signer2"],
            "ESCROW_ADDR",
            True,
            False,
        ),
        (
            "IBC transfer with internal transfer to the escrow addr after the precompile call",
            ADDRS["signer2"],
            "ESCROW_ADDR",
            False,
            True,
        ),
    ],
)
def test_ibc_transfer_from_eoa_with_internal_transfer(
    ibc, name, src_addr, other_addr, transfer_before, transfer_after
):
    """
    Test ibc transfer from contract
    with an internal transfer within the contract call
    """
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    w3 = evmos.w3

    channel = "channel-0"
    amt = 1000000000000000000
    src_denom = "aevmos"
    gas_limit = 200_000
    evmos_gas_price = w3.eth.gas_price

    dst_addr = ibc.chains["chainmain"].cosmos_cli().address("signer2")
    # address that will escrow the aevmos tokens when transferring via IBC
    escrow_bech32 = evmos.cosmos_cli().escrow_address(channel)

    if other_addr is None:
        other_addr = src_addr
    elif other_addr == "ESCROW_ADDR":
        other_addr = decode_bech32(escrow_bech32)

    # Deployment of contracts and initial checks
    initial_contract_balance = 100
    eth_contract = setup_interchain_sender_contract(
        evmos, ACCOUNTS["signer2"], amt, initial_contract_balance
    )

    # get starting balances to check after the tx
    src_bech32 = evmos.cosmos_cli().address("signer2")
    contract_bech32 = eth_to_bech32(eth_contract.address)
    other_addr_bech32 = eth_to_bech32(other_addr)

    src_starting_balance = get_balance(ibc.chains["evmos"], src_bech32, "aevmos")
    dest_starting_balance = get_balance(
        ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM
    )
    other_addr_initial_balance = get_balance(
        ibc.chains["evmos"], other_addr_bech32, src_denom
    )
    escrow_initial_balance = get_balance(ibc.chains["evmos"], escrow_bech32, src_denom)

    # Calling the actual transfer function on the custom contract
    send_tx = eth_contract.functions.testTransferFundsWithTransferToOtherAcc(
        other_addr,
        src_addr,
        "transfer",
        channel,
        src_denom,
        amt,
        dst_addr,
        transfer_before,
        transfer_after,
    ).build_transaction(
        {"from": ADDRS["signer2"], "gasPrice": evmos_gas_price, "gas": gas_limit}
    )
    receipt = send_transaction(ibc.chains["evmos"].w3, send_tx, KEYS["signer2"])
    assert receipt.status == 1, f"Failed {name}"
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

    # estimate the expected amt transferred and counter state
    # based on the test case
    exp_amt_tranferred_internally = (
        30
        if transfer_after and transfer_before
        else 0 if not transfer_after and not transfer_before else 15
    )
    exp_counter_after = (
        2
        if transfer_after and transfer_before
        else 0 if not transfer_after and not transfer_before else 1
    )

    exp_src_final_bal = src_starting_balance - amt - fees
    exp_escrow_final_bal = escrow_initial_balance + amt
    if other_addr == src_addr:
        exp_src_final_bal += exp_amt_tranferred_internally
    elif other_addr == decode_bech32(escrow_bech32):
        other_addr_final_balance = get_balance(
            ibc.chains["evmos"], escrow_bech32, src_denom
        )
        # check the escrow account escrowed the coins successfully
        # and received the transferred tokens during the contract call
        exp_escrow_final_bal += exp_amt_tranferred_internally
    else:
        other_addr_final_balance = get_balance(
            ibc.chains["evmos"], other_addr_bech32, src_denom
        )
        assert (
            other_addr_final_balance
            == other_addr_initial_balance + exp_amt_tranferred_internally
        )

    # check final balance for source address (tx signer)
    src_final_amount_evmos = get_balance(ibc.chains["evmos"], src_bech32, src_denom)
    assert src_final_amount_evmos == exp_src_final_bal, f"Failed {name}"

    # Check contracts final state (balance & counter)
    contract_final_balance = get_balance(
        ibc.chains["evmos"], contract_bech32, src_denom
    )
    assert (
        contract_final_balance
        == initial_contract_balance - exp_amt_tranferred_internally
    ), f"Failed {name}"

    counter_after = eth_contract.functions.counter().call()
    assert counter_after == exp_counter_after, f"Failed {name}"

    # check escrow account balance is updated properly
    escrow_final_balance = get_balance(ibc.chains["evmos"], escrow_bech32, src_denom)
    assert escrow_final_balance == exp_escrow_final_bal


@pytest.mark.parametrize(
    "name, src_addr, ibc_transfer_amt, transfer_before, "
    + "transfer_between, transfer_after, err_contains",
    [
        (
            "Two IBC transfers with internal transfer before, "
            + "between and after the precompile call",
            ADDRS["signer2"],
            int(1e18),
            True,
            True,
            True,
            None,
        ),
        (
            "Two IBC transfers with internal transfer before and between the precompile call",
            ADDRS["signer2"],
            int(1e18),
            True,
            True,
            False,
            None,
        ),
        (
            "Two IBC transfers with internal transfer between and after the precompile call",
            ADDRS["signer2"],
            int(1e18),
            False,
            True,
            True,
            None,
        ),
        (
            "Two IBC transfers with internal transfer before and after the precompile call",
            ADDRS["signer2"],
            int(1e18),
            True,
            False,
            True,
            None,
        ),
        (
            "Two IBC transfers with internal transfer before and between, "
            + "and second IBC transfer fails",
            ADDRS["signer2"],
            int(35000e18),
            True,
            True,
            False,
            "insufficient funds",
        ),
    ],
)
def test_ibc_multi_transfer_from_eoa_with_internal_transfer(
    ibc,
    name,
    src_addr,
    ibc_transfer_amt,
    transfer_before,
    transfer_between,
    transfer_after,
    err_contains,
):
    """
    Test ibc transfer from contract
    with an internal transfer within the contract call
    """
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    w3 = evmos.w3

    src_denom = "aevmos"
    gas_limit = 800_000
    evmos_gas_price = w3.eth.gas_price

    dst_addr = ibc.chains["chainmain"].cosmos_cli().address("signer2")
    escrow_bech32 = evmos.cosmos_cli().escrow_address("channel-0")

    # Deployment of contracts and initial checks
    initial_contract_balance = 100
    eth_contract = setup_interchain_sender_contract(
        evmos, ACCOUNTS["signer2"], ibc_transfer_amt, initial_contract_balance
    )

    # send some funds (100 aevmos) to the contract to perform
    # internal transfer within the tx
    src_bech32 = eth_to_bech32(src_addr)
    contract_bech32 = eth_to_bech32(eth_contract.address)

    # get starting balances to check after the tx
    src_starting_balance = get_balance(evmos, src_bech32, "aevmos")
    dest_starting_balance = get_balance(
        ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM
    )
    escrow_initial_balance = get_balance(evmos, escrow_bech32, src_denom)

    # Calling the actual transfer function on the custom contract
    send_tx = eth_contract.functions.testMultiTransferWithInternalTransfer(
        src_addr,
        "transfer",
        "channel-0",
        src_denom,
        ibc_transfer_amt,
        dst_addr,
        transfer_before,
        transfer_between,
        transfer_after,
    ).build_transaction(
        {"from": ADDRS["signer2"], "gasPrice": evmos_gas_price, "gas": gas_limit}
    )
    receipt = send_transaction(ibc.chains["evmos"].w3, send_tx, KEYS["signer2"])

    escrow_final_balance = get_balance(ibc.chains["evmos"], escrow_bech32, src_denom)

    if err_contains is not None:
        assert receipt.status == 0
        # check error msg
        # get the corresponding error message from the trace
        trace = debug_trace_tx(ibc.chains["evmos"], receipt.transactionHash.hex())
        # stringify the tx trace to look for the expected error message
        trace_str = json.dumps(trace, separators=(",", ":"))
        assert err_contains in trace_str

        # check balances where reverted accordingly
        # Check contracts final state (balance & counter)
        contract_final_balance = get_balance(
            ibc.chains["evmos"], contract_bech32, src_denom
        )
        assert contract_final_balance == initial_contract_balance, f"Failed {name}"
        counter_after = eth_contract.functions.counter().call()
        assert counter_after == 0, f"Failed {name}"

        # check sender balance was decreased only by fees paid
        src_final_amount_evmos = get_balance(ibc.chains["evmos"], src_bech32, src_denom)
        fees = receipt.gasUsed * evmos_gas_price

        exp_src_final_bal = src_starting_balance - fees
        assert src_final_amount_evmos == exp_src_final_bal, f"Failed {name}"

        # check balance on destination chain
        # wait a couple of blocks to check on the other chain
        wait_for_new_blocks(ibc.chains["chainmain"].cosmos_cli(), 5)
        dest_final_balance = get_balance(
            ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM
        )
        assert dest_final_balance == dest_starting_balance

        # escrow account balance should be same as initial balance
        assert escrow_final_balance == escrow_initial_balance

        return

    assert receipt.status == 1, debug_trace_tx(
        ibc.chains["evmos"], receipt.transactionHash.hex()
    )
    fees = receipt.gasUsed * evmos_gas_price

    final_dest_balance = dest_starting_balance

    def check_dest_balance():
        nonlocal final_dest_balance
        final_dest_balance = get_balance(
            ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM
        )
        return final_dest_balance > dest_starting_balance

    wait_for_fn("destination balance change", check_dest_balance)
    assert final_dest_balance == dest_starting_balance + ibc_transfer_amt

    # check the balance was escrowed
    assert escrow_final_balance == escrow_initial_balance + ibc_transfer_amt

    # estimate the expected amt transferred and counter state
    # based on the test case
    exp_amt_tranferred_internally = 0
    exp_counter_after = 0
    for transferred in [transfer_after, transfer_between, transfer_before]:
        if transferred:
            exp_amt_tranferred_internally += 15
            exp_counter_after += 1

    # check final balance for source address (tx signer)
    src_final_amount_evmos = get_balance(ibc.chains["evmos"], src_bech32, src_denom)
    exp_src_final_bal = (
        src_starting_balance - ibc_transfer_amt - fees + exp_amt_tranferred_internally
    )
    assert src_final_amount_evmos == exp_src_final_bal, f"Failed {name}"

    # Check contracts final state (balance & counter)
    contract_final_balance = get_balance(
        ibc.chains["evmos"], contract_bech32, src_denom
    )
    assert (
        contract_final_balance
        == initial_contract_balance - exp_amt_tranferred_internally
    ), f"Failed {name}"

    counter_after = eth_contract.functions.counter().call()
    assert counter_after == exp_counter_after, f"Failed {name}"


def test_multi_ibc_transfers_with_revert(ibc):
    """
    Test ibc transfer from contract
    with an internal transfer within the contract call
    """
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    w3 = evmos.w3

    src_denom = "aevmos"
    gas_limit = 800_000
    ibc_transfer_amt = int(1e18)
    src_addr = ADDRS["signer2"]
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price

    dst_addr = ibc.chains["chainmain"].cosmos_cli().address("signer2")
    escrow_bech32 = ibc.chains["evmos"].cosmos_cli().escrow_address("channel-0")

    # Deployment of contracts and initial checks
    initial_contract_balance = 100
    interchain_sender_contract = setup_interchain_sender_contract(
        evmos, ACCOUNTS["signer2"], ibc_transfer_amt * 2, initial_contract_balance
    )

    counter = interchain_sender_contract.functions.counter().call()
    assert counter == 0

    caller_contract, tx_receipt = deploy_contract(
        w3, CONTRACTS["InterchainSenderCaller"], [interchain_sender_contract.address]
    )
    assert tx_receipt.status == 1

    counter = caller_contract.functions.counter().call()
    assert counter == 0

    src_bech32 = eth_to_bech32(src_addr)
    interchain_sender_contract_bech32 = eth_to_bech32(
        interchain_sender_contract.address
    )

    # get starting balances to check after the tx
    src_starting_balance = get_balance(ibc.chains["evmos"], src_bech32, "aevmos")
    dest_starting_balance = get_balance(
        ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM
    )
    escrow_initial_balance = get_balance(ibc.chains["evmos"], escrow_bech32, src_denom)

    # Calling the actual transfer function on the custom contract
    send_tx = caller_contract.functions.transfersWithRevert(
        src_addr,
        "transfer",
        "channel-0",
        src_denom,
        ibc_transfer_amt,
        dst_addr,
    ).build_transaction(
        {"from": ADDRS["signer2"], "gasPrice": evmos_gas_price, "gas": gas_limit}
    )
    receipt = send_transaction(ibc.chains["evmos"].w3, send_tx, KEYS["signer2"])

    escrow_final_balance = get_balance(ibc.chains["evmos"], escrow_bech32, src_denom)

    assert receipt.status == 1, debug_trace_tx(
        ibc.chains["evmos"], receipt.transactionHash.hex()
    )
    fees = receipt.gasUsed * evmos_gas_price

    final_dest_balance = dest_starting_balance

    def check_dest_balance():
        nonlocal final_dest_balance
        final_dest_balance = get_balance(
            ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM
        )
        return final_dest_balance > dest_starting_balance

    wait_for_fn("destination balance change", check_dest_balance)
    assert final_dest_balance == dest_starting_balance + ibc_transfer_amt

    # check the balance was escrowed
    # note that 4 ibc-transfers where included in the tx for a total
    # of 2 x ibc_transfer_amt, but 2 of these transfers where reverted
    assert escrow_final_balance == escrow_initial_balance + ibc_transfer_amt

    # estimate the expected amt transferred and counter state
    # based on the test case
    exp_amt_tranferred_internally = 45
    exp_interchain_sender_counter_after = 3

    # check final balance for source address (tx signer)
    src_final_amount_evmos = get_balance(ibc.chains["evmos"], src_bech32, src_denom)
    exp_src_final_bal = (
        src_starting_balance - fees - ibc_transfer_amt + exp_amt_tranferred_internally
    )
    assert src_final_amount_evmos == exp_src_final_bal

    # Check contracts final state (balance & counter)
    contract_final_balance = get_balance(
        ibc.chains["evmos"], interchain_sender_contract_bech32, src_denom
    )
    assert (
        contract_final_balance
        == initial_contract_balance - exp_amt_tranferred_internally
    )

    counter_after = interchain_sender_contract.functions.counter().call()
    assert counter_after == exp_interchain_sender_counter_after

    counter_after = caller_contract.functions.counter().call()
    assert counter_after == 2


def test_multi_ibc_transfers_with_nested_revert(ibc):
    """
    Test multiple ibc transfer from contract
    with an internal transfer within the contract call
    and a nested revert
    """
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    w3 = evmos.w3

    src_denom = "aevmos"
    gas_limit = 800_000
    ibc_transfer_amt = int(1e18)
    src_addr = ADDRS["signer2"]
    evmos_gas_price = ibc.chains["evmos"].w3.eth.gas_price

    dst_addr = ibc.chains["chainmain"].cosmos_cli().address("signer2")
    escrow_bech32 = ibc.chains["evmos"].cosmos_cli().escrow_address("channel-0")

    # Deployment of contracts and initial checks
    initial_contract_balance = 100
    interchain_sender_contract = setup_interchain_sender_contract(
        evmos, ACCOUNTS["signer2"], ibc_transfer_amt * 2, initial_contract_balance
    )

    counter = interchain_sender_contract.functions.counter().call()
    assert counter == 0

    caller_contract, tx_receipt = deploy_contract(
        w3, CONTRACTS["InterchainSenderCaller"], [interchain_sender_contract.address]
    )
    assert tx_receipt.status == 1

    counter = caller_contract.functions.counter().call()
    assert counter == 0

    src_bech32 = eth_to_bech32(src_addr)
    interchain_sender_contract_bech32 = eth_to_bech32(
        interchain_sender_contract.address
    )

    # get starting balances to check after the tx
    src_starting_balance = get_balance(ibc.chains["evmos"], src_bech32, "aevmos")
    dest_starting_balance = get_balance(
        ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM
    )
    escrow_initial_balance = get_balance(ibc.chains["evmos"], escrow_bech32, src_denom)

    # Calling the actual transfer function on the custom contract
    send_tx = caller_contract.functions.transfersWithNestedRevert(
        src_addr,
        "transfer",
        "channel-0",
        src_denom,
        ibc_transfer_amt,
        dst_addr,
    ).build_transaction(
        {"from": ADDRS["signer2"], "gasPrice": evmos_gas_price, "gas": gas_limit}
    )
    receipt = send_transaction(ibc.chains["evmos"].w3, send_tx, KEYS["signer2"])

    escrow_final_balance = get_balance(ibc.chains["evmos"], escrow_bech32, src_denom)

    # tx should be successfull, but all transfers should be reverted
    # only the contract's state should've changed (counter)
    exp_caller_contract_final_counter = 2
    assert receipt.status == 1, debug_trace_tx(
        ibc.chains["evmos"], receipt.transactionHash.hex()
    )
    fees = receipt.gasUsed * evmos_gas_price

    final_dest_balance = dest_starting_balance

    tries = 0

    def check_dest_balance():
        nonlocal tries
        nonlocal final_dest_balance
        tries += 1
        final_dest_balance = get_balance(
            ibc.chains["chainmain"], dst_addr, EVMOS_IBC_DENOM
        )
        if tries == 7:
            return True
        return final_dest_balance > dest_starting_balance

    wait_for_fn("destination balance change", check_dest_balance)
    assert final_dest_balance == dest_starting_balance

    # check the balance was escrowed
    # note that 4 ibc-transfers where included in the tx for a total
    # of 2 x ibc_transfer_amt, but 2 of these transfers where reverted
    assert escrow_final_balance == escrow_initial_balance

    # estimate the expected amt transferred and counter state
    # based on the test case
    exp_interchain_sender_counter_after = 0

    # check final balance for source address (tx signer)
    src_final_amount_evmos = get_balance(ibc.chains["evmos"], src_bech32, src_denom)
    exp_src_final_bal = src_starting_balance - fees
    assert src_final_amount_evmos == exp_src_final_bal

    # Check contracts final state (balance & counter)
    contract_final_balance = get_balance(
        ibc.chains["evmos"], interchain_sender_contract_bech32, src_denom
    )
    assert contract_final_balance == initial_contract_balance

    counter_after = interchain_sender_contract.functions.counter().call()
    assert counter_after == exp_interchain_sender_counter_after

    counter_after = caller_contract.functions.counter().call()
    assert counter_after == exp_caller_contract_final_counter


def setup_interchain_sender_contract(
    evmos, src_acc, transfer_amt, initial_contract_balance
):
    """
    Helper function to setup the InterchainSender contract
    used in tests. It deploys the contract, creates an authorization
    for the contract to use the provided src_acc's funds,
    and sends some funds to the InterchainSender contract
    for internal transactions within its methods.
    It returns the deployed contract
    """
    src_denom = "aevmos"
    gas_limit = 200_000
    evmos_gas_price = evmos.w3.eth.gas_price
    # Deployment of contracts and initial checks
    eth_contract, tx_receipt = deploy_contract(evmos.w3, CONTRACTS["InterchainSender"])
    assert tx_receipt.status == 1

    counter = eth_contract.functions.counter().call()
    assert counter == 0

    pc = get_precompile_contract(evmos.w3, "ICS20I")
    # Approve the contract to spend the src_denom
    approve_tx = pc.functions.approve(
        eth_contract.address,
        [["transfer", "channel-0", [[src_denom, transfer_amt]], [], []]],
    ).build_transaction(
        {
            "from": src_acc.address,
            "gasPrice": evmos_gas_price,
            "gas": gas_limit,
        }
    )
    tx_receipt = send_transaction(evmos.w3, approve_tx, src_acc.key)
    assert tx_receipt.status == 1

    def check_allowance_set():
        new_allowance = pc.functions.allowance(
            eth_contract.address, src_acc.address
        ).call()
        return new_allowance != []

    wait_for_fn("allowance has changed", check_allowance_set)

    # send some funds (100 aevmos) to the contract to perform
    # internal transfer within the tx
    src_bech32 = eth_to_bech32(src_acc.address)
    contract_bech32 = eth_to_bech32(eth_contract.address)
    fund_tx = evmos.cosmos_cli().transfer(
        src_bech32,
        contract_bech32,
        f"{initial_contract_balance}aevmos",
        gas_prices=f"{evmos_gas_price + 100000}aevmos",
        generate_only=True,
    )

    fund_tx = evmos.cosmos_cli().sign_tx_json(fund_tx, src_bech32, max_priority_price=0)
    rsp = evmos.cosmos_cli().broadcast_tx_json(fund_tx, broadcast_mode="sync")
    assert rsp["code"] == 0, rsp["raw_log"]
    txhash = rsp["txhash"]
    wait_for_new_blocks(evmos.cosmos_cli(), 2)
    receipt = evmos.cosmos_cli().tx_search_rpc(f"tx.hash='{txhash}'")[0]
    assert receipt["tx_result"]["code"] == 0, rsp["raw_log"]

    return eth_contract
