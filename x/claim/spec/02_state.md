<!--
order: 2
-->

# State

### Claim Records

```protobuf
// A Claim Records is the metadata of claim data per address
message ClaimRecord {
  // address of claim user
  string address = 1 [ (gogoproto.moretags) = "yaml:\"address\"" ];

  // total initial claimable amount for the user
  repeated cosmos.base.v1beta1.Coin initial_claimable_amount = 2 [
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
    (gogoproto.nullable) = false,
    (gogoproto.moretags) = "yaml:\"initial_claimable_amount\""
  ];

  // true if action is completed
  // index of bool in array refers to action enum #
  repeated bool action_completed = 3 [
    (gogoproto.moretags) = "yaml:\"action_completed\"",
    (gogoproto.nullable) = false
  ];
}
```
When a user get airdrop for his/her action, claim record is created to prevent duplicated actions on future actions.

### State

```protobuf
message GenesisState {
  // balance of the claim module's account
  cosmos.base.v1beta1.Coin module_account_balance = 1 [
    (gogoproto.moretags) = "yaml:\"module_account_balance\"",
    (gogoproto.nullable) = false
  ];

  // params defines all the parameters of the module.
  Params params = 2 [
    (gogoproto.moretags) = "yaml:\"params\"",
    (gogoproto.nullable) = false
  ];

  // list of claim records, one for every airdrop recipient
  repeated ClaimRecord claim_records = 3 [
    (gogoproto.moretags) = "yaml:\"claim_records\"",
    (gogoproto.nullable) = false
  ];
}
```

Claim module's state consists of `params`, `claim_records`, and `module_account_balance`.
