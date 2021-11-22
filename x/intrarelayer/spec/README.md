<!--
order: 0
title: "Intrarelayer Overview"
parent:
  title: "intrarelayer"
-->

# `intrarelayer`

## Abstract

This document specifies the internal `x/intrarelayer` module of the Evmos Hub.

The Intrarelayer module enables the Evmos Hub to support a trustless, on-chain bidirectional internal relaying (*aka* intrarelaying) of tokens between Evmos' EVM and Cosmos runtimes, specifically the  `x/evm` and `x/bank` modules. This functionality allows token holders on Evmos to instantaneously convert their native Cosmos `sdk.Coins` (referred also in this document as "Coin(s)") to ERC20 (aka "Token(s)") and vice versa.

This intrarelaying functionality is fully governed by native $EVMOS token holders who manage canonical `TokenPair` registrations (ie, ERC20 ←→ Coin mappings). This governance functionality is implemented using the Cosmos-SDK `gov` module with custom proposal types for registering and updating the canonical mappings respectively.

Developers on Evmos may wish to deploy smart-contracts using the ERC20 token representation, so that they can transfer existing tokens on Ethereum and other EVM-based chains to Evmos and the rest of the Cosmos ecosystem. The `x/intrarelayer` module allows for an implementation of ERC20 ←→ Coin conversion by leveraging the governance and EVM functionalities of the Evmos Hub, while retaining fungibility with the original asset on the issuing environment/runtime (EVM or Cosmos) and preserving ownership of the contract.

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[State Transitions](03_state_transitions.md)**
4. **[Transactions](04_transactions.md)**
5. **[Hooks](05_hooks.md)**
6. **[Events](06_events.md)**
7. **[Parameters](07_params.md)**
8. **[Clients](08_clients.md)**
