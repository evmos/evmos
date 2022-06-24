<!--
order: 2
-->

# State

## State Objects

The `x/fees` module keeps the following objects in state:

| State Object   | Description                           | Key                                                               | Value         | Store |
| :------------- | :------------------------------------ | :---------------------------------------------------------------- | :------------ | :---- |
| `Fee`          | Fee bytecode                          | `[]byte{1} + []byte(contract_address)`                            | `[]byte{fee}` | KV    |
| `DeployerFees` | Contract by deployer address bytecode | `[]byte{2} + []byte(deployer_address) + []byte(contract_address)` | `[]byte{1}`   | KV    |
| `WithdrawFees` | Contract by withdraw address bytecode | `[]byte{3} + []byte(withdraw_address) + []byte(contract_address)` | `[]byte{1}`   | KV    |

### DeployerAddress

A `DeployerAddress` is the EOA address for a registered contract.

### WithdrawAddress

The `WithdrawAddress` is the address that receives transaction fees for a registered contract.

### ContractAddresses

`ContractAddresses` defines a slice of all contract addresses registered by a deployer.

## Genesis State

The `x/fees` module's `GenesisState` defines the state necessary for initializing the chain from a previous exported height. It contains the module parameters and the fee for registered contracts:

```go
// GenesisState defines the module's genesis state.
type GenesisState struct {
	// module parameters
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// active registered contracts
	Fees []Fee `protobuf:"bytes,2,rep,name=dev_fee_infos,json=devFeeInfos,proto3" json:"dev_fee_infos"`
}
```
