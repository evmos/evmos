<!--
order: 0
-->

# Concepts

The `x/mint` module is designed to handle the regular printing of new tokens within a chain.
The design taken within Osmosis is to

- Mint new tokens once per epoch (default one week)
- To have a "Reductioning factor" every period, which reduces the amount of rewards per epoch.
    (default: period is 3 years, where a year is 52 epochs. The next period's rewards are 2/3 of the prior period's rewards)

## Reductioning factor

This is a generalization over the Bitcoin style halvenings.
Every year, the amount of rewards issued per week will reduce by a governance specified factor, instead of a fixed `1/2`.
So `RewardsPerEpochNextPeriod = ReductionFactor * CurrentRewardsPerEpoch)`.
When `ReductionFactor = 1/2`, the Bitcoin halvenings are recreated.
We default to having a reduction factor of `2/3`, and thus reduce rewards at the end of every year by `33%`.

The implication of this is that the total supply is finite, according to the following formula:

$$Total\ Supply = InitialSupply + EpochsPerPeriod * \frac{InitialRewardsPerEpoch}{1 - ReductionFactor} $$
