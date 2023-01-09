<!--
order: 5
-->

# Hooks

The `x/inflation` module implements the `AfterEpochEnd`  hook from the
`x/epoch` module in order to allocate inflation.

## Epoch Hook: Inflation

The epoch hook handles the inflation logic which is run at the end of each
epoch. It is responsible for minting and allocating the epoch mint provision as
well as updating it:

1. Check if inflation is disabled. If it is, skip inflation, increment number
   of skipped epochs and return without proceeding to the next steps.
2. A block is committed, that signalizes that an `epoch` has ended (block
   `header.Time` has surpassed `epoch_start` + `epochIdentifier`).
3. Mint coin in amount of calculated `epochMintProvision` and allocate according to
   inflation distribution to staking rewards, usage incentives and community
   pool.
4. If a period ends with the current epoch, increment the period by `1` and set new value to the store.
