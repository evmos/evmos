<!--
order: 3
-->

# ERC-20 Registration

Learn how to register interoperable ERC-20s through [Evmos Governance](../users/governance/overview.md). {synopsis}

## Stage 1: Drafting the ERC-20 Proposal

The following topics must be addressed when drafting an ERC-20 Proposal:

1. Applicant(s) - the profile of the person(s)/entity making the proposal
- who you are and your involvement in Cosmos and/or other blockchain networks
- an overview of team members involved and their relevant experience
- brief mission statement for your organization/business (if applicable) eg. website
- past work you've done eg. include your Github
- some sort of proof of who you are eg. Keybase
2. Background Information - promote understanding of the ERC-20 Module
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

Remember to provide links to the relevant [Commonwealth Evmos community](https://commonwealth.im/evmos) discussions concerning your proposal, as well as the [proposal on testnet](#submit-the-proposal-to-the-testnet).

## Stage 2: Submitting the ERC-20 Proposal

After the drafting process, the ERC-20 Proposal can be submitted.

### Formatting the Proposal's Text
The ideal format for a proposal is as a Markdown file (ie. `.md`) in a Github repo or [HackMd](https://hackmd.io/). Markdown
is a simple and accessible format for writing plain text files that is easy to
<!-- markdown-link-check-disable-next-line -->
learn. See the [Github Markdown Guide](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax) for details on
writing markdown files.

### Submit the Proposal to Testnet

:::tip
**Note**: Not sure how to submit a proposal to either testnet or mainnet? Check out [this document](../users/governance/submitting.md).
:::

You may want to submit your proposal to the testnet chain before the mainnet for a number of reasons, such as wanting to see what the proposal description will look like, to share what the proposal will look like in advance with stakeholders, and to signal that your proposal is about to go live on the mainnet.

Submitting your proposal to the testnet increases the likelihood of engagement and the possibility that you will be alerted to a flaw before deploying your proposal to mainnet.

## Stage 3: The On-Chain ERC-20 Proposal

A majority of the voting community should probably be aware of the proposal and have considered it before the proposal goes live on-chain. If you're taking a conservative approach, you should have reasonable confidence that your proposal will pass before risking deposit contributions. Make revisions to your draft proposal after each stage of engagement.

See the [submitting guide](../users/governance/submitting.md) for more on submitting proposals.

### The Deposit Period

The deposit period currently lasts 14 days. If you submitted your transaction with the minimum deposit (64 EVMOS), your proposal will immediately enter the voting period. If you didn't submit the minimum deposit amount (currently 64 EVMOS), then this may be an opportunity for others to show their support by contributing (and risking) their EVMOS as a bond for your proposal. You can request contributions openly and also contact stakeholders directly (particularly stakeholders who are enthusiastic about your proposal). Remember that each contributor is risking their funds, and you can [read more about the conditions for burning deposits here](../users/governance/process.md#burned-deposits).

This is a stage where proposals may begin to get broader attention. Most popular explorers currently display proposals that are in the deposit period, but due to proposal spamming, this may change.

A large cross-section of the blockchain/cryptocurrency community exists on Twitter. Having your proposal in the deposit period is a good time to engage the Evmos community to prepare validators to vote and EVMOS-holders that are staking.

### The Voting Period

At this point you'll want to track which validator has voted and which has not. You'll want to re-engage directly with top stake-holders, ie. the highest-ranking validator operators, to ensure that:

1. they are aware of your proposal;
2. they can ask you any questions about your proposal; and
3. they are prepared to vote.

Remember that any voter may change their vote at any time before the voting period ends. That historically doesn't happen often, but there may be an opportunity to convince a voter to change their vote. The biggest risk is that stakeholders won't vote at all (for a number of reasons). Validator operators tend to need multiple reminders to vote. How you choose to contact validator operators, how often, and what you say is up to you--remember that no validator is obligated to vote, and that operators are likely occupied by competing demands for their attention. Take care not to stress any potential relationship with validator operators.
