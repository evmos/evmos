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

## Invariants

The `x/claims` module registers an [`Invariant`](https://docs.cosmos.network/main/building-modules/invariants.html#invariants) to ensure that a property is true at any given time. These functions are useful to detect bugs early on and act upon them to limit their potential consequences (e.g. by halting the chain).

### ClaimsInvariant

The `ClaimsInvariant` checks that the total amount of all unclaimed coins held
in claims records is equal to the escrowed balance held in the claims module
account. This is important to ensure that there are sufficient coins to claim for all claims records.

```go
balance := k.bankKeeper.GetBalance(ctx, moduleAccAddr, params.ClaimsDenom)
isInvariantBroken := !expectedUnclaimed.Equal(balance.Amount.ToDec())
```
