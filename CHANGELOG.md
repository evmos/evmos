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

## [v8.0.0] - 2022-09-16

### State Machine Breaking

- (deps) Bump Ethermint version to [`v0.19.0`](https://github.com/evmos/ethermint/releases/tag/v0.19.0)
- (feesplit) Add `x/feesplit` module

### Improvements

- (deps) [\#839](https://github.com/evmos/evmos/pull/839) Bump ibc-go to [`v3.2.0`](https://github.com/cosmos/ibc-go/releases/tag/v3.2.0) and Cosmos SDK to [`v0.45.7`](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.7).
- (build) [\#725](https://github.com/evmos/evmos/pull/725) Migrate Protobuf code generator to [Protobuf Workspaces](https://docs.buf.build/reference/workspaces)

## [v7.0.0] - 2022-08-04

### State Machine Breaking

- (deps) Bump Ethermint version to [`v0.18.0`](https://github.com/evmos/ethermint/releases/tag/v0.18.0)

### Bug Fixes

- (app) [\#760](https://github.com/evmos/evmos/pull/760) Migrate inaccessible balance of testnet faucet account to new address
- (inflation) [\#748](https://github.com/evmos/evmos/pull/748) Remove overcounted epochs from `skippedEpochs` value in store

## [v6.0.3] - 2022-07-26

### Bug Fixes

- (deps) [\#803](https://github.com/evmos/evmos/pull/803) Bump Ethermint version to [`v0.17.2`](https://github.com/evmos/ethermint/releases/tag/v0.17.2)

## [v6.0.2] - 2022-07-13

### Bug Fixes

- (deps) [\#769](https://github.com/evmos/evmos/pull/769) Bump Ethermint version to [`v0.17.1`](https://github.com/evmos/ethermint/releases/tag/v0.17.1)

## [v6.0.1] - 2022-06-28

### Improvements

- (ci) [\#729](https://github.com/evmos/evmos/pull/729) Remove unshallow action in goreleaser.

## [v6.0.0] - 2022-06-28

### State Machine Breaking

- (deps) [\#719](https://github.com/evmos/evmos/pull/719) Bump Ethermint version to [`v0.17.0`](https://github.com/evmos/ethermint/releases/tag/v0.17.0)

### API Breaking

- (all) [\#701](https://github.com/evmos/evmos/pull/703) Rename Go module to `evmos/evmos`

### Improvements

- (deps) [\#714](https://github.com/evmos/evmos/pull/714) Bump Go version to `1.18`.
- (cmd) [\#696](https://github.com/evmos/evmos/pull/696) Set a custom tendermint node configuration on initialization.
- (feesplit) [\#685](https://github.com/evmos/evmos/pull/685) Internal Specification audit.
- (feesplit) [\#691](https://github.com/evmos/evmos/pull/691) Internal API audit.
- (feesplit) [\#715](https://github.com/evmos/evmos/pull/715) Internal state machine audit.

## [v5.0.0] - 2022-06-14

### State Machine Breaking

- (deps) [\#684](https://github.com/evmos/evmos/pull/684) Bump ibc-go version to [`v3.1.0`](https://github.com/cosmos/ibc-go/releases/tag/v3.1.0)
- (vesting) [\#666](https://github.com/evmos/evmos/pull/666) Remove support of Cosmos SDK `VestingAccount` types.
- (deps) [\#663](https://github.com/evmos/evmos/pull/663) Bump Ethermint version to [`v0.16.1`](https://github.com/evmos/ethermint/releases/tag/v0.16.1)
- (claims) [\#605](https://github.com/evmos/evmos/pull/605) Remove duplicated `SetClaimsRecord`.
- (erc20) [\#602](https://github.com/evmos/evmos/pull/602) Modified `RegisterERC20` proposals. Fix erc20 name sanitization to allow spaces on token name.

### API Breaking

- (claims) [\#605](https://github.com/evmos/evmos/pull/605) Remove `claims-` prefix in CLI query commands.
- (erc20) [\#592](https://github.com/evmos/evmos/pull/592) Finish module completeness audit.
- (analytics) [\#637](https://github.com/evmos/evmos/pull/637) Add telemetry to Evmos modules.
- (vesting) [\#643](https://github.com/evmos/evmos/pull/643) Remove the `create-vesting-account` CLI command from Cosmos SDK in favor of the clawback vesting accounts.

### Improvements

- (erc20) [\#677](https://github.com/evmos/evmos/pull/677) Add Amino registration to `ConvertCoin` and `ConvertERC20` msgs for ERC712 compatibility.
- (deps) [\#668](https://github.com/evmos/evmos/pull/668) Bump Cosmos SDK to [`v0.45.5`](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.5)
- (erc20) [\#642](https://github.com/evmos/evmos/pull/642) Remove enforcing ibc and channel names during `RegisterCoin`

### Bug Fixes

- (app) [\#682](https://github.com/evmos/evmos/pull/682) Fix Tendermint consensus params (Evidence `MaxAgeNumBlocks` and `MaxAgeDuration`)
- (incentives) [\#656](https://github.com/evmos/evmos/pull/656) Fix incentives that were previously only allocated to `EthAccount`s.
- (feesplit) [\#612](https://github.com/evmos/evmos/pull/612) Fix fees registration cli command and description
- (inflation) [\#554](https://github.com/evmos/evmos/pull/554) Changing erroneous epoch skips to `daily` instead of `weekly`
- (claims) [\#626](https://github.com/evmos/evmos/pull/626) fix durations denominated in `nanoseconds`
- (epochs) [\#629](https://github.com/evmos/evmos/pull/629) fix epochs durations denominated in `nanoseconds`

## [v4.0.1] - 2022-05-10

### Bug Fixes

(erc20) [\#588](https://github.com/evmos/evmos/pull/588) Revert PR [\#556](https://github.com/evmos/evmos/pull/556).

## [v4.0.0] - 2022-05-09

### State Machine Breaking

- (app) [\#537](https://github.com/evmos/evmos/pull/537) Fix router key for IBC client proposals.
- (erc20) [\#530](https://github.com/evmos/evmos/pull/530) Use the highest denom unit when deploying an ERC20 contract.

### API Breaking

- (upgrade) [\#557](https://github.com/evmos/evmos/pull/557) Update Evmos go.mod version `v3` -> `v4`
- (erc20) [\#544](https://github.com/evmos/evmos/pull/544) Remove `updateTokenPairERC20Proposal` functionality rename `relay` to `conversion`
- (inflation) [\#536](https://github.com/evmos/evmos/pull/536) Rename inflation endpoint `/evmos/inflation/v1/total_supply` -> `/evmos/inflation/v1/circulating_supply`

### Improvements

- (deps) [\#580](https://github.com/evmos/evmos/pull/580) Bump Ethermint to [`v0.15.0`](https://github.com/evmos/ethermint/releases/tag/v0.15.0)
- (gitpod) [\#564](https://github.com/evmos/evmos/pull/564) Add one-click development environment
- (erc20) [\#556](https://github.com/evmos/evmos/pull/556) Remove deprecated migrations.
- (incentives) [\#551](https://github.com/evmos/evmos/pull/551) Add additional check to only distribute incentives to EOAs.
- (cmd) [\#543](https://github.com/evmos/evmos/pull/543) Update mainnet default `min-gas-price` to `0.0025aevmos`.
- (epochs) [\#539](https://github.com/evmos/evmos/pull/539) Use constants for epoch identifiers.

### Bug Fixes

- (erc20) [\#530](https://github.com/evmos/evmos/pull/530) Fix `Metadata` equal check for denom units.
- (app) [\#523](https://github.com/evmos/evmos/pull/523) Fix testnet upgrade store loader.

## [v3.0.1] - 2022-05-09

### Improvements

- (app) [\#555](https://github.com/evmos/evmos/pull/555) `v4.0.0` upgrade logic.
- (p2p) [\#541](https://github.com/evmos/evmos/pull/541) Increase default inbound connections and use 8:1 ratio of inbound:outbound. Add default seeds to reduce the need for configuration.
- (deps) [\#528](https://github.com/evmos/evmos/pull/528) Bump Cosmos SDK to [`v0.45.4`](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.4)

## [v3.0.0] - 2022-04-22

### State Machine Breaking

- [\#342](https://github.com/evmos/evmos/pull/342) Implement IBC middleware to recover stuck funds

### API Breaking

- [\#415](https://github.com/evmos/evmos/pull/415) Bump Evmos go version to v3

### Bug Fixes

- (vesting) [\#502](https://github.com/evmos/evmos/pull/502) Fix gas exhaustion bug by removing `SpendableCoins` during vesting account clawback.
- (vesting) [\#483](https://github.com/evmos/evmos/pull/483) Fix balance clawback when vesting start time is in the future
- (claims) [\#381](https://github.com/evmos/evmos/pull/381) Fix claim migration and deletion for EVM chains via IBC attestation.
- (claims) [\#374](https://github.com/evmos/evmos/pull/374) Fix balance invariant in Claims `InitGenesis`
- (erc20) [\#366](https://github.com/evmos/evmos/issues/366) Delete ERC20 denom map when deleting pair.
- (claims) [\#505](https://github.com/evmos/evmos/pull/505) Fix IBC attestation ordering

### Improvements

- (vesting) [\#486](https://github.com/evmos/evmos/pull/486) Refactor `x/vesting` types and tests.
- (erc20) [\#484](https://github.com/evmos/evmos/pull/484) Avoid unnecessary commits to the StateDB and don't estimate gas when performing a query
- (deps) [\#478](https://github.com/evmos/evmos/pull/478) Bump Cosmos SDK to [`v0.45.3`](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.3)
- (deps) [\#478](https://github.com/evmos/evmos/pull/478) Bump Ethermint to [`v0.14.0`](https://github.com/evmos/ethermint/releases/tag/v0.14.0)
- (vesting) [\#468](https://github.com/evmos/evmos/pull/468) Use coins `Min` function from Cosmos SDK.
- (cmd) [\#446](https://github.com/evmos/evmos/pull/446) Update `migrate` command to migrate Evmos, Ethermint and Cosmos SDK modules.
- (app) [\#446](https://github.com/evmos/evmos/pull/446) Refactor upgrade code.
- (ibc) [\#412](https://github.com/evmos/evmos/pull/412) Introduce boilerplate struct for IBC applications.
- (deps) [\#402](https://github.com/evmos/evmos/pull/402) Bump IBC go to [`v3.0.0`](https://github.com/cosmos/ibc-go/releases/tag/v3.0.0)
- (ibctesting) [\#388](https://github.com/evmos/evmos/pull/388) Support Cosmos and EVM chains in IBC testing `Coordinator`.
- (claims) [\#385](https://github.com/evmos/evmos/pull/385) Add claims invariant.
- (inflation) [\#383](https://github.com/evmos/evmos/pull/383) Add gRPC endpoints for inflation rate and total supply
- (inflation) [\#369](https://github.com/evmos/evmos/pull/369) Add `enableInflation` parameter.
- (claims) [\#432](https://github.com/evmos/evmos/pull/432) Add IBC trigger amount to claims merge/migrate IBC callbacks.
- (claims) [\#442](https://github.com/evmos/evmos/pull/443) Remove claims merge/migrate cases where sender already completed an action as they are never reached
- (claims) [\#507](https://github.com/evmos/evmos/pull/507) Always return claimable amount on grpc queries regarding of claims status.
- (claims) [\#516](https://github.com/evmos/evmos/pull/516) Retain claims records when all actions have been completed.

## [v2.0.1] - 2022-03-06

### Bug Fixes

- (upgrade) [#\363](https://github.com/evmos/evmos/pull/363) Don't use `GetParams` for upgrades.

## [v2.0.0] - 2022-03-06

### State Machine Breaking

- (claims) Restrict claiming to a list of authorized IBC channels.

### Improvements

- (deps) [\#360](https://github.com/evmos/evmos/pull/360) Bump Ethermint to [`v0.11.0`](https://github.com/evmos/ethermint/releases/tag/v0.11.0)
- (deps) [\#282](https://github.com/evmos/evmos/pull/282) Bump IBC go to [`v3.0.0-rc1`](https://github.com/cosmos/ibc-go/releases/tag/v3.0.0-rc1)

### Bug Fixes

- (erc20) [\#337](https://github.com/evmos/evmos/pull/337) Ignore errors in ERC20 module's EVM hook.
- (erc20) [\#336](https://github.com/evmos/evmos/pull/336) Return `nil` for disabled ERC20 module or ERC20 EVM hook.

## [v1.1.2] - 2022-03-06

### Bug Fixes

- (app) [\#354](https://github.com/evmos/evmos/pull/354) Add v2 version upgrade logic

## [v1.1.1] - 2022-03-04

### Improvements

- (deps) [\#345](https://github.com/evmos/evmos/pull/345) Bump Ethermint to [`v0.10.2`](https://github.com/evmos/ethermint/releases/tag/v0.10.2)

### Bug Fixes

- (app) [\#341](https://github.com/evmos/evmos/pull/341) Return error when `--ledger` flag is passed in CLI

## [v1.1.0] - 2022-03-02

### Bug Fixes

- (ante) [\#318](https://github.com/evmos/evmos/pull/318) Add authz check in vesting and min commission `AnteHandler` decorators.
- (vesting) [\#317](https://github.com/evmos/evmos/pull/317) Fix clawback for vested coins.

## [v1.0.0] - 2022-02-28

### State Machine Breaking

- (ante) [\#302](https://github.com/evmos/evmos/pull/302) Add AnteHandler decorator to enforce global min validator commission rate.
- (app) [\#224](https://github.com/evmos/evmos/pull/224) Fix power reduction my setting the correct value on app initialization.
- (keys) [\#189](https://github.com/evmos/evmos/pull/189) Remove support for Tendermint's `secp256k1` keys.
- [\#173](https://github.com/evmos/evmos/pull/173) Rename `intrarelayer` module to `erc20`
- [\#190](https://github.com/evmos/evmos/pull/190) Remove governance hook from `erc20` module

### Features

- [\#286](https://github.com/evmos/evmos/pull/286) Add `x/vesting` module.
- [\#184](https://github.com/evmos/evmos/pull/184) Add claims module for claiming the airdrop tokens.
- [\#183](https://github.com/evmos/evmos/pull/183) Add epoch module for incentives.
- [\#202](https://github.com/evmos/evmos/pull/202) Add custom configuration for state sync snapshots and tendermint p2p peers. This introduces a custom `InitCmd` function.
- [\#176](https://github.com/evmos/evmos/pull/176) Add `x/incentives` module.

### Improvements

- (deps) Bumped Ethermint to [`v0.10.0`](https://github.com/evmos/ethermint/releases/tag/v0.10.0)
- (deps) Bumped IBC-go to `v3.0.0-rc0`
- (deps) Bumped Cosmos SDK to [`v0.45.1`](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.1)
- (deps) bumped Tendermint to `v0.34.15` and tm-db to `v0.6.6`.
- [\#180](https://github.com/evmos/evmos/pull/180) Delete `TokenPair` if ERC20 contract has been selfdestructed.

### Bug Fixes

- (erc20) [\#169](https://github.com/evmos/evmos/pull/169) Fixes several testnet bugs:
  - Check if supply exists for a token before when submitting a `RegisterCoinProposal`, allowing users to create an ERC20 representation of an invalid Cosmos Coin.
  - Sanitize the ERC20 token name when creating coin metadata on ER `RegisterERC20Proposal`.
  - Fix coin metadata validation error when registering an ERC20 with 0 denom units.
- (erc20) [\#191](https://github.com/evmos/evmos/pull/191) Add direct balance protection (IF-ETHERMINT-06).
- (erc20) [\#192](https://github.com/evmos/evmos/pull/192) Add delayed malicious effect protection (IF-ETHERMINT-06).
- (erc20) [\#200](https://github.com/evmos/evmos/pull/200) Match coin and token decimals for erc20 deployment during registerCoin
- (erc20) [\#201](https://github.com/evmos/evmos/pull/201) bug(erc-20): Compile built-in contracts in the build process (IF-ETHERMINT-02).

## [v0.4.2] - 2021-12-11

### Bug Fixes

- (app) [\#166](https://github.com/evmos/evmos/pull/166) Fix `UpgradeHandler`.

## [v0.4.1] - 2021-12-07

### Improvements

- (build) [\#143](https://github.com/evmos/evmos/pull/143) Added `build-reproducible` rule in `Makefile` to build docker containers

### Bug Fixes

- (build) [\#151](https://github.com/evmos/evmos/pull/151) Fixes `version` command by picking the latest tag in the current branch instead of across all branches as the current version

## [v0.4.0] - 2021-12-02

### State Machine Breaking

- (erc20) [\#119](https://github.com/evmos/evmos/issues/119) Register `x/erc20` proposal types on governance module.

### Improvements

- (app) [\#128](https://github.com/evmos/evmos/pull/128) Add ibc-go `TestingApp` interface.
- (ci) [\#117](https://github.com/evmos/evmos/pull/117) Enable automatic backport of PRs.
- (deps) [\#135](https://github.com/evmos/evmos/pull/135) Bump Ethermint version to [`v0.9.0`](https://github.com/evmos/ethermint/releases/tag/v0.9.0)
- (ci) [\#136](https://github.com/evmos/evmos/pull/136) Deploy `evmos` docker container to [docker hub](https://hub.docker.com/u/tharsishq) for every versioned releases

### Bug Fixes

- (build) [\#116](https://github.com/evmos/evmos/pull/116) Fix `build-docker` command

## [v0.3.0] - 2021-11-24

### API Breaking

- (erc20) [\#99](https://github.com/evmos/evmos/pull/99) Rename `enable_e_v_m_hook` json parameter to `enable_evm_hook`.

### Improvements

- (deps) [\#110](https://github.com/evmos/evmos/pull/110) Bump Ethermint version to [`v0.8.1`](https://github.com/evmos/ethermint/releases/tag/v0.8.1)
- (erc20) [\#107](https://github.com/evmos/evmos/pull/107) Add IBC validation
- (cmd) [\#105](https://github.com/evmos/evmos/pull/105) Improve testnet command to include JSON-RPC client.

## Bug Fixes

- (erc20) [\#109](https://github.com/evmos/evmos/pull/109) Fix hardcoded erc20 nonce and `UpdateTokenPairERC20` proposal to support ERC20s with 0 decimals.
- (erc20) [\#102](https://github.com/evmos/evmos/pull/102) Add `convert-erc20` cmd

## [v0.2.0] - 2021-11-17

### Features

- (erc20) [\#82](https://github.com/evmos/evmos/pull/82) ERC20 module
- (cmd) [\#32](https://github.com/evmos/evmos/pull/32) Create `testnet` command that spins up a new local testnet with N nodes.

### Improvements

- (deps) [\#94](https://github.com/evmos/evmos/pull/94) Bump Ethermint version to [`v0.8.0`](https://github.com/evmos/ethermint/releases/tag/v0.8.0)
- (deps) [\#80](https://github.com/evmos/evmos/pull/80) Bump ibc-go to [`v2.0.0`](https://github.com/cosmos/ibc-go/releases/tag/v2.0.0)

## [v0.1.3] - 2021-10-24

### Improvements

- (deps) [\#64](https://github.com/evmos/evmos/pull/64) Bump Ethermint version to `v0.7.2`

### Bug Fixes

- (cmd) [\#41](https://github.com/evmos/evmos/pull/41) Fix `debug` command.

## [v0.1.2] - 2021-10-08

### Improvements

- (deps) [\#34](https://github.com/evmos/evmos/pull/34) Bump Ethermint version to `v0.7.1`

## [v0.1.1] - 2021-10-07

### Bug Fixes

- (build) [\#30](https://github.com/evmos/evmos/pull/30) Fix `version` command.

## [v0.1.0] - 2021-10-07

### Improvements

- (cmd) [\#26](https://github.com/evmos/evmos/pull/26) Use config on genesis accounts.
- (deps) [\#28](https://github.com/evmos/evmos/pull/28) Bump Ethermint version to `v0.7.0`
