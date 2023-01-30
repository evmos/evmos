<!--
order: 0
title: "Epochs Overview"
parent:
  title: "epochs"
-->

# `epochs`

## Abstract

This document specifies the internal `x/epochs` module of the Evmos Hub.

Often, when working with the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk),
we would like to run certain pieces of code every so often.

The purpose of the `epochs` module is to allow other modules to maintain
that they would like to be signaled once in a time period.
So, another module can specify it wants to execute certain code once a week, starting at UTC-time = x.
`epochs` creates a generalized epoch interface to other modules so they can be more easily signaled upon such events.

## Contents

1. **[Concept](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Events](03_events.md)**
4. **[Keeper](04_keeper.md)**
5. **[Hooks](05_hooks.md)**
6. **[Queries](06_queries.md)**
7. **[Future improvements](07_future_improvements.md)**
