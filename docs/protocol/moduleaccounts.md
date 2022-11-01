<!--
order: 1
-->

# Module Accounts

Some modules have their own module account. Think of this as a wallet that can only be controlled by that module.
Below is a table of modules, their respective wallet addresses and permissions.

### Account Permisions and their meaning

The `burner` permission means this account has the permission to burn or destroy tokens.
The `minter` permission means this account has permission to mint or create new tokens.
The `staking` permission means this account has permission to stake tokens on behalf of it's owner.

| Name                    | Address                                             | Permissions        |
| :---------------------- | :-------------------------------------------------- | :----------------- |
| `claims`                | [evo15cvq3ljql6utxseh0zau9m8ve2j8erz89m5wkz](https://www.mintscan.io/evoblock/account/evo15cvq3ljql6utxseh0zau9m8ve2j8erz89m5wkz)   | `none`             |
| `erc20`                 | [evo1glht96kr2rseywuvhhay894qw7ekuc4qg9z5nw](https://www.mintscan.io/evoblock/account/evo1glht96kr2rseywuvhhay894qw7ekuc4qg9z5nw)   | `minter` `burner`  |
| `fee_collector`         | [evo17xpfvakm2amg962yls6f84z3kell8c5ljcjw34](https://www.mintscan.io/evoblock/account/evo17xpfvakm2amg962yls6f84z3kell8c5ljcjw34)   | `none`             |
| `incentives`            | [evo1krxwf5e308jmclyhfd9u92kp369l083wn67k4q](https://www.mintscan.io/evoblock/account/evo1krxwf5e308jmclyhfd9u92kp369l083wn67k4q)   | `minter` `burner`  |
| `inflation`             | [evo1d4e35hk3gk4k6t5gh02dcm923z8ck86qygxf38](https://www.mintscan.io/evoblock/account/evo1d4e35hk3gk4k6t5gh02dcm923z8ck86qygxf38)   | `minter`           |
| `transfer`              | [evo1yl6hdjhmkf37639730gffanpzndzdpmhv788dt](https://www.mintscan.io/evoblock/account/evo1yl6hdjhmkf37639730gffanpzndzdpmhv788dt)   | `minter` `burner`  |
| `bonded_tokens_pool`    | [evo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3h6cprl](https://www.mintscan.io/evoblock/account/evo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3h6cprl)   | `burner` `staking` |
| `not_bonded_tokens_pool`| [evo1tygms3xhhs3yv487phx3dw4a95jn7t7lr6ys4t](https://www.mintscan.io/evoblock/account/evo1tygms3xhhs3yv487phx3dw4a95jn7t7lr6ys4t)   | `burner` `staking` |
| `gov`                   | [evo10d07y265gmmuvt4z0w9aw880jnsr700jcrztvm](https://www.mintscan.io/evoblock/account/evo10d07y265gmmuvt4z0w9aw880jnsr700jcrztvm)   | `burner`           |
| `distribution`          | [evo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8974jnh](https://www.mintscan.io/evoblock/account/evo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8974jnh)   | `none`             |
