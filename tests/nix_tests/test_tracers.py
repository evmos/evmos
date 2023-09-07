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
    compare_fields,
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
    compare_fields(
        tx_res["result"],
        EXPECTED_STRUCT_TRACER,
        ["failed", "gas", "returnValue", "structLogs"],
    )

    tx_res = eth_rpc.make_request(
        "debug_traceTransaction", [tx_hash, {"tracer": "callTracer"}]
    )

    fields = ["to", "from", "gas", "gasUsed", "input", "output", "type", "value"]
    compare_fields(tx_res["result"], EXPECTED_CALLTRACERS, fields)

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

    compare_fields(tx_res["result"], EXPECTED_CALLTRACERS, fields)

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
    compare_fields(tx_res["result"], EXPECTED_CONTRACT_CREATE_TRACER, fields)
