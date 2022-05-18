<!--
order: 2
-->

# State

## State Objects

The `x/erc20` module keeps the following objects in state:

| State Object       | Description                                    | Key                         | Value               | Store    |
| ------------------ | ---------------------------------------------- | --------------------------- | ------------------- | --- |
| `TokenPair`        | Token Pair bytecode                            | `[]byte{1} + []byte(id)`    | `[]byte{tokenPair}` | KV    |
| `TokenPairByERC20` | Token Pair id bytecode by erc20 contract bytes | `[]byte{2} + []byte(erc20)` | `[]byte(id)`        | KV    |
| `TokenPairByDenom` | Token Pair id bytecode by denom string         | `[]byte{3} + []byte(denom)` | `[]byte(id)`        | KV    |

### Token Pair

One-to-one mapping of native Cosmos coin denomination to ERC20 token contract addresses (i.e `sdk.Coin` ←→ ERC20).

```go
type TokenPair struct {
	// address of ERC20 contract token
	Erc20Address string `protobuf:"bytes,1,opt,name=erc20_address,json=erc20Address,proto3" json:"erc20_address,omitempty"`
	// cosmos base denomination to be mapped to
	Denom string `protobuf:"bytes,2,opt,name=denom,proto3" json:"denom,omitempty"`
	// shows token mapping enable status
	Enabled bool `protobuf:"varint,3,opt,name=enabled,proto3" json:"enabled,omitempty"`
	// ERC20 owner address ENUM (0 invalid, 1 ModuleAccount, 2 external address
	ContractOwner Owner `protobuf:"varint,4,opt,name=contract_owner,json=contractOwner,proto3,enum=evmos.erc20.v1.Owner" json:"contract_owner,omitempty"`
}
```

### Token pair ID

The unique identifier of a `TokenPair` is obtained by obtaining the SHA256 hash of the ERC20 hex contract address and the Coin denomination using the following function:

```tsx
tokenPairId = sha256(erc20 + "|" + denom)
```

### Token Origin

The `ConvertCoin` and `ConvertERC20` functionalities use the owner field to check whether the token being used is a native Coin or a native ERC20. The field is based on the token registration proposal type (`RegisterCoinProposal` = 1, `RegisterERC20Proposal` = 2).

The `Owner` enumerates the ownership of a ERC20 contract.

```go
type Owner int32

const (
	// OWNER_UNSPECIFIED defines an invalid/undefined owner.
	OWNER_UNSPECIFIED Owner = 0
	// OWNER_MODULE erc20 is owned by the erc20 module account.
	OWNER_MODULE Owner = 1
	// EXTERNAL erc20 is owned by an external account.
	OWNER_EXTERNAL Owner = 2
)
```

The `Owner` can be checked with the following helper functions:

```go
// IsNativeCoin returns true if the owner of the ERC20 contract is the
// erc20 module account
func (tp TokenPair) IsNativeCoin() bool {
	return tp.ContractOwner == OWNER_MODULE
}

// IsNativeERC20 returns true if the owner of the ERC20 contract not the
// erc20 module account
func (tp TokenPair) IsNativeERC20() bool {
	return tp.ContractOwner == OWNER_EXTERNAL
}
```

### Token Pair by ERC20 and by Denom

`TokenPairByERC20` and `TokenPairByDenom` are additional state objects for querying a token pair id.

## Genesis State

The `x/erc20` module's `GenesisState` defines the state necessary for initializing the chain from a previous exported height. It contains the module parameters and the registered token pairs :

```go
// GenesisState defines the module's genesis state.
type GenesisState struct {
	// module parameters
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// registered token pairs
	TokenPairs []TokenPair `protobuf:"bytes,2,rep,name=token_pairs,json=tokenPairs,proto3" json:"token_pairs"`
}
```
