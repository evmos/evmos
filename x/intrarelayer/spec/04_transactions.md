<!--
order: 4
-->

# Transactions

This section defines the `sdk.Msg` concrete types that result in the state transitions defined on the previous section.

## `RegisterCoinProposal`

A gov `Content` type to register a token pair from a Cosmos Coin. Governance users vote on this proposal and it automatically executes the custom handler for `RegisterCoinProposal` when the vote passes.

```go
type RegisterCoinProposal struct {
	// title of the proposal
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	// proposal description
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	// token pair of Cosmos native denom and ERC20 token address
	Metadata types.Metadata `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata"`
}
```

The proposal content stateless validation fails if:

- Title is invalid (length or char)
- Description is invalid (length or char)
- Metadata is invalid
    - Name and Symbol are not blank
    - Base and Display denominations are valid coin denominations
    - Base and Display denominations are present in the DenomUnit slice
    - Base denomination has exponent 0
    - Denomination units are sorted in ascending order
    - Denomination units not duplicated

## `RegisterERC20Proposal`

A gov `Content` type to register a token pair from an ERC20 Token. Governance users vote on this proposal and it automatically executes the custom handler for `RegisterERC20Proposal` when the vote passes.

```go
type RegisterERC20Proposal struct {
	// title of the proposal
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	// proposal description
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	// contract address of ERC20 token
	Erc20Address string `protobuf:"bytes,3,opt,name=erc20address,proto3" json:"erc20address,omitempty"`
}
```

The proposal Content stateless validation fails if:

- Title is invalid (length or char)
- Description is invalid (length or char)
- ERC20Address is invalid

## `MsgConvertCoin`

A user broadcasts a `MsgConvertCoin` message to convert a Cosmos Coin to a ERC20 token.

```go
type MsgConvertCoin struct {
	// Cosmos coin which denomination is registered on intrarelayer bridge.
	// The coin amount defines the total ERC20 tokens to convert.
	Coin types.Coin `protobuf:"bytes,1,opt,name=coin,proto3" json:"coin"`
	// recipient hex address to receive ERC20 token
	Receiver string `protobuf:"bytes,2,opt,name=receiver,proto3" json:"receiver,omitempty"`
	// cosmos bech32 address from the owner of the given ERC20 tokens
	Sender string `protobuf:"bytes,3,opt,name=sender,proto3" json:"sender,omitempty"`
}
```

Message stateless validation fails if:

- Coin is invalid (invalid denom or non-positive amount)
- Receiver hex address is invalid
- Sender bech32 address is invalid

## `MsgConvertERC20`

A user broadcasts a `MsgConvertERC20` message to convert a ERC20 token to a native Cosmos coin.

```go
type MsgConvertERC20 struct {
	// ERC20 token contract address registered on intrarelayer bridge
	ContractAddress string `protobuf:"bytes,1,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
	// amount of ERC20 tokens to mint
	Amount github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,2,opt,name=amount,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"amount"`
	// bech32 address to receive SDK coins.
	Receiver string `protobuf:"bytes,3,opt,name=receiver,proto3" json:"receiver,omitempty"`
	// sender hex address from the owner of the given ERC20 tokens
	Sender string `protobuf:"bytes,4,opt,name=sender,proto3" json:"sender,omitempty"`
}
```

Message stateless validation fails if:

- Contract address is invalid
- Amount is not positive
- Receiver bech32 address is invalid
- Sender hex address is invalid

## `ToggleTokenRelayProposal`

A gov Content type to toggle the internal relaying of a token pair.

```go
type ToggleTokenRelayProposal struct {
	// title of the proposal
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	// proposal description
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	// token identifier can be either the hex contract address of the ERC20 or the
	// Cosmos base denomination
	Token string `protobuf:"bytes,3,opt,name=token,proto3" json:"token,omitempty"`
}
```

## `UpdateTokenPairERC20Proposal`

A gov Content type to update a token pair's ERC20 contract address.

```go
type UpdateTokenPairERC20Proposal struct {
	// title of the proposal
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	// proposal description
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	// contract address of ERC20 token
	Erc20Address string `protobuf:"bytes,3,opt,name=erc20_address,json=erc20Address,proto3" json:"erc20_address,omitempty"`
	// new address of ERC20 token contract
	NewErc20Address string `protobuf:"bytes,4,opt,name=new_erc20_address,json=newErc20Address,proto3" json:"new_erc20_address,omitempty"`
}
```
