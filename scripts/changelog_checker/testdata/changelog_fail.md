<!--
Some comments at head of file...
-->
# Changelog

## Unreleased

### State Machine Breaking

- (p256-precompile) [#1922](https://github.com/evmos/evmos/pull/1922) Add `secp256r1` curve precompile.
- (distribution-precompile) [#1948](https://github.com/evmos/evmos/pull/1949) Add `ClaimRewards` custom transaction.

### API Breaking

- (inflation) [#2015](https://github.com/evmos/evmos/pull/2015) Rename `inflation` module to `inflation/v1`.
- (ante) [#2078](https://github.com/evmos/evmos/pull/2078) Deprecate legacy EIP-712 ante handler.

### Improvements

- (testnet) [\#1864](https://github.com/evmos/evmos/pull/1864) Add `--base-fee` and `--min-gas-price` flags.
- (stride-outpost) [#1912](https://github.com/evmos/evmos/pull/1912) Add Stride Outpost interface and ABi.

### Bug Fixes

- (evm) [#1801](https://github.com/evmos/evmos/pull/1801) Fixed the problem `gas_used` is 0

### Invalid Category

- (evm) [#1802](https://github.com/evmos/evmos/pull/1802) Fixed the problem `gas_used` is 0.

### Bug Fixes

- (evm) [#1803](https://github.com/evmos/evmos/pull/1803) Fixed the problem `gas_used` is 0.

## [v15.0.0](https://github.com/evmos/evmos/releases/tag/v15.0.0) - 2023-10-31

### API Breaking

- (vesting) [#1862](https://github.com/evmos/evmos/pull/1862) Add Authorization Grants to the Vesting extension.
- (evm) [#1801](https://github.com/evmos/evmos/pull/1801) Fixed the problem `gas_used` is 0.

## [v15.0.0](https://github.com/evmos/evmos/releases/tag/v15.0.0) - 2023-10-31

### API Breaking

- (vesting) [#1862](https://github.com/evmos/evmos/pull/1862) Add Authorization Grants to the Vesting extension.
- malformed entry in changelog
