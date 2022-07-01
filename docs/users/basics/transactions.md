<!--
order: 1
-->

# Transactions

Learn more about transactions on Evmos {synopsis}

::: tip
🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧

This documentation page is currently under work in progress.

🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧
:::

<!-- 
TODO: explain what transactions are on Evmos and blockchains. 
Explain that transactions can be identified by hashes and that they can 
contain multiple messages. Why can transactions fail? 

Explain that transactions can interoperate with other blockchains.
-->

## Transaction Confirmations

<!-- TODO: why are Ethereum transactions different than Cosmos -->

## Transaction Types

<!-- TODO: explain which transactions types does Evmos support (i.e modules and changes) and provide a few examples. -->

<!-- TODO: why are Ethereum transactions different than Cosmos -->

### Cosmos Transactions

### Ethereum Transactions

Evmos supports the following Ethereum transactions.

#### Legacy Transactions

Legacy transactions are the transaction format used prior to [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718), the Ethereum Improvement Proposal which transitioned from [RLP-encoded](https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp/) transactions to a new generalized envelope for typed transactions.

Ethereum legacy transactions take the following form:

```bash
RLP([nonce, gasPrice, gasLimit, to, value, data, v, r, s])
```

while the new standard for transactions takes this form:

```bash
TransactionType || TransactionPayload
```

where `TransactionType` is a number between `0x0` and `0x7f`, for a total of 128 possible transaction types, and `TransactionPayload` is an arbitrary byte array, defined by the transaction type.

EIP-2718 is backwards-compatible, meaning legacy transactions are still valid as transactions on the Ethereum network. RLP-encoded transactions always begin with a byte larger than or equal to `0xc0`, so typed transactions do not collide with legacy transactions, and differentiating between them is simple.

Evmos supports not only legacy transactions, but several typed transactions ([see below](#access-list-transactions-eip-2930httpseipsethereumorgeipseip-2930))!

#### Access List Transactions ([EIP-2930](https://eips.ethereum.org/EIPS/eip-2930))

Evmos supports access list transactions, which were introduced in EIP-2930 and take the following form:

```bash
0x01 || RLP([chainId, nonce, gasPrice, gasLimit, to, value, data, accessList, signatureYParity, signatureR, signatureS])
```

This transaction type contains an `accessList`, which is a list of addresses and storage keys that the transaction plans to access. Gas costs for transactions are still charged, but at a discount relative to the cost of accessing outside the list.

#### Dynamic Fee Transactions ([EIP-1559](https://eips.ethereum.org/EIPS/eip-1559))

Evmos supports dyanamic fee transactions, which were introduced in EIP-1559 and take the following form:

```bash
0x02 || RLP([chain_id, nonce, max_priority_fee_per_gas, max_fee_per_gas, gas_limit, destination, amount, data, access_list, signature_y_parity, signature_r, signature_s])
```

With this proposal, Ethereum introduced a base fee per gas in the protocol, which is meant to centralize at a certain target. These transactions specify the maximum fee per gas users are willing to give to miners (`max_priority_fee_per_gas`) to incentivize them to include their transaction in the next mined block, along with the maximum fee per gas users are willing to pay total (`max_fee_per_gas`). The base fee per gas, combined with the `max_priority_fee_per_gas`, should stay below the `max_fee_per_gas`.

### Interchain Transactions

<!-- TODO: transactions that use IBC or bridges to send them to other chains -->

## Transaction Receipts

<!-- TODO: explain Ethereum transaction receipts -->
