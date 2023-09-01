from web3 import Web3

from .expected_constants import (
    EXPECTED_CALLTRACERS,
    EXPECTED_CONTRACT_CREATE_TRACER,
    EXPECTED_STRUCT_TRACER,
)
from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    deploy_contract,
    send_transaction,
    w3_wait_for_new_blocks,
)


def test_tracers(cluster):
    w3: Web3 = cluster.w3
    eth_rpc = w3.provider
    gas_price = w3.eth.gas_price
    tx = {"to": ADDRS["community"], "value": 100, "gasPrice": gas_price}
    tx_hash = send_transaction(w3, tx, KEYS["validator"])["transactionHash"].hex()

    tx_res = eth_rpc.make_request("debug_traceTransaction", [tx_hash])

    # Order of fields is different on evmos and geth
    assert (
        tx_res["result"]["failed"] == EXPECTED_STRUCT_TRACER["failed"]
    ), "Failed field mismatch"
    assert (
        tx_res["result"]["gas"] == EXPECTED_STRUCT_TRACER["gas"]
    ), "Gas field mismatch"
    assert (
        tx_res["result"]["returnValue"] == EXPECTED_STRUCT_TRACER["returnValue"]
    ), "ReturnValue field mismatch"
    assert (
        tx_res["result"]["structLogs"] == EXPECTED_STRUCT_TRACER["structLogs"]
    ), "StructLogs field mismatch"

    tx_res = eth_rpc.make_request(
        "debug_traceTransaction", [tx_hash, {"tracer": "callTracer"}]
    )

    assert tx_res["result"]["to"] == EXPECTED_CALLTRACERS["to"], "To field mismatch"
    assert (
        tx_res["result"]["from"] == EXPECTED_CALLTRACERS["from"]
    ), "From field mismatch"
    assert (
        tx_res["result"]["gasUsed"] == EXPECTED_CALLTRACERS["gasUsed"]
    ), "GasUsed field mismatch"
    assert (
        tx_res["result"]["input"] == EXPECTED_CALLTRACERS["input"]
    ), "Input field mismatch"
    assert (
        tx_res["result"]["output"] == EXPECTED_CALLTRACERS["output"]
    ), "Output field mismatch"
    assert (
        tx_res["result"]["type"] == EXPECTED_CALLTRACERS["type"]
    ), "Type field mismatch"
    assert (
        tx_res["result"]["value"] == EXPECTED_CALLTRACERS["value"]
    ), "Value field mismatch"

    # geth works with this format, while evmos throws a parsing error
    tx_res = eth_rpc.make_request(
        "debug_traceTransaction",
        [tx_hash, {"tracer": "callTracer", "tracerConfig": {"onlyTopCall": True}}],
    )

    if "error" in tx_res and "cannot unmarshal object" in tx_res["error"]["message"]:
        tx_res = eth_rpc.make_request(
            "debug_traceTransaction",
            [tx_hash, {"tracer": "callTracer", "tracerConfig": "{'onlyTopCall':True}"}],
        )

    assert tx_res["result"]["to"] == EXPECTED_CALLTRACERS["to"], "To field mismatch"
    assert (
        tx_res["result"]["from"] == EXPECTED_CALLTRACERS["from"]
    ), "From field mismatch"
    assert tx_res["result"]["gas"] == EXPECTED_CALLTRACERS["gas"], "Gas field mismatch"
    assert (
        tx_res["result"]["gasUsed"] == EXPECTED_CALLTRACERS["gasUsed"]
    ), "GasUsed field mismatch"
    assert (
        tx_res["result"]["input"] == EXPECTED_CALLTRACERS["input"]
    ), "Input field mismatch"
    assert (
        tx_res["result"]["output"] == EXPECTED_CALLTRACERS["output"]
    ), "Output field mismatch"
    assert (
        tx_res["result"]["type"] == EXPECTED_CALLTRACERS["type"]
    ), "Type field mismatch"
    assert (
        tx_res["result"]["value"] == EXPECTED_CALLTRACERS["value"]
    ), "Value field mismatch"

    _, tx = deploy_contract(
        w3,
        CONTRACTS["TestERC20A"],
    )
    tx_hash = tx["transactionHash"].hex()

    w3_wait_for_new_blocks(w3, 1)

    tx_res = eth_rpc.make_request(
        "debug_traceTransaction", [tx_hash, {"tracer": "callTracer"}]
    )

    tx_res["result"]["to"] = EXPECTED_CONTRACT_CREATE_TRACER["to"]
    assert (
        tx_res["result"]["from"] == EXPECTED_CONTRACT_CREATE_TRACER["from"]
    ), "From field mismatch"
    assert tx_res["result"]["gas"] == EXPECTED_CONTRACT_CREATE_TRACER["gas"], "Gas field mismatch"
    assert (
        tx_res["result"]["gasUsed"] == EXPECTED_CONTRACT_CREATE_TRACER["gasUsed"]
    ), "GasUsed field mismatch"
    assert (
        tx_res["result"]["input"] == EXPECTED_CONTRACT_CREATE_TRACER["input"]
    ), "Input field mismatch"
    assert (
        tx_res["result"]["output"] == EXPECTED_CONTRACT_CREATE_TRACER["output"]
    ), "Output field mismatch"
    assert (
        tx_res["result"]["type"] == EXPECTED_CONTRACT_CREATE_TRACER["type"]
    ), "Type field mismatch"
    assert (
        tx_res["result"]["value"] == EXPECTED_CONTRACT_CREATE_TRACER["value"]
    ), "Value field mismatch"
