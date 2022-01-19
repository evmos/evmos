<!--
order: 4
-->

# Parameters

The minting module contains the following parameters:

| Key                                        | Type         | Example                                |
| ------------------------------------------ | ------------ | -------------------------------------- |
| mint_denom                                 | string       | "uosmo"                                |
| genesis_epoch_provisions                   | string (dec) | "500000000"                            |
| epoch_identifier                           | string       | "weekly"                               |
| reduction_period_in_epochs                 | int64        | 156                                    |
| reduction_factor                           | string (dec) | "0.6666666666666"                      |
| distribution_proportions.staking           | string (dec) | "0.4"                                  |
| distribution_proportions.pool_incentives   | string (dec) | "0.3"                                  |
| distribution_proportions.developer_rewards | string (dec) | "0.2"                                  |
| distribution_proportions.community_pool    | string (dec) | "0.1"                                  |
| weighted_developer_rewards_receivers       | array        | [{"address": "osmoxx", "weight": "1"}] |
| minting_rewards_distribution_start_epoch   | int64        | 10                                     |

**Notes**
1. `mint_denom` defines denom for minting token - uosmo
2. `genesis_epoch_provisions` provides minting tokens per epoch at genesis.
3. `epoch_identifier` defines the epoch identifier to be used for mint module e.g. "weekly"
4. `reduction_period_in_epochs` defines the number of epochs to pass to reduce mint amount
5. `reduction_factor` defines the reduction factor of tokens at every `reduction_period_in_epochs`
6. `distribution_proportions` defines distribution rules for minted tokens, when developer rewards address is empty, it distribute tokens to community pool.
7. `weighted_developer_rewards_receivers` provides the addresses that receives developer rewards by weight
8. `minting_rewards_distribution_start_epoch` defines the start epoch of minting to make sure minting start after initial pools are set
