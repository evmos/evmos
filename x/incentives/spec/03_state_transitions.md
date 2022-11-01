<!--
order: 3
-->

# State Transitions

The `x/incentive` module allows for two types of registration state transitions:  `RegisterIncentiveProposal` and `CancelIncentiveProposal`. The logic for *gas metering* and *distributing rewards*, is handled through [Hooks](05_hooks.md).

## Incentive Registration

A user registers an incentive defining the contract, allocations, and number of epochs. Once the proposal passes (i.e is approved by governance), the incentive module creates the incentive and distributes rewards.

1. User submits a `RegisterIncentiveProposal`.
2. Validators of the Evoblock Hub vote on the proposal using `MsgVote` and proposal passes.
3. Create incentive for the contract with a `TotalGas = 0` and set its `startTime` to `ctx.Blocktime` if the following conditions are met:
    1. Incentives param is globally enabled
    2. Incentive is not yet registered
    3. Balance in the inflation pool is > 0 for each allocation denom except for the mint denomination. We know that the amount of the minting denom (eg: EVO) will be added to every block but for other denoms (IBC vouchers, ERC20 tokens using the `x/erc20` module) the module account needs to have a positive amount to distribute the incentives
    4. The sum of all registered allocations for each denom (current + proposed) is < 100%
