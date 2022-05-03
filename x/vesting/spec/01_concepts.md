<!--
order: 1
-->

# Concepts

## Vesting

Vesting describes the process of converting `unvested` into `vested` tokens without transferring the ownership of those tokens. In an unvested state, tokens cannot be transferred to other accounts, delegated to validators, or used for governance. A vesting schedule describes the amount and time at which tokens are vested. The duration until which the first tokens are vested is called the `cliff`.

## Lockup

The lockup describes the schedule by which tokens are converted from a  `locked` to an `unlocked` state. As long as all tokens are locked, the account cannot perform any Ethereum transactions using the `x/evm` module. Additionally, locked tokens cannot be transferred to other accounts. In the case in which tokens are both locked and vested at the same time, it is possible to delegate them to validators, but not transfer them to other accounts.

## Schedules

Vesting and lockup schedules specify the amount and time at which tokens are vested or unlocked. They are defined as [`periods`](https://docs.cosmos.network/v0.42/modules/auth/05_vesting.html#period) where each period has its own length and amount. A typical vesting schedule for instance would be defined starting with a one-year period to represent the vesting cliff, followed by several monthly vesting periods until the total allocated vesting amount is vested.

Vesting or lockup schedules can be easily created with Agoric’s [`vestcalc`](https://github.com/agoric-labs/cosmos-sdk/tree/Agoric/x/auth/vesting/cmd/vestcalc) tool. E.g. to calculate a four-year vesting schedule with a one year cliff, starting in January 2022, you can run vestcalc with:

```bash
vestcalc --write --start=2022-01-01 --coins=200000000000000000000000aevmos --months=48 --cliffs=2023-01-01
```

## Clawback

In case a `ClawbackVestingAccount`'s underlying commitment or contract is breached, the clawback provides a mechanism to return unvested funds to the original funder. The funder of the `ClawbackVestingAccount` is the address that sends tokens to the account at account creation. Only the funder can perform the clawback to return the funds to their account. Alternatively, they can specify a destination address to send unvested funds to.
