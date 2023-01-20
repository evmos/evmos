<!--
order: 3
-->

# Gas and Fees

Learn about the differences between `Gas` and `Fees` in Ethereum and Cosmos. {synopsis}

## Pre-requisite Readings

- [Cosmos SDK Gas](https://docs.cosmos.network/main/basics/gas-fees.html) {prereq}
- [Ethereum Gas](https://ethereum.org/en/developers/docs/gas/) {prereq}

## Basics

### Why do Transactions Need Fees?

If anyone can submit transactions to a network at no cost, the network can be overrun by a handful of actors sending large numbers of fraudulent transactions to clog up the network and stop it from working.

The solution to this is a concept called “gas," which is a resource consumed throughout transaction execution. In practice, a small amount of gas is spent on each step of code execution, thus effectively charging for use of a validator’s resources and preventing malicious actors from halting a network at will.

### What is Gas?

In general, gas is a unit that measures the computational intensity of a particular transaction—in other words, how much work would be required to evaluate and perform the job. Complex, multi-step transactions, such as a Cosmos transaction that delegates to a dozen validators, require more gas than simple, single-step transactions, such as a Cosmos transaction to send tokens to another address.

When referring to a transaction, “gas” refers to the total quantity of gas required for the transaction. For example, a transaction may require 300,000 units of gas to be executed.

Gas can be thought of as electricity (kWh) within a house or factory, or fuel for automobiles. The idea is that it costs something to get somewhere.

More on Gas:

- [Cosmos Gas Fees](https://docs.cosmos.network/main/basics/gas-fees)
- [Cosmos Tx Lifecycle](https://docs.cosmos.network/main/basics/tx-lifecycle.html)
- [Ethereum Gas](https://ethereum.org/en/developers/docs/gas/)

### How is Gas Calculated?

In general, there’s no way to know exactly how much gas a transaction will cost without simply running it. Using the Cosmos SDK, this can be done by [simulating the Tx](https://docs.cosmos.network/main/run-node/txs#simulating-a-transaction). Otherwise, there are ways to estimate the amount of gas a transaction will require, based on the details of the transaction fields, and data. In the case of the EVM, for example, each bytecode operation has a [corresponding amount of gas](https://ethereum.org/en/developers/docs/evm/opcodes/).

More on Gas Calculations:

- [Estimate Gas](https://docs.ethers.org/v5/api/providers/provider/#Provider-estimateGas)
- [Executing EVM Bytecode](https://ethereum.org/en/developers/docs/evm/opcodes/)
- [Simulate a Cosmos SDK Tx](https://docs.cosmos.network/main/run-node/txs#simulating-a-transaction)

### How does Gas Relate to Fees?

While gas refers to the computational work required for execution, fees refer to the amount of the tokens you actually spend to execute the transaction. They are derived using the following formula:

```markdown
Total Fees = Gas * Gas Price (the price per unit of gas)
```

If “gas” was measured in kWh, the “gas price” would be the rate (in dollars per kWh) determined by your energy provider, and the “fees” would be your bill. Just as with electricity, gas price is liable to fluctuate over a given day, depending on network traffic.

More on Gas vs. Fees:

- [Cosmos Gas and Fees](https://docs.cosmos.network/main/basics/gas-fees)
- [Ethereum Gas and Fees](https://ethereum.org/en/developers/docs/gas/)

### How are Fees Handled on Cosmos?

Gas fees on Cosmos are relatively straightforward. As a user, you specify two fields:

1. A `GasLimit` corresponding to an upper bound on execution gas, defined as `GasWanted`
2. One of `Fees` or `GasPrice`, which will be used to specify or calculate the transaction fees

The node will entirely consume the fees provided, then begin to execute the transaction. If the `GasLimit` is found to be insufficient during execution, the transaction will fail and roll back any changes made, without refunding the fees provided.

Validators for Cosmos SDK-based chains can specify their `min-gas-prices` that they will enforce when selecting transactions to include in blocks. Thus, transactions with insufficient fees will encounter delays or fail outright.

At the beginning of each block, fees from the previous block are [allocated to validators and delegators](https://docs.cosmos.network/main/modules/distribution), after which they can be withdrawn and spent.

### How are Fees Handled on Ethereum?

Fees on Ethereum include multiple implementations that were introduced over time.

Originally, a user would specify a `GasPrice` and `GasLimit` within a transaction—much like a Cosmos SDK transaction. A block proposer would receive the entire gas fee from each transaction in the block, and they would select transactions to include accordingly.

With proposal EIP-1559 and the London Hard fork, gas calculation changed. The `GasPrice` from above has now been split into two separate components: a `BaseFee` and `PriorityFee`. The `BaseFee` is calculated automatically based on the block size and is burned once the block is mined. The `PriorityFee` goes to the proposer and represents a tip, or an incentive for a proposer to include the transaction in a block.

```markdown
Gas Price = Base Fee + Priority Fee
```

Within a transaction, users can specify a `max_fee_per_gas` corresponding to the total `GasPrice` and a `max_priority_fee_per_gas` corresponding to a maximum `PriorityFee`, in addition to specifying the `gas_limit` as before. All surplus gas that was not required for execution is refunded to the user.

More on Ethereum Fees:

- [Gas Calculation Docs](https://ethereum.org/en/developers/docs/gas/)
- [Proposal EIP-1559](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1559.md)

## Implementation

### How are Gas and Fees Handled on Evmos?

Fundamentally, Evmos is a Cosmos SDK chain that enables EVM compatibility as part of a Cosmos SDK module. As a result of this architecture, all EVM transactions are ultimately encoded as Cosmos SDK transactions and update a Cosmos SDK-managed state.

Since all transactions are represented as Cosmos SDK transactions, transaction fees can be treated identically across execution layers. In practice, dealing with fees includes standard Cosmos SDK logic, some Ethereum logic, and custom Evmos logic. For the most part, fees are collected by the `fee_collector` module, then paid out to validators and delegators. A few key distinctions are as follows:

1. Fee Market Module

    In order to support EIP-1559 gas and fee calculation on Evmos’ EVM layer, Evmos tracks the gas supplied for each block and uses that to calculate a base fee for future EVM transactions, thus enabling EVM dynamic fees and transaction prioritization as specified by EIP-1559.

    For EVM transactions, each node bypasses their local `min-gas-prices` configuration, and instead applies EIP-1559 fee logic—the gas price simply must be greater than both the global `min-gas-price` and the block's `BaseFee`, and the surplus is considered a priority tip. This allows validators to compute Ethereum fees without applying Cosmos SDK fee logic.

    Unlike on Ethereum, the `BaseFee` on Evmos is not burned, and instead is distributed to validators and delegators. Furthermore, the `BaseFee` is lower-bounded by the global `min-gas-price` (currently, the global `min-gas-price` parameter is set to zero, although it can be updated via Governance).

2. EVM Gas Refunds

    Evmos refunds a fraction (at least 50% by default) of the unused gas for EVM transactions to approximate the current behavior on Ethereum. [Why not always 100%?](https://github.com/evmos/ethermint/issues/1085)

3. Revenue Module

    Evmos developed the Revenue Module as a way to reward developers for creating useful dApps—any contract that is registered with Evmos’ Revenue Module rewards a fraction of the transaction fee (currently 95%) from each transaction that interacts with the contract to the contract developer. Validators and delegators earn the remaining portion.

### Detailed Timeline

1. Nodes execute the previous block and run the `EndBlock` hook
    * As part of this hook, the FeeMarket (EIP-1559) module tracks the total `TransientGasWanted` from the transactions on this block. This will be used for the next block’s `BaseFee`.
2. Nodes receive transactions for a subsequent block and gossip these transactions to peers
    * These can be sorted and prioritized by the included fee price (using EIP-1559 fee priority mechanics for EVM transactions - [code snippet](https://github.com/evmos/ethermint/blob/57ed355c985d9f3116aba6aabfa2ee0f3f38e966/app/ante/eth.go#L137)), to be included in the next block
3. Nodes run `BeginBlock` for the subsequent block
    * The FeeMarket module calculates the `BaseFee` ([code snippet](https://github.com/evmos/ethermint/blob/89fdd1984826ea524cb9b8feb089a99b6cfe8ace/x/feemarket/keeper/abci.go#L14)) to be applied for this block using the total `GasWanted` from the previous block.
    * The Distribution module [distributes](https://docs.cosmos.network/main/modules/distribution#begin-block) the previous block’s fee rewards to validators and delegators
4. For each valid transaction that will be included in this block, nodes perform the following:
    * They run an `AnteHandler` corresponding to the transaction type. This process:
        1. Performs basic transaction validation
        2. Verifies the fees provided are greater than the global and local minimum validator values *and* greater than the `BaseFee` calculated
        3. (For Ethereum transactions) Preemptively consumes gas for the EVM transaction
        4. Deducts the transaction fees from the user and transfers them to the `fee_collector` module
        5. Increments the `TransientGasWanted` in the current block, to be used to calculate the next block’s `BaseFee`
    * Then, for standard Cosmos Transactions, nodes:
        1. Execute the transaction and update the state
        2. Consume gas for the transaction
    * For Ethereum Transactions, nodes:
        1. Execute the transaction and update the state
        2. Calculate the gas used and compare it to the gas supplied, then refund a designated portion of the surplus
        3. Send a fraction of the fees used as revenue to contract developers as part of the Revenue Module, if the transaction interacted with a registered smart contract
5. Nodes run `EndBlock` for this block and store the block’s `GasWanted`

## Detailed Mechanics

### Cosmos `Gas`

In the Cosmos SDK, gas is tracked in the main `GasMeter` and the `BlockGasMeter`:

- `GasMeter`: keeps track of the gas consumed during executions that lead to state transitions. It is reset on every transaction execution.
- `BlockGasMeter`: keeps track of the gas consumed in a block and enforces that the gas does not go over a predefined limit. This limit is defined in the Tendermint consensus parameters and can be changed via governance parameter change proposals.

Since gas is priced per-byte, the same interaction is more gas-intensive with larger parameter values than smaller (unlike Ethereum's `uint256` values, Cosmos SDK numericals are represented using [Big.Int](https://pkg.go.dev/math/big#Int) types, which are dynamically sized).

More information regarding gas as part of the Cosmos SDK can be found [here](https://docs.cosmos.network/main/basics/gas-fees.html).

### Matching EVM Gas consumption

Evmos is an EVM-compatible chain that supports Ethereum Web3 tooling. For this reason, gas consumption must be equatable with other EVMs, most importantly Ethereum.

The main difference between EVM and Cosmos state transitions, is that the EVM uses a [gas table](https://github.com/ethereum/go-ethereum/blob/master/params/protocol_params.go) for each OPCODE, whereas Cosmos uses a `GasConfig` that charges gas for each CRUD operation by setting a flat and per-byte cost for accessing the database.

+++ https://github.com/cosmos/cosmos-sdk/blob/3fd376bd5659f076a4dc79b644573299fd1ec1bf/store/types/gas.go#L187-L196

In order to match the gas consumed by the EVM, the gas consumption logic from the SDK is ignored, and instead the gas consumed is calculated by subtracting the state transition leftover gas plus refund from the gas limit defined on the message.

To ignore the SDK gas consumption, we reset the transaction `GasMeter` count to 0 and manually set it to the `gasUsed` value computed by the EVM module at the end of the execution.

+++ https://github.com/evmos/ethermint/blob/098da6d0cc0e0c4cefbddf632df1057383973e4a/x/evm/keeper/state_transition.go#L188

### `AnteHandler`

The Cosmos SDK [`AnteHandler`](https://docs.cosmos.network/main/basics/gas-fees.html#antehandler)
performs basic checks prior to transaction execution. These checks are usually signature
verification, transaction field validation, transaction fees, etc.

Regarding gas consumption and fees, the `AnteHandler` checks that the user has enough balance to
cover for the tx cost (amount plus fees) as well as checking that the gas limit defined in the
message is greater or equal than the computed intrinsic gas for the message.

### Gas Refunds

In the EVM, gas can be specified prior to execution. The totality of the gas specified is consumed at the beginning of the execution (during the `AnteHandler` step) and the remaining gas is refunded back to the user if any gas is left over after the execution. Additionally the EVM can also define gas to be refunded back to the user but those will be capped to a fraction of the used gas depending on the fork/version being used.

### Zero-Fee Transactions

In Cosmos, a minimum gas price is not enforced by the `AnteHandler` as the `min-gas-prices` is
checked against the local node/validator. In other words, the minimum fees accepted are determined
by the validators of the network, and each validator can specify a different minimum value for their fees.
This potentially allows end users to submit 0 fee transactions if there is at least one single
validator that is willing to include transactions with `0` gas price in their blocks proposed.

For this same reason, in Evmos it is possible to send transactions with `0` fees for transaction
types other than the ones defined by the `evm` module. EVM module transactions cannot have `0` fees
as gas is required inherently by the EVM. This check is done by the EVM transactions stateless validation
(i.e `ValidateBasic`) function as well as on the custom `AnteHandler` defined by Evmos.

### Gas Estimation

Ethereum provides a JSON-RPC endpoint `eth_estimateGas` to help users set up a correct gas limit in their transactions.

For that reason, a specific query API `EstimateGas` is implemented in Evmos. It will apply the transaction against the current block/state and perform a binary search in order to find the optimal gas value to return to the user (the same transaction will be applied over and over until we find the minimum gas needed before it fails). The reason we need to use a binary search is that the gas required for the
transaction might be higher than the value returned by the EVM after applying the transaction, so we need to try until we find the optimal value.

A cache context will be used during the whole execution to avoid changes be persisted in the state.

+++ https://github.com/evmos/ethermint/blob/098da6d0cc0e0c4cefbddf632df1057383973e4a/x/evm/keeper/grpc_query.go#L100

For Cosmos Tx's, developers can use Cosmos SDK's [transaction simulation](https://docs.cosmos.network/main/run-node/txs#simulating-a-transaction) to create an accurate estimate.

### Cross-Chain Gas and Fees

Let’s say a user transfers tokens from Chain A to Evmos via IBC-transfer and wants to execute an Evmos transaction—however, they don’t have any Evmos tokens to cover fees. The Cosmos SDK introduced `Tips` as a solution to this issue; a user can cover fees using a different token—in this case, tokens from Chain A.

To cover transaction fees using a tip, this user can sign a transaction with a tip and no fees, then send the transaction to a fee relayer. The fee relayer will then cover the fee in the native currency (Evmos in this case), and receive the tip in payment, behaving as an intermediary exchange.

More on Cosmos Tips:

- [Cosmos Tips Docs](https://docs.cosmos.network/main/core/tips)
