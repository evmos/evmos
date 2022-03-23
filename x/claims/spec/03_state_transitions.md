<!--
order: 3
-->

# State Transitions

## ABCI

### End Block

The ABCI EndBlock checks if the airdrop has ended in order to process the clawback of unclaimed tokens.

1. Check if the airdrop has concluded. This is the case if:
    - the global flag is enabled
    - the current block time is greater than the airdrop end time
2. Clawback tokens from the escrow account that holds the unclaimed tokens by transferring its balance to the community pool
3. Clawback tokens from empty user accounts by transferring the balance from empty user accounts with claims records to the community pool if:
    - the account has a sequence number of 0, i.e. no transactions submitted, and
    - the balance amount is positive
    - the account is not a vesting account
4. Prune all the claim records from the state
5. Disable any further claim by setting the global parameter to `false`
