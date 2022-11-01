<!--
order: 5
-->

# Hooks

The `x/incentives` module implements two transaction hooks from the `x/evm` and `x/epoch` modules.

## EVM Hook - Gas Metering

The EVM hook updates the logs that keep track of much gas was used for interacting with an incentived contract during one epoch. An [EVM hook](https://evoblock.dev/modules/evm/06_hooks.html) executes custom logic after each successful evm transaction. In this case it updates the incentive’s total gas count and the participant's own gas count.

1. User submits an EVM transaction to an incentivized smart contract and the transaction is finished successfully.
2. The EVM hook’s `PostTxProcessing` method is called on the incentives module. It is passed a transaction receipt that includes the cumulative gas used by the transaction sender to pay for the gas fees. The hook
    1. adds `gasUsed` to an incentive's cumulated `totalGas` and
    2. adds `gasUsed` to a participant's gas meter's cumulative gas used.

## Epoch Hook - Distribution of Rewards

The Epoch hook triggers the distribution of usage rewards for all registered incentives at the end of each epoch (one day or one week). This distribution process first 1) allocates the rewards for each incentive from the allocation pool and then 2) distributes these rewards to all partticipants of each incentive.

1. A `RegisterIncentiveProposal` passes and an `incentive` for the proposed contract is created.
2. An `epoch` begins and `rewards` ($EVO and other denoms) that are minted on every block for inflation are added to the inflation pool every block.
3. Users submit transactions and call functions on the incentivized smart contracts to interact and gas gets logged through the EVM Hook.
4. A block, which signalizes the end of an `epoch`, is proposed and the `DistributeIncentives` method is called through `AfterEpochEnd` hook. This method:
    1. Allocates the amount to be distributed from the inflation pool
    2. Distributes the rewards to all participants. The rewards of each participant are limited by the amount of gas they spent on transaction fees during the current epoch and the reward scaler parameter.
    3. Deletes all gas meters for the contract
    4. Updates the remaining epochs of each incentive. If an incentive’s remaining epochs equals to zero, the incentive is removed and the allocation meters are updated.
    5. Sets the cumulative totalGas to zero for the next epoch
5. Rewards for a given denomination accumulate in the inflation pool if the denomination’s allocation capacity is not fully exhaused and the sum of all active incentivized contracts' allocation is < 100%. The accumulated rewards are added to the allocation in the following epoch.
