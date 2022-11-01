<!--
order: 7
-->

# Clients

A user can query the Evoblock `x/vesting` module using the CLI, gRPC, or REST.

## CLI

Find below a list of `evoblockd` commands added with the `x/vesting` module. You can obtain the full list by using the `evoblockd -h` command.

### Genesis

The genesis configuration commands allow users to configure the genesis `vesting` account state.

`add-genesis-account`

Allows users to set up clawback vesting accounts at genesis, funded with an allocation of tokens, subject to clawback. Must provide a lockup periods file (`--lockup`), a vesting periods file (`--vesting`), or both.

If both files are given, they must describe schedules for the same total amount.
If one file is omitted, it will default to a schedule that immediately unlocks or vests the entire amount. The described amount of coins will be transferred from the --from address to the vesting account. Unvested coins may be "clawed back" by the funder with the clawback command. Coins may not be transferred out of the account if they are locked or unvested. Only vested coins may be staked. For an example of how to set this see [this link](https://github.com/evoblockchain/evoblock/pull/303).

```go
evoblockd add-genesis-account [address_or_key_name] [coin][,[coin]] [flags]
```

### Queries

The `query` commands allow users to query `vesting` account state.

**`balances`**

Allows users to query the locked, unvested and vested tokens for a given vesting account

```go
evoblockd query vesting balances [address] [flags]
```

### Transactions

The `tx` commands allow users to create and clawback `vesting` account state.

**`create-clawback-vesting-account`**

Allows users to create a new vesting account funded with an allocation of tokens, subject to clawback. Must provide a lockup periods file (--lockup), a vesting periods file (--vesting), or both.

If both files are given, they must describe schedules for the same total amount.
If one file is omitted, it will default to a schedule that immediately unlocks or vests the entire amount. The described amount of coins will be transferred from the --from address to the vesting account. Unvested coins may be "clawed back" by the funder with the clawback command. Coins may not be transferred out of the account if they are locked or unvested. Only vested coins may be staked. For an example of how to set this see [this link](https://github.com/evoblockchain/evoblock/pull/303).

```go
evoblockd tx vesting create-clawback-vesting-account [to_address] [flags]
```

**`clawback`**

Allows users to create a transfer unvested amount out of a ClawbackVestingAccount. Must be requested by the original funder address (--from) and may provide a destination address (--dest), otherwise the coins return to the funder. Delegated or undelegating staking tokens will be transferred in the delegated (undelegating) state. The recipient is vulnerable to slashing, and must act to unbond the tokens if desired.

```go
evoblockd tx vesting clawback [address] [flags]
```

## gRPC

### Queries

| Verb   | Method                                 | Description                            |
| ------ | -------------------------------------- | -------------------------------------- |
| `gRPC` | `evoblock.vesting.v1.Query/Balances`      | Gets locked, unvested and vested coins |
| `GET`  | `/evoblock/vesting/v1/balances/{address}` | Gets locked, unvested and vested coins |

### Transactions

| Verb   | Method                                                 | Description                      |
| ------ | ------------------------------------------------------ | -------------------------------- |
| `gRPC` | `evoblock.vesting.v1.Msg/CreateClawbackVestingAccount`    | Creates clawback vesting account |
| `gRPC` | `/evoblock.vesting.v1.Msg/Clawback`                       | Performs clawback                |
| `GET`  | `/evoblock/vesting/v1/tx/create_clawback_vesting_account` | Creates clawback vesting account |
| `GET`  | `/evoblock/vesting/v1/tx/clawback`                        | Performs clawback                |
