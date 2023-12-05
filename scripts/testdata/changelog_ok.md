<!--
Some comments at head of file...
-->
# Changelog

## Unreleased

### State Machine Breaking

- (p256-precompile) [#1922](https://github.com/evmos/evmos/pull/1922) [EIP-7212](https://eips.ethereum.org/EIPS/eip-7212) `secp256r1` curve precompile.
- (distribution-precompile) [#1949](https://github.com/evmos/evmos/pull/1949) Add `ClaimRewards` custom transaction.

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
- (staking) [#2105](https://github.com/evmos/evmos/pull/2105) Detect the length of the ed25519 pubkey in precompile CreateValidator to prevent panic.
- (p256-precompile) [#2110](https://github.com/evmos/evmos/pull/2110) Adjust `p256` precompile address.
- (upgrade) [#2117](https://github.com/evmos/evmos/pull/2117) Enable Stride and Osmosis outposts in v16 upgrade handler.
- (osmosis-outpost) [#2109](https://github.com/evmos/evmos/pull/2109) Add Osmosis outpost to available EVM extensions.

### Bug Fixes

- (evm) [#1801](https://github.com/evmos/evmos/pull/1801) Fixed the problem gas_used is 0 when using evm type tx to transfer token to module account.
- (evm) [#1943](https://github.com/evmos/evmos/pull/1943) Fix gas estimation (`eth_estimateGas`) to be consistent with gas used in EVM extensions transactions.
- (test) [#1989](https://github.com/evmos/evmos/pull/1989) Fix the problem about deliverTxSimulate in test app/ante/cosmos/min_price_test.go
- (db) [#2072](https://github.com/evmos/evmos/pull/2072) Change VersionDb directory permission from 777 (insecure) to 750 (general)

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