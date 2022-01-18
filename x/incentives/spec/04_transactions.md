<!--
order: 4
-->

# Transactions

This section defines the `sdk.Msg` concrete types that result in the state transitions defined on the previous section.

## `RegisterIncentiveProposal`

A gov `Content` type to register an Incentive for a given contract for the duration of a certain number of epochs. Governance users vote on this proposal and it automatically executes the custom handler for `RegisterIncentiveProposal` when the vote passes.

```go
type RegisterIncentiveProposal struct {
	// title of the proposal
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	// proposal description
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	// contract address
	Contract string `protobuf:"bytes,3,opt,name=contract,proto3" json:"contract,omitempty"`
	// denoms and percentage of rewards to be allocated
	Allocations github_com_cosmos_cosmos_sdk_types.DecCoins `protobuf:"bytes,4,rep,name=allocations,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.DecCoins" json:"allocations"`
	// number of remaining epochs
	Epochs uint32 `protobuf:"varint,5,opt,name=epochs,proto3" json:"epochs,omitempty"`
}
```

The proposal content stateless validation fails if:

- Title is invalid (length or char)
- Description is invalid (length or char)
- Contract address is invalid
- Allocations are invalid
    - no allocation included in Allocations
    - invalid amount of at least one allocation (below 0 or above 1)
- Epochs are invalid (zero)

## `CancelIncentiveProposal`

A gov `Content` type to remove an Incentive. Governance users vote on this proposal and it automatically executes the custom handler for `CancelIncentiveProposal` when the vote passes.

```go
type CancelIncentiveProposal struct {
	// title of the proposal
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	// proposal description
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	// contract address
	Contract string `protobuf:"bytes,3,opt,name=contract,proto3" json:"contract,omitempty"`
}
```

The proposal content stateless validation fails if:

- Title is invalid (length or char)
- Description is invalid (length or char)
- Contract address is invalid
