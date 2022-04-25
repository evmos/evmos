<!--
order: 5
-->

# Hooks

The fees module implements one transaction hook, from the `x/evm` module.

## EVM Hook

An [EVM hook](https://evmos.dev/modules/evm/06_hooks.html) executes custom logic after each successful EVM transaction.

All fees paid by a user for transaction execution are sent to the `FeeCollector` Module Account during the `AnteHandler` execution.

If the `x/fees` module is disabled or the EVM transaction targets an unregistered contract, the EVM hook returns `nil`, without performing any actions. In this case, 100% of the transaction fees remain in the `FeeCollector` module, to be distributed to the block proposer.

If the `x/fees` module is enabled, the EVM hook sends a percentage of the fees paid by the user for a transaction to a registered contract, to the withdraw address set for that contract, or to the contract deployer.

1. User submits an EVM transaction to a smart contract that has been registered to receive fees and the transaction is finished successfully.
2. The EVM hook’s `PostTxProcessing` method is called on the fees module. It is passed the initial transaction message, that includes the gas price paid by the user and the transaction receipt, which includes the gas used by the transaction. The hook calculates

    ```go
    devFees := receipt.GasUsed * msg.GasPrice * params.DeveloperShares
    ```

    and sends these dev fees from the `FeeCollector` (Cosmos SDK `auth` module account) to the registered withdraw address for that contract. The remaining amount in the `FeeCollector` is allocated towards the [SDK  Distribution Scheme](https://docs.cosmos.network/main/modules/distribution/03_begin_block.html#the-distribution-scheme). If there is no withdraw address, fees are sent to contract deployer's address.
