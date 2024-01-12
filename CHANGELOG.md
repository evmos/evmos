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

- (p256-precompile) [#1922](https://github.com/evmos/evmos/pull/1922) [EIP-7212](https://eips.ethereum.org/EIPS/eip-7212) `secp256r1` curve precompile.
- (distribution-precompile) [#1949](https://github.com/evmos/evmos/pull/1949) Add `ClaimRewards` custom transaction.
- (bech32-precompile) [#2038](https://github.com/evmos/evmos/pull/2038) Add `bech32` conversion precompile.
- (bech32-precompile) [#2124](https://github.com/evmos/evmos/pull/2124) Activate `bech32` precompile.
- (staking-precompile) [#2030](https://github.com/evmos/evmos/pull/2030) Implement the `CreateValidator` function for staking precompiled contract.
- (fees) [#1998](https://github.com/evmos/evmos/pull/1998) Restrict transaction fee tokens.
- (staking-precompile) [#2053](https://github.com/evmos/evmos/pull/2053) Change the validator address in the events from string type to address type.
- (distribution-precompile) [#2055](https://github.com/evmos/evmos/pull/2055) Change the validator address in the events from string type to address type.
- (recovery) [#2087](https://github.com/evmos/evmos/pull/2087) Remove `x/recovery` module.
- (bank-precompile) [#2095](https://github.com/evmos/evmos/pull/2095) Add `bank` precompile.
- (incentives) [#2070](https://github.com/evmos/evmos/pull/2070) Remove `x/incentives` module and burn incentives pool balance.
- (evm) [#2084](https://github.com/evmos/evmos/pull/2084) Remove `x/claims` params and migrate the `EVMChannels` param to the `x/evm` module params.
- (post) [#2128](https://github.com/evmos/evmos/pull/2128) Add `BurnDecorator` to `PostHandler` to burn cosmos transaction fees.
- (fee-collector) [#2129] (https://github.com/evmos/evmos/pull/2129) Add `Burner` role to `FeeCollector` module.
- (erc20) [#2154] (https://github.com/evmos/evmos/pull/2154) Remove evm hooks from `erc20` module.
- (erc20) [#2155] (https://github.com/evmos/evmos/pull/2155) Remove CLI commands for register and convert Coin.
- (erc20) [#2157] (https://github.com/evmos/evmos/pull/2157) Remove `RegisterCoinProposal` logic, protos and related tests.
- (osmosis-outpost) [#2206](https://github.com/evmos/evmos/pull/2206) Update swap function input to custom struct.
- (stride-outpost) [#2207](https://github.com/evmos/evmos/pull/2207) Update Stride Outpost to include additional arguments.
- (incentives) [#2221](https://github.com/evmos/evmos/pull/2221) Burn the usage incentives pool balance during v16 upgrade.
- (erc20) [#2217](https://github.com/evmos/evmos/pull/2217) Add logic to deploy token pairs contracts on genesis.

### API Breaking

- (inflation) [#2015](https://github.com/evmos/evmos/pull/2015) Rename `inflation` module to `inflation/v1`.
- (ante) [#2078](https://github.com/evmos/evmos/pull/2078) deprecate legacy EIP712 ante handler

### Improvements

- (testnet) [#1864](https://github.com/evmos/evmos/pull/1864) Add `--base-fee` and `--min-gas-price` flags to the command `evmosd testnet init-files`.
- (stride-outpost) [#1912](https://github.com/evmos/evmos/pull/1912) Add Stride Outpost interface and ABI.
- (stride-outpost) [#1913](https://github.com/evmos/evmos/pull/1913) Add Run function, Precompile struct and tx definitions.
- (stride-outpost) [#1914](https://github.com/evmos/evmos/pull/1914) Add types, events and common util function.
- (osmosis-outpost) [#1915](https://github.com/evmos/evmos/pull/1915) Add Osmosis Outpost interface and ABI.
- (ics20-precompile) [#1916](https://github.com/evmos/evmos/pull/1916) Make ICS20 Transfer event a common function.
- (ics20-precompile) [#1917](https://github.com/evmos/evmos/pull/1917) Make timeout height a const in the ics20 precompile.
- (stride-outpost) [#1926](https://github.com/evmos/evmos/pull/1926) Refactor event names and definitions.
- (stride-outpost) [#1931](https://github.com/evmos/evmos/pull/1931) Add Stride unit testing setup.
- (stride-outpost) [#1932](https://github.com/evmos/evmos/pull/1932) Add transaction implementation and events unit tests.
- (stride-outpost) [#1935](https://github.com/evmos/evmos/pull/1935) Refactor RedeemStake method.
- (osmosis-outpost) [#1921](https://github.com/evmos/evmos/pull/1921) Add Osmosis outpost types and errors.
- (rpc) [#1952](https://github.com/evmos/evmos/pull/1952) Add tests for EVM extensions transactions gas estimation (related to changes in [#1943](https://github.com/evmos/evmos/pull/1943)).
- (distribution-precompile) [#1954](https://github.com/evmos/evmos/pull/1954) Add `ClaimRewards` unit and event tests.
- (osmosis-outpost) [#1944](https://github.com/evmos/evmos/pull/1944) Add more validation to Osmosis outpost.
- (precompiles) [#1973](https://github.com/evmos/evmos/pull/1973) Use `LoadABI` from precompiles common package in outposts.
- (erc20-precompile) [#1983](https://github.com/evmos/evmos/pull/1983) Add ERC-20 Precompile queries.
- (erc20-precompile) [#1984](https://github.com/evmos/evmos/pull/1984) Add tests for ERC20 precompile events.
- (ci) [#1985](https://github.com/evmos/evmos/pull/1985) Add CI action to check for a Changelog entry on PRs to `main`.
- (erc20-precompile) [#1983](https://github.com/evmos/evmos/pull/1983) Add ERC-20 Precompile queries.
- (osmosis-outpost) [#1986](https://github.com/evmos/evmos/pull/1986) Add Osmosis outpost transaction.
- (werc20-precompile) [#1990](https://github.com/evmos/evmos/pull/1990) Add WERC-20 Precompile events.
- (werc20-precompile) [#1991](https://github.com/evmos/evmos/pull/1991) Add WERC-20 Precompile transactions.
- (distribution-precompile) [#1992](https://github.com/evmos/evmos/pull/1992) Remove outdated check utility for distribution approval.
- (erc20-precompile) [#1993](https://github.com/evmos/evmos/pull/1993) Add ERC-20 Precompile transactions.
- (erc20-precompile) [#1994](https://github.com/evmos/evmos/pull/1994) Add tests for ERC20 precompile type utilities.
- (erc20-precompile) [#1995](https://github.com/evmos/evmos/pull/1995) Add ERC-20 precompile approvals and authorizations.
- (erc20-precompile) [#1997](https://github.com/evmos/evmos/pull/1997) Add logic for ERC-20 precompile registration.
- (erc20-precompile) [#2005](https://github.com/evmos/evmos/pull/2005) Add tests for ERC20 precompile approvals.
- (bin) [#2007](https://github.com/evmos/evmos/pull/2007) Add commands in Makefile for building binary with rocksDB and pebbleDB
- (erc20-precompile) [#2009](https://github.com/evmos/evmos/pull/2009) Add ERC20 precompile transaction unit tests.
- (osmosis-outpost) [#2011](https://github.com/evmos/evmos/pull/2011) Update outpost swap ABI to return IBC next sequence.
- (utils) [#2010](https://github.com/evmos/evmos/pull/2010) Add utils function to create ibc denom trace.
- (erc20-precompile) [#2012](https://github.com/evmos/evmos/pull/2012) Adjust ERC20 extension approvals to handle multiple denominations.
- (osmosis-outpost) [#2017](https://github.com/evmos/evmos/pull/2017) Refactor types, errors and precompile struct.
- (erc20-precompile) [#2023](https://github.com/evmos/evmos/pull/2023) Add tests for ERC20 precompile queries.
- (osmosis-outpost) [#2025](https://github.com/evmos/evmos/pull/2025) Use a struct to wrap parsed parameters from Solidity.
- (ante) [#2028](https://github.com/evmos/evmos/pull/2028) MonoAnteHandler for EVM transaction.
- (erc20-precompile) [#2037](https://github.com/evmos/evmos/pull/2037) Add IsTransactions and RequiredGas tests for the ERC20 extension.
- (bank-precompile) [#2040](https://github.com/evmos/evmos/pull/2040) Add bank extension unit tests for queries.
- (bank-precompile) [#2041](https://github.com/evmos/evmos/pull/2041) Add `supplyOf` query to bank extension.
- (staking-precompile) [#2046](https://github.com/evmos/evmos/pull/2046) Format any type ConsensusPubkey into a base64 string.
- (bank-precompile) [#2045](https://github.com/evmos/evmos/pull/2045) Add unit tests for `supplyOf` query.
- (osmosis-outpost) [#2042](https://github.com/evmos/evmos/pull/2042) Add Osmosis's wasm hook validation functions to test
- (staking-precompile) [#2050](https://github.com/evmos/evmos/pull/2050) Add CreateValidator unit test for precompiled contract staking.
- (make) [#2052](https://github.com/evmos/evmos/pull/2052) Fix Makefile targets to compile ERC20 contracts.
- (werc20-precompile) [#2057](https://github.com/evmos/evmos/pull/2057) WERC20 refactors and handling of fallback and receive functions.
- (werc20-precompile) [#2059](https://github.com/evmos/evmos/pull/2059) Add WERC20 base integration tests.
- (werc20-precompile) [#2062](https://github.com/evmos/evmos/pull/2062) Remove name checking for `fallback` and `receive` functions.
- (osmosis-outpost) [#2063](https://github.com/evmos/evmos/pull/2063) Check that receiver address is bech32 with "evmos" as human readable part.
- (cmn-precompile) [#2064](https://github.com/evmos/evmos/pull/2064) Handle all `fallback` and `receive` function cases
- (erc20-precompile) [#2066](https://github.com/evmos/evmos/pull/2066) Adjust ERC20 EVM extension allowance behavior to align with standard ERC20 smart contracts.
- (erc20-precompile) [#2067](https://github.com/evmos/evmos/pull/2067) Further alignments between ERC20 smart contracts and EVM extension.
- (stride-outpost) [#2069](https://github.com/evmos/evmos/pull/2069) Refactor `osmosis` and `stride` testing.
- (erc20-precompile) [#2073](https://github.com/evmos/evmos/pull/2073) Align metadata query errors with ERC20 contracts.
- (werc20-precompile) [#2074](https://github.com/evmos/evmos/pull/2074) Add `werc20` EVM Extension acceptance tests with `WEVMOS` contract.
- (erc20-precompile) [#2075](https://github.com/evmos/evmos/pull/2075) Align allowance adjustment errors with ERC20 contracts.
- (erc20-precompile) [#2080](https://github.com/evmos/evmos/pull/2080) Add ERC20 integration test setup.
- (erc20-precompile) [#2081](https://github.com/evmos/evmos/pull/2081) Add ERC20 query integration tests.
- (erc20-precompile) [#2083](https://github.com/evmos/evmos/pull/2083) Add integration tests for ERC20 allowance transactions.
- (erc20-precompile) [#2085](https://github.com/evmos/evmos/pull/2085) Add ERC20 transfer integration tests.
- (erc20-precompile) [#2086](https://github.com/evmos/evmos/pull/2086) Add ERC20 metadata query integration tests.
- (erc20-precompile) [#2088](https://github.com/evmos/evmos/pull/2088) Emit additional approval event in case of `transferFrom`.
- (erc20-precompile) [#2097](https://github.com/evmos/evmos/pull/2097) Adjust ERC20 allowance behavior for same spender.
- (bank-precompile) [#2096](https://github.com/evmos/evmos/pull/2096) Add `bank` precompile integration tests.
- (osmosis-outpost) [#2077](https://github.com/evmos/evmos/pull/2077) Update Osmosis outpost to use cross-chain swap contract V1.
- (app) [#2104](https://github.com/evmos/evmos/pull/2104) Refactor code to use `sdkmath.Int` and `sdkmath.LegacyDec` instead of the SDK types
- (staking-precompile) [#2105](https://github.com/evmos/evmos/pull/2105) Detect the length of the ed25519 pubkey in precompile `CreateValidator` to prevent panic.
- (p256-precompile) [#2110](https://github.com/evmos/evmos/pull/2110) Adjust `p256` precompile address.
- (upgrade) [#2117](https://github.com/evmos/evmos/pull/2117) Enable Stride and Osmosis outposts in v16 upgrade handler.
- (osmosis-outpost) [#2109](https://github.com/evmos/evmos/pull/2109) Add Osmosis outpost to available EVM extensions.
- (osmosis-outpost) [#2029](https://github.com/evmos/evmos/pull/2029) Add Osmosis outpost end-to-end tests.
- (upgrade) [#2131](https://github.com/evmos/evmos/pull/2131) Remove incentives pool burning logic from upgrade handler.
- (staking-precompile) [#2122](https://github.com/evmos/evmos/pull/2122) Replace bech32 address with EVM hex address for `CreateValidator` function and remove delegator address argument.
- (inflation) [#2137](https://github.com/evmos/evmos/pull/2137) Reduce daily inflation by 2/3s.
- (app) [#2138](https://github.com/evmos/evmos/pull/2138) Replace imports of store related types and functions to use Cosmos-SDK `store/types` package. 
- (app) [#2139](https://github.com/evmos/evmos/pull/2139) Remove old upgrade handlers logic.
- (outpost) [#2171](https://github.com/evmos/evmos/pull/2171) Add channelID selector based on the ChainID.
- (db) [#2182](https://github.com/evmos/evmos/pull/2182) Bump RocksDB version to `v8.8.1`.
- (api) [#2198] (https://github.com/evmos/evmos/pull/2198) Remove deprecated `RegisterCoinProposal` from store to avoid breaking the governance proposals query.
- (outposts) [#2215](https://github.com/evmos/evmos/pull/2215) backport handle cases for input and output denoms without token pair lookup
- (cli) [#2212](https://github.com/evmos/evmos/pull/2212) Remove ValidateBasic in cli.
- (swagger) [#2219](https://github.com/evmos/evmos/pull/2219) Update Swagger configuration to remove outdated entries and update vesting module version.
- (outposts) [#2223](https://github.com/evmos/evmos/pull/2223) Update Outposts struct documentation and `ValidateBasic`.
- (p256-precompile) [#2228](https://github.com/evmos/evmos/pull/2228) Adjust p256 precompile address from `0x0b` to `0x100`.
- (distribution-precompile) [#2240](https://github.com/evmos/evmos/pull/2240) Replace hardcoded expected balance in distribution precompile tests with queried balance.
- (staking-precompile) [#2234](https://github.com/evmos/evmos/pull/2234) Fix wrong error messages in `NewMsgCreateValidator`.
- (metrics) [#2246](https://github.com/evmos/evmos/pull/2246) Add burned cosmos transactions fees metric.

### Bug Fixes

- (evm) [#1801](https://github.com/evmos/evmos/pull/1801) Fixed the problem gas_used is 0 when using evm type tx to transfer token to module account.
- (evm) [#1943](https://github.com/evmos/evmos/pull/1943) Fix gas estimation (`eth_estimateGas`) to be consistent with gas used in EVM extensions transactions.
- (test) [#1989](https://github.com/evmos/evmos/pull/1989) Fix the problem about deliverTxSimulate in test app/ante/cosmos/min_price_test.go
- (db) [#2072](https://github.com/evmos/evmos/pull/2072) Change VersionDb directory permission from 777 (insecure) to 750 (general)
- (api) [#2196](https://github.com/evmos/evmos/pull/2196) Remove governance proposals related to the deprecated `x/incentives` module to fix the governance proposals query.
- (swagger) [#2218](https://github.com/evmos/evmos/pull/2218) Use correct version of proto dependencies to generate swagger

## [v15.0.0] - 2023-10-31

### API Breaking

- (vesting) [#1862](https://github.com/evmos/evmos/pull/1862) Add Authorization Grants to the Vesting extension.
- (ics20) [#1848](https://github.com/evmos/evmos/pull/1848) Refactor ICS20 Authorization and remove Revoke Event.
- (staking) [#2076](https://github.com/evmos/evmos/pull/2076) Replace bech32 address with evm hex address for query validator.

### State Machine Breaking

- (staking)[#1734](https://github.com/evmos/evmos/pull/1734) Return single struct from staking precompile queries.
- (deps) [#1780](https://github.com/evmos/evmos/pull/1780) Bump ibc-go version to `v7.3.0`.
- (upgrade) [#1845](https://github.com/evmos/evmos/pull/1845) Include remaining strategic reserve migrations.
- (imp) [#1847](https://github.com/evmos/evmos/pull/1847) Remove crisis module in v15 upgrade handler.
- (evm) [#1851](https://github.com/evmos/evmos/pull/1851) Enable [EIP 3855](https://eips.ethereum.org/EIPS/eip-3855) (`PUSH0` opcode) during upgrade.
- (sdk) [#1869](https://github.com/evmos/evmos/pull/1869) Bump Cosmos SDK to v0.47.5.
- (evm) [#1900](https://github.com/evmos/evmos/pull/1900) Enable [EIP 3855](https://eips.ethereum.org/EIPS/eip-3855) (`PUSH0` opcode) by default.
- (db) [1964](https://github.com/evmos/evmos/pull/1964) Bump versionDB dependecy to fix memory store bug.

### Improvements

- (ics20) [#1850](https://github.com/evmos/evmos/pull/1850) Extract common Grant checking and accepting methods.
- (ics20) [#1849](https://github.com/evmos/evmos/pull/1849) Extract common approval methods for ICS20 Authorizations.
- (test) [#1728](https://github.com/evmos/evmos/pull/1728) Add integration test suite using network methods.
- (ci) [#1725](https://github.com/evmos/evmos/pull/1725) Add nix integration test setup to CI flow
- (evm) [#1737](https://github.com/evmos/evmos/pull/1737) Update EVM extensions file name to match interface naming convention.
- (app) [#1842](https://github.com/evmos/evmos/pull/1842) Add support for [MemIAVL](https://github.com/crypto-org-chain/cronos/wiki/MemIAVL) and [versionDB](https://github.com/crypto-org-chain/cronos/blob/main/versiondb/README.md).
- (upgrade) [#1834](https://github.com/evmos/evmos/pull/1834) Improve v14 migration tests and utilities.
- (app) [#1835](https://github.com/evmos/evmos/pull/1835) Remove migration logic from the app's `BeginBlocker`.
- (tests) [#1805](https://github.com/evmos/evmos/pull/1805) Improve local node script by using predefined keys and adding configuration flags.
- (docker) [#1743](https://github.com/evmos/evmos/pull/1743) Add rclone binary to Docker image.
- (config) [#1893](https://github.com/evmos/evmos/pull/1893) Add default config for MemIAVL.

### Bug Fixes

- (consensus) [#1740](https://github.com/evmos/evmos/pull/1740) Enable setting block gas limit to max by specifying it as -1 in the genesis file.
- (ante) [#1753](https://github.com/evmos/evmos/pull/1753) Handle zero fee case on evm transactions.
- (rpc) [#1829](https://github.com/evmos/evmos/pull/1829) Bump IAVL to v0.20.1 to fix concurrency issue
- (testnet) [#1857](https://github.com/evmos/evmos/pull/1857) Remove the crisis module causing an error when using the `evmosd testnet init-files` command.
- (rpc) [#1863](https://github.com/evmos/evmos/pull/1863) Handle error gracefully on RPC calls when node is not persisting ABCI responses.
- (ibc) [#1918](https://github.com/evmos/evmos/pull/1918) Upgrade ibc-go to `v7.3.1`, which (among other things) fixes the `DenomTraces` REST endpoint.
- (gov) [#1981](https://github.com/evmos/evmos/pull/1981) Remove deprecated `cosmos.params.v1beta1/ParameterChangeProposal` handler
- (revenue) [#2032](https://github.com/evmos/evmos/pull/2032) Fixed the problem that users cannot send transactions with gasPrice of 0 to precompiled contracts.

## [v14.1.0] - 2023-09-25

### Bug Fixes

- (upgrade) [#1803](https://github.com/evmos/evmos/pull/1803) Fix the upgrade procedure on v14.0.0.

## [v14.0.0] - 2023-09-19

### State Machine Breaking

- (vesting) [#1754](https://github.com/evmos/evmos/pull/1754) Implement further vesting module refactors.
- (deps) [#1732](https://github.com/evmos/evmos/pull/1732) Bump ibc-go version with error message fix.
- (vesting) [#1730](https://github.com/evmos/evmos/pull/1730) Remove smart contract conversion to `ClawbackVestingAccount`
- (evm) [#1727](https://github.com/evmos/evmos/pull/1727) Return an error when calling inactive EVM extensions
- (deps) [#1662](https://github.com/evmos/evmos/pull/1662) Bump Cosmos-SDK to v0.47.4 and ibc-go to v7.2.0.

### Improvements

- (gov) [#1791](https://github.com/evmos/evmos/pull/1791) Extend maximum proposal metadata length.
- (cli) [#1786](https://github.com/evmos/evmos/pull/1786) Add `block` CLI command to query a block from local db.
- (cli) [#1714](https://github.com/evmos/evmos/pull/1714) Use empty string as default value in `chain-id` flag to use the chain id from the genesis file when not specified.
- (vesting) [#1709](https://github.com/evmos/evmos/pull/1709) Add clawed back coins to `MsgClawbackResponse`
- (vesting) [#1708](https://github.com/evmos/evmos/pull/1708) Minor improvements to `vesting` module
- (cli) [#1706](https://github.com/evmos/evmos/pull/1706) Update `DefaultGasAdjustment` factor used in transactions.
- (staking) [#1702](https://github.com/evmos/evmos/pull/1702) Change authorization names to `grantee` / `granter` in the `staking` precompile
- (ics20) [#1688](https://github.com/evmos/evmos/pull/1688) Change authorization names to `grantee` / `granter` in the `ICS20` precompile
- (mod) [#1687](https://github.com/evmos/evmos/pull/1687) Bump Evmos version to v14.
- (proto) [#1684](https://github.com/evmos/evmos/pull/1684) Update proto files to use `evmos/v14`
- (deps) [#1682](https://github.com/evmos/evmos/pull/1682) Migrate `evmos-ledger-go` logic to this repository.
- (cli) [#1677](https://github.com/evmos/evmos/pull/1677) Update docs for `vesting` cli
- (mod) [#1674](https://github.com/evmos/evmos/pull/1674) Update `evmos` module name to `evmos/v14`
- (vesting)[#1672](https://github.com/evmos/evmos/pull/1672) Port `vesting` precompile code and refactor integration tests
- (vesting)[#1667](https://github.com/evmos/evmos/pull/1667) Add support for vesting precompile in the `evm` module
- (cli) [#1647](https://github.com/evmos/evmos/pull/1647) Update defaults on `evmosd start` flags.
- (vesting) Refactor vesting flow

### Bug Fixes

- (deps) [#1718](https://github.com/evmos/evmos/pull/1718) Update rosetta types import.
- (proto) [#1713](https://github.com/evmos/evmos/pull/1713) Add proto file for v1 vesting module account
- (evm) [#1703](https://github.com/evmos/evmos/pull/1703) Prevent panic on uint64 conversion in EVM keeper `ApplyMessageWithConfig` function.
- (ante) [#1693](https://github.com/evmos/evmos/pull/1693) Prevent panic on int64 conversion in EVM fees antehandler.
- (cli) [#1681](https://github.com/evmos/evmos/pull/1681) Add `bootstrap-state` command.
- (e2e) [#1678](https://github.com/evmos/evmos/pull/1678) Fix e2e tests after recent changes to `evmosd start` default flags
- (rpc) [#1676](https://github.com/evmos/evmos/pull/1676) Fix gas meter stacking gas from predecessors in `TraceTx` & `TraceBlock` functions.
- (rpc) [#1663](https://github.com/evmos/evmos/pull/1663) Fix block number returned in opcode for debug trace related api.
- (revenue) [#1659](https://github.com/evmos/evmos/pull/1659) Check if DevelopersShares are set to 0
- (rpc) [#1655](https://github.com/evmos/evmos/pull/1655) Avoid channel get changed when concurrent subscribe happens.
- (rpc) [#1650](https://github.com/evmos/evmos/pull/1650) Fix racing conditions on RPC PubSub logic
- (vesting) Fix vesting bug.

## [v13.0.2] - 2023-07-05

### Bug Fixes

- (upgrade) [#1644](https://github.com/evmos/evmos/pull/1644) Adjust upgrade constants
- (ci) [#1642](https://github.com/evmos/evmos/pull/1642) Fix target folder in GH action to push to the [Buf Schema Registry](https://buf.build/evmos/evmos) upon release

## [v13.0.1] - 2023-07-04

### Improvements

- (deps) [#1635](https://github.com/evmos/evmos/pull/1635) Update cometbft `v0.34.29` with several minor bug fixes and low-severity security-fixes.

## [v13.0.0] - 2023-07-03

### Improvements

- (app) [#1623](https://github.com/evmos/evmos/pull/1623) Adjust default app config to disable all server options
- (app) [#1619](https://github.com/evmos/evmos/pull/1619) Add snapshot commands to CLI
- (revenue) [#1607](https://github.com/evmos/evmos/pull/1607) Route gas fees from calling EVM extensions to community pool
- (docker) [#1606](https://github.com/evmos/evmos/pull/1606) Improve Dockerfile to reduce image size
- (deps) [#1597](https://github.com/evmos/evmos/pull/1597) Bump geth fork
- (deps) [#1595](https://github.com/evmos/evmos/pull/1595) Bump cometbft and goleveldb
- (evm) [#1578](https://github.com/evmos/evmos/pull/1578#) Add support of ICS20 transfer extension
- (test) [#1486](https://github.com/evmos/evmos/pull/1486) Add benchmark tests for `DeductFeeDecorator` and `EthGasConsumeDecorator` ante handler decorators
- (deps) [#1488](https://github.com/evmos/evmos/pull/1488) Bump btcd version to [`v0.23.3`](https://github.com/btcsuite/btcd/releases/tag/v0.23.3)
- (deps) [#1492](https://github.com/evmos/evmos/pull/1492) Bump Cosmos SDK version to [`v0.46.11-alpha.ledger`](https://github.com/evmos/cosmos-sdk/releases/tag/v0.46.11-alpha.ledger) & use cometbft [`v0.34.27`](https://github.com/cometbft/cometbft/releases/tag/v0.34.27) replacement for Tendermint import

### State Machine Breaking

- (evm) [#1625](https://github.com/evmos/evmos/pull/1625) Migrate updated EVM extensions
- (vesting) [#1596](https://github.com/evmos/evmos/pull/1596) Add MsgCreateClawbackVestingAccount period validation
- (evm) [#1535](https://github.com/evmos/evmos/pull/1535) Add EVM extensions support

### Bug Fixes

- (evm) [1602](https://github.com/evmos/evmos/pull/1602) Fixed hard coded BaseDenom and wrong comparison for MaxUint256
- (deps) [#1567](https://github.com/evmos/evmos/pull/1567) Bump cosmos-sdk version to `v0.46.11-alpha.ledger.7`.
  Fix memory leak in `cosmos/iavl` package.

## [v12.1.6] - 2023-07-04

### Improvements

- (deps) [#1635](https://github.com/evmos/evmos/pull/1635) Update cometbft `v0.34.29` with several minor bug fixes and low-severity security-fixes.

## [v12.1.5] - 2023-06-08

### Bug Fixes

- (vesting) [GHSA-2q3r-p2m3-898g](https://github.com/evmos/evmos/commit/39b750cdaf1d69158ab93da85bd43ae4a7da1456) Apply ClawbackVestingAccount Barberry patch & Bump SDK to v0.46.13

## [v12.1.4] - 2023-05-26

### Improvements

- (deps) [#1571](https://github.com/evmos/evmos/pull/1571) Bump IBC-go version to [`v6.1.1`](https://github.com/cosmos/ibc-go/releases/tag/v6.1.1)

### Bug Fixes

- (ci) [#1546](https://github.com/evmos/evmos/pull/1546) Fix docker image push on release action
- (ci) [#1475](https://github.com/evmos/evmos/pull/1475) Fix version of GitHub action to push to the [Buf Schema Registry](https://buf.build/evmos/evmos) upon releases

## [v12.1.3] - 2023-05-24

### Improvements

- (cli) [#1556](https://github.com/evmos/evmos/pull/1556) Add CLI subcommand to debug legacy EIP712 transaction data

### Bug Fixes

- (deps) [#1566](https://github.com/evmos/evmos/pull/1566) Bump cosmos-sdk version to `v0.46.10-ledger.3`.
  Fix memory leak in `cosmos/iavl` package.

## [v12.1.2] - 2023-04-14

### Bug Fixes

- (rpc) [#1431](https://github.com/evmos/evmos/pull/1431) Fix websocket connection id parsing

## [v12.1.1] - 2023-04-14

### Improvements

- (config) [#1513](https://github.com/evmos/evmos/pull/1513) Set default `timeout_commit` to `3s`

## [v12.1.0] - 2023-03-24

### Improvements

- (deps) [#1498](https://github.com/evmos/evmos/pull/1498) Bump Cosmos SDK version to [v0.46.10-ledger.1](https://github.com/evmos/cosmos-sdk/releases/tag/v0.46.10-ledger.1)
- (lint) [#1487](https://github.com/evmos/evmos/pull/1487) Fix lint issues created by new `golangci-lint` version

## [v12.0.0] - 2023-03-23

### State Machine Breaking

- (evm)[#1308](https://github.com/evmos/evmos/pull/1308) Migrate `evm` and `feemarket` types
- (contracts) [#1306](https://github.com/evmos/evmos/pull/1306) Migrate `contracts` directory to evmos repository
- (proto) [#1305](https://github.com/evmos/evmos/pull/1305) Migrate Ethermint proto files
- (ante) [#1266](https://github.com/evmos/evmos/pull/1266) Use `DynamicFeeChecker` for Cosmos txs.
- (ante) [#1403](https://github.com/evmos/evmos/pull/1403) Update `AnteHandler` decorator for `x/authz` messages to run in deliverTx mode
- (eip712) [#1390](https://github.com/evmos/evmos/pull/1390) Refactor EIP-712 message handling to support multiple message schemas
- (ante) [#1405](https://github.com/evmos/evmos/pull/1405) Enable fees to be deducted from unclaimed staking rewards

### API Breaking

- [#1426](https://github.com/evmos/evmos/pull/1426) Move `revenue` module files into `v1` directory.
- [#1355](https://github.com/evmos/evmos/pull/1355) Remove `vm` package from EVM.

### Improvements

- (tests) [#1434](https://github.com/evmos/evmos/pull/1434) Set default staking denom to `aevmos` in `evm` and `feemarket` tests
- (test) [#1402](https://github.com/evmos/evmos/pull/1402) Refactor NewTx function arguments
- (test) [#1415](https://github.com/evmos/evmos/pull/1415) Refactor InvalidTx type and NextFn used in AnteHandler tests
- (vesting) [#1400](https://github.com/evmos/evmos/pull/1400) Add convert vesting account message
- (test) [#1393](https://github.com/evmos/evmos/pull/1393) Move utilities from `tests` folder to `testutil` package
- (test) [\#1391](https://github.com/evmos/evmos/pull/1391) Refactor test files
- (claims) [#1378](https://github.com/evmos/evmos/pull/1378) Validate authorized channels when updating claims params
- (test) [#1348](https://github.com/evmos/evmos/pull/1348) Add query executions to e2e upgrade test suite
- (deps) [#1370](https://github.com/evmos/evmos/pull/1370) Bump Cosmos SDK version to [`v0.46.9-ledger`](https://github.com/evmos/cosmos-sdk/releases/tag/v0.46.9-ledger)
- (deps) [#1370](https://github.com/evmos/evmos/pull/1370) Bump Tendermint version to [`v0.34.26`](https://github.com/informalsystems/tendermint/releases/tag/v0.34.26)
- (evm) [#1354](https://github.com/evmos/evmos/pull/1354) Expose `Context` from the `StateDB` instance.
- (proto)[#1311](https://github.com/evmos/evmos/pull/1311) Also generate common types with `make proto-gen`
- (revenue)[#1153](https://github.com/evmos/evmos/pull/1153) Migrate revenue module event emitting to `TypedEvent`
- (erc20) [#1152](https://github.com/evmos/evmos/pull/1152) Migrate event emitting to `TypedEvent`
- (claims) [#1126](https://github.com/evmos/evmos/pull/1126) Remove old x/params migration logic
- (vesting) [#1155](https://github.com/evmos/evmos/pull/1155) Migrate deprecated event emitting to new `TypedEvent`
- (docs) [#1361](https://github.com/evmos/evmos/pull/1361) Update `vesting` module docs with new behavior for `ClawbackVestingAccounts`
- (evm) [#1349](https://github.com/evmos/evmos/pull/1349) Restrict the Evmos codebase from working with chain IDs other than `9000` and `9001`
- (test) [#1352](https://github.com/evmos/evmos/pull/1352) Deprecate usage of `aphoton` as denomination on tests
- (test) [#1369](https://github.com/evmos/evmos/pull/1369) Refactor code to use `BaseDenom` for simplification
- (cli) [#1371](https://github.com/evmos/evmos/pull/1371) Improve cli error messages
- (ante) [#1380](https://github.com/evmos/evmos/pull/1380) Split vesting decorators between `evm` and `cosmos` packages
- (cli) [#1386](https://github.com/evmos/evmos/pull/1386) Use required fees (i.e `--fees=auto`) as default if fees are not specified
- (test) [#1408](https://github.com/evmos/evmos/pull/1408) Refactor `DeployContract` and `DeployContractWithFactory` functions used for tests
- (test) [#1417](https://github.com/evmos/evmos/pull/1417) Refactor EIP-712 transactions helper functions used on tests
- (ante) [#1468](https://github.com/evmos/evmos/pull/1468) Add TxFeeChecker requirement
- (deps) [#1473](https://github.com/evmos/evmos/pull/1473) Bump Cosmos SDK version to [v0.46.10-alpha.ledger.2](https://github.com/evmos/cosmos-sdk/releases/tag/v0.46.10-alpha.ledger.2)
- (ante) [#1470](https://github.com/evmos/evmos/pull/1470) Improve error message on `DynamicFeeChecker` ante handler
- (test) [#1484](https://github.com/evmos/evmos/pull/1484) Update e2e test: refactor Makefile command and use latest changes for the tests

### Bug Fixes

- (ante) [#1433](https://github.com/evmos/evmos/pull/1433) Add max priority fee check on `FeeChecker`.
- (ci) [#1383](https://github.com/evmos/evmos/pull/1383) Fix go-releaser error when building macOS binaries
- (ante) [#1435](https://github.com/evmos/evmos/pull/1435) Add block gas limit check for cosmos transactions
- (evm) [#1452](https://github.com/evmos/evmos/pull/1452) Consider refund amount on `gasUsed` calculation
- (evm) [#1466](https://github.com/evmos/evmos/pull/1466) Add `gasUsed` field in Ethereum transaction receipt
- (cli) [#1467](https://github.com/evmos/evmos/pull/1467) Rollback fees `auto` flag logic
- (ci) [#1476](https://github.com/evmos/evmos/pull/1476) Fix go-releaser configuration to be consistent with previous version binaries naming
- (upgrade) [#1493](https://github.com/evmos/evmos/pull/1493) Add decay bug affected accounts

## [v11.0.2] - 2023-02-10

### Improvements

- (deps) [#1370](https://github.com/evmos/evmos/pull/1370) Bump Cosmos SDK version to [`v0.46.9-ledger`](https://github.com/evmos/cosmos-sdk/releases/tag/v0.46.9-ledger)
- (deps) [#1370](https://github.com/evmos/evmos/pull/1370) Bump Tendermint version to [`v0.34.26`](https://github.com/informalsystems/tendermint/releases/tag/v0.34.26)
- (deps) [#1374](https://github.com/evmos/evmos/pull/1374) Bump Gin version to [`v1.7.7`](https://github.com/gin-gonic/gin/releases/tag/v1.7.7)
- (ante) [#1382](https://github.com/evmos/evmos/pull/1382) Add `AnteHandler` decorator for `x/authz` messages

## [v11.0.1] - 2023-02-04

### Improvements

- (deps) [#1248](https://github.com/evmos/evmos/pull/1248) Use the Informal Systems Tendermint Core fork

### Bug Fixes

- (deps) [#1342](https://github.com/evmos/evmos/pull/1342) Bump `tendermint` to [`v0.34.25`](https://github.com/informalsystems/tendermint/releases/tag/v0.34.25)

## [v11.0.0] - 2023-01-27

### State Machine Breaking

- (deps) [#1288](https://github.com/evmos/evmos/pull/1288) Bump `ethermint` to [`v0.21.0`](https://github.com/evmos/ethermint/releases/v0.21.0)
- (ica) [#1101](https://github.com/evmos/evmos/pull/1101) Add ICA host submodule
- (inflation) [#1210](https://github.com/evmos/evmos/pull/1210) Delete `EpochMintProvision` from `KVStore` in a migration
- (deps) [\#1196](https://github.com/evmos/evmos/pull/1196) Bump `ibc-go` to [`v6.1.0`](https://github.com/cosmos/ibc-go/releases/tag/v6.1.0)
- (inflation) [#1193](https://github.com/evmos/evmos/pull/1193) Remove `EpochMintProvision` setters and getters to calculate on the fly
- (erc20) [#1100](https://github.com/evmos/evmos/pull/1100) Deprecate usage of `x/params` in `x/erc20`
- (inflation) [#1107](https://github.com/evmos/evmos/pull/1107) Deprecate usage of `x/params` in `x/inflation`
- (incentives) [#1130](https://github.com/evmos/evmos/pull/1130) Deprecate usage of `x/params` in `x/incentives`
- (claims) [#1125](https://github.com/evmos/evmos/pull/1125) Deprecate usage of `x/params` in `x/claims`
- (revenue) [#1129](https://github.com/evmos/evmos/pull/1129) Deprecate usage of `x/params` in `x/revenue`
- (vesting) [#1268](https://github.com/evmos/evmos/pull/1268) Allow usage of vested and unlocked tokens in EVM interactions

### Features

- (upgrade) [#1209](https://github.com/evmos/evmos/pull/1209) Incentivized testnet reward distribution logic.

### Improvements

- (tests) [#1283](https://github.com/evmos/evmos/pull/1283) Enable multiple upgrades for automated upgrade tests
- (deps) [#1279](https://github.com/evmos/evmos/pull/1279) Bump Cosmos SDK version to [`v0.46.8-ledger`](https://github.com/evmos/cosmos-sdk/releases/tag/v0.46.8-ledger)
- (inflation) [#1258](https://github.com/evmos/evmos/pull/1258) Remove unnecessary `Coin` validation and store calls for `Params`

### Bug Fixes

- (app) [#1276](https://github.com/evmos/evmos/pull/1276) Fix store uploader for `x/recovery` module.
- (inflation) [#1259](https://github.com/evmos/evmos/pull/1259) Re-add missing key to not disrupt order in store
- (upgrade) [#1257](https://github.com/evmos/evmos/pull/1257) Add `recovery` module store to `StoreUpgrades`
- (upgrade) [#1252](https://github.com/evmos/evmos/pull/1252) Add account number and sequence to migrated IBC transfer escrow accounts.
- (upgrade) [#1242](https://github.com/evmos/evmos/pull/1242) Fix Ethermint params upgrade
- (ibc) [#1156](https://github.com/evmos/evmos/pull/1156) Migrate IBC transfer escrow accounts to `ModuleAccount` type.
- (upgrade) [#1252](https://github.com/evmos/evmos/pull/1252) Add account number and sequence to migrated IBC transfer escrow accounts.

## [v10.0.1] - 2023-01-03

### Improvements

- (deps) [#1201](https://github.com/evmos/evmos/pull/1201) Bump `ics23/go` to v0.9.0

## [v10.0.0] - 2022-12-28

### State Machine Breaking

- (deps) [#1184](https://github.com/evmos/evmos/pull/1184) Bump Ethermint version to [`v0.20.0-rc5`](https://github.com/evmos/ethermint/releases/tag/v0.20.0-rc5)
- (deps) [\#1176](https://github.com/evmos/evmos/pull/1176) Bump `ibc-go` to [`v5.2.0`](https://github.com/cosmos/ibc-go/releases/tag/v5.2.0)
- (vesting) [\#1070](https://github.com/evmos/evmos/pull/1070) Add Amino encoding support to the vesting module for EIP-712 signing.
- (ante) [#1054](https://github.com/evmos/evmos/pull/1054) Remove validator commission `AnteHandler` decorator and replace it with the new `MinCommissionRate` staking parameter.
- (deps) [\#1041](https://github.com/evmos/evmos/pull/1041) Add ICS-23 dragon-berry replace in `go.mod` as mentioned in the [Cosmos SDK release](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.4)

### API Breaking

- (erc20) [\#914](https://github.com/evmos/evmos/pull/914) Support registering multiple assets on `RegisterCoinProposal` and `RegisterERC20Proposal`

### Features

- (app) [\#1114](https://github.com/evmos/evmos/pull/1114) Add default File store listener for application from [ADR38](https://docs.cosmos.network/v0.47/architecture/adr-038-state-listening)
- (transfer, erc20) [\#1085](https://github.com/evmos/evmos/pull/1085) Added wrapper for ICS-20 `transfer` to automatically convert ERC-20 tokens to native Cosmos coins.

### Improvements

- (tests) [\1194](https://github.com/evmos/evmos/pull/1194) Lint tests so they are consistent with non-test code.
- (deps) [\#1176](https://github.com/evmos/evmos/pull/1176) Bump Cosmos SDK to [`v0.46.7`](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.7)
- (ci) [#1138](https://github.com/evmos/evmos/pull/1138) Add Golang dependency vulnerability checker.
- (docs) [\#1090](https://github.com/evmos/evmos/pull/1090) Add audits page to documentation.
- (vesting) [\#1087](https://github.com/evmos/evmos/pull/1087) Add new `MsgUpdateVestingFunder` that updates the `Funder` field of a given clawback vesting account
- (ibc) [\#1081](https://github.com/evmos/evmos/pull/1081) Added utils functions for ibc denoms.
- (erc20) [\#1059](https://github.com/evmos/evmos/pull/1059) Add util functions (iterator and params) for ERC20 module.
- (go) [\#1039](https://github.com/evmos/evmos/pull/1039) Bump go v1.19
- (test) [#1028](https://github.com/evmos/evmos/pull/1028) Add node upgrade end-to-end test suite.
- (cmd) [\#1027](https://github.com/evmos/evmos/pull/1027) Apply Google CLI Syntax for required and optional args.
- (ante) [\#993](https://github.com/evmos/evmos/pull/993) Re-order AnteHandlers for better performance
- (docs) [\#985](https://github.com/evmos/evmos/pull/985) Specify repo branch name on markdown-link-check configuration.
- (docs) [\#883](https://github.com/evmos/evmos/pull/883) Add Ethereum tx indexer documentation.
- (docs) [\#980](https://github.com/evmos/evmos/pull/980) Fix documentation links to cosmos-sdk docs.
- (cmd) [\#974](https://github.com/evmos/evmos/pull/974) Add `prune` command.
- (cli) [#816](https://github.com/evmos/evmos/pull/816) Add Ledger CLI support.

### Bug Fixes

- (app) [#1165](https://github.com/evmos/evmos/pull/1165) Update Ledger supported algorithms to only consist of `EthSecp256k1`
- (cmd) [#1172](https://github.com/evmos/evmos/pull/1172) Update default node snapshot interval to `5000`
- (cmd) [\#1121](https://github.com/evmos/evmos/pull/1121) Fix `evmosd version` to show either tag or last commit.
- (cmd) [\#1120](https://github.com/evmos/evmos/pull/1120) Fix snapshot configuration
- (app) [\#1118](https://github.com/evmos/evmos/pull/1118) Setup gRPC node service with the application.
- (analytics) [\#1094](https://github.com/evmos/evmos/pull/1094) Fix unbound metrics and remove labels that keep increasing.

## [v9.1.0] - 2022-10-25

### Improvements

- (deps) [\#1011](https://github.com/evmos/evmos/pull/1011) Bump Cosmos SDK to [`v0.45.10`](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.10)

## [v9.0.0] - 2022-10-21

### Bug Fixes

- (claims) [#951](https://github.com/evmos/evmos/pull/951) Fix ClawbackEmptyAccounts logic

## [v8.2.3] - 2022-10-15

### Improvements

- (cmd) [#988](https://github.com/evmos/evmos/pull/988) Set IAVL config
- (cli) [#971](https://github.com/evmos/evmos/pull/971) Add `prune` command.

## [v8.2.2] - 2022-10-14

### Improvements

- (deps)[#965](https://github.com/evmos/evmos/pull/965) Bump SDK to v0.45.9 and Ethermint to v0.19.3

## [v8.2.0] - 2022-09-23

### State Machine Breaking

- (app) [\#918](https://github.com/evmos/evmos/pull/918) Fix unregistered `revenue` module for `v8.1.0` store upgrade

### Bug Fixes

- (app,docs) [\#933](https://github.com/evmos/evmos/pull/933) Replace invalid linux `x86_64` [architecture](https://go.dev/doc/install/source#environment) to `amd64`.

## [v8.1.1] - 2022-09-23

### Bug Fixes

- (app) [\#922](https://github.com/evmos/evmos/pull/922) Add hard fork logic for `v8.2.0`

## [v8.1.0] - 2022-08-30

### State Machine Breaking

- (revenue) [\#859](https://github.com/evmos/evmos/pull/859) Add amino codecs to `x/revenue` module to support EIP-712 signatures.
- (deps) Bump Ethermint version to [`v0.19.2`](https://github.com/evmos/ethermint/releases/tag/v0.19.2)

## [v8.0.0] - 2022-08-16

### State Machine Breaking

- (deps) [\#845](https://github.com/evmos/evmos/pull/845) Bump Ethermint version to [`v0.19.0`](https://github.com/evmos/ethermint/releases/tag/v0.19.0)
- (revenue) Add `x/revenue` module

### Improvements

- (deps) [\#839](https://github.com/evmos/evmos/pull/839) Bump ibc-go to [`v3.2.0`](https://github.com/cosmos/ibc-go/releases/tag/v3.2.0) and Cosmos SDK to [`v0.45.7`](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.7).
- (build) [\#725](https://github.com/evmos/evmos/pull/725) Migrate Protobuf code generator to [Protobuf Workspaces](https://docs.buf.build/reference/workspaces)

### Bug Fixes

- (build) [\#856](https://github.com/evmos/evmos/pull/856) Update docker base image to use golang:1.18.5-bullseye and expose other relevant ports

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
- (revenue) [\#685](https://github.com/evmos/evmos/pull/685) Internal Specification audit.
- (revenue) [\#691](https://github.com/evmos/evmos/pull/691) Internal API audit.
- (revenue) [\#715](https://github.com/evmos/evmos/pull/715) Internal state machine audit.

## [v5.0.0] - 2022-06-14

### State Machine Breaking

- (deps) [\#684](https://github.com/evmos/evmos/pull/684) Bump ibc-go version to [`v3.1.0`](https://github.com/cosmos/ibc-go/releases/tag/v3.1.0)
- (vesting) [\#666](https://github.com/evmos/evmos/pull/666) Remove support of Cosmos SDK `VestingAccount` types.
- (deps) [\#663](https://github.com/evmos/evmos/pull/663) Bump Ethermint version to [`v0.16.1`](https://github.com/evmos/ethermint/releases/tag/v0.16.1)
- (claims) [\#605](https://github.com/evmos/evmos/pull/605) Remove duplicated `SetClaimsRecord`.
- (erc20) [\#602](https://github.com/evmos/evmos/pull/602) Modified `RegisterERC20` proposals.
  Fix erc20 name sanitization to allow spaces on token name.

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
- (revenue) [\#612](https://github.com/evmos/evmos/pull/612) Fix fees registration cli command and description
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
- (p2p) [\#541](https://github.com/evmos/evmos/pull/541) Increase default inbound connections and use 8:1 ratio of inbound:outbound.
  Add default seeds to reduce the need for configuration.
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
- [\#202](https://github.com/evmos/evmos/pull/202) Add custom configuration for state sync snapshots and tendermint p2p peers.
  This introduces a custom `InitCmd` function.
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
