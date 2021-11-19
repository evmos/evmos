<!--
order: 7
-->

# Parameters

The intrarelayer module contains the following parameters:

[Untitled](https://www.notion.so/09ff02e6cf524480a497a48fedf174ff)

## Enable Intrarelayer

The `EnableIntrarelayer` parameter toggles all state transitions in the module. When the parameter is disabled, it will prevent all Tokenpair registration and conversion functionality.

## Token Pair Voting Period

The `TokenPairVotingPeriod` parameter defines the period of time in which validators can submit their vote for a token pair registration proposal. This value overrides the default value of the governance module.

### Enable EVM Hook

The `EnableEVMHook` parameter enables the EVM hook to convert an ERC20 token to a Cosmos Coin by transferring the Tokens through a `MsgEthereumTx`  to the `ModuleAddress` Ethereum address.
