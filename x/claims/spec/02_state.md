<!--
order: 2
-->

# State

## State Objects

The `x/claims` module keeps the following objects in state:

| State Object   | Description            | Key                           | Value                  | Store |
|----------------|------------------------|-------------------------------|------------------------|-------|
| `ClaimsRecord` | Claims record bytecode | `[]byte{1} + []byte(address)` | `[]byte{claimsRecord}` | KV    |

### Claim Record

A `ClaimRecord` defines the initial claimable airdrop amount and the list of completed actions to claim the tokens.

```protobuf
message ClaimsRecord {
  // total initial claimable amount for the user
  string initial_claimable_amount = 1 [
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable) = false
  ];
  // slice of the available actions completed
  repeated bool actions_completed = 2;
}
```

## Genesis State

The `x/claims` module's `GenesisState` defines the state necessary for initializing the chain from a previously exported height. It contains the module parameters and a slice containing all the claim records by user address:

```go
// GenesisState defines the claims module's genesis state.
type GenesisState struct {
	// params defines all the parameters of the module.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// list of claim records with the corresponding airdrop recipient
	ClaimsRecords []ClaimsRecordAddress `protobuf:"bytes,2,rep,name=claims_records,json=claimsRecords,proto3" json:"claims_records"`
}
```
