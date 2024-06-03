import pytest

from .ibc_utils import EVMOS_IBC_DENOM, assert_ready, get_balance, prepare_network
from .network import Evmos
from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    deploy_contract,
    get_precompile_contract,
    send_transaction,
    wait_for_fn,
)


@pytest.fixture(scope="module", params=["evmos", "evmos-rocksdb"])
def ibc(request, tmp_path_factory):
    """
    Prepares the network.
    """
    name = "ibc-precompile"
    evmos_build = request.param
    path = tmp_path_factory.mktemp(name)
    network = prepare_network(path, name, [evmos_build, "chainmain"])
    yield from network


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
    """Test ibc transfer from contract"""
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
