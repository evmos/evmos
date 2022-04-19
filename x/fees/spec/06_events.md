<!--
order: 6
-->

# Events

The `x/fees` module emits the following events:

## Register Contract Fee Info

| Type                    | Attribute Key         | Attribute Value                    |
| :---------------------- | :---------------------| :--------------------------------- |
| `register_dev_fee_info` | `"contract"`          | `{msg.ContractAddress}`            |
| `register_dev_fee_info` | `"sender"`            | `{msg.DeployerAddress}`            |
| `register_dev_fee_info` | `"withdraw_address"`  | `{msg.WithdrawAddress}`            |

## Update Contract Fee Info

| Type                   | Attribute Key                 | Attribute Value             |
| :--------------------- | :---------------------------- | :-------------------------- |
| `update_dev_fee_info`  | `"contract"`                  | `{msg.ContractAddress}`     |
| `update_dev_fee_info`  | `"sender"`                    | `{msg.DeployerAddress}`     |
| `update_dev_fee_info`  | `"withdraw_address"`          | `{msg.WithdrawAddress}`     |

## Cancel Contract Fee Info

| Type                   | Attribute Key                 | Attribute Value             |
| :--------------------- | :---------------------------- | :-------------------------- |
| `cancel_dev_fee_info`  | `"contract"`                  | `{msg.ContractAddress}`     |
| `cancel_dev_fee_info`  | `"sender"`                    | `{msg.DeployerAddress}`     |
