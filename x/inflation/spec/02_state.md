<!--
order: 2
-->

# State

## State Objects

The `x/inflation` module keeps the following objects in state:

| State Object       | Description                    | Key         | Value                        | Store |
| ------------------ | ------------------------------ | ----------- | ---------------------------- | ----- |
| Period             | Period Counter                 | `[]byte{1}` | `[]byte{period}`             | KV    |
| EpochIdentifier    | Epoch identifier bytes         | `[]byte{3}` | `[]byte{epochIdentifier}`    | KV    |
| EpochsPerPeriod    | Epochs per period bytes        | `[]byte{4}` | `[]byte{epochsPerPeriod}`    | KV    |
| SkippedEpochs      | Number of skipped epochs bytes | `[]byte{5}` | `[]byte{skippedEpochs}`      | KV    |

### Period

Counter to keep track of amount of past periods, based on the epochs per period.

### EpochIdentifier

Identifier to trigger epoch hooks.

### EpochsPerPeriod

Amount of epochs in one period

## Genesis State

The `x/inflation` module's `GenesisState` defines the state necessary for
initializing the chain from a previously exported height. It contains the module
parameters and the list of active incentives and their corresponding gas meters
:

```go
type GenesisState struct {
	// params defines all the parameters of the module.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// amount of past periods, based on the epochs per period param
	Period uint64 `protobuf:"varint,2,opt,name=period,proto3" json:"period,omitempty"`
	// inflation epoch identifier
	EpochIdentifier string `protobuf:"bytes,3,opt,name=epoch_identifier,json=epochIdentifier,proto3" json:"epoch_identifier,omitempty"`
	// number of epochs after which inflation is recalculated
	EpochsPerPeriod int64 `protobuf:"varint,4,opt,name=epochs_per_period,json=epochsPerPeriod,proto3" json:"epochs_per_period,omitempty"`
	// number of epochs that have passed while inflation is disabled
	SkippedEpochs uint64 `protobuf:"varint,5,opt,name=skipped_epochs,json=skippedEpochs,proto3" json:"skipped_epochs,omitempty"`
}
```
