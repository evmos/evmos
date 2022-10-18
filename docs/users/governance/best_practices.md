<!--
order: 3
-->

# Best Practices

::: tip
**Note:**

- If users are creating governance proposals which require community pool funding (such as those of type `CommunityPoolSpendProposal`), refer to [this section](#community-spend-proposal).
- If users are creating governance proposals concerned with the ERC-20 Module (such as those of type `RegisterCoinProposal`), refer to [this section](#erc-20-proposal).
- If users are creating governance proposals concerned with changing parameters (such as those of type `ParameterChangeProposal`), refer to [this section](#parameter-change-proposal).

:::

## General Advice: Community Outreach

Engagement is likely to be critical to the success of a proposal. The degree to which you engage with Evmos community should be relative to the potential impact that your proposal may have on the stakeholders. This guide does not cover all ways of engaging: you could bring your idea to a podcast or a hackathon, host an AMA on [Reddit](https://www.reddit.com/r/evmos) or host a Q&A (questions & answers). We encourage you to experiment and use your strengths to introduce proposal ideas and gather feedback.

There are many different ways to engage. One strategy involves a few stages of engagement before and after submitting a proposal on chain. **Why do it in stages?** It's a more conservative approach to save resources. The idea is to check in with key stakeholders at each stage before investing more resources into developing your proposal.

In the first stage of this strategy, you should engage people (ideally experts) informally about your idea. You'll want to start with the minimal, critical components (name, value to cosmos hub, timeline, any funding needs) and check:

- Does it make sense?
- Are there critical flaws?
- Does it need to be reconsidered?

You should be able engaging with key stakeholders (eg. a large validator operator) with a few short sentences to measure their support. Here's an example:

> "We are considering a proposal for funding to work on (project). We think it will help Evmos to (outcome). Timeline is (x), and we're asking for (y) amount. Do you think that this is a proposal that (large validator) may support?"

**Why a large validator?** They tend to be the de facto decision-makers on Evmos, since their delegators also delegate their voting power. If you can establish a base layer of off-chain support, you can be more confident that it's worth proceeding to the next stage.

::: tip
**Note:** many will likely hesitate to commit support, and that's okay. It will be important to reassure these stakeholders that this isn't a binding a commitment. You're just canvassing the community to get a feel for whether it's worthwhile to proceed. It's also an opportunity to connect with new people and to answer their questions about what it is you're working on. It will be important for them to clearly understand why you think what you're proposing will be valuable to Evmos, and if possible, why it will be valuable to them as long-term stakeholders.
:::

- If you're just developing your idea, [start at Stage 1](#stage-1-your-idea).
- If you're already confident about your idea, [skip to Stage 2](#stage-2-your-draft-proposal).
- If you've drafted your proposal, engaged with the community, and submitted your proposal to the testnet, [skip to Stage 3](#stage-3-your-on-chain-proposal).

## Stage 1: Your Idea

### Not yet confident about your idea?

Great! Governance proposals potentially impact many stakeholders. Introduce your idea with known members of the community before investing resources into drafting a proposal. Don't let negative feedback dissuade you from exploring your idea if you think that it's still important.

If you know people who are very involved with Evmos, send them a private message with a concise overview of what you think will result from your idea or proposed changes. Wait for them to ask questions before providing details. Do the same in semi-private channels where people tend to be respectful (and hopefully supportive).

### Confident with your idea?

Great! However, remember that governance proposals potentially impact many stakeholders, which can happen in unexpected ways. Introduce your idea with members of the community before investing resources into drafting a proposal. At this point you should seek out and carefully consider critical feedback in order to protect yourself from [confirmation bias](https://en.wikipedia.org/wiki/Confirmation_bias). This is the ideal time to see a critical flaw, because submitting a flawed proposal will waste resources.

### Are you ready to draft a governance proposal?

There will likely be differences of opinion about the value of what you're proposing to do and the strategy by which you're planning to do it. If you've considered feedback from broad perspectives and think that what you're doing is valuable and that your strategy should work, and you believe that others feel this way as well, it's likely worth drafting a proposal. However, remember that the largest EVMOS stakers have the biggest vote, so a vocal minority isn't necessarily representative or predictive of the outcome of an on-chain vote.

A conservative approach is to have some confidence that you roughly have initial support from a majority of the voting power before proceeding to drafting your proposal. However, there are likely other approaches, and if your idea is important enough, you may want to pursue it regardless of whether or not you are confident that the voting power will support it.

## Stage 2: Your Draft Proposal

The next major section outlines and describes some potential elements of drafting a proposal. Ensure that you have considered your proposal and anticipated questions that the community will likely ask. Once your proposal is on-chain, you will not be able to change it.

### Proposal Elements

It will be important to balance two things: being detailed and being concise. You'll want to be concise so that people can assess your proposal quickly. You'll want to be detailed so that voters will have a clear, meaningful understanding of what the changes are and how they are likely to be impacted.

Every proposal should contain a summary with key details:

- who is submitting the proposal
- the amount of the proposal or parameter(s) being changed;
- and deliverables and timeline
- a reason for the proposal and potential impacts
- a short summary of the history (what compelled this proposal), solution that's being presented, and future expectations

Assume that many people will stop reading at this point. However, it is important to provide in-depth information, so a few more pointers for Parameter-Change, Community Spend, and ERC-20 Module proposals are below.

#### Parameter-Change Proposal

1. Problem/Value - generally the problem or value that's motivating the parameter change(s)
2. Solution - generally how changing the parameter(s) will address the problem or improve the network
   - the beneficiaries of the change(s) (ie. who will these changes impact and how?)
      - voters should understand the importance of the change(s) in a simple way
3. Risks & Benefits - clearly describe how making this/these change(s) may expose stakeholders to new benefits and/or risks
4. Supplementary materials - optional materials eg. models, graphs, tables, research, signed petition, etc

#### Community Spend Proposal

1. Applicant(s) - the profile of the person(s)/entity making the proposal
   - who you are and your involvement in Cosmos and/or other blockchain networks
   - an overview of team members involved and their relevant experience
   - brief mission statement for your organization/business (if applicable) eg. website
   - past work you've done eg. include your Github
   - some sort of proof of who you are eg. Keybase
2. Problem - generally what you're solving and/or opportunity you're addressing
   - provide relevant information about both past and present issues created by this problem
   - give suggestions as to the state of the future if this work is not completed
3. Solution - generally how you're proposing to deliver the solution
   - your plan to fix the problem or deliver value
   - the beneficiaries of this plan (ie. who will your plan impact and how?)
     - follow the "as a user" template ie. write a short user story about the problem you are trying to solve and how users will interact with what you're proposing to deliver (eg. benefits and functionality from a user’s perspective)
     - voters should understand the value of what you're providing in a simple way
   - your reasons for selecting this plan
   - your motivation for delivering this solution/value
4. Funding - amount and denomination proposed eg. 5000 EVMOS
   - the entity controlling the account receiving the funding
   - consider an itemized breakdown of funding per major deliverable
   - consider outlining how the funds will be spent
5. Deliverables and timeline - the specifics of what you're delivering and how, and what to expect
   - what are the specific deliverables? (be detailed)
   - when will each of these be delivered?
   - will there be a date at which the project will be considered failed if the deliverables have not been met?
   - how will each of these be delivered?
   - what will happen if you do not deliver on time?
     - what is the deadline for the project to be considered failed?
     - do you have a plan to return the funds?
   - how will you be accountable to Evmos stakeholders?
     - how will you communicate updates and how often?
     - how can the community observe your progress?
     - how can the community provide feedback?
   - how should the quality of deliverables be assessed? eg. metrics
5. Relationships and disclosures
   - have you received or applied for grants or funding? for similar work? eg. from the [Evmos Grants Program](https://medium.com/evmos/announcing-evmos-grants-78aa28562db6)
   - how will you and/or your organization benefit?
   - do you see this work continuing in the future and is there a plan?
   - what are the risks involved with this work?
   - do you have conflicts of interest to declare?

#### ERC-20 Proposal

1. Applicant(s) - the profile of the person(s)/entity making the proposal
   - who you are and your involvement in Cosmos and/or other blockchain networks
   - an overview of team members involved and their relevant experience
   - brief mission statement for your organization/business (if applicable) eg. website
   - past work you've done eg. include your Github
   - some sort of proof of who you are eg. Keybase
2. Background information - promote understanding of the ERC-20 Module
   - a mention of the original [blog post](https://medium.com/evmos/introducing-evmos-erc20-module-f40a61e05273) that introduced the ERC-20 Module
   - a brief explanation of what the ERC-20 Module does
   - a mention of the [ERC-20 Module documentation](https://docs.evmos.org/modules/erc20/)
3. Solution - generally how ERC-20 Module changes will be made
   - a brief explanation of what the proposal will do if it passes
   - a brief explanation of the precautions taken, how it was tested, and who was consulted prior to making the proposal
   - a breakdown of the proposal's payload, and third-party review
   - a brief explanation of the risks involved (depending on the direction of IBC Coin, ERC-20)
   - ensure the following are both adhered to and documented:
      - the contracts are verified (either through the [EVM explorer](https://evm.evmos.org) or via [Sourcify](https://sourcify.dev))
      - the contracts are deployed open-source
      - the contracts do not extend the `IERC20.sol` interface through a malicious implementation
      - the contracts use the main libraries for ERC-20s (eg. [OpenZeppelin](https://docs.openzeppelin.com/contracts/4.x/erc20), [dapp.tools](https://dapp.tools/))
      - the transfer logic is not modified (i.e. transfer logic is not directly manipulated)
      - no malicious `Approve` events can directly manipulate users' balance through a delayed granted allowance

Remember to provide links to the relevant [Commonwealth Evmos community](https://commonwealth.im/evmos) discussions concerning your proposal, as well as the [proposal on testnet](#submit-your-proposal-to-the-testnet).

### Begin with a well-considered draft proposal

The ideal format for a proposal is as a Markdown file (ie. `.md`) in a Github repo or [HackMd](https://hackmd.io/). Markdown
is a simple and accessible format for writing plain text files that is easy to
<!-- markdown-link-check-disable-next-line -->
learn. See the [Github Markdown Guide](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax) for details on
writing markdown files.

### Engage the community with your draft proposal

1. Post a discussion in the [Commonwealth Evmos community](https://commonwealth.im/evmos). Ideally this should contain a link to this repository, either directly to your proposal if it has been merged, or else to a pull-request containing your proposal if it has not been merged yet.
2. Directly engage key members of the community for feedback. These could be large contributors, those likely to be most impacted by the proposal, and entities with high stake-backing (eg. high-ranked validators; large stakers).
3. Target members of the community in a semi-public way before bringing the draft to a full public audience. The burden of public scrutiny in a semi-anonymized environment (eg. Twitter) can be stressful and overwhelming without establishing support. Solicit opinions in places with people who have established reputations first.

### Submit your proposal to the testnet

:::tip
**Note**: Not sure how to submit a proposal to either testnet or mainnet? Check out [this document](./submitting.md).
:::

You may want to submit your proposal to the testnet chain before the mainnet for a number of reasons, such as wanting to see what the proposal description will look like, to share what the proposal will look like in advance with stakeholders, and to signal that your proposal is about to go live on the mainnet.

Perhaps most importantly, for parameter change proposals, you can test the parameter changes in advance (if you have enough support from the voting power on the testnet).

Submitting your proposal to the testnet increases the likelihood of engagement and the possibility that you will be alerted to a flaw before deploying your proposal to mainnet.

## Stage 3: Your On-Chain Proposal

A majority of the voting community should probably be aware of the proposal and have considered it before the proposal goes live on-chain. If you're taking a conservative approach, you should have reasonable confidence that your proposal will pass before risking deposit contributions. Make revisions to your draft proposal after each stage of engagement.

See the [submitting guide](./submitting.md) for more on submitting proposals.

### The Deposit Period

The deposit period currently lasts 3 days. If you submitted your transaction with the minimum deposit (192 EVMOS), your proposal will immediately enter the voting period. If you didn't submit the minimum deposit amount (currently 192 EVMOS), then this may be an opportunity for others to show their support by contributing (and risking) their EVMOS as a bond for your proposal. You can request contributions openly and also contact stakeholders directly (particularly stakeholders who are enthusiastic about your proposal). Remember that each contributor is risking their funds, and you can [read more about the conditions for burning deposits here](./process.md#burned-deposits).

This is a stage where proposals may begin to get broader attention. Most popular explorers currently display proposals that are in the deposit period, but due to proposal spamming, this may change.

A large cross-section of the blockchain/cryptocurrency community exists on Twitter. Having your proposal in the deposit period is a good time to engage the Evmos community to prepare validators to vote and EVMOS-holders that are staking.

### The Voting Period

At this point you'll want to track which validator has voted and which has not. You'll want to re-engage directly with top stake-holders, ie. the highest-ranking validator operators, to ensure that:

1. they are aware of your proposal;
2. they can ask you any questions about your proposal; and
3. they are prepared to vote.

Remember that any voter may change their vote at any time before the voting period ends. That historically doesn't happen often, but there may be an opportunity to convince a voter to change their vote. The biggest risk is that stakeholders won't vote at all (for a number of reasons). Validator operators tend to need multiple reminders to vote. How you choose to contact validator operators, how often, and what you say is up to you--remember that no validator is obligated to vote, and that operators are likely occupied by competing demands for their attention. Take care not to stress any potential relationship with validator operators.
