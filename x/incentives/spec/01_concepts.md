<!--
order: 1
-->

# Concepts


## Incentive

The purpose of the `x/incentives` module is to provide incentives to users who interact with smart contracts. An incentive allows users to earn rewards up to `rewards = k * sum(tx fees)`, where `k` defines a reward scaler parameter that caps the incentives allocated to a single user by multiplying it with the sum of transaction fees that they’ve spent in the current epoch.

An `incentive` describes the conditions under which rewards are allocated and distributed for a given smart contract. At the end of every epoch, rewards are allocated from an Inflation pool and distributed to participants of the incentive, depending on how much gas every participant spent and the scaling parameter.

The incentive for a given smart contract can be enabled or disabled via governance.

## Inflation Pool

The inflation pool holds `rewards` that can be allocated to incentives. On every block, inflation rewards are minted and added to the inflation pool. Additionally, rewards may also be transferred to the inflation pool on top of inflation. The details of how rewards are added to the inflation pool are described in the `x/inflation` module.

## Epoch

Rewarding users for smart contract interaction is organized in epochs. An `epoch` is a fixed duration in which rewards are added to the inflation pool and smart contract interaction is logged. At the end of an epoch, rewards are allocated and distributed to all participants. This creates a user experience, where users check their balance for new rewards regularly (e.g. every day at the same time).

## Allocation

Before rewards are distributed to users, each incentive allocates rewards from the inflation pool.  The `allocation` describes the portion of rewards in the inflation pool, that is allocated to an incentive for a specified coin.

Users can be rewarded in several coin denominations. These are organized in `allocations`.  An allocation includes the coin denomination and the percentage of rewards that are allocated from the inflation pool.

- There is a cap on how high the reward percentage can be per allocation. It is defined via the chain parameters and can be modified via governance
- The amount of incentives is limited by the sum of all active incentivized contracts' allocations. If the sum is > 100%, no further incentive can be proposed until another allocation becomes inactive.

## Distribution

The allocated rewards for an incentive are distributed according to how much gas participants spent on interaction with the contract during an epoch. The gas used per address is recorded using transaction hooks and stored on the KV store.  At the end of an epoch, the allocated rewards in the incentive are distributed by transferring them to the participants accounts.

::: tip
💡 We use hooks instead of the transaction hash to measure the gas spent because the hook has access to the actual gas spent and the hash only includes the gas limit.
:::
