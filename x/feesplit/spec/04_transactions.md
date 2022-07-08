<!--
order: 4
-->

# Transactions

This section defines the `sdk.Msg` concrete types that result in the state transitions defined on the previous section.

### `MsgRegisterFeeSplit`

Defines a transaction signed by a developer to register a contract for transaction fee distribution. The sender must be an EOA that corresponds to the contract deployer address.

```go
type MsgRegisterFeeSplit struct {
	// contract hex address
	ContractAddress string `protobuf:"bytes,1,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
	// bech32 address of message sender, must be the same as the origin EOA
	// sending the transaction which deploys the contract
	DeployerAddress string `protobuf:"bytes,2,opt,name=deployer_address,json=deployerAddress,proto3" json:"deployer_address,omitempty"`
	// bech32 address of account receiving the transaction fees
	WithdrawerAddress string `protobuf:"bytes,3,opt,name=withdraw_address,json=withdrawerAddress,proto3" json:"withdraw_address,omitempty"`
	// array of nonces from the address path, where the last nonce is
	// the nonce that determines the contract's address - it can be an EOA nonce
	// or a factory contract nonce
	Nonces []uint64 `protobuf:"varint,4,rep,packed,name=nonces,proto3" json:"nonces,omitempty"`
}
```

The message content stateless validation fails if:

- Contract hex address is invalid
- Contract hex address is zero
- Deployer bech32 address is invalid
- Withdraw bech32 address is invalid
- Nonces array is empty

### `MsgUpdateFeeSplit`

Defines a transaction signed by a developer to update the withdraw address of a contract registered for transaction fee distribution. The sender must be an EOA that corresponds to the contract deployer address.

```go
type MsgUpdateFeeSplit struct {
	// contract hex address
	ContractAddress string `protobuf:"bytes,1,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
	// deployer bech32 address
	DeployerAddress string `protobuf:"bytes,2,opt,name=deployer_address,json=deployerAddress,proto3" json:"deployer_address,omitempty"`
	// new withdraw bech32 address for receiving the transaction fees
	WithdrawerAddress string `protobuf:"bytes,3,opt,name=withdraw_address,json=withdrawerAddress,proto3" json:"withdraw_address,omitempty"`
}
```

The message content stateless validation fails if:

- Contract hex address is invalid
- Contract hex address is zero
- Deployer bech32 address is invalid
- Withdraw bech32 address is invalid
- Withdraw bech32 address is same as deployer address

### `MsgCancelFeeSplit`

Defines a transaction signed by a developer to remove the information for a registered contract. Transaction fees will no longer be distributed to the developer, for this smart contract. The sender must be an EOA that corresponds to the contract deployer address.

```go
type MsgCancelFeeSplit struct {
	// contract hex address
	ContractAddress string `protobuf:"bytes,1,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
	// deployer bech32 address
	DeployerAddress string `protobuf:"bytes,2,opt,name=deployer_address,json=deployerAddress,proto3" json:"deployer_address,omitempty"`
}
```

The message content stateless validation fails if:

- Contract hex address is invalid
- Contract hex address is zero
- Deployer bech32 address is invalid
