<!--
order: 1
-->

# Concepts

## EOA

An Externally Owned Account ([EOA](https://ethereum.org/en/whitepaper/#ethereum-accounts)) is an account controlled by a private key, that can sign transactions.

## Deployer Address

The EOA address that deployed the smart contract being registered for fee distribution.

## Withdraw Address

The address set by a contract deployer to receive transaction fees for a registered smart contract. If not set, it defaults to the deployer’s address.

## Developer

The entity that has control over the deployer account.

## Registration of a Contract

Any contract can be registered by a developer by submitting a signed transaction. The signer of this transaction must match the address of the deployer of the contract in order for the registration to succeed. After the transaction is executed successfully, the developer will start receiving a portion of the transaction fees paid when a user interacts with the registered contract.

### Fee Distribution

As described above, developers will earn a portion of the transaction fee after they register their contracts. The transactions eligible are only EVM transactions (`MsgEthereumTx`). Cosmos SDK transactions are not eligible at this time.

#### EVM Transaction Fees

When a transaction is executed, the entire fee amount `gasLimit * gasPrice` is sent to the `FeeCollector` Module Account during the `AnteHandler` execution. After the EVM executes the transaction, the user receives a refund of `(gasLimit - gasUsed) * gasPrice`.

Therefore, the user only pays for the execution: `txFee = gasUsed * gasPrice`. This is the transaction fee distributed between developers and validators, in accordance with the `x/fees` module parameters: `DeveloperShares`, `ValidatorShares`. This distribution is handled through the `PostTxProcessing` [Hook](./05_hooks.md).

### Address Derivation

When registering a smart contract, the deployer provides an array of nonces, used to [derive the contract’s address](https://github.com/ethereum/go-ethereum/blob/d8ff53dfb8a516f47db37dbc7fd7ad18a1e8a125/crypto/crypto.go#L107-L111). The smart contract can be directly deployed by the deployer's EOA or created through one or more [factory](https://en.wikipedia.org/wiki/Factory_method_pattern) pattern smart contracts.

If `MyContract` is deployed directly by `DeployerEOA`, in a transaction sent with nonce `5`, then the array of nonces is `[5]`.

If the contract was created by a smart contract, through the `CREATE` opcode, we need to provide all the nonces from the creation path. Let's take the example of `DeployerEOA` deploying a `Factory0` smart contract with nonce `5`. Then, `DeployerEOA` sends a transaction to `Factory0` through which a `Factory1` smart contract is created. Let us assume `Factory1` is the second contract created by `Factory0` - the nonce is `1`. Then, `DeployerEOA` sends a transaction to the `Factory1` contract, through which `MyContract` is created. Let us assume this is the first contract created by `Factory1` - the nonce is `0`. We now have an address derivation path of `DeployerEOA` -> `Factory0` -> `Factory1` -> `MyContract`. To be able to verify that `DeployerEOA` can register `MyContract`, we need to provide the following nonces: `[5, 1, 0]`.

::: tip
**Note**: Even if `MyContract` is created from `Factory1` through a transaction sent by an account different from `DeployerEOA`, only `DeployerEOA` can register `MyContract`.
:::
