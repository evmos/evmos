<!--
order: 6
-->

# Events

The `x/fees` module emits the following events:

## Register Fee

| Type           | Attribute Key          | Attribute Value           |
| :------------- | :--------------------- | :------------------------ |
| `register_fee` | `"contract"`           | `{msg.ContractAddress}`   |
| `register_fee` | `"sender"`             | `{msg.DeployerAddress}`   |
| `register_fee` | `"withdrawer_address"` | `{msg.WithdrawerAddress}` |

## Update Fee

| Type         | Attribute Key          | Attribute Value           |
| :----------- | :--------------------- | :------------------------ |
| `update_fee` | `"contract"`           | `{msg.ContractAddress}`   |
| `update_fee` | `"sender"`             | `{msg.DeployerAddress}`   |
| `update_fee` | `"withdrawer_address"` | `{msg.WithdrawerAddress}` |

## Cancel Fee

| Type         | Attribute Key | Attribute Value         |
| :----------- | :------------ | :---------------------- |
| `cancel_fee` | `"contract"`  | `{msg.ContractAddress}` |
| `cancel_fee` | `"sender"`    | `{msg.DeployerAddress}` |
