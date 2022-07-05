<!--
order: 6
-->

# Events

The `x/feesplit` module emits the following events:

## Register Fee Split

| Type                 | Attribute Key          | Attribute Value           |
| :------------------- | :--------------------- | :------------------------ |
| `register_fee_split` | `"contract"`           | `{msg.ContractAddress}`   |
| `register_fee_split` | `"sender"`             | `{msg.DeployerAddress}`   |
| `register_fee_split` | `"withdrawer_address"` | `{msg.WithdrawerAddress}` |

## Update Fee Split

| Type               | Attribute Key          | Attribute Value           |
| :----------------- | :--------------------- | :------------------------ |
| `update_fee_split` | `"contract"`           | `{msg.ContractAddress}`   |
| `update_fee_split` | `"sender"`             | `{msg.DeployerAddress}`   |
| `update_fee_split` | `"withdrawer_address"` | `{msg.WithdrawerAddress}` |

## Cancel Fee Split

| Type               | Attribute Key | Attribute Value         |
| :----------------- | :------------ | :---------------------- |
| `cancel_fee_split` | `"contract"`  | `{msg.ContractAddress}` |
| `cancel_fee_split` | `"sender"`    | `{msg.DeployerAddress}` |
