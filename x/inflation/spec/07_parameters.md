<!--
order: 7
-->

# Parameters

The `x/inflation` module contains the parameters described below. All parameters
can be modified via governance.

[Untitled](https://www.notion.so/8a33df381c0242e089c105c6d8afdf74)

| Key                         | Type    | Default Value                      |
| --------------------------- | ------- | ---------------------------------- |
| `EnableIncentives`          | bool    | `true`                             |
| `AllocationLimit`           | sdk.Dec | `sdk.NewDecWithPrec(5,2)` // 5%    |
| `IncentivesEpochIdentifier` | string  | `week`                             |
| `rewardScaler`              | sdk.Dec | `sdk.NewDecWithPrec(12,1)` // 120% |


## Mint Denom

The `MintDenom` parameter sets the denomination in which new coins are minted.

## Exponential Calculation

The `ExponentialCalculation` parameter holds all values required for the
calculation of the `epochMintProvision`. The values `A`, `R` and `C` describe
the descrease of inflation over time. The `BondingTarget` and `MaxVariance`
allow for an increase in inflation, which is automatically regulated by the
`bonded ratio`, the portion of staked tokens in the network. The exact formula
can be found under
[Concepts](https://www.notion.so/Inflation-Module-2fa8b7ae430d47e697164fcdb59b5c55).

## Inflation Distribution

The `IinflationDistribution` parameter defines the distribution in which
inflation is allocated through minting on each epoch (`stakingRewards`,
`usageIncentives`,  `CommunityPool`). The `x/inflation` excludes the team
vesting distribution, as team vesting is minted once at genesis. To reflect this
the distribution from the Evmos Token Model is recalculated into a distribution
that excludes team vesting. Note, that this does not change the inflation
proposed in the Evmos Token Model. Each `InflationDistribution` can be
calculated like this:

```markdown
stakingRewards = evmosTokenModelDistribution / (1 - teamVestingDistribution)
0.5333333      = 40%                         / (1 - 25%)
```
