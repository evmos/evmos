<!--
order: 0
title: "Inflation Overview"
parent:
  title: "inflation"
-->

# `inflation`

## Abstract

The `x/inflation` module mints new Evmos tokens and allocates them in daily
epochs according to the [Evmos Token
Model](https://evmos.blog/the-evmos-token-model-edc07014978b) distribution to
* Staking Rewards `40%`,
* Team Vesting `25%`,
* Usage Incentives: `25%`,
* Community Pool `10%`.

It replaces the currently used Cosmos SDK `x/mint` module.

The allocation of new coins incentivizes specific behaviour in the Evmos
network. Inflation allocates funds to 1) the `Fee Collector account` (in the sdk
`x/auth` module) to increase staking rewards, 2) the  `x/incentives` module
account  to provide supply for usage incentives and 3) the community pool
(managed by sdk `x/distr` module) to fund spending proposals.

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Hooks](03_hooks.md)**
4. **[Events](04_events.md)**
5. **[Parameters](05_parameters.md)**
6. **[Clients](06_clients.md)**
