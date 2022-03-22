<!--
order: 5
-->

# Events

The `x/claims` module emits the following events:

## Claim

| Type    | Attribute Key | Attribute Value                                                         |
| ------- | ------------- | ----------------------------------------------------------------------- |
| `claim` | `"sender"`    | `{address}`                                                             |
| `claim` | `"amount"`    | `{amount}`                                                              |
| `claim` | `"action"`    | `{"ACTION_VOTE"/ "ACTION_DELEGATE"/"ACTION_EVM"/"ACTION_IBC_TRANSFER"}` |

## Merge Claims Records

| Type                   | Attribute Key                 | Attribute Value             |
| ---------------------- | ----------------------------- | --------------------------- |
| `merge_claims_records` | `"recipient"`                 | `{recipient.String()}`      |
| `merge_claims_records` | `"claimed_coins"`             | `{claimed_coins.String()}`  |
| `merge_claims_records` | `"fund_community_pool_coins"` | `{remainderCoins.String()}` |
