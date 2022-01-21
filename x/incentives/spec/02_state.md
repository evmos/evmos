<!--
order: 2
-->

# State

## State Objects

The `x/incentives` module keeps the following objects in state:

| State Object    | Description                                   | Key                                                    | Value               | Store |
| --------------- | --------------------------------------------- | ------------------------------------------------------ | ------------------- | ----- |
| Incentive       | Incentive bytecode                            | `[]byte{1} + []byte(contract)`                         | `[]byte{incentive}` | KV    |
| GasMeter        | Incentive id bytecode by erc20 contract bytes | `[]byte{2} + []byte(contract) + []byte(participant)  ` | `[]byte{gasMeter}`  | KV    |
| AllocationMeter | Total allocation bytes by denom bytes         | `[]byte{3} + []byte(denom)`                            | `[]byte{sdk.Dec}`   | KV    |

### Incentive

An instance that organizes distribution conditions for a given smart contract.

```go
type Incentive struct {
	// contract address
	Contract string `protobuf:"bytes,1,opt,name=contract,proto3" json:"contract,omitempty"`
	// denoms and percentage of rewards to be allocated
	Allocations github_com_cosmos_cosmos_sdk_types.DecCoins `protobuf:"bytes,2,rep,name=allocations,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.DecCoins" json:"allocations"`
	// number of remaining epochs
	Epochs uint32 `protobuf:"varint,3,opt,name=epochs,proto3" json:"epochs,omitempty"`
	// distribution start time
	StartTime time.Time `protobuf:"bytes,4,opt,name=start_time,json=startTime,proto3,stdtime" json:"start_time"`
	// cumulative gas spent by all gasmeters of the incentive during the epoch
	TotalGas uint64 `protobuf:"varint,5,opt,name=total_gas,json=totalGas,proto3" json:"total_gas,omitempty"`
}
```

As long as an incentive has remaining epochs, it distributes rewards according to its allocations. The allocations are stored as `sdk.DecCoins` where each containing [`sdk.DecCoin`](https://github.com/cosmos/cosmos-sdk/blob/master/types/dec_coin.go) describes the percentage of rewards (`Amount`) that are allocated to the contract for a given coin denomination (`Denom`). An incentive can contain several allocations, resulting in users to receive rewards in form of several different denominations.

### GasMeter

Tracks the cumulative gas spent in a contract per participant during one epoch.

```go
type GasMeter struct {
	// hex address of the incentivized contract
	Contract string `protobuf:"bytes,1,opt,name=contract,proto3" json:"contract,omitempty"`
	// participant address that interacts with the incentive
	Participant string `protobuf:"bytes,2,opt,name=participant,proto3" json:"participant,omitempty"`
	// cumulative gas spent during the epoch
	CumulativeGas uint64 `protobuf:"varint,3,opt,name=cumulative_gas,json=cumulativeGas,proto3" json:"cumulative_gas,omitempty"`
}
```

### AllocationMeter

An allocation meter stores the sum of all registered incentives’ allocations for a given denomination and is used to limit the amount of registered incentives.

Say, there are several incentives that have registered an allocation for the $EVMOS coin and the allocation meter for $EVMOS is at 97%. Then a new incentve proposal can only include an $EVMOS allocation at up to 3%, claiming the last remaining allocation capcaity from the $EVMOS rewards in the inflation pool.

## Genesis State

The `x/incentives` module's `GenesisState` defines the state necessary for initializing the chain from a previously exported height. It contains the module parameters and the list of active incentives and their corresponding gas meters:

```go
// GenesisState defines the module's genesis state.
type GenesisState struct {
	// module parameters
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// active incentives
	Incentives []Incentive `protobuf:"bytes,2,rep,name=incentives,proto3" json:"incentives"`
	// active Gasmeters
	GasMeters []GasMeter `protobuf:"bytes,3,rep,name=gas_meters,json=gasMeters,proto3" json:"gas_meters"`
}
```
