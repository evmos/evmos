<!--
order: 2
-->

# dApp Store Contract Registration

This guide explains how to register your smart contract in the Evmos dApp store, and start earning income every time a user interacts with your smart contract. {synopsis}

The Evmos dApp store is a revenue-per-transaction model, which allows developers to get paid for deploying their decentralized application (dApps) on Evmos. Developers generate revenue every time a user interacts with their dApp in the dApp store, providing them a steady income. Users can discover new applications in the dApp store and pay for the transaction fees that finance the dApp's revenue. This value-reward exchange of dApp services for transaction fees is implemented by the [x/revenue module](../../../x/revenue/spec/01_concepts.md).

## Requirements

- Address of a deployed smart contract.
- Capability to sign transactions with the address that deployed the contract. If your smart contract was deployed by a contract using a [factory pattern](https://en.wikipedia.org/wiki/Factory_method_pattern), then the signing capaility is required for the address that deployed the factory.
- The nonce of the contract deployment transaction. You can query the nonce, e.g. using the `eth_getTransactionByHash` JSON-RPC endpoint.
- Withdrawer address, in case you wish to receive your earnings at a specified address.

::: warning
**IMPORTANT**: If your contract is part of a development project, please ensure that the deployer of the contract (or the factory that deploys the contract) is an account that is owned by that project. This avoids the situation of a malicious individual/employee deployer (including former contributors) who leaves your project and could later change the withdrawal address unilaterally.
:::

## Register Contract

To add your contract in the Evmos dApp Store, you need to register a `revenue` for that contract. The `revenue` includes the details for receiving a cut of the transaction fees, which users pay for interacting with your smart contract. Every time a user submits a transaction to your registered smart contract, a part of the transaction fees (50% by default) is transferred to the withdrawer address specified in the `revenue`. If the withdrawer is not specified, the transaction fees are sent to the contract deployer.

You can register a contract by signing a transaction with the address that originally deployed the contract. You can use the following CLI command, where

- `$NONCE` is the nonce of transaction that deployed the contract (e.g. `0`),
- `$CONTRACT` is the hex address of the deployed contract (e.g `0x5f6659B6F712c729c46786bA9562eC50907c67CF`) and
- (optional) `$WITHDRAWER` is the bech32 address of the address to receive the transaction fees (e.g. `evmos1keyy3teyq7t7kuxgeek3v65n0j27k20v2ugysf`):

```bash
# Register a revenue for your contract
evmosd tx revenue register $CONTRACT $NONCE $WITHDRAWER \
--from=dev0 \ # contract deployer key
--gas=700000 --gas-prices=10000aevmos \ # can vary depending on the network
```

After your transaction is submitted successfully, you can query your `revenue` with :

```bash
# Check revenues
evmosd q revenue contract $CONTRACT
```

Congrats ☄️☄️☄️ Now that you've registered a revenue for your contract, it is part of the Evmos dApp store and you will receive a cut of the transaction fees every time a user interacts with your contract. If you wondering how large your cut is, have a look at the [revenue parameter `DeveloperShares`](../../../x/revenue/spec/07_parameters.md#developer-shares-amount), which is controlled through governance. You can query the parameters using our [OpenAPI documentation](https://api.evmos.org).

### Deployed Factory Pattern

You can also register a contract which has been deployed by a smart contract instead of an [EOA](https://docs.evmos.org/modules/evm/01_concepts.html#accounts). In this case, you need to provide a sequence of nonces that proves the trace from an original deployer who deployed the factory to the contract that is being registered.

**Example** `DeployerEOA` -> `FactoryA` -> `FactoryB`-> `MyContract`: `DeployerEOA` deploys a `FactoryA` smart contract with nonce `5`. Then, `DeployerEOA` sends a transaction to `FactoryA` through which a `FactoryB` smart contract is created. If we assume `FactoryB` is the second contract created by `FactoryA`, then `FactoryA`'s nonce is `2`. Then, `DeployerEOA` sends a transaction to the `FactoryB` contract, through which `MyContract` is created. If this is the first contract created by FactoryB - the nonce is `1`. To be able to verify that `DeployerEOA` can register `MyContract`, we need to provide the following nonces: `[5, 2, 1]`.

## Update Contract

Registered contracts can also be updated. To update the withdrawer address of your `revenue`, use the following CLI command:

```bash
# Update withdrawer for your contract
evmosd tx revenue update $CONTRACT $WITHDRAWER \
--gas=700000 --gas-prices=10000aevmos \
--from=mm
```

If the specified withdrawer is the same address as the deployer, then the revenue is updated with an empty withdrawer address, so that all transaction fees are sent to the deployer address.

## Cancel Contract

Revenues can also be canceled. In order to stop receiving transaction fees for interaction with your contract, use the following CLI command:

```bash
# Cancel revenue for your contract
evmosd tx revenue cancel $CONTRACT \
--gas=700000 --gas-prices=10000aevmos \
--from=mm
```
