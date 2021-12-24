<!--
order: 7
-->

# Params

Claim module provides below params

```protobuf
// Params defines the claim module's parameters.
message Params {
  google.protobuf.Timestamp airdrop_start_time = 1 [
    (gogoproto.stdtime) = true,
    (gogoproto.nullable) = false,
    (gogoproto.moretags) = "yaml:\"airdrop_start_time\""
  ];
  google.protobuf.Duration duration_until_decay = 2 [
    (gogoproto.nullable) = false,
    (gogoproto.stdduration) = true,
    (gogoproto.jsontag) = "duration_until_decay,omitempty",
    (gogoproto.moretags) = "yaml:\"duration_until_decay\""
  ];
  google.protobuf.Duration duration_of_decay = 3 [
    (gogoproto.nullable) = false,
    (gogoproto.stdduration) = true,
    (gogoproto.jsontag) = "duration_of_decay,omitempty",
    (gogoproto.moretags) = "yaml:\"duration_of_decay\""
  ];
  // denom of claimable asset
  string claim_denom = 4;
}
```

1. `airdrop_start_time` refers to the time when user can start to claim airdrop.
2. `duration_until_decay` refers to the duration from start time to decay start time.
3. `duration_of_decay` refers to the duration from decay start time to claim end time. Users are not able to claim airdrop after this.
4. `claim_denom` refers to the denomination of claiming tokens. As a default, it's `uosmo`.