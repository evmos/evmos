<!--
order: 0
title: "Incentives Overview"
parent:
  title: "incentives"
-->

# `incentives`

## Abstract

This document specifies the internal `x/incentives` module of the Evmos Hub.

The `x/incentives` module is part of the Evmos tokenomics and aims
to increase the growth of the network by distributing rewards
to users who interact with incentivized smart contracts.
The rewards drive users to interact with applications on Evmos and reinvest their rewards in more services in the network.

The usage incentives are taken from block reward emission (inflation)
and are pooled up in the Incentives module account (escrow address).
The incentives functionality is fully governed by native $EVMOS token holders
who manage the registration of `Incentives`,
so that native $EVMOS token holders decide which application should be part of the usage incentives.
This governance functionality is implemented using the Cosmos-SDK `gov` module
with custom proposal types for registering the incentives.

Users participate in incentives by submitting transactions to an incentivized contract.
The module keeps a record of how much gas the participants spent on their transactions and stores these in gas meters.
Based on their gas meters, participants in the incentive are rewarded in regular intervals (epochs).

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[State Transitions](03_state_transitions.md)**
4. **[Transactions](04_transactions.md)**
5. **[Hooks](05_hooks.md)**
6. **[Events](06_events.md)**
7. **[Parameters](07_parameters.md)**
8. **[Clients](08_clients.md)**
