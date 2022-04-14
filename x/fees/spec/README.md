<!--
order: 0
title: "Fees Overview"
parent:
  title: "fees"
-->

# `fees`

Split EVM transaction fees between block proposer and smart contract developers.

## Abstract

The current document specifies the internal `x/fees` module of the Evmos Hub.

The `x/fees` module is part of the Evmos tokenomics ([Evmos Token Model Blog](https://evmos.blog/the-evmos-token-model-edc07014978b)) and aims to increase the growth of the network by splitting the transaction fees between block proposer and smart contract deployers (or their assigned withdraw account). This mechanism is introduced to increase the adoption of the Evmos Hub by offering a new stable source of income for smart contract developers.

This is the web3 dApp Store: paying developers and network operators for their services via built-in shared fee revenue model.

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
