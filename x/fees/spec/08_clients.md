<!--
order: 8
-->

# Clients

## CLI

Find below a list of  `evmosd` commands added with the  `x/fees` module. You can obtain the full list by using the `evmosd -h` command. A CLI command can look like this:

```bash
evmosd query fees params
```

### Queries

| Command        | Subcommand           | Description                                       |
| :------------- | :------------------- | :------------------------------------------------ |
| `query` `fees` | `params`             | Get fees params                                   |
| `query` `fees` | `fee-info`           | Get registered fee info                           |
| `query` `fees` | `fee-infos`          | Get all registered fee infos                      |
| `query` `fees` | `fee-infos-deployer` | Get all contracts that a deployer has registered |

### Transactions

| Command     | Subcommand     | Description                                |
| :---------- | :------------- | :----------------------------------------- |
| `tx` `fees` | `register-fee` | Register a contract for receiving fees     |
| `tx` `fees` | `update-fee`   | Update the withdraw address for a contract |
| `tx` `fees` | `cancel-fee`   | Remove the fee info for a contract         |

## gRPC

### Queries

| Verb   | Method                                   | Description                                 |
| :----- | :--------------------------------------- | :------------------------------------------ |
| `gRPC` | `evmos.fees.v1.Query/Params`             | Get fees params                             |
| `gRPC` | `evmos.fees.v1.Query/Fee`                | Get registered fee info                     |
| `gRPC` | `evmos.fees.v1.Query/Fees`               | Get all registered fee infos                |
| `gRPC` | `evmos.fees.v1.Query/DeployerFees`       | Get all registered fee infos for a deployer |
| `GET`  | `/evmos/fees/v1/params`                  | Get fees params                             |
| `GET`  | `/evmos/fees/v1/fees/{contract_address}` | Get registered fee info                     |
| `GET`  | `/evmos/fees/v1/fees`                    | Get all registered fee infos                |
| `GET`  | `/evmos/fees/v1/fees/{deployer_address}` | Get all registered fee infos for a deployer |

### Transactions

| Verb   | Method                                    | Description                                |
| :----- | :---------------------------------------- | :----------------------------------------- |
| `gRPC` | `evmos.fees.v1.Msg/RegisterFee`           | Register a contract for receiving fees     |
| `gRPC` | `evmos.fees.v1.Msg/UpdateFee`             | Update the withdraw address for a contract |
| `gRPC` | `evmos.fees.v1.Msg/CancelFee`             | Remove the fee info for a contract         |
| `POST` | `/evmos/fees/v1/tx/register_dev_fee_info` | Register a contract for receiving fees     |
| `POST` | `/evmos/fees/v1/tx/update_dev_fee_info`   | Update the withdraw address for a contract |
| `POST` | `/evmos/fees/v1/tx/cancel_dev_fee_info`   | Remove the fee info for a contract         |
