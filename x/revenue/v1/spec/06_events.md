<!--
order: 6
-->

# Events

The `x/revenue` module emits the following events:

## Register Fee Split

| Type                 | Attribute Key          | Attribute Value           |
| :------------------- | :--------------------- | :------------------------ |
| `register_revenue` | `"contract"`           | `{msg.ContractAddress}`   |
| `register_revenue` | `"sender"`             | `{msg.DeployerAddress}`   |
| `register_revenue` | `"withdrawer_address"` | `{msg.WithdrawerAddress}` |

## Update Fee Split

| Type               | Attribute Key          | Attribute Value           |
| :----------------- | :--------------------- | :------------------------ |
| `update_revenue` | `"contract"`           | `{msg.ContractAddress}`   |
| `update_revenue` | `"sender"`             | `{msg.DeployerAddress}`   |
| `update_revenue` | `"withdrawer_address"` | `{msg.WithdrawerAddress}` |

## Cancel Fee Split

| Type               | Attribute Key | Attribute Value         |
| :----------------- | :------------ | :---------------------- |
| `cancel_revenue` | `"contract"`  | `{msg.ContractAddress}` |
| `cancel_revenue` | `"sender"`    | `{msg.DeployerAddress}` |
