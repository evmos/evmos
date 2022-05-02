<!--
order: 6
-->

# Events

The `x/erc20` module emits the following events:

## Register Coin Proposal

| Type            | Attribute Key   | Attribute Value   |
| --------------- | --------------- | ----------------- |
| `register_coin` | `"cosmos_coin"` | `{denom}`         |
| `register_coin` | `"erc20_token"` | `{erc20_address}` |

## Register ERC20 Proposal

| Type             | Attribute Key   | Attribute Value   |
| ---------------- | --------------- | ----------------- |
| `register_erc20` | `"cosmos_coin"` | `{denom}`         |
| `register_erc20` | `"erc20_token"` | `{erc20_address}` |

## Toggle Token Conversion

| Type                      | Attribute Key   | Attribute Value   |
| ------------------------- | --------------- | ----------------- |
| `toggle_token_conversion` | `"erc20_token"` | `{erc20_address}` |
| `toggle_token_conversion` | `"cosmos_coin"` | `{denom}`         |

## Convert Coin

| Type           | Attribute Key   | Attribute Value              |
| -------------- | --------------- | ---------------------------- |
| `convert_coin` | `"sender"`      | `{msg.Sender}`               |
| `convert_coin` | `"receiver"`    | `{msg.Receiver}`             |
| `convert_coin` | `"amount"`      | `{msg.Coin.Amount.String()}` |
| `convert_coin` | `"cosmos_coin"` | `{denom}`                    |
| `convert_coin` | `"erc20_token"` | `{erc20_address}`            |

## Convert ERC20

| Type            | Attribute Key   | Attribute Value         |
| --------------- | --------------- | ----------------------- |
| `convert_erc20` | `"sender"`      | `{msg.Sender}`          |
| `convert_erc20` | `"receiver"`    | `{msg.Receiver}`        |
| `convert_erc20` | `"amount"`      | `{msg.Amount.String()}` |
| `convert_erc20` | `"cosmos_coin"` | `{denom}`               |
| `convert_erc20` | `"erc20_token"` | `{msg.ContractAddress}` |
