<!--
order: 8
-->

# Clients

## CLI

Find below a list of  `evmosd` commands added with the  `x/feesplit` module. You can obtain the full list by using the `evmosd -h` command. A CLI command can look like this:

```bash
evmosd query fees params
```

### Queries

| Command            | Subcommand             | Description                              |
| :----------------- | :--------------------- | :--------------------------------------- |
| `query` `feesplit` | `params`               | Get fees params                          |
| `query` `feesplit` | `contract`             | Get the fee split for a given contract   |
| `query` `feesplit` | `contracts`            | Get all fee splits                       |
| `query` `feesplit` | `deployer-contracts`   | Get all fee splits of a given deployer   |
| `query` `feesplit` | `withdrawer-contracts` | Get all fee splits of a given withdrawer |

### Transactions

| Command         | Subcommand | Description                                |
| :-------------- | :--------- | :----------------------------------------- |
| `tx` `feesplit` | `register` | Register a contract for receiving fees     |
| `tx` `feesplit` | `update`   | Update the withdraw address for a contract |
| `tx` `feesplit` | `cancel`   | Remove the fee split for a contract        |

## gRPC

### Queries

| Verb   | Method                                            | Description                              |
| :----- | :------------------------------------------------ | :--------------------------------------- |
| `gRPC` | `evmos.feesplit.v1.Query/Params`                  | Get fees params                          |
| `gRPC` | `evmos.feesplit.v1.Query/FeeSplit`                | Get the fee split for a given contract   |
| `gRPC` | `evmos.feesplit.v1.Query/FeeSplits`               | Get all fee splits                       |
| `gRPC` | `evmos.feesplit.v1.Query/DeployerFeeSplits`       | Get all fee splits of a given deployer   |
| `gRPC` | `evmos.feesplit.v1.Query/WithdrawerFeeSplits`     | Get all fee splits of a given withdrawer |
| `GET`  | `/evmos/feesplit/v1/params`                       | Get fees params                          |
| `GET`  | `/evmos/feesplit/v1/feesplits/{contract_address}`  | Get the fee split for a given contract   |
| `GET`  | `/evmos/feesplit/v1/feesplits`                    | Get all fee splits                       |
| `GET`  | `/evmos/feesplit/v1/feesplits/{deployer_address}` | Get all fee splits of a given deployer   |
| `GET`  | `/evmos/feesplit/v1/feesplits/{withdraw_address}` | Get all fee splits of a given withdrawer |

### Transactions

| Verb   | Method                                     | Description                                |
| :----- | :----------------------------------------- | :----------------------------------------- |
| `gRPC` | `evmos.feesplit.v1.Msg/RegisterFeeSplit`   | Register a contract for receiving fees     |
| `gRPC` | `evmos.feesplit.v1.Msg/UpdateFeeSplit`     | Update the withdraw address for a contract |
| `gRPC` | `evmos.feesplit.v1.Msg/CancelFeeSplit`     | Remove the fee split for a contract        |
| `POST` | `/evmos/feesplit/v1/tx/register_fee_split` | Register a contract for receiving fees     |
| `POST` | `/evmos/feesplit/v1/tx/update_fee_split`   | Update the withdraw address for a contract |
| `POST` | `/evmos/feesplit/v1/tx/cancel_fee_split`   | Remove the fee split for a contract        |
