<!--
order: 0
title: "ERC20 Overview"
parent:
  title: "erc20"
-->

# `erc20`

::: tip
**Note:** Working on a governance proposal related to the ERC-20 Module? Make sure to look at [Evoblock Governance](../../validators/governance/overview.md), and specifically the [best practices](../../validators/governance/best_practices#erc-20-proposal).
:::

## Abstract

This document specifies the internal `x/erc20` module of the Evoblock Hub.

The `x/erc20` module enables the Evoblock Hub to support a trustless, on-chain bidirectional internal conversion of tokens between Evoblock' EVM and Cosmos runtimes, specifically the `x/evm` and `x/bank` modules. This allows token holders on Evoblock to instantaneously convert their native Cosmos `sdk.Coins` (in this document referred to as "Coin(s)") to ERC-20 (aka "Token(s)") and vice versa, while retaining fungibility with the original asset on the issuing environment/runtime (EVM or Cosmos) and preserving ownership of the ERC-20 contract.

This conversion functionality is fully governed by native $EVO token holders who manage the canonical `TokenPair` registrations (ie, ERC20 ←→ Coin mappings). This governance functionality is implemented using the Cosmos-SDK `gov` module with custom proposal types for registering and updating the canonical mappings respectively.

Why is this important? Cosmos and the EVM are two runtimes that are not compatible by default. The native Cosmos Coins cannot be used in applications that require the ERC-20 standard. Cosmos coins are held on the `x/bank` module (with access to module methods like querying the supply or balances) and ERC-20 Tokens live on smart contracts. This problem is similar to [wETH](https://weth.io/), with the difference,  that it not only applies to gas tokens (like $EVO), but to all Cosmos Coins (IBC vouchers, staking and gov coins, etc.) as well.

With the `x/erc20` users on Evoblock can

- use existing native cosmos assets (like $OSMO or $ATOM) on EVM-based chains, e.g. for Trading IBC tokens on DeFi protocols, buying NFT, etc.
- transfer existing tokens on Ethereum and other EVM-based chains to Evoblock to take advantage of application-specific chains in the Cosmos ecosystem
- build new applications that are based on ERC-20 smart contracts and have access to the Cosmos ecosystem.

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[State Transitions](03_state_transitions.md)**
4. **[Transactions](04_transactions.md)**
5. **[Hooks](05_hooks.md)**
6. **[Events](06_events.md)**
7. **[Parameters](07_parameters.md)**
8. **[Clients](08_clients.md)**
