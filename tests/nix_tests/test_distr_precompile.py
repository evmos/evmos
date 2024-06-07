import json
import pytest

from .ibc_utils import get_balance
from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    debug_trace_tx,
    deploy_contract,
    send_transaction,
    wait_for_fn,
)


# fund from EOA (msg.sender)
# fund from contract (deposit first)
# fund from another EOA
@pytest.mark.parametrize(
    "name, deposit_amt, args, err_contains",
    [
        (
            "fund from another EOA",
            None,
            [ADDRS["signer2"], int(1e18)],
            "does not match the spender address",
        ),
        (
            "fund from EOA (origin)",
            None,
            [ADDRS["signer1"], int(1e18)],
            None,
        ),
        (
            "fund from contract (deposit first)",
            int(1e18),
            ["CONTRACT_ADDR", int(1e18)],
            None,
        ),
    ],
)
def test_fund_community_pool(evmos_cluster, name, deposit_amt, args, err_contains):
    """
    Test the fundCommunityPool function of the distribution
    precompile calling it from another precompile
    """
    denom = "aevmos"
    gas_limit = 200_000
    gas_price = evmos_cluster.w3.eth.gas_price

    # Deployment of distr precompile caller and initial checks
    eth_contract, tx_receipt = deploy_contract(
        evmos_cluster.w3, CONTRACTS["DistributionCaller"]
    )
    assert tx_receipt.status == 1

    counter = eth_contract.functions.counter().call()
    assert counter == 0

    if deposit_amt is not None:
        # deposit some funds to the contract
        deposit_tx = eth_contract.functions.deposit().build_transaction(
            {
                "from": ADDRS["signer1"],
                "value": deposit_amt,
                "gas": gas_limit,
                "gasPrice": gas_price,
            }
        )
        deposit_receipt = send_transaction(
            evmos_cluster.w3, deposit_tx, KEYS["signer1"]
        )
        assert deposit_receipt.status == 1, f"Failed: {name}"

        def check_contract_balance():
            new_contract_balance = evmos_cluster.w3.eth.get_balance(
                eth_contract.address
            )
            return new_contract_balance > 0

        wait_for_fn("contract balance change", check_contract_balance)

    signer1_prev_balance = get_balance(
        evmos_cluster, evmos_cluster.cosmos_cli().address("signer1"), denom
    )
    signer2_prev_balance = get_balance(
        evmos_cluster, evmos_cluster.cosmos_cli().address("signer2"), denom
    )

    community_prev_balance = evmos_cluster.cosmos_cli().distribution_community()

    # call the smart contract to fund the community pool
    if args[0] == "CONTRACT_ADDR":
        args[0] = eth_contract.address
    send_tx = eth_contract.functions.testFundCommunityPool(*args).build_transaction(
        {
            "from": ADDRS["signer1"],
            "gasPrice": gas_price,
            "gas": gas_limit,
        }
    )
    receipt = send_transaction(evmos_cluster.w3, send_tx, KEYS["signer1"])

    if err_contains is not None:
        assert receipt.status == 0
        trace = debug_trace_tx(evmos_cluster, receipt.transactionHash.hex())
        # stringify the tx trace to look for the expected error message
        trace_str = json.dumps(trace, separators=(",", ":"))
        assert err_contains in trace_str, f"Failed: {name}"
        return

    assert receipt.status == 1, f"Failed: {name}"
    fees = receipt.gasUsed * gas_price

    # check that contract's balance is 0
    final_contract_balance = evmos_cluster.w3.eth.get_balance(eth_contract.address)
    assert final_contract_balance == 0, f"Failed: {name}"

    # check counter of contract
    counter_after = eth_contract.functions.counter().call()
    assert counter_after == 0, f"Failed: {name}"

    # check that community pool balance increased
    funds_sent_amt = args[1]
    community_final_balance = evmos_cluster.cosmos_cli().distribution_community()
    assert community_final_balance >= community_prev_balance + funds_sent_amt, f"Failed: {name}"

    # signer2 balance should remain unchanged
    signer2_final_balance = get_balance(
        evmos_cluster, evmos_cluster.cosmos_cli().address("signer2"), denom
    )

    assert signer2_final_balance == signer2_prev_balance, f"Failed: {name}"

    signer1_final_balance = get_balance(
        evmos_cluster, evmos_cluster.cosmos_cli().address("signer1"), denom
    )
    # if there was a deposit in the contract
    # the community pool was funded by the contract
    # and the tx sender (signer1) only paid the fees
    if deposit_amt is not None:
        assert signer1_final_balance == signer1_prev_balance - fees, f"Failed: {name}"
        return

    # signer1 account sent the funds to the community pool
    assert (
        signer1_final_balance == signer1_prev_balance - funds_sent_amt - fees
    ), f"Failed: {name}"
