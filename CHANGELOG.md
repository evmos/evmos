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

## API Breaking

* (intrarelayer) [\#99](https://github.com/tharsis/evmos/pull/99) Rename `enable_e_v_m_hook` json parameter to `enable_evm_hook`.

## [v0.2.0] - 2021-11-17

### Features

* (intrarelayer) [\#82](https://github.com/tharsis/evmos/pull/82) Intrarelayer module
* (cmd) [\#32](https://github.com/tharsis/evmos/pull/32) Create `testnet` command that spins up a new local testnet with N nodes.

### Improvements

* (deps) [\#94](https://github.com/tharsis/evmos/pull/94) Bump Ethermint version to [`v0.8.0`](https://github.com/tharsis/ethermint/releases/tag/v0.8.0)
* (deps) [\#80](https://github.com/tharsis/evmos/pull/80) Bump ibc-go to [`v2.0.0`](https://github.com/cosmos/ibc-go/releases/tag/v2.0.0)

## [v0.1.3] - 2021-10-24

### Improvements

* (deps) [\#64](https://github.com/tharsis/evmos/pull/64) Bump Ethermint version to `v0.7.2`

## Bug Fixes

* (cmd) [\#41](https://github.com/tharsis/evmos/pull/41) Fix `debug` command.

## [v0.1.2] - 2021-10-08

### Improvements

* (deps) [\#34](https://github.com/tharsis/evmos/pull/34) Bump Ethermint version to `v0.7.1`

## [v0.1.1] - 2021-10-07

## Bug Fixes

* (build) [\#30](https://github.com/tharsis/evmos/pull/30) Fix `version` command.

## [v0.1.0] - 2021-10-07

## Improvements

* (cmd) [\#26](https://github.com/tharsis/evmos/pull/26) Use config on genesis accounts.
* (deps) [\#28](https://github.com/tharsis/evmos/pull/28) Bump Ethermint version to `v0.7.0`
