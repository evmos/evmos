<!--
Some comments at head of file...
-->
# Changelog

## Unreleased

### State Machine Breaking

- (p256-precompile) [#1922](https://github.com/evmos/evmos/pull/1922) Add `secp256r1` curve precompile.
- (distribution-precompile) [#1949](https://github.com/evmos/evmos/pull/1949) Add `ClaimRewards` custom transaction.
- (swagger) [#2218](https://github.com/evmos/evmos/pull/2218) Use correct version of proto dependencies to generate swagger.
- (go) [#1687](https://github.com/evmos/evmos/pull/1687) Bump Evmos version to v14.

### API Breaking

- (inflation) [#2015](https://github.com/evmos/evmos/pull/2015) Rename `inflation` module to `inflation/v1`.
- (ante) [#2078](https://github.com/evmos/evmos/pull/2078) Deprecate legacy EIP-712 ante handler.
- (evm) [#1851](https://github.com/evmos/evmos/pull/1851) Enable [EIP 3855](https://eips.ethereum.org/EIPS/eip-3855) (`PUSH0` opcode) during upgrade.

### Improvements

- (testnet) [#1864](https://github.com/evmos/evmos/pull/1864) Add `--base-fee` and `--min-gas-price` flags.
- (stride-outpost) [#1912](https://github.com/evmos/evmos/pull/1912) Add Stride outpost interface and ABI.
- (app) [#2104](https://github.com/evmos/evmos/pull/2104) Refactor to use `sdkmath.Int` and `sdkmath.LegacyDec` instead of SDK types.
- (all) [#701](https://github.com/evmos/evmos/pull/701) Rename Go module to `evmos/evmos`.

### Bug Fixes

- (evm) [#1801](https://github.com/evmos/evmos/pull/1801) Fixed the problem `gas_used` is 0.
- (erc20) [#109](https://github.com/evmos/evmos/pull/109) Fix hardcoded ERC-20 nonce and `UpdateTokenPairERC20` proposal to support ERC-20s with 0 decimals.

## [v15.0.0](https://github.com/evmos/evmos/releases/tag/v15.0.0) - 2023-10-31

### API Breaking

- (vesting) [#1862](https://github.com/evmos/evmos/pull/1862) Add Authorization Grants to the Vesting extension.
- (app) [#555](https://github.com/evmos/evmos/pull/555) `v4.0.0` upgrade logic.

## [v2.0.0](https://github.com/evmos/evmos/releases/tag/v2.0.0) - 2021-10-31

### State Machine Breaking

- legacy entries do not have to be fully correct
