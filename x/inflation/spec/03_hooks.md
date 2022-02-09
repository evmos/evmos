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

1. A block is commited, that signalizes that an `epoch` has ended (block
   `header.Time` has surpassed `epoch_start` + `epochIdentifier`).
2. Mint coin in amount of `epochMintProvision` and allocate according to
   inflation distribution to staking rewards, usage incentives and community
   pool.
3. If a period ends with current epoch,
    1. increment the period by 1 and set to store and
    2. recalculate epochMintProvision and set to store.
