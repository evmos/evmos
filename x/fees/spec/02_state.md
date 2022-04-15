<!--
order: 2
-->

# State

## State Objects

The `x/fees` module keeps the following objects in state:

| State Object        | Description                 | Key                                    | Value                        | Store |
| :------------------ | :-------------------------- | :------------------------------------- | :--------------------------- | :---- |
| `DeployerAddress`   | Deployer address bytecode   | `[]byte{1} + []byte(contract_address)` | `[]byte{deployer_address}`   | KV    |
| `WithdrawAddress`   | Withdraw address bytecode   | `[]byte{2} + []byte(contract_address)` | `[]byte{withdraw_address}`   | KV    |
| `ContractAddresses` | Contract addresses bytecode | `[]byte{3} + []byte(deployer_address)` | `[]byte{contract_addresses}` | KV    |

### DeployerAddress

Deployer address for a registered contract.

### WithdrawAddress

Address that will receive transaction fees for a registered contract.

### ContractAddresses

Slice of contract addresses registered by a developer.

## Genesis State

The `x/fees` module's `GenesisState` defines the state necessary for initializing the chain from a previous exported height. It contains the module parameters and the fee information for registered contracts:

```go
// GenesisState defines the module's genesis state.
type GenesisState struct {
	// module parameters
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// active registered contracts
	DevFeeInfos []DevFeeInfo `protobuf:"bytes,2,rep,name=dev_fee_infos,json=devFeeInfos,proto3" json:"dev_fee_infos"`
}
```
