<!--
order: 0
title: "Revenue Overview"
parent:
  title: "revenue"
-->

# `revenue`

## Abstract

This document specifies the internal `x/revenue` module of the Evoblock Hub.

The `x/revenue` module enables the Evoblock Hub to support splitting transaction fees between block proposer and smart contract deployers. As a part of the [Evoblock Token Model](https://evoblock.blog/the-evoblock-token-model-edc07014978b), this mechanism aims to increase the adoption of the Evoblock Hub by offering a new stable source of income for smart contract deployers. Developers can register their smart contracts and everytime someone interacts with a registered smart contract, the contract deployer or their assigned withdrawal account receives a part of the transaction fees.

Together, all registered smart contracts make up the Evoblock dApp Store: paying developers and network operators for their services via built-in shared fee revenue model.

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[State Transitions](03_state_transitions.md)**
4. **[Transactions](04_transactions.md)**
5. **[Hooks](05_hooks.md)**
6. **[Events](06_events.md)**
7. **[Parameters](07_parameters.md)**
8. **[Clients](08_clients.md)**
9. **[Future Improvements](09_improvements.md)**
