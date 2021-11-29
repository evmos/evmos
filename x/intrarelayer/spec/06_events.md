<!--
order: 6
-->

# Events

The `x/intrarelayer` module emits the following events:

## Register Coin Proposal

| Type            | Attibute Key    | Attibute Value    |
| --------------- | --------------- | ----------------- |
| `register_coin` | `"cosmos_coin"` | `{denom}`         |
| `register_coin` | `"erc20_token"` | `{erc20_address}` |

## Register ERC20 Proposal

| Type             | Attibute Key    | Attibute Value    |
| ---------------- | --------------- | ----------------- |
| `register_erc20` | `"cosmos_coin"` | `{denom}`         |
| `register_erc20` | `"erc20_token"` | `{erc20_address}` |

## Toggle Token Relay

| Type                 | Attibute Key    | Attibute Value    |
| -------------------- | --------------- | ----------------- |
| `toggle_token_relay` | `"erc20_token"` | `{erc20_address}` |
| `toggle_token_relay` | `"cosmos_coin"` | `{denom}`         |

## Update Token Pair ERC20

| Type                      | Attibute Key    | Attibute Value    |
| ------------------------- | --------------- | ----------------- |
| `update_token_pair_erc20` | `"erc20_token"` | `{erc20_address}` |
| `update_token_pair_erc20` | `"cosmos_coin"` | `{denom}`         |

## Convert Coin

| Type           | Attibute Key    | Attibute Value              |
| -------------- | --------------- | --------------------------- |
| `convert_coin` | `"sender"`      | `{msg.Sender}`              |
| `convert_coin` | `"receiver"`    | `{msg.Receiver}`            |
| `convert_coin` | `"amount"`      | `{msg.Coin.Amount.String()}` |
| `convert_coin` | `"cosmos_coin"` | `{denom}`                   |
| `convert_coin` | `"erc20_token"` | `{erc20_address}`           |

## Convert ERC20

| Type            | Attibute Key    | Attibute Value              |
| --------------- | --------------- | --------------------------- |
| `convert_erc20` | `"sender"`      | `{msg.Sender}`              |
| `convert_erc20` | `"receiver"`    | `{msg.Receiver}`            |
| `convert_erc20` | `"amount"`      | `{msg.Amount.String()}`     |
| `convert_erc20` | `"cosmos_coin"` | `{denom}`                   |
| `convert_erc20` | `"erc20_token"` | `{msg.ContractAddress}`     |


