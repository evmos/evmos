<!--
order: 8
-->

# Clients

## CLI

Find below a list of  `evmosd` commands added with the  `x/intrarelayer` module. You can obtain the full list by using the `evmosd -h` command. A CLI command can look like this:

```bash
evmosd query intrarelayer params
```

### Queries

| Command                | Subcommand    | Description                    |
| ---------------------- | ------------- | ------------------------------ |
| `query` `intrarelayer` | `params`      | Get intrarelayer params        |
| `query` `intrarelayer` | `token-pair`  | Get registered token pair      |
| `query` `intrarelayer` | `token-pairs` | Get all registered token pairs |

### Transactions

| Command             | Subcommand      | Description                    |
| ------------------- | --------------- | ------------------------------ |
| `tx` `intrarelayer` | `convert-coin`  | Convert a Cosmos Coin to ERC20 |
| `tx` `intrarelayer` | `convert-erc20` | Convert a ERC20 to Cosmos Coin |

## gRPC

### Queries

| Verb   | Method                                   | Description                    |
| ------ | ---------------------------------------- | ------------------------------ |
| `gRPC` | `evmos.intrarelayer.v1.Query/Params`     | Get intrarelayer params        |
| `gRPC` | `evmos.intrarelayer.v1.Query/TokenPair`  | Get registered token pair      |
| `gRPC` | `evmos.intrarelayer.v1.Query/TokenPairs` | Get all registered token pairs |
| `GET`  | `/evmos/intrarelayer/v1/params`          | Get intrarelayer params        |
| `GET`  | `/evmos/intrarelayer/v1/token_pair`      | Get registered token pair      |
| `GET`  | `/evmos/intrarelayer/v1/token_pairs`     | Get all registered token pairs |

### Transactions

| Verb   | Method                                    | Description                    |
| ------ | ----------------------------------------- | ------------------------------ |
| `gRPC` | `evmos.intrarelayer.v1.Msg/ConvertCoin`   | Convert a Cosmos Coin to ERC20 |
| `gRPC` | `evmos.intrarelayer.v1.Msg/ConvertERC20`  | Convert a ERC20 to Cosmos Coin |
| `GET`  | `/evmos/intrarelayer/v1/tx/convert_coin`  | Convert a Cosmos Coin to ERC20 |
| `GET`  | `/evmos/intrarelayer/v1/tx/convert_erc20` | Convert a ERC20 to Cosmos Coin |

<!-- ## JSON-RPC

TODO

- Prereq: intrarelaying enabled, pair enabled, evm hook enabled
- Transfer registered ERC20 to module address
- Should update balance on the bank module -->
