<!--
order: 3
-->

# State Transitions

## ABCI

### End Block

The ABCI EndBlock checks if the airdrop has ended in order to process the clawback of unclaimed tokens.

1. Check if the airdrop has has concluded
    - if the global flag is enabled
    - if the current block time is greater than the airdrop end time
2. Clawback tokens from the module escrow account by transferring the escrow account balance to the community pool
3. Transfer balance from empty accounts to community pool
    - if the account has sequence of 0, i.e. no transactions submitted
    - if balance amount is positive
4. Prune all the claim records from the state
5. Disable any further claim by setting the global parameter to `false`
