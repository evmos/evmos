<!--
order: 0
title: "Incentives Overview"
parent:
  title: "incentives"
-->

# `inflation`

## Abstract

The `x/inflation` module mints new Evmos tokens and allocates them in daily
epochs according to the [Evmos Token
Model](https://evmos.blog/the-evmos-token-model-edc07014978b) distribution to
Staking Rewards ( `40%` ), Team Vesting ( `25%`) Usage Incentives: `25%` and
Community Pool ( `10%`). It replaces the currently used Cosmos SDK `x/mint`
module.

The allocation of new coins incentivizes specific behaviour in the Evmos
network. Inflation allocates funds to 1) the `Fee Collector account` (in the sdk
`x/auth` module) to increase staking rewards, 2) the  `x/incentives` module
account  to provide supply for usage incentives and 3) the community pool
(managed by sdk `x/distr` module) to fund spending proposals.

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[State Transitions](03_state_transitions.md)**
4. **[Transactions](04_transactions.md)**
5. **[Hooks](05_hooks.md)**
6. **[Events](06_events.md)**
7. **[Parameters](07_params.md)**
8. **[Clients](08_clients.md)**
