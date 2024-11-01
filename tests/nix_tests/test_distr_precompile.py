import json

import pytest

from .ibc_utils import get_balances
from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    amount_of,
    amount_of_dec_coin,
    debug_trace_tx,
    deploy_contract,
    eth_to_bech32,
    get_fee,
    send_transaction,
    wait_for_fn,
    wait_for_new_blocks,
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
    gas_limit = 200_000
    gas_price = evmos_cluster.w3.eth.gas_price
    comm_pool_denom = "aevmos"
    cli = evmos_cluster.cosmos_cli()
    fee_denom = cli.evm_denom()

    # Deployment of distr precompile caller and initial checks
    eth_contract, tx_receipt = deploy_contract(
        evmos_cluster.w3, CONTRACTS["DistributionCaller"]
    )
    assert tx_receipt.status == 1

    counter = eth_contract.functions.counter().call()
    assert counter == 0

    if deposit_amt is not None:
        # deposit some funds to the contract.
        # In case evm_denom != 'aevmos', we'll need
        # to fund the contract via a cosmos tx,
        # because the precompile only funds the community pool with 'aevmos'
        if fee_denom == comm_pool_denom:
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
        else:
            # fund contract via cosmos tx
            sender = eth_to_bech32(ADDRS["signer1"])
            tx = cli.transfer(
                sender,
                eth_to_bech32(eth_contract.address),
                f"{deposit_amt}{comm_pool_denom}",
                generate_only=True,
            )
            tx = cli.sign_tx_json(tx, sender, max_priority_price=0)
            rsp = cli.broadcast_tx_json(tx, broadcast_mode="sync")
            assert rsp["code"] == 0, rsp["raw_log"]
            txhash = rsp["txhash"]
            wait_for_new_blocks(cli, 2)
            receipt = cli.tx_search_rpc(f"tx.hash='{txhash}'")[0]
            assert receipt["tx_result"]["code"] == 0, debug_trace_tx(
                evmos_cluster, txhash
            )

    signer1_prev_balances = get_balances(
        evmos_cluster, evmos_cluster.cosmos_cli().address("signer1")
    )
    signer2_prev_balances = get_balances(
        evmos_cluster, evmos_cluster.cosmos_cli().address("signer2")
    )

    # comm pool can have other coins balance (e.g. when evm_denom != 'aevmos')
    # For the precompiles, we only care about the 'aevmos' balance
    # because is the denomination used on the precompile tx
    pool_balances = evmos_cluster.cosmos_cli().distribution_community()
    community_prev_balance = amount_of_dec_coin(pool_balances, comm_pool_denom)

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

    assert receipt.status == 1, debug_trace_tx(
        evmos_cluster, receipt.transactionHash.hex()
    )

    # check that contract's balance is 0
    final_contract_balance = evmos_cluster.w3.eth.get_balance(eth_contract.address)
    assert final_contract_balance == 0, f"Failed: {name}"

    # check counter of contract
    counter_after = eth_contract.functions.counter().call()
    assert counter_after == 0, f"Failed: {name}"

    # check that community pool balance increased
    funds_sent_amt = args[1]

    # sent amount and fees are within the EVM 18 decimals
    # If the evm denom has 6 decimals, we need to scale this
    # when comparing with cosmos balances.
    # Check if evm has 6 dec,
    # actual fees will have 6 dec
    # instead of 18
    fees = get_fee(evmos_cluster.cosmos_cli(), gas_price, gas_limit, receipt.gasUsed)

    final_pool_balances = evmos_cluster.cosmos_cli().distribution_community()
    community_final_balance = amount_of_dec_coin(final_pool_balances, comm_pool_denom)
    assert (
        community_final_balance >= community_prev_balance + funds_sent_amt
    ), f"Failed: {name}"

    # signer2 balance should remain unchanged
    signer2_final_balances = get_balances(
        evmos_cluster, evmos_cluster.cosmos_cli().address("signer2")
    )

    assert signer2_final_balances == signer2_prev_balances, f"Failed: {name}"

    signer1_final_balances = get_balances(
        evmos_cluster, evmos_cluster.cosmos_cli().address("signer1")
    )
    # if there was a deposit in the contract
    # the community pool was funded by the contract
    # and the tx sender (signer1) only paid the fees
    fee_denom_amt_initial = amount_of(signer1_prev_balances, fee_denom)
    fee_denom_amt_final = amount_of(signer1_final_balances, fee_denom)
    if deposit_amt is not None:
        assert fee_denom_amt_final == fee_denom_amt_initial - fees, f"Failed: {name}"
        return

    sent_denom_amt_initial = amount_of(signer1_prev_balances, comm_pool_denom)
    sent_denom_amt_final = amount_of(signer1_final_balances, comm_pool_denom)
    if comm_pool_denom != fee_denom:
        assert fee_denom_amt_final == fee_denom_amt_initial - fees, f"Failed: {name}"
        assert sent_denom_amt_final == sent_denom_amt_initial - funds_sent_amt
        return

    # signer1 account sent the funds to the community pool
    assert (
        sent_denom_amt_final == sent_denom_amt_initial - funds_sent_amt - fees
    ), f"Failed: {name}"
