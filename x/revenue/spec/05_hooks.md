<!--
order: 5
-->

# Hooks

The fees module implements one transaction hook from the `x/evm` module in order to distribute fees between developers and validators.

## EVM Hook

A [`PostTxProcessing` EVM hook](https://evoblock.dev/modules/evm/06_hooks.html) executes custom logic after each successful EVM transaction. All fees paid by a user for transaction execution are sent to the `FeeCollector` module account during the `AnteHandler` execution before being distributed to developers and validators.

If the `x/revenue` module is disabled or the EVM transaction targets an unregistered contract, the EVM hook returns `nil`, without performing any actions. In this case, 100% of the transaction fees remain in the `FeeCollector` module, to be distributed to the block proposer.

If the `x/revenue` module is enabled and a EVM transaction targets a registered contract, the EVM hook sends a percentage of the transaction fees (paid by the user) to the withdraw address set for that contract, or to the contract deployer.

1. User submits EVM transaction (`MsgEthereumTx`) to a smart contract and transaction is executed successfully
2. Check if
   * fees module is enabled
   * smart contract is registered to receive fees
3. Calculate developer fees according to the `DeveloperShares` parameter. The initial transaction message includes the gas price paid by the user and the transaction receipt, which includes the gas used by the transaction.

   ```go
    devFees := receipt.GasUsed * msg.GasPrice * params.DeveloperShares
    ```

4. Transfer developer fee from the `FeeCollector` (Cosmos SDK `auth` module account) to the registered withdraw address for that contract. If there is no withdraw address, fees are sent to contract deployer's address.
5. Distribute the remaining amount in the `FeeCollector` to validators according to the [SDK  Distribution Scheme](https://docs.cosmos.network/main/modules/distribution/03_begin_block.html#the-distribution-scheme).
