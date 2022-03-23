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

- [\#342](https://github.com/tharsis/evmos/pull/342) Implement IBC middleware to recover stuck funds

### API Breaking

- [\#415](https://github.com/tharsis/evmos/pull/415) Bump Evmos go version to v3

### Bug Fixes

- (claims) [\#381](https://github.com/tharsis/evmos/pull/381) Fix claim migration and deletion for EVM chains via IBC attestation.
- (claims) [\#374](https://github.com/tharsis/evmos/pull/374) Fix balance invariant in Claims `InitGenesis`
- (erc20) [\#366](https://github.com/tharsis/evmos/issues/366) Delete ERC20 denom map when deleting pair.

### Improvements

- (ibc) [\#412](https://github.com/tharsis/evmos/pull/412) Introduce boilerplate struct for IBC applications.
- (deps) [\#402](https://github.com/tharsis/evmos/pull/402) Bump IBC go to [`v3.0.0`](https://github.com/cosmos/ibc-go/releases/tag/v3.0.0)
- (ibctesting) [\#388](https://github.com/tharsis/evmos/pull/388) Support Cosmos and EVM chains in IBC testing `Coordinator`.
- (claims) [\#385](https://github.com/tharsis/evmos/pull/385) Add claims invariant.
- (inflation) [\#383](https://github.com/tharsis/evmos/pull/383) Add gRPC endpoints for inflation rate and total supply
- (inflation) [\#369](https://github.com/tharsis/evmos/pull/369) Add `enableInflation` parameter.

## [v2.0.1] - 2022-03-06

### Bug Fixes

- (upgrade) [#\363](https://github.com/tharsis/evmos/pull/363) Don't use `GetParams` for upgrades.

## [v2.0.0] - 2022-03-06

### State Machine Breaking

- (claims) Restrict claiming to a list of authorized IBC channels.

### Improvements

- (deps) [\#360](https://github.com/tharsis/evmos/pull/360) Bump Ethermint to [`v0.11.0`](https://github.com/tharsis/ethermint/releases/tag/v0.11.0)
- (deps) [\#282](https://github.com/tharsis/evmos/pull/282) Bump IBC go to [`v3.0.0-rc1`](https://github.com/cosmos/ibc-go/releases/tag/v3.0.0-rc1)

### Bug Fixes

- (erc20) [\#337](https://github.com/tharsis/evmos/pull/337) Ignore errors in ERC20 module's EVM hook.
- (erc20) [\#336](https://github.com/tharsis/evmos/pull/336) Return `nil` for disabled ERC20 module or ERC20 EVM hook.

## [v1.1.2] - 2022-03-06

### Bug Fixes

- (app) [\#354](https://github.com/tharsis/evmos/pull/354) Add v2 version upgrade logic

## [v1.1.1] - 2022-03-04

### Improvements

- (deps) [\#345](https://github.com/tharsis/evmos/pull/345) Bump Ethermint to [`v0.10.2`](https://github.com/tharsis/ethermint/releases/tag/v0.10.2)

### Bug Fixes

- (app) [\#341](https://github.com/tharsis/evmos/pull/341) Return error when `--ledger` flag is passed in CLI

## [v1.1.0] - 2022-03-02

### Bug Fixes

- (ante) [\#318](https://github.com/tharsis/evmos/pull/318) Add authz check in vesting and min commission `AnteHandler` decorators.
- (vesting) [\#317](https://github.com/tharsis/evmos/pull/317) Fix clawback for vested coins.

## [v1.0.0] - 2022-02-28

### State Machine Breaking

- (ante) [\#302](https://github.com/tharsis/evmos/pull/302) Add AnteHandler decorator to enforce global min validator commission rate.
- (app) [\#224](https://github.com/tharsis/evmos/pull/224) Fix power reduction my setting the correct value on app initialization.
- (keys) [\#189](https://github.com/tharsis/evmos/pull/189) Remove support for Tendermint's `secp256k1` keys.
- [\#173](https://github.com/tharsis/evmos/pull/173) Rename `intrarelayer` module to `erc20`
- [\#190](https://github.com/tharsis/evmos/pull/190) Remove governance hook from `erc20` module

### Features

- [\#286](https://github.com/tharsis/evmos/pull/286) Add `x/vesting` module.
- [\#184](https://github.com/tharsis/evmos/pull/184) Add claims module for claiming the airdrop tokens.
- [\#183](https://github.com/tharsis/evmos/pull/183) Add epoch module for incentives.
- [\#202](https://github.com/tharsis/evmos/pull/202) Add custom configuration for state sync snapshots and tendermint p2p peers. This introduces a custom `InitCmd` function.
- [\#176](https://github.com/tharsis/evmos/pull/176) Add `x/incentives` module.

### Improvements

- (deps) Bumped Ethermint to [`v0.10.0`](https://github.com/tharsis/ethermint/releases/tag/v0.10.0)
- (deps) Bumped IBC-go to `v3.0.0-rc0`
- (deps) Bumped Cosmos SDK to [`v0.45.1`](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.1)
- (deps) bumped Tendermint to `v0.34.15` and tm-db to `v0.6.6`.
- [\#180](https://github.com/tharsis/evmos/pull/180) Delete `TokenPair` if ERC20 contract has been selfdestructed.

### Bug Fixes

- (erc20) [\#169](https://github.com/tharsis/evmos/pull/169) Fixes several testnet bugs:
  - Check if supply exists for a token before when submitting a `RegisterCoinProposal`, allowing users to create an ERC20 representation of an invalid Cosmos Coin.
  - Sanitize the ERC20 token name when creating coin metadata on ER `RegisterERC20Proposal`.
  - Fix coin metadata validation error when registering an ERC20 with 0 denom units.
- (erc20) [\#191](https://github.com/tharsis/evmos/pull/191) Add direct balance protection (IF-ETHERMINT-06).
- (erc20) [\#192](https://github.com/tharsis/evmos/pull/192) Add delayed malicious effect protection (IF-ETHERMINT-06).
- (erc20) [\#200](https://github.com/tharsis/evmos/pull/200) Match coin and token decimals for erc20 deployment during registerCoin
- (erc20) [\#201](https://github.com/tharsis/evmos/pull/201) bug(erc-20): Compile built-in contracts in the build process (IF-ETHERMINT-02).

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
