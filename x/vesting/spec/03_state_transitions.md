<!--
order: 3
-->

# State Transitions

The `x/vesting` module allows for state transitions that create and update a clawback vesting account with `CreateClawbackVestingAccount` or perform a clawback of unvested funds with `Clawback`.

## Create Clawback Vesting Account

A funder creates a new clawback vesting account defining the address to fund as well as the vesting/lockup schedules. Additionally, new grants can be added to existing clawback vesting accounts with the same message.

1. Funder submits a `MsgCreateClawbackVestingAccount` through one of the clients.
2. Check if
   1. the vesting account address is not blocked
   2. there is at least one vesting or lockup schedule provided. If one of them is absent, default to instant vesting or unlock schedule.
   3. lockup and vesting total amounts are equal
3. Create or update a clawback vesting account and send coins from the funder to the vesting account
   1. if the clawback vesting account already exists and `--merge` is set to true, add a grant to the existing total vesting amount and update the vesting and lockup schedules.
   2. else create a new clawback vesting account

## Clawback

The funding address is the only address that can perform the clawback.

1. Funder submits a `MsgClawback` through one of the clients.
2. Check if
   1. a destination address is given and default to funder address if not
   2. the destination address is not blocked
   3. the account exists and is a clawback vesting account
   4. account funder is same as in msg
3. Transfer unvested tokens from the clawback vesting account to the destination address, update the lockup schedule and remove future vesting events.

## Update Clawback Vesting Account Funder

The funding address of an existing clawback vesting account can be updated only by the current funder.

1. Funder submits a `MsgUpdateVestingFunder` through one of the clients.
2. Check if
   1. the new funder address is not blocked
   2. the vesting account exists and is a clawback vesting account
   3. account funder is same as in msg
3. Update the vesting account funder with the new funder address.
