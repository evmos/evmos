<!--
order: 1
-->

# Overview

Learn about validating on Evmos {synopsis}

## Introduction

Evmos is based on [Tendermint Core](https://github.com/tendermint/tendermint/blob/master/docs/introduction/what-is-tendermint.md), which relies on a set of validators that are responsible for committing new blocks in the blockchain. These validators participate in the consensus protocol by broadcasting votes which contain cryptographic signatures signed by each validator's private key.

Validator candidates can bond their own staking tokens and have the tokens "delegated", or staked, to them by token holders. The **{{ $themeConfig.project.denom }}** is Evmos's native token. At its onset, Evmos will launch with 150 validators. The validators are determined by who has the most stake delegated to them - the top 150 validator candidates with the most stake will become Evmos validators.

Validators and their delegators will earn {{ $themeConfig.project.denom }} as block provisions and tokens as transaction fees through execution of the Tendermint consensus protocol. Initially, transaction fees will be paid in EVMOS but in the future, any token in the Cosmos ecosystem will be valid as fee tender if it is whitelisted by governance. Note that validators can set commission on the fees their delegators receive as additional incentive.

If validators double sign, are frequently offline or do not participate in governance, their staked {{ $themeConfig.project.denom }} (including {{ $themeConfig.project.denom }} of users that delegated to them) can be slashed. The penalty depends on the severity of the violation.

## Hardware

Validators should set up a physical operation secured with restricted access. A good starting place, for example, would be co-locating in secure data centers.

Validators should expect to equip their datacenter location with redundant power, connectivity, and storage backups. Expect to have several redundant networking boxes for fiber, firewall and switching and then small servers with redundant hard drive and failover. Hardware can be on the low end of datacenter gear to start out with.

We anticipate that network requirements will be low initially. Bandwidth, CPU and memory requirements will rise as the network grows. Large hard drives are recommended for storing years of blockchain history.

### Supported OS

We officially support macOS, Windows and Linux only in the following architectures:

* `darwin/arm64`
* `darwin/x86_64`
* `linux/arm64`
* `linux/amd64`
* `windows/x86_64`

### Minimum Requirements

To run mainnet or testnet validator nodes, you will need a machine with the following minimum hardware requirements:

* 4 or more physical CPU cores
* At least 500GB of SSD disk storage
* At least 32GB of memory (RAM)
* At least 100mbps network bandwidth

As the usage of the blockchain grows, the server requirements may increase as well, so you should have a plan for updating your server as well.

## Get Involved

::: tip
Seek legal advice if you intend to run a validator.
:::

Set up a dedicated validator's website, social profile (eg: Twitter) and signal your intention to become a validator on Discord. This is important since users will want to have information about the entity they are staking their EVMOS to.

## Community

Discuss the finer details of being a validator and seek advise from the rest of the validator community on our [Discord](https://discord.gg/evmos).
