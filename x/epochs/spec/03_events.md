<!--
order: 3
-->

# Events

The `x/epochs` module emits the following events:

## BeginBlocker

| Type          | Attribute Key     | Attribute Value   |
| ------------- | ----------------- | ----------------- |
| `epoch_start` | `"epoch_number"`  | `{epoch_number}`  |
| `epoch_start` | `"start_time"`    | `{start_time}`    |

## EndBlocker

| Type           | Attribute Key    | Attribute Value   |
| ------------- | ----------------- | ----------------- |
| `epoch_end`   | `"epoch_number"`  | `{epoch_number}`  |
