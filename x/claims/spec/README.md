<!--
order: 0
title: "Claims Overview"
parent:
  title: "claims"
-->

# `claims`

## Abstract

This document specifies the internal `x/claims` module of the Evmos Hub.

The `x/claims` module is part of the Evmos [Rektdrop](https://evmos.blog/the-evmos-rektdrop-abbe931ba823) and aims to increase the distribution of the network tokens to a large number of users.

Users are assigned with an initial amount of tokens from the airdrop allocation, and then are able to automatically claim higher percentages as they perform certain tasks on-chain.

For the Evmos Rektdrop, users are required to claim their airdrop by participating in core network activities. A Rektdrop recipient has to perform the following activities to get the allocated tokens:

* 25% is claimed by staking
* 25% is claimed by voting in governance
* 25% is claimed by using the EVM (deploy or interact with contract, transfer EVMOS through a web3 wallet)
* 25% is claimed by by sending or receiving an IBC transfer

Furthermore, these claimable assets 'expire' if not claimed. Users have two months (`DurationUntilDecay`) to claim their full airdrop amount. After two months, the reward amount available will decline over 1 month (`DurationOfDecay`) in real time, until it hits `0%` at 3 months from launch (`DurationUntilDecay + DurationOfDecay`).

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[State Transitions](03_state_transitions.md)**
4. **[Hooks](04_hooks.md)**
5. **[Events](05_events.md)**
6. **[Parameters](06_parameters.md)**
7. **[Clients](07_clients.md)**
