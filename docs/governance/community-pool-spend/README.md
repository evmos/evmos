---
order: 1
parent:
  order: 5
---

# Evmos Community Pool

Evmos launched with community-spend capabilities on December 11, 2019, effectively unlocking the potential for token-holders to vote to approve spending from the Community Pool.

🇪🇸 Esta página también está [disponible en español](https://github.com/raquetelio/CosmosCommunitySpend/blob/master/README%5BES_es%5D.md).

## Why create a proposal to use Community Pool funds?

There are other funding options, most notably the Interchain Foundation's grant program. Why create a community-spend proposal?

**As a strategy: you can do both.** You can submit your proposal to the Interchain Foundation, but also consider submitting your proposal publicly on-chain. If the Hub votes in favour, you can withdraw your Interchain Foundation application.

**As a strategy: funding is fast.** Besides the time it takes to push your proposal on-chain, the only other limiting factor is a fixed 14-day voting period. As soon as the proposal passes, your account will be credited the full amount of your proposal request.

**To build rapport.** Engaging publicly with the community is the opportunity to develop relationships with stakeholders and to educate them about the importance of your work. Unforeseen partnerships could arise, and overall the community may value your work more if they are involved as stakeholders.

**To be more independent.** The Interchain Foundation (ICF) may not always be able to fund work. Having a more consistently funded source and having a report with its stakeholders means you can use your rapport to have confidence in your ability to secure funding without having to be dependent upon the ICF alone.

## Drafting a Community-spend Proposal

Drafting and submitting a proposal is a process that takes time, attention, and involves risk. The objective of this documentation is to make this process easier by preparing participants for what to pay attention to, the information that should be considered in a proposal, and how to reduce the risk of losing deposits. Ideally, a proposal should only fail to pass because the voters 1) are aware and engaged and 2) are able to make an informed decision to vote down the proposal.

If you are considering drafting a proposal, you should review the general
background on drafting and submitting a proposal:

1. [How the voting process and governance mechanism works](../process.md)
1. [How to draft your proposal and engage with the Cosmos community about it](../best-practices.md)
1. [How to format proposals](../formatting.md)
1. [How to submit your proposal](../submitting.md)

You should also review details specific to Community Pool Spend proposals below.

## Learn About the Community Pool

### How is the Community Pool funded?

2% of all staking rewards generated (via block rewards & transaction fees) are continually transferred to and accrue within the Community Pool. For example, from Dec 19, 2019 until Jan 20, 2020 (32 days), 28,726 EVMOS were generated and added to the pool.

### How can funding for the Community Pool change?

Though the rate of funding is currently fixed at 2% of staking rewards, the effective rate is dependent upon Evmos's staking rewards, which can change with inflation and block times.

The current 2% tax rate of funding may be modified with a governance proposal and enacted immediately after the proposal passes.

Currently, funds cannot be sent to the Community Pool, but we should expect this to change with the next upgrade. Read more about this new functionality [here](https://github.com/cosmos/cosmos-sdk/pull/5249). What makes this functionality important?

1. Funded projects that fail to deliver may return funding to Community Pool;
2. Entities may help fund the Community Pool by depositing funds directly to the account.

### What is the balance of the Community Pool?

You may directly query Evmos 3 for the balance of the Community Pool:

```evmosd q distribution community-pool --chain-id cosmoshub-3 --node cosmos-node-1.figment.network:26657```

Alternatively, popular Cosmos explorers such as [Big Dipper](https://cosmos.bigdipper.live) and [Hubble](https://hubble.figment.network/cosmos/chains/cosmoshub-3) display the ongoing Community Pool balance.

### How can funds from the Community Pool be spent?

Funds from the Cosmos Community Pool may be spent via successful governance proposal.

### How should funds from the Community Pool be spent?

We don't know 🤷

The prevailing assumption is that funds should be spent in a way that brings value to Evmos. However, there is debate about how to keep the fund sustainable. There is also some debate about who should receive funding. For example, part of the community believes that the funds should only be used for those who need funding most. Other topics of concern include:

- retroactive grants
- price negotiation
- fund disbursal (eg. payments in stages; payments pegged to reduce volitiliy)
- radical overhaul of how the community-spend mechanism functions

We can expect this to take shape as proposals are discussed, accepted, and rejected by Evmos community.

### How are funds disbursed after a community-spend proposal is passed?

If a community-spend proposal passes successfully, the number of EVMOS encoded in the proposal will be transferred from the community pool to the address encoded in the proposal, and this will happen immediately after the voting period ends.
