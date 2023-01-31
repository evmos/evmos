<!--
order: 5
-->

# Hooks

The `x/epochs` module implements hooks so that other modules can use epochs
to allow facets of the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) to run on specific schedules.

## Hooks Implementation

```go
// combine multiple epoch hooks, all hook functions are run in array sequence
type MultiEpochHooks []types.EpochHooks

// AfterEpochEnd is called when epoch is going to be ended, epochNumber is the
// number of epoch that is ending
func (mh MultiEpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {...}

// BeforeEpochStart is called when epoch is going to be started, epochNumber is
// the number of epoch that is starting
func (mh MultiEpochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {...}

// AfterEpochEnd executes the indicated hook after epochs ends
func (k Keeper) AfterEpochEnd(ctx sdk.Context, identifier string, epochNumber int64) {...}

// BeforeEpochStart executes the indicated hook before the epochs
func (k Keeper) BeforeEpochStart(ctx sdk.Context, identifier string, epochNumber int64) {...}
```

## Recieving Hooks

When other modules (outside of `x/epochs`) recieve hooks,
they need to filter the value `epochIdentifier`, and only do executions for a specific `epochIdentifier`.

The filtered values from `epochIdentifier` could be stored in the `Params` of other modules,
so they can be modified by governance.

Governance can change epoch periods from `week` to `day` as needed.
