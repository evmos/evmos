<!--
order: 0
title: "Vesting Overview"
parent:
  title: "vesting"
-->

# `vesting`

## Abstract

This document specifies the internal `x/vesting` module of the Evmos Hub.

The `x/vesting` module introduces the `ClawbackVestingAccount`,  a new vesting account type that implements the Cosmos SDK [`VestingAccount`](https://docs.cosmos.network/v0.42/modules/auth/05_vesting.html#vesting-account-types) interface. This account is used to allocate tokens that are subject to vesting, lockup, and clawback.

The `ClawbackVestingAccount` allows any two parties to agree on a future rewarding schedule, where tokens are granted permissions over time. The parties can use this account to enforce legal contracts or commit to mutual long-term interests.

In this commitment, vesting is the mechanism for gradually earning permission to transfer and delegate allocated tokens. Additionally, the lockup provides a mechanism to prevent the right to transfer allocated tokens and perform Ethereum transactions from the account. Both vesting and lockup are defined in schedules at account creation. At any time, the funder of a `ClawbackVestingAccount` can perform a clawback to retrieve unvested tokens. The circumstances under which a clawback should be performed can be agreed upon in a contract (e.g. smart contract).

For Evmos, the `ClawbackVestingAccount` is used to allocate tokens to core team members and advisors to incentivize long-term participation in the project.

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[State Transitions](03_state_transitions.md)**
4. **[Transactions](04_transactions.md)**
5. **[AnteHandlers](05_antehandlers.md)**
6. **[Events](06_events.md)**
7. **[Clients](07_clients.md)**
