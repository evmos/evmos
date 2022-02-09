<!--
order: 5
-->

# Events

The `x/claims` module emits the following events:

## Claim

| Type    | Attribute Key | Attribute Value                                                          |
|---------|--------------|-------------------------------------------------------------------------|
| `claim` | `"sender"`   | `{address}`                                                             |
| `claim` | `"amount"`   | `{amount}`                                                              |
| `claim` | `"action"`   | `{"ACTION_VOTE"/ "ACTION_DELEGATE"/"ACTION_EVM"/"ACTION_IBC_TRANSFER"}` |
