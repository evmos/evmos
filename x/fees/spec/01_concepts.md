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

If the contract was created by a smart contract, through the `CREATE` opcode, we need to provide all the nonces from the creation path. Let's take the example of `DeployerEOA` deploying a `FactoryA` smart contract with nonce `5`. Then, `DeployerEOA` sends a transaction to `FactoryA` through which a `FactoryB` smart contract is created. Let us assume `FactoryB` is the second contract created by `FactoryA` - the nonce is `2`. Then, `DeployerEOA` sends a transaction to the `FactoryB` contract, through which `MyContract` is created. Let us assume this is the first contract created by `FactoryB` - the nonce is `1`. We now have an address derivation path of `DeployerEOA` -> `FactoryA` -> `FactoryB` -> `MyContract`. To be able to verify that `DeployerEOA` can register `MyContract`, we need to provide the following nonces: `[5, 2, 1]`.

::: tip
**Note**: Even if `MyContract` is created from `FactoryB` through a transaction sent by an account different from `DeployerEOA`, only `DeployerEOA` can register `MyContract`.
:::

## Global Minimum Gas Price

The minimum gas price needed for transactions to be processed by Evmos. It applies to both Cosmos and EVM transactions. Governance can change this `fees` module parameter value. If the effective gas price or the minimum gas price is lower than the global `MinGasPrice` (`min-gas-price (local) < MinGasPrice (global) OR EffectiveGasPrice < MinGasPrice`), then `MinGasPrice` is used as a lower bound. If transactions are rejected due to having a gas price lower than `MinGasPrice`, users need to resend the transactions with a gas price higher than `MinGasPrice`. In the case of EIP-1559 (dynamic fee transactions), users must increase the priority fee for their transactions to be valid.
