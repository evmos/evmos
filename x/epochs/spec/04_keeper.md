<!--
order: 4
-->

# Keepers

The `x/epochs` module only exposes one keeper, the epochs keeper, which can be used to manage epochs.

## Epochs Keeper

Presently only one fully-permissioned epochs keeper is exposed, which has the ability to both read and write the `EpochInfo` for all epochs, and to iterate over all stored epochs.

```go
// Keeper of epoch nodule maintains collections of epochs and hooks.
type Keeper struct {
	cdc      codec.Codec
	storeKey sdk.StoreKey
	hooks    types.EpochHooks
}
```

```go
// Keeper is the interface for epoch module keeper
type Keeper interface {
  // GetEpochInfo returns epoch info by identifier
  GetEpochInfo(ctx sdk.Context, identifier string) types.EpochInfo

  // SetEpochInfo set epoch info
  SetEpochInfo(ctx sdk.Context, epoch types.EpochInfo)

  // DeleteEpochInfo delete epoch info
  DeleteEpochInfo(ctx sdk.Context, identifier string)

  // IterateEpochInfo iterate through epochs
  IterateEpochInfo(ctx sdk.Context, fn func(index int64, epochInfo types.EpochInfo) (stop bool))

  // Get all epoch infos
  AllEpochInfos(ctx sdk.Context) []types.EpochInfo
}
```
