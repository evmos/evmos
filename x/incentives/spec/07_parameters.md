<!--
order: 7
-->

# Parameters

The `x/incentives` module contains the parameters described below. All parameters can be modified via governance.

| Key                         | Type    | Default Value                      |
| --------------------------- | ------- | ---------------------------------- |
| `EnableIncentives`          | bool    | `true`                             |
| `AllocationLimit`           | sdk.Dec | `sdk.NewDecWithPrec(5,2)` // 5%    |
| `IncentivesEpochIdentifier` | string  | `week`                             |
| `rewardScaler`              | sdk.Dec | `sdk.NewDecWithPrec(12,1)` // 120% |

## Enable Incentives

The `EnableIncentives` parameter toggles all state transitions in the module. When the parameter is disabled, it will prevent all Incentive registration and cancellation and distribution functionality.

## Allocation Limit

The `AllocationLimit` parameter defines the maximum allocation that each incentive can define per denomination. For example, with an `AllocationLimit` of 5%, there can be at most 20 active incentives per denom if they all max out the limit.

There is a cap on how high the reward percentage can be per allocation.

## Incentives Epoch Identifier

The `IncentivesEpochIdentifier` parameter specifies the length of an epoch. It is the interval at which incentive rewards are regularly distributed.

## Reward Scaler

The `rewardScaler` parameter defines  each participant’s reward limit, relative to their gas used. An incentive allows users to earn rewards up to `rewards = k * sum(txFees)`, where `k` defines the reward scaler parameter that caps the incentives allocated to a single user by multiplying it to the sum of transaction fees that they’ve spent in the current epoch.
