<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## Unreleased

### State Machine Breaking

- [\#173](https://github.com/tharsis/evmos/pull/173) Rename `intrarelayer` module to `erc20`

### Improvements

*- (deps) bumped Tendermint to `v0.34.15` and tm-db to `v0.6.6`.

### Bug Fixes 

- (erc20) [\#169](https://github.com/tharsis/evmos/pull/169) Fixes several testnet bugs:
  - Check if supply exists for a token before when submitting a `RegisterCoinProposal`, allowing users to create an ERC20 representation of an invalid Cosmos Coin.
  - Sanitize the ERC20 token name when creating coin metadata on ER `RegisterERC20Proposal`.
  - Fix coin metadata validation error when registering an ERC20 with 0 denom units. 

## [v0.4.2] - 2021-12-11

### Bug Fixes

- (app) [\#166](https://github.com/tharsis/evmos/pull/166) Fix `UpgradeHandler`.

## [v0.4.1] - 2021-12-07

### Improvements

- (build) [\#143](https://github.com/tharsis/evmos/pull/143) Added `build-reproducible` rule in `Makefile` to build docker containers

### Bug Fixes

- (build) [\#151](https://github.com/tharsis/evmos/pull/151) Fixes `version` command by picking the latest tag in the current branch instead of across all branches as the current version

## [v0.4.0] - 2021-12-02

### State Machine Breaking

- (erc20) [\#119](https://github.com/tharsis/evmos/issues/119) Register `x/erc20` proposal types on governance module.

### Improvements

- (app) [\#128](https://github.com/tharsis/evmos/pull/128) Add ibc-go `TestingApp` interface.
- (ci) [\#117](https://github.com/tharsis/evmos/pull/117) Enable automatic backport of PRs.
- (deps) [\#135](https://github.com/tharsis/evmos/pull/135) Bump Ethermint version to [`v0.9.0`](https://github.com/tharsis/ethermint/releases/tag/v0.9.0)
- (ci) [\#136](https://github.com/tharsis/evmos/pull/136) Deploy `evmos` docker container to [docker hub](https://hub.docker.com/u/tharsishq) for every versioned releases

### Bug Fixes

- (build) [\#116](https://github.com/tharsis/evmos/pull/116) Fix `build-docker` command

## [v0.3.0] - 2021-11-24

### API Breaking

- (erc20) [\#99](https://github.com/tharsis/evmos/pull/99) Rename `enable_e_v_m_hook` json parameter to `enable_evm_hook`.

### Improvements

- (deps) [\#110](https://github.com/tharsis/evmos/pull/110) Bump Ethermint version to [`v0.8.1`](https://github.com/tharsis/ethermint/releases/tag/v0.8.1)
- (erc20) [\#107](https://github.com/tharsis/evmos/pull/107) Add IBC validation
- (cmd) [\#105](https://github.com/tharsis/evmos/pull/105) Improve testnet command to include JSON-RPC client.

## Bug Fixes

- (erc20) [\#109](https://github.com/tharsis/evmos/pull/109) Fix hardcoded erc20 nonce and `UpdateTokenPairERC20` proposal to support ERC20s with 0 decimals.
- (erc20) [\#102](https://github.com/tharsis/evmos/pull/102) Add `convert-erc20` cmd

## [v0.2.0] - 2021-11-17

### Features

- (erc20) [\#82](https://github.com/tharsis/evmos/pull/82) ERC20 module
- (cmd) [\#32](https://github.com/tharsis/evmos/pull/32) Create `testnet` command that spins up a new local testnet with N nodes.

### Improvements

- (deps) [\#94](https://github.com/tharsis/evmos/pull/94) Bump Ethermint version to [`v0.8.0`](https://github.com/tharsis/ethermint/releases/tag/v0.8.0)
- (deps) [\#80](https://github.com/tharsis/evmos/pull/80) Bump ibc-go to [`v2.0.0`](https://github.com/cosmos/ibc-go/releases/tag/v2.0.0)

## [v0.1.3] - 2021-10-24

### Improvements

- (deps) [\#64](https://github.com/tharsis/evmos/pull/64) Bump Ethermint version to `v0.7.2`

### Bug Fixes

- (cmd) [\#41](https://github.com/tharsis/evmos/pull/41) Fix `debug` command.

## [v0.1.2] - 2021-10-08

### Improvements

- (deps) [\#34](https://github.com/tharsis/evmos/pull/34) Bump Ethermint version to `v0.7.1`

## [v0.1.1] - 2021-10-07

### Bug Fixes

- (build) [\#30](https://github.com/tharsis/evmos/pull/30) Fix `version` command.

## [v0.1.0] - 2021-10-07

### Improvements

- (cmd) [\#26](https://github.com/tharsis/evmos/pull/26) Use config on genesis accounts.
- (deps) [\#28](https://github.com/tharsis/evmos/pull/28) Bump Ethermint version to `v0.7.0`
