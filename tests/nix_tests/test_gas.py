from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    deploy_contract,
    send_transaction,
    w3_wait_for_new_blocks,
)


def test_gas_eth_tx(geth, evmos_cluster):
    tx_value = 10

    # send a transaction with geth
    geth_gas_price = geth.w3.eth.gas_price
    tx = {"to": ADDRS["community"], "value": tx_value, "gasPrice": geth_gas_price}
    geth_receipt = send_transaction(geth.w3, tx, KEYS["validator"])

    # send an equivalent transaction with evmos
    evmos_gas_price = evmos_cluster.w3.eth.gas_price
    tx = {"to": ADDRS["community"], "value": tx_value, "gasPrice": evmos_gas_price}
    evmos_receipt = send_transaction(evmos_cluster.w3, tx, KEYS["validator"])

    assert geth_receipt.gasUsed == evmos_receipt.gasUsed


def test_gas_deployment(geth, evmos_cluster):
    # deploy an identical contract on geth and evmos
    # ensure that the gasUsed is equivalent
    _, geth_contract_receipt = deploy_contract(geth.w3, CONTRACTS["TestERC20A"])
    _, evmos_contract_receipt = deploy_contract(
        evmos_cluster.w3, CONTRACTS["TestERC20A"]
    )
    assert geth_contract_receipt.gasUsed == evmos_contract_receipt.gasUsed


def test_gas_call(geth, evmos_cluster):
    function_input = 10

    # deploy an identical contract on geth and evmos
    # ensure that the contract has a function which consumes non-trivial gas
    geth_contract, _ = deploy_contract(geth.w3, CONTRACTS["BurnGas"])
    evmos_contract, _ = deploy_contract(evmos_cluster.w3, CONTRACTS["BurnGas"])

    # call the contract and get tx receipt for geth
    geth_gas_price = geth.w3.eth.gas_price
    geth_txhash = geth_contract.functions.burnGas(function_input).transact(
        {"from": ADDRS["validator"], "gasPrice": geth_gas_price}
    )
    geth_call_receipt = geth.w3.eth.wait_for_transaction_receipt(geth_txhash)

    # repeat the above for evmos
    evmos_gas_price = evmos_cluster.w3.eth.gas_price
    evmos_txhash = evmos_contract.functions.burnGas(function_input).transact(
        {"from": ADDRS["validator"], "gasPrice": evmos_gas_price}
    )
    evmos_call_receipt = evmos_cluster.w3.eth.wait_for_transaction_receipt(evmos_txhash)

    # ensure that the gasUsed is equivalent
    assert geth_call_receipt.gasUsed == evmos_call_receipt.gasUsed


def test_block_gas_limit(evmos_cluster):
    tx_value = 10

    # get the block gas limit from the latest block
    w3_wait_for_new_blocks(evmos_cluster.w3, 5)
    block = evmos_cluster.w3.eth.get_block("latest")
    exceeded_gas_limit = block.gasLimit + 100

    # send a transaction exceeding the block gas limit
    evmos_gas_price = evmos_cluster.w3.eth.gas_price
    tx = {
        "to": ADDRS["community"],
        "value": tx_value,
        "gas": exceeded_gas_limit,
        "gasPrice": evmos_gas_price,
    }

    # expect an error due to the block gas limit
    try:
        send_transaction(evmos_cluster.w3, tx, KEYS["validator"])
    except Exception as error:
        assert "exceeds block gas limit" in error.args[0]["message"]

    # deploy a contract on evmos
    evmos_contract, _ = deploy_contract(evmos_cluster.w3, CONTRACTS["BurnGas"])

    # expect an error on contract call due to block gas limit
    try:
        evmos_txhash = evmos_contract.functions.burnGas(exceeded_gas_limit).transact(
            {
                "from": ADDRS["validator"],
                "gas": exceeded_gas_limit,
                "gasPrice": evmos_gas_price,
            }
        )
        (evmos_cluster.w3.eth.wait_for_transaction_receipt(evmos_txhash))
    except Exception as error:
        assert "exceeds block gas limit" in error.args[0]["message"]

    return
