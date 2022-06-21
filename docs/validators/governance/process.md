<!--
order: 2
-->

# Proposal Process

## Deposit Period

The deposit period lasts either 5 days or until the proposal deposit totals 64 EVMOS, whichever happens first.

### Deposits

Deposit amounts are at risk of being burned. Prior to a governance proposal entering the voting period (ie. for the proposal to be voted upon), there must be at least a minimum number of EVMOS deposited (64). Anyone may contribute to this deposit. Deposits of passed and failed proposals are returned to the contributors.

In the past, different people have considered contributions amounts differently. There is some consensus that this should be a personal choice. There is also some consensus that this can be an opportunity for supporters to signal their support by adding to the deposit amount, so a proposer may choose to leave contribution room (ie. a deposit below 64 EVMOS) so that others may participate. It is important to remember that any contributed EVMOS are at risk of being burned.

### Burned deposits

Deposits are burned when proposals:

1. **Expire** - deposits will be burned if the deposit period ends before reaching the minimum deposit (64 EVMOS)
2. **Fail to reach quorum** - deposits will be burned for proposals that do not reach quorum ie. 40% of all staked EVMOS must vote
3. **Are vetoed** - deposits for proposals with 33.4% of voting power backing the `NoWithVeto` option are also burned

## Voting Period

The voting period is currently a fixed 5-day period. During the voting period, participants may select a vote of either `Yes`, `No`, `Abstain`, or `NoWithVeto`. Voters may change their vote at any time before the voting period ends.

## What do the voting options mean?

1. **`Abstain`**: indicates that the voter is impartial to the outcome of the proposal.
2. **`Yes`**: indicates approval of the proposal in its current form.
3. **`No`**: indicates disapproval of the proposal in its current form.
4. **`NoWithVeto`**: indicates stronger opposition to the proposal than simply voting `No`. If the number of `NoWithVeto` votes is greater than a third of total votes excluding `Abstain` votes, the proposal is rejected and the deposits are [burned](#burned-deposits).

As accepted by the community in [Proposal 6](https://ipfs.io/ipfs/QmRtR7qkeaZCpCzHDwHgJeJAZdTrbmHLxFDYXhw7RoF1pp), voters are expected to vote `NoWithVeto` if a proposal leads to undesirable outcomes for the community. It states “if a proposal seems to be spam or is deemed to have caused a negative externality to Cosmos community, voters should vote `NoWithVeto`.”

Voting `NoWithVeto` provides a mechanism for a minority group representing a *third* of the participating voting power to reject a proposal that would otherwise pass. This makes explicit an aspect of the consensus protocol: it works as long as only up to [a third of nodes fail](https://docs.tendermint.com/v0.35/introduction/what-is-tendermint.html). In other words, greater than a third of validators are always in a position to cause a proposal to fail outside the formalized governance process and the network's norms, such as by censoring transactions. The purpose of internalizing this aspect of the consensus protocol into the governance process is to discourage validators from relying on collusion and censorship tactics to influence voting outcomes.

## What determines whether or not a governance proposal passes?

There are four criteria:

1. A minimum deposit of 64 EVMOS is required for the proposal to enter the voting period
   - anyone may contribute to this deposit
   - the deposit must be reached within 14 days (this is the deposit period)
2. A minimum of 40% of the network's voting power (quorum) is required to participate to make the proposal valid
3. A simple majority (greater than 50%) of the participating voting power must back the `Yes` vote during the 14-day voting period
4. Less than 33.4% of participating voting power votes `NoWithVeto`

Currently, the criteria for submitting and passing/failing all proposal types is the same.

### How is voting tallied?

Voting power is determined by stake weight at the end of the 14-day voting period and is proportional to the number of total EVMOS participating in the vote. Only bonded EVMOS count towards the voting power for a governance proposal. Liquid EVMOS will not count toward a vote or quorum.

Inactive validators can cast a vote, but their voting power (including the backing of their delegators) will not count toward the vote if they are not in the active set when the voting period ends. That means that if I delegate to a validator that is either jailed, tombstoned, or ranked lower than 125 in stake-backing at the time that the voting period ends, my stake-weight will not count in the vote.

Though a simple majority `Yes` vote (ie. 50% of participating voting power) is required for a governance proposal vote to pass, a `NoWithVeto` vote of 33.4% of participating voting power or greater can override this outcome and cause the proposal to fail. This enables a minority group representing greater than 1/3 of voting power to fail a proposal that would otherwise pass.

### How is quorum determined?

Voting power, whether backing a vote of `Yes`, `Abstain`, `No`, or `NoWithVeto`, counts toward quorum. Quorum is required for the outcome of a governance proposal vote to be considered valid and for deposit contributors to recover their deposit amounts. If the proposal vote does not reach quorum (ie. less than 40% of the network's voting power is participating) within 14 days, any deposit amounts will be burned and the proposal outcome will not be considered to be valid.
