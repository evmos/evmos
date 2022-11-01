<!--
order: 1
-->

# Concepts

## Rektdrop

The Evoblock [Rektdrop](https://evoblock.blog/the-evoblock-rektdrop-abbe931ba823) is the genesis airdrop for the EVO token to Cosmos Hub, Osmosis and Ethereum users.

> The end goal of Evoblock is to bring together the Cosmos and Ethereum community and thus the Rektdrop has been designed to reward past participation in both networks under this theme of “getting rekt”.

The Rektdrop is the first airdrop that:

- Implements the [gasdrop](https://www.sunnya97.com/blog/gasdrop) mechanism by Sunny Aggarwal
- Covers the most number of chains and applications involved in an airdrop
- Airdrops to bridge users
- Includes reparations for users in exploits and negative market externalities (i.e. MEV)

The snapshot of the airdrop was on **November 25th, 2021 at 19:00 UTC**

## Actions

An `Action` corresponds to a given transaction that the user must perform to receive the allocated tokens from the airdrop.

There are 4 types of actions, each of which release 25% of their remaining corresponding airdrop allocation. The 4 actions are as follows (`ActionUnspecified` is not considered for claiming):

```go
// UNSPECIFIED defines an invalid action. NOT claimable
ActionUnspecified Action = 0
// VOTE defines a proposal vote.
ActionVote Action = 1
// DELEGATE defines an staking delegation.
ActionDelegate Action = 2
// EVM defines an EVM transaction.
ActionEVM Action = 3
// IBC Transfer defines a fungible token transfer transaction via IBC.
ActionIBCTransfer Action = 4
```

These actions are monitored by registering claim post transaction **hooks** to the governance, staking, and EVM modules. Once the user performs an action, the `x/claims` module will unlock the corresponding portion of the assets and transfer them to the balance of the user.

These actions can be performed in any order and the claims module will not grant any additional tokens after the corresponding action is performed.

### Vote Action

After voting on a proposal, the corresponding proportion will be airdropped to the user's balance by performing a transfer from the claim escrow account (`ModuleAccount`) to the user.

### Staking (i.e Delegate) Action

After staking Evoblock tokens (i.e delegating), the corresponding proportion will be airdropped to the user's balance by performing a transfer from the claim escrow account (`ModuleAccount`) to the user.

### EVM Action

If the user deploys or interacts with a smart contract (via an application or wallet integration), the corresponding proportion will be airdropped to the user's balance by performing a transfer from the claim escrow account (`ModuleAccount`) to the user. This also applies when the user performs a transfer using Metamask or another web3 wallet of their preference.

### IBC Transfer Action

If a user submits an IBC transfer to a recipient on a counterparty chain or receives an IBC transfer from a counterparty chain, the corresponding proportion will be airdropped to the user's balance submitting or receiving the transfer.

## Claim Records

A Claims Records is the metadata of claim data per address. It keeps track of all the actions performed by the the user as well as the total amount of tokens allocated to them. All users that have an address with a corresponding `ClaimRecord` are eligible to claim the airdrop.

## Claiming Process

As described in the [Action](#action) section, a user must submit transactions to receive the allocated tokens from the airdrop. However, since Evoblock only supports Ethereum keys and not de default Tendermint keys, this process differs for Ethereum and Cosmos eligible users.

### Ethereum Users

Evoblock shares the coin type (`60`) and key derivation (Ethereum `secp256k1`) with Ethereum. This allows users (EOA accounts) that have been allocated EVO tokens to directly claim their tokens using their preferred web3 wallet.

### Cosmos Hub and Osmosis Users

Cosmos Hub and Osmosis users who use the default Tendermint `secp256k1` keys, need to perform a "cross-chain attestation" of their Evoblock address.

This can be done by submitting an IBC transfer from Cosmos Hub and Osmosis, which is signed by the addresses, that have been allocated the tokens.

The recipient Evoblock address of this IBC transfer is the address, that the tokens will be airdropped to.

::: warning
**IMPORTANT**

Only submit an IBC transfer to an Evoblock address that you own. Otherwise, you will lose your airdrop allocation.
:::

## Decay Period

A decay period defines the duration of the period during which the amount of claimable tokens by the user decays decrease linearly over time. It's goal is to incentivize users to claim their tokens and interact with the blockchain early.

The start is of this period is defined by the by the addition of the `AirdropStartTime` and `DurationUntilDecay` parameter and the duration of the linear decay is defined by `DurationOfDecay`, as described below:

```go
decayStartTime = AirdropStartTime + DurationUntilDecay
decayEndTime = decayStartTime + DurationOfDecay
```

By default, users have two months (`DurationUntilDecay`) to claim their full airdrop amount. After two months, the reward amount available will decline over 1 month (`DurationOfDecay`) in real time, until it hits `0%` at 3 months from launch (end).

## Airdrop Clawback

After the claim period ends, the tokens that were not claimed by users will be transferred to the community pool treasury. In the same way, users with tokens allocated but no transactions (i.e nonce = 0), will have their balance clawbacked to the community pool.
