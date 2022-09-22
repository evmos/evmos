<!--
order: 8
-->

# Clients

## CLI

Find below a list of  `evmosd` commands added with the  `x/revenue` module. You can obtain the full list by using the `evmosd -h` command. A CLI command can look like this:

```bash
evmosd query fees params
```

### Queries

| Command            | Subcommand             | Description                              |
| :----------------- | :--------------------- | :--------------------------------------- |
| `query` `revenue` | `params`               | Get fees params                          |
| `query` `revenue` | `contract`             | Get the fee split for a given contract   |
| `query` `revenue` | `contracts`            | Get all fee splits                       |
| `query` `revenue` | `deployer-contracts`   | Get all fee splits of a given deployer   |
| `query` `revenue` | `withdrawer-contracts` | Get all fee splits of a given withdrawer |

### Transactions

| Command         | Subcommand | Description                                |
| :-------------- | :--------- | :----------------------------------------- |
| `tx` `revenue` | `register` | Register a contract for receiving fees     |
| `tx` `revenue` | `update`   | Update the withdraw address for a contract |
| `tx` `revenue` | `cancel`   | Remove the fee split for a contract        |

## gRPC

### Queries

| Verb   | Method                                            | Description                              |
| :----- | :------------------------------------------------ | :--------------------------------------- |
| `gRPC` | `evmos.revenue.v1.Query/Params`                  | Get fees params                          |
| `gRPC` | `evmos.revenue.v1.Query/Revenue`                | Get the fee split for a given contract   |
| `gRPC` | `evmos.revenue.v1.Query/Revenues`               | Get all fee splits                       |
| `gRPC` | `evmos.revenue.v1.Query/DeployerRevenues`       | Get all fee splits of a given deployer   |
| `gRPC` | `evmos.revenue.v1.Query/WithdrawerRevenues`     | Get all fee splits of a given withdrawer |
| `GET`  | `/evmos/revenue/v1/params`                       | Get fees params                          |
| `GET`  | `/evmos/revenue/v1/revenues/{contract_address}`  | Get the fee split for a given contract   |
| `GET`  | `/evmos/revenue/v1/revenues`                    | Get all fee splits                       |
| `GET`  | `/evmos/revenue/v1/revenues/{deployer_address}` | Get all fee splits of a given deployer   |
| `GET`  | `/evmos/revenue/v1/revenues/{withdraw_address}` | Get all fee splits of a given withdrawer |

### Transactions

| Verb   | Method                                     | Description                                |
| :----- | :----------------------------------------- | :----------------------------------------- |
| `gRPC` | `evmos.revenue.v1.Msg/RegisterRevenue`   | Register a contract for receiving fees     |
| `gRPC` | `evmos.revenue.v1.Msg/UpdateRevenue`     | Update the withdraw address for a contract |
| `gRPC` | `evmos.revenue.v1.Msg/CancelRevenue`     | Remove the fee split for a contract        |
| `POST` | `/evmos/revenue/v1/tx/register_revenue` | Register a contract for receiving fees     |
| `POST` | `/evmos/revenue/v1/tx/update_revenue`   | Update the withdraw address for a contract |
| `POST` | `/evmos/revenue/v1/tx/cancel_revenue`   | Remove the fee split for a contract        |
