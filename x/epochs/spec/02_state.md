<!--
order: 2
-->

# State

Epochs module keeps `EpochInfo` objects and modify the information as epochs info changes.
Epochs are initialized as part of genesis initialization, and modified on begin blockers or end blockers.

### Epoch information type

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

EpochInfo keeps `identifier`, `start_time`,`duration`, `current_epoch`, `current_epoch_start_time`,  `epoch_counting_started`, `current_epoch_start_height`.

1. `identifier` keeps epoch identification string.
2. `start_time` keeps epoch counting start time, if block time passes `start_time`, `epoch_counting_started` is set.
3. `duration` keeps target epoch duration.
4. `current_epoch` keeps current active epoch number.
5. `current_epoch_start_time` keeps the start time of current epoch.
6. `epoch_number` is counted only when `epoch_counting_started` flag is set.
7. `current_epoch_start_height` keeps the start block height of current epoch.
