<!--
order: 1
-->

# Concepts

## Actions

An `Action` corresponds to a given transaction that the user must perform to receive the allocated tokens from the airdrop.

All accounts start out with 1% of their entire airdrop allocation.

There are 4 types of actions, each of which release 20% of their remaining corresponding airdrop allocation. The 4 actions are as follows:

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

After staking Evmos tokens (i.e delegating), the corresponding proportion will be airdropped to the user's balance by performing a transfer from the claim escrow account (`ModuleAccount`) to the user.

### EVM Action

If the user deploys or interacts with a smart contract (via an application or wallet integration), the corresponding proportion will be airdropped to the user's balance by performing a transfer from the claim escrow account (`ModuleAccount`) to the user. This also applies when the user performs a transfer using Metamask or another web3 wallet of their preference.

### IBC Transfer Action

## Claim Records

A Claim Records is the metadata of claim data per address. It keeps track of all the actions performed by the the user as well as the total amount of tokens allocated to them. All users that have an address with a corresponding `ClaimRecord` are eligible to claim the airdrop.

## Claiming Process



### Ethereum Users

### Cosmos Hub and Osmosis Users



## Decay Period

A decay period is defined by the module parameters in order to incentivize users to claim their tokens and interact with the blockchain early. Once this period starts, it decreases the amount of claimable tokens by the user linearly over time. The start is of this period is defined by the by the addition of the `` and `DurationUntilDecay` parameter and the duration of the linear decay is defined by `DurationOfDecay`, as described below:

```python
decay_start = claim_start + duration_until_decay

decay_end = decay_start + duration_of_decay
```

By default, users have two months (`DurationUntilDecay`) to claim their full airdrop amount. After two months, the reward amount available will decline over 1 month (`DurationOfDecay`) in real time, until it hits `0%` at 3 months from launch (end).

## Airdrop Clawback

After the claim period ends (defined by the module parameters), the tokens that were not claimed will be transferred to the community pool treasury. In the same way, users with tokens allocated.
