<!--
order: 1
-->

# How to register your Smart Contract in the dApp store

This guide explains how to register your smart contract in the Evmos dApp store and start earning an income every time a user interacts with your smart contract {synopsis}

The Evmos dApp store is a revenue-per-transaction model, which allows developers to get payed for deploying their decentralized application (dApps) on Evmos. Developers generate revenue, every time a user interacts with their dApp in the dApp store, gaining them a steady income. Users can discover new applications in the dApp store and pay for the transaction fees that finance the dApp's revenue. This value-reward exchange of dApp services for transaction fees is implemented by the [x/feesplit module](https://github.com/evmos/evmos/blob/main/x/feesplit/spec/01_concepts.md).

## Prerequisites

* Address of a deployed smart contract
* Capability to sign transactions with the address that deployed the contract. If your smart contract was deployed by a another contract using a [factory pattern](https://en.wikipedia.org/wiki/Factory_method_pattern), then this capaility is required for the address that deployed the factory.
* Withdrawer address, in case you wish to receive your earnings at a specified address

::: warning
If your contract is part of a developer project, please ensure that the deployer of the contract (or the factory that deployes the contract) is an account that is owned by that project. This avoids the situtation, that an individual deployer who leaves your project could become malicious.
:::

## Register Contract

To add your contract in the Evmos dApp store, you need to register a `feesplit` for the contract. The `feesplit` includes the necessary details for receiving a part of the transaction fees, that users pay for interacting with your smart contract. Every time a user submits a transaction to your registered contract a part of the transaction fess is transferred to the withdrawer address specified in the `feesplit`. If the withdrawer is not specified, then the transaction fees are sent to the contract deployer.

You can register a contract by signing a transaction with the address that deployed the contract. You can use the following CLI command, where

* `$NONCE` is the nonce of transaction that deployed the contract (e.g. `0`),
* `$CONTRACT` is the hex address of the deployed contract (e.g `0x5f6659B6F712c729c46786bA9562eC50907c67CF`) and
* (optional) `$WITHDRAWER` is the bech32 address of the address to receive the transaction fees (e.g. `evmos1keyy3teyq7t7kuxgeek3v65n0j27k20v2ugysf`):

```bash
# Register a feesplit for your the contract
evmosd tx feesplit register $CONTRACT $NONCE $WITHDRAWER \
--from=mm \ # the name
--fees=20aevmos \
```
After your transaction is submitted successfully you can query your `feesplit`:

```bash
# Check feesplits
evmosd q feesplit contract $CONTRACT
```
Congrats ☄️ Now that you've registered a feesplit for your contract, your contract is part of the Evmos dApp store and you will receive a part of the transaction fees, every time a user interacts with your contract. If you wondering how large your cut is, have a look at TODO.

### Deployed Factory Pattern TODO

## Update Contract

A registered contract can also be updated. To update the withdrawer address of your `feesplit` you can the following CLI command:

```bash
# Update withdrawer of existing contract to new account, e.g. A3
evmosd tx feesplit update $CONTRACT $WITHDRAWER \
--fees=40aevmos \
--from=mm
```

## Cancel Contract

A feesplit can also we canceled in order to stop receiving transaction fees for interaction with your contract using:

```bash
# Cancel feesplit for a given contract
evmosd tx feesplit cancel $CONTRACT \
--fees=40aevmos \
--from=mm
```
