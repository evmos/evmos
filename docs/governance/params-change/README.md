---
order: 1
parent:
  order: 7
---

# Evmos Parameters

This Evmos educational documentation aims to outline the [Hub's parameters](#params-wiki), describe their functions, and describe the potential implications of modifying each parameter. This documentation also aims to provide [guidelines for creating and assessing parameter-change proposals](#drafting-a-parameter-change-proposal).

**This documentation has not had adequate review from experts or testing from participants, so please be cautious when using it.** [Discuss its development here](https://forum.cosmos.network/t/gwg-cosmos-hub-parameters-wiki/3170) and please provide feedback either in the forum or by opening a Github issue. If you are technically inclined, this is the full [list of modules](https://github.com/cosmos/cosmos-sdk/tree/master/x) in the Cosmos SDK.

## Drafting a Parameter Change Proposal

Drafting and submitting a parameter-change governance proposal involves two kinds of risk: losing proposal deposit amounts and the potential to alter the function of Evmos network in an undesirable way. The objective of this documentation is to reduce these risks by preparing participants for what to pay attention to and for what information to consider in a proposal. Ideally, a proposal should only fail to pass because the voters 1) are aware and engaged and 2) are able to make an informed decision to vote down the proposal.

If you are considering drafting a proposal, you should review the general
background on drafting and submitting a proposal:

1. [How the voting process and governance mechanism works](../process.md)
1. [How to draft your proposal and engage with the Cosmos community about it](../best-practices.md)
1. [How to format proposals](../formatting.md)
1. [How to submit your proposal](../submitting.md)

You should also review details specific to Parameter Change proposals below.

## Params Wiki

The complete parameters of Evmos are split up into different modules.
Each module has its own set of parameters. Any of them can be updated with a
Params Change Proposal. There is an [index of these parameters here](./param-index.md).

There are currently 8 modules active in Evmos with parameters that may be altered via governance proposal:

1. [auth](./Auth.md) - Authentication of accounts and transactions
2. [bank](./Bank.md) - Token transfer functionalities
3. [gov](./Governance.md) - On-chain governance proposals and voting
4. [staking](./Staking.md) - Proof-of-stake layer
5. [slashing](./Slashing.md) - Validator punishment mechanisms
6. [distribution](./Distribution.md) - Fee distribution and staking token provision distribution
7. [crisis](./Crisis.md) - Halting the blockchain under certain circumstances (ie. if an invariant is broken)
8. [mint](./Mint.md) - Creation of new units of staking token

The value or setting for each parameter may be verified in the chain's genesis file, [found here](https://raw.githubusercontent.com/cosmos/launch/master/genesis.json). These are the parameter settings that the latest Evmos chain launched with, and will remain so unless a governance proposal or software upgrade changes them.

There are also ways to query the current settings for each module's parameter(s). Some can be queried with the command line program [`evmosd`](../../getting-started/installation.md), but I'm still exploring the ways that these settings can be queried.

You can begin by using the command `gaia q [module] -h` to get help about the subcommands for the module you want to query. For example, `evmosd q staking params --chain-id evmos_9001-3 --node http://51.79.82.228:26657` returns the settings of four parameters:

```
unbonding_time: 504h0m0s
max_validators: 125
max_entries: 7
bond_denom: aevmos
```

## The Voting Process & Governance Mechanism

The criteria for submitting a parameter-change proposal and the subsequent voting conditions are the same as those for signalling (text-based) proposals and community-spend proposals. Details about the deposit period can be found [here](../process.md#_1-deposit-period), and voting period [here](../process.md#what-determines-whether-or-not-a-governance-proposal-passes).

If a paramater-change proposal is successful, the change takes effect immediately upon completion of the voting period.

### Note

- You cannot currently query the `bank` module's parameter, which is `sendenabled`. You also cannot query the `crisis` module's parameters.
- You will need to compile [`evmosd`](../../getting-started/installation.md) from source into a binary file executable by your operating system eg. MacOS, Windows, Linux
- You will need to indicate which chain you are querying, and currently this is `--chain-id evmos_9001-4`
- You will need to connect to a full node. If evmosd isn't already configured for this, you can use this tag in your command `--node [address]:26657`.

## Full nodes

Running a full node can be difficult for those not technically-inclined, so you may choose to use a third-party's full node. In this case, the primary security risk is that of censorship: it's the single place where you have a single gateway to the network, and any messages submitted through an untrusted node could be censored.

You can find a list of available Evmos endpoints under the [API section](https://github.com/cosmos/chain-registry/blob/master/evmos_9001/chain.json) in the [Chain Registry](https://github.com/cosmos/chain-registry).

## Credits

This documentation was originally created by Gavin Birch ([Figment Networks](https://figment.network)). Its development was supported by funding approved on January 29, 2020 by the Evmos via Community Spend [Proposal 23](https://hubble.figment.network/cosmos/chains/evmos_9001-3/governance/proposals/23) ([full Proposal PDF here](https://ipfs.io/ipfs/QmSMGEoY2dfxADPfgoAsJxjjC6hwpSNx1dXAqePiCEMCbY)).
