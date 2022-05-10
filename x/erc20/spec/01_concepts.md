<!--
order: 1
-->

# Concepts

## Token Pair

The `x/erc20` module maintains a canonical one-to-one mapping of native Cosmos Coin denomination to ERC20 Token contract addresses (i.e `sdk.Coin` ←→ ERC20), called `TokenPair`.  The conversion of the ERC20 tokens ←→ Coin of a given pair can be enabled or disabled via governance.

## Token Pair Registration

Users can register a new token pair proposal through the governance module and initiate a vote to include the token pair in the module. Depending on which exists first, the coin or the token, you can either register a Cosmos Coin or a ERC20 Token to create a token pair.

When the proposal passes, the erc20 module registers the Cosmos Coin and ERC20 Token mapping on the application's store.

### Registration of a Cosmos Coin

A native Cosmos Coin corresponds to an `sdk.Coin` that is native to the bank module. It can be either the native staking/gas denomination (eg: EVMOS, ATOM, etc) or an IBC fungible token voucher (i.e with denom format of `ibc/{hash}`).

When a proposal is initiated for an existing native Cosmos Coin, the erc20 module will deploy a factory ERC20 contract, representing the ERC20 token for the token pair, giving the module ownership of that contract.

### Registration of an ERC20 token

A proposal for an existing (i.e already deployed) ERC20 contract can be initiated too. In this case, the ERC20 maintains the original owner of the contract and uses an escrow & mint / burn & unescrow mechanism similar to the one defined by the [ICS20 - Fungible Token Transfer](https://github.com/cosmos/ibc/blob/master/spec/app/ics-020-fungible-token-transfer) specification. The token pair is composed of the original ERC20 token and a corresponding native Cosmos coin denomination.

### Token details and metadata

Coin metadata is derived from the ERC20 token details (name, symbol, decimals) and vice versa. A special case is also described below that for the ERC20 representation of IBC fungible token (ICS20) vouchers.

#### Coin Metadata to ERC20 details

During the registration of a Cosmos Coin the following bank `Metadata` is used to deploy a ERC20 contract:

- **Name**
- **Symbol**
- **Decimals**

The native Cosmos Coin contains a more extensive metadata than the ERC20 and includes all necessary details for the conversion into a ERC20 Token, which requires no additional population of data.

#### IBC voucher Metadata to ERC20 details

IBC vouchers should comply to the following standard:

- **Name**: `{NAME} channel-{channel}`
- **Symbol**:  `ibc{NAME}-{channel}`
- **Decimals**:  derived from bank `Metadata`

#### ERC20 details to Coin Metadata

During the Registration of an ERC20 Token the Coin metadata is derived from the ERC20 metadata and the bank metadata:

- **Description**: `Cosmos coin token representation of {contractAddress}`
- **DenomUnits**:
    - Coin: `0`
    - ERC20: `{uint32(erc20Data.Decimals)}`
- **Base**: `{"erc20/%s", address}`
- **Display**: `{erc20Data.Name}`
- **Name**: `{types.CreateDenom(strContract)}`
- **Symbol:** `{erc20Data.Symbol}`

## Token Pair Modifiers

A valid token pair can be modified through several governance proposals. The internal conversion of a token pair can be toggled with `ToggleTokenConversionProposal`, so that the conversions between the token pair's tokens can be enabled or disabled.

## Token Conversion

Once a token pair proposal passes, the module allows for the conversion of that token pair. Holders of native Cosmos coins and IBC vouchers on the Evmos chain can convert their Coin into ERC20 Tokens, which can then be used in Evmos EVM, by creating a `ConvertCoin` Tx. Vice versa, the `ConvertERC20` Tx allows holders of ERC20 tokens on the Evmos chain to convert ERC-20 tokens back to their native Cosmos Coin representation.

Depending on the ownership of the ERC20 contract, the ERC20 tokens either follow a burn/mint or a transfer/escrow mechanism during conversion.

## Malicious Contracts

The ERC20 standard is an interface that defines a set of method signatures (name, arguments and output) without defining its methods' internal logic. Therefore it is possible for developers to deploy contracts that contain hidden malicious behaviour within those methods. For instance, the ERC20 `transfer` method, which is responsible for sending an `amount` of tokens to a given `recipient` could include code to siphon some amount of tokens intended for the recipient into a different predefined account, which is owned by the malicious contract deployer.

More sophisticated malicious implementations might also inherit code from customized ERC20 contracts that include malicous behaviour. For an overview of more extensive examples, plese review the x/erc20 audit, section `IF-EVMOS-06: IERC20 Contracts may execute arbitrary code`.

As the `x/erc20` module allows any arbitrary ERC20 contract to be registered through governance, it is essential that the proposer or the voters manually verify during voting phase that the proposed contract uses the default ERC20.sol implementation.

Here are our recommendations for the reviewing process:

- contract solidity code should be verified and accessable (e.g. using an explorer)
- contract should be audited by a reputabele auditor
- inherited contracts need to be verified for correctness
