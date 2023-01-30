<!--
order: 5
-->

# AnteHandlers

The `x/vesting` module provides `AnteDecorator`s that are recursively chained together
into a single [`Antehandler`](https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-alpha1/docs/architecture/adr-010-modular-antehandler.md).
These decorators perform basic validity checks on an Ethereum or SDK transaction,
such that it could be thrown out of the transaction Mempool.

Note that the `AnteHandler` is called on both `CheckTx` and `DeliverTx`,
as Tendermint proposers presently have the ability to include in their proposed block transactions that fail `CheckTx`.

## Decorators

The following decorators implement the vesting logic for token delegation and performing EVM transactions.

### `VestingDelegationDecorator`

Validates if a transaction contains a staking delegation of unvested coins. This AnteHandler decorator will fail if:

- the message is not a `MsgDelegate`
- sender account cannot be found
- sender account is not a `ClawbackVestingAccount`
- the bond amount is greater than the coins already vested

### `EthVestingTransactionDecorator`

Validates if a clawback vesting account is permitted to perform Ethereum transactions,
based on if it has its vesting schedule has surpassed the vesting cliff and first lockup period.
This AnteHandler decorator will fail if:

- the message is not a `MsgEthereumTx`
- sender account cannot be found
- sender account is not a `ClawbackVestingAccount`
- block time is before surpassing vesting cliff end (with zero vested coins) AND
- block time is before surpassing all lockup periods (with non-zero locked coins)
