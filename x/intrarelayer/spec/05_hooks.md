<!--
order: 5
-->

# Hooks

The intrarelayer module implements two transaction hooks from the EVM and Governance modules

## EVM Hooks

::: tip
👉 **Purpose**: Allow for users to convert ERC20s to Cosmos Coins by sending an Ethereum tx transfer to the module account address. This enables native conversion of tokens via Metamask and EVM-enabled wallets.
:::

### Registered Coin: ERC20 to Coin

1. User transfers ERC20 tokens to the `ModuleAccount` address to escrow (lock)
2. Check if the ERC20 Token that was transferred from the sender is a native ERC20 or a native cosmos coin by looking at the ethereum Logs
3. If the token contract address is corresponds to the ERC20 representation of a native Cosmos Coin
    1. Call `burn()` ERC20 method from the  `ModuleAccount`
        1. NOTE: This is the same as 1.2, but since the tokens are already on the ModuleAccount balance, we burn the tokens from the module address instead of calling `burnFrom()`
        2. NOTE: We don't need to mint because (1.1) escrows the coin
    2. Transfer Cosmos Coin to the bech32 account address of the sender hex address (1.)

### Registered ERC20: ERC20 to Coin

1. User transfers coins to the Module Account to escrow (lock)
2. Check if the ERC20 Token that was transferred is a native ERC20 or a native cosmos coin
3. If the token contract address is a native ERC20 token
    1. Mint Cosmos Coin
    2. Transfer Cosmos Coin to the bech32 account address of the sender hex (1.)

## Governance Hooks

::: tip
👉 **Purpose:** speed up the approval process of a token pair registration by defining a custom `VotingPeriod` duration for the `RegisterCoinProposal` and `RegisterERC20Proposal`.
:::

### Overwriting the Voting Period

By Implementing the [GovHooks](https://github.com/cosmos/cosmos-sdk/blob/86474748888204515f59aaeab9be295066563f46/x/gov/types/expected_keepers.go#L57) Interface from the Cosmos-SDK, the voting period for all proposals of the Intrarelayer module can be customized using the `AfterProposalDeposit` hook.

1. Set the voting period  on the intrarelayer module parameters at genesis or through governance
2. Submit a new governance proposal, e.g. `RegisterERC20Proposal`
3. The `AfterProposalDeposit` hook is automatically called and overrides the voting period for all proposals to the value defined on the intrarelayer module parameters.
