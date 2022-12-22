<!--
order: 0
title: "Recovery Overview"
parent:
  title: "recovery"
-->

# `recovery`

Recover tokens that are stuck on unsupported Evmos accounts.

## Abstract

This document specifies the  `x/recovery` module of the Evmos Hub.

The `x/recovery` module enables users on Evmos to recover locked funds that were transferred to accounts whose keys are not supported on Evmos. This happened in particular after the initial Evmos launch (`v1.1.2`), where users transferred tokens to a `secp256k1` Evmos address via IBC in order to [claim their airdrop](https://docs.evmos.org/modules/claims/). To be EVM compatible, [keys on Evmos](https://docs.evmos.org/users/technical_concepts/accounts.html#evmos-accounts) are generated using the `eth_secp256k1` key type which results in a different address derivation than e.g. the `secp256k1` key type used by other Cosmos chains.

At the time of Evmos’ relaunch, the value of locked tokens on unsupported accounts sits at $36,291.28 worth of OSMO and $268.86 worth of ATOM tokens according to the [Mintscan](https://www.mintscan.io/evmos/assets) block explorer. With the `x/recovery` module, users can recover these tokens back to their own addresses in the originating chains by performing IBC transfers from authorized IBC channels (i.e Osmosis for OSMO, Cosmos Hub for ATOM).

## Contents

1. **[Concepts](01_concepts.md)**
2. **[Hooks](02_hooks.md)**
3. **[Events](03_events.md)**
4. **[Parameters](04_parameters.md)**
5. **[Clients](05_clients.md)**
