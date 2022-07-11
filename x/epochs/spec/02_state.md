<!--
order: 2
-->

# State

## State Objects

The `x/epochs` module keeps the following `objects in state`:

| State Object | Description         | Key                  | Value               | Store |
|--------------|---------------------|----------------------|---------------------|-------|
| `EpochInfo`  | Epoch info bytecode | `[]byte{identifier}` | `[]byte{epochInfo}` | KV    |

### EpochInfo

An `EpochInfo` defines several variables:

1. `identifier` keeps an epoch identification string
2. `start_time` keeps the start time for epoch counting: if block height passes `start_time`, then `epoch_counting_started` is set
3. `duration` keeps the target epoch duration
4. `current_epoch` keeps the current active epoch number
5. `current_epoch_start_time` keeps the start time of the current epoch
6. `epoch_counting_started` is a flag set with `start_time`, at which point `epoch_number` will be counted
7. `current_epoch_start_height` keeps the start block height of the current epoch

```protobuf
message EpochInfo {
    string identifier = 1;
    google.protobuf.Timestamp start_time = 2 [
        (gogoproto.stdtime) = true,
        (gogoproto.nullable) = false,
        (gogoproto.moretags) = "yaml:\"start_time\""
    ];
    google.protobuf.Duration duration = 3 [
        (gogoproto.nullable) = false,
        (gogoproto.stdduration) = true,
        (gogoproto.jsontag) = "duration,omitempty",
        (gogoproto.moretags) = "yaml:\"duration\""
    ];
    int64 current_epoch = 4;
    google.protobuf.Timestamp current_epoch_start_time = 5 [
        (gogoproto.stdtime) = true,
        (gogoproto.nullable) = false,
        (gogoproto.moretags) = "yaml:\"current_epoch_start_time\""
    ];
    bool epoch_counting_started = 6;
    reserved 7;
    int64 current_epoch_start_height = 8;
}
```

The `epochs` module keeps these `EpochInfo` objects in state, which are initialized at genesis and are modified on begin blockers or end blockers.

### Genesis State

The `x/epochs` module's `GenesisState` defines the state necessary for initializing the chain from a previously exported height. It contains a slice containing all the `EpochInfo` objects kept in state:

```go
// Genesis State defines the epoch module's genesis state
type GenesisState struct {
    // list of EpochInfo structs corresponding to all epochs
	Epochs []EpochInfo `protobuf:"bytes,1,rep,name=epochs,proto3" json:"epochs"`
}
```
