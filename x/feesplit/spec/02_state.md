<!--
order: 2
-->

# State

## State Objects

The `x/feesplit` module keeps the following objects in state:

| State Object          | Description                           | Key                                                               | Value              | Store |
| :-------------------- | :------------------------------------ | :---------------------------------------------------------------- | :----------------- | :---- |
| `FeeSplit`            | Fee split bytecode                     | `[]byte{1} + []byte(contract_address)`                            | `[]byte{feeSplit}` | KV    |
| `DeployerFeeSplits`   | Contract by deployer address bytecode | `[]byte{2} + []byte(deployer_address) + []byte(contract_address)` | `[]byte{1}`        | KV    |
| `WithdrawerFeeSplits` | Contract by withdraw address bytecode | `[]byte{3} + []byte(withdraw_address) + []byte(contract_address)` | `[]byte{1}`        | KV    |

### FeeSplit

A FeeSplit defines an instance that organizes fee distribution conditions for
the owner of a given smart contract

```go
type FeeSplit struct {
	// hex address of registered contract
	ContractAddress string `protobuf:"bytes,1,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
	// bech32 address of contract deployer
	DeployerAddress string `protobuf:"bytes,2,opt,name=deployer_address,json=deployerAddress,proto3" json:"deployer_address,omitempty"`
	// bech32 address of account receiving the transaction fees it defaults to
	// deployer_address
	WithdrawerAddress string `protobuf:"bytes,3,opt,name=withdrawer_address,json=withdrawerAddress,proto3" json:"withdrawer_address,omitempty"`
}
```

### ContractAddress

`ContractAddress` defines the contract address that has been registered for fee distribution.

### DeployerAddress

A `DeployerAddress` is the EOA address for a registered contract.

### WithdrawerAddress

The `WithdrawerAddress` is the address that receives transaction fees for a registered contract.

## Genesis State

The `x/feesplit` module's `GenesisState` defines the state necessary for initializing the chain from a previous exported height. It contains the module parameters and the feesplits for registered contracts:

```go
// GenesisState defines the module's genesis state.
type GenesisState struct {
	// module parameters
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// active registered contracts for fee distribution
	FeeSplits []FeeSplit `protobuf:"bytes,2,rep,name=fee_splits,json=feeSplits,proto3" json:"fee_splits"`
}

```
