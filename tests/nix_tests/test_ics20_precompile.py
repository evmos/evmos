import json

import pytest

from .ibc_utils import EVMOS_IBC_DENOM, assert_ready, get_balance, prepare_network
from .network import Evmos
from .utils import (
    ACCOUNTS,
    ADDRS,
    CONTRACTS,
    KEYS,
    debug_trace_tx,
    decode_bech32,
    deploy_contract,
    eth_to_bech32,
    get_precompile_contract,
    send_transaction,
    wait_for_fn,
    wait_for_new_blocks,
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
