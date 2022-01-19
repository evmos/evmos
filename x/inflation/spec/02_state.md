<!--
order: 2
-->

# State

## Minter

The minter is a space for holding current rewards information.

```go
type Minter struct {
    EpochProvisions sdk.Dec   // Rewards for the current epoch
}
```

## Params

Minting params are held in the global params store.

```go
type Params struct {
    MintDenom               string                  // type of coin to mint
    GenesisEpochProvisions  sdk.Dec                 // initial epoch provisions at genesis
    EpochIdentifier         string                  // identifier of epoch
    ReductionPeriodInEpochs int64                   // number of epochs between reward reductions
    ReductionFactor         sdk.Dec                 // reduction multiplier to execute on each period
	DistributionProportions DistributionProportions // distribution_proportions defines the proportion of the minted denom
	WeightedDeveloperRewardsReceivers    []WeightedAddress // address to receive developer rewards
	MintingRewardsDistributionStartEpoch int64             // start epoch to distribute minting rewards
}
```

## LastHalvenEpoch

Last halven epoch stores the epoch number when the last reduction of coin mint amount per epoch has happened.

**TODO:**
- Update the name to LastReductionEpoch as the reduction amount could be set by governance.