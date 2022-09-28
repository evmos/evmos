<!--
order: 1
-->

# Concepts

## Evmos dApp Store

The Evmos dApp store is a revenue-per-transaction model, which allows developers to get payed for deploying their decentralized application (dApps) on Evmos. Developers generate revenue, every time a user interacts with their dApp in the dApp store, gaining them a steady income. Users can discover new applications in the dApp store and pay for the transaction fees that finance the dApp's revenue. This value-reward exchange of dApp services for transaction fees is implemented by the `x/revenue` module.

## Registration

Developers register their application in the dApp store by registering their application's smart contracts. Any contract can be registered by a developer by submitting a signed transaction. The signer of this transaction must match the address of the deployer of the contract in order for the registration to succeed. After the transaction is executed successfully, the developer will start receiving a portion of the transaction fees paid when a user interacts with the registered contract.

::: tip
 **NOTE**: If your contract is part of a developer project, please ensure that the deployer of the contract (or the factory that deployes the contract) is an account that is owned by that project. This avoids the situtation, that an individual deployer who leaves your project could become malicious.
:::

## Fee Distribution

As described above, developers will earn a portion of the transaction fee after registering their contracts. To understand how transaction fees are distributed, we look at the following two things in detail:

* The transactions eligible are only [EVM transactions](https://docs.evmos.org/modules/evm/) (`MsgEthereumTx`). Cosmos SDK transactions are not eligible at this time.
* The registration of factory contracts (smart contracts that have been deployed by other contracts) requires the identification original contract's deployer. This is done through address derivation.

### EVM Transaction Fees

Users pay transaction fees to pay interact with smart contracts using the EVM. When a transaction is executed, the entire fee amount (`gasLimit * gasPrice`) is sent to the `FeeCollector` module account during the [Cosmos SDK AnteHandler](https://docs.cosmos.network/v0.44/modules/auth/03_antehandlers.html) execution. After the EVM executes the transaction, the user receives a refund of `(gasLimit - gasUsed) * gasPrice`. In result a user pays a total transaction fee of `txFee = gasUsed * gasPrice` for the execution.

This transaction fee is distributed between developers and validators, in accordance with the `x/revenue` module parameters: `DeveloperShares`, `ValidatorShares`. This distribution is handled through the EVM's [`PostTxProcessing` Hook](./05_hooks.md).

### Address Derivation

dApp developers might use a [factory pattern](https://en.wikipedia.org/wiki/Factory_method_pattern) to implement their application logic through smart contracts. In this case a smart contract can be either deployed by an Externally Owned Account ([EOA](https://ethereum.org/en/whitepaper/#ethereum-accounts): an account controlled by a private key, that can sign transactions) or through another contract.

In both cases, the fee distribution requires the identification a deployer address that is an EOA address, unless a withdrawal address is set by the contract deployer during registration to receive transaction fees for a registered smart contract. If a withdrawal address is not set, it defaults to the deployer’s address.

The identification of the deployer address is done through address derivation. When registering a smart contract, the deployer provides an array of nonces, used to [derive the contract’s address](https://github.com/ethereum/go-ethereum/blob/d8ff53dfb8a516f47db37dbc7fd7ad18a1e8a125/crypto/crypto.go#L107-L111):

* If `MyContract` is deployed directly by `DeployerEOA`, in a transaction sent with nonce `5`, then the array of nonces is `[5]`.
* If the contract was created by a smart contract, through the `CREATE` opcode, we need to provide all the nonces from the creation path. E.g. if `DeployerEOA` deploys a `FactoryA` smart contract with nonce `5`. Then, `DeployerEOA` sends a transaction to `FactoryA` through which a `FactoryB` smart contract is created. If we assume `FactoryB` is the second contract created by `FactoryA`, then `FactoryA`'s nonce is `2`. Then, `DeployerEOA` sends a transaction to the `FactoryB` contract, through which `MyContract` is created. If this is the first contract created by `FactoryB` - the nonce is `1`. We now have an address derivation path of `DeployerEOA` -> `FactoryA` -> `FactoryB` -> `MyContract`. To be able to verify that `DeployerEOA` can register `MyContract`, we need to provide the following nonces: `[5, 2, 1]`.

::: tip
**Note**: Even if `MyContract` is created from `FactoryB` through a transaction sent by an account different from `DeployerEOA`, only `DeployerEOA` can register `MyContract`.
:::
