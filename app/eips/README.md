# Evmos Custom EIPs

This document explain how **evmOS** allows chain built on top of it to define custom EIPs to modify the behavior of EVM
opcodes.

## Custom EIPs

Inside an EVM, every state transition or query is executed by evaluating opcodes. Custom EIPs are functions used to
change the behavior of these opcodes to tailor the EVM functionalities to fit the app-chain requirements.

Custom EIPs should be defined in an `eips` package inside the `./app/eips/` folder of chains using the **evmOS**
framework. This organization of custom implementations is not a strict requirement, but is the suggested approach to
have a clean organization of functionalities. In this file, only the custom modifier should be defined.

Inside this package, custom EIP should be defined in a file called `eips.go`. In this file, the EIPs modifier should be
defined with the signature:

```go
func(jt *vm.JumpTable) {}
```

where `vm` is the package `"github.com/evmos/evmos/v20/x/evm/core/vm"`.

Custom EIPs are used to modify the behavior of opcodes, which are described by the `operation` structure:

```go
type operation struct {
	// execute is the operation function
	execute     executionFunc
	constantGas uint64
	dynamicGas  gasFunc
	// minStack tells how many stack items are required
	minStack int
	// maxStack specifies the max length the stack can have for this operation
	// to not overflow the stack.
	maxStack int

	// memorySize returns the memory size required for the operation
	memorySize memorySizeFunc
}
```

With the **evmOS** framework, it is possible to modify any of the fields defined in the type via the `operation` setter
methods:

- `SetExecute`: update the execution logic for the opcode.

- `SetConstantGas`: update the value used for the constant gas cost.

- `SetDynamicGas`: update the function used to compute the dynamic gas cost.

- `SetMinStack`: update the minimum number of items in the stack required to execute the `operation`.

- `SetMaxStack`: update the maximum number of items that will be in the stack after executing the `operation`.

- `SetMemorySize`: the memory size required by the `operation`.

An example for an EIP which modifies the constant gas used for the `CREATE` opcode is reported below:

```go
// Enable a custom EIP-0000
func Enable0000(jt *vm.JumpTable) {
	jt[vm.CREATE].SetConstantGas(1)
}
```

In the same folder should also be defined tests and contracts used to verify the EIPs logic.

## Activate Custom EIPs

The activation of custom EIPs should be done inside the `config.go` file defined in the `./app/` folder. This file has
the role of the single source for modify the EVM implementation which is defined in the
[`x/evm/`](https://github.com/evmos/evmos/tree/main/x/evm) folder
of **evmOS**.

In this file, 3 main components should be defined:

- The custom EIPs, also called activators.
- The additional default EIPs enabled.
- The EVM configurator instance.

All these components will be described in the following sections.

### Opcode & EIP Activators

Activators is the name provided by [Go-ethereum](https://geth.ethereum.org/) to the definition of the structure
grouping all possible non-default EIPs:

```go
var activators = map[int]func(*JumpTable){
	3855: enable3855,
    ...
}
```

It can be interpreted as a list of available functionalities that can be toggled to change opcodes behavior. The
structure is a map where the key is the EIP number in the octal representation, and the value is the custom EIP
function that has to be evaluated.

In **evmOS**, custom activators should be defined in a structure with the same data type, like in the example below:

```go
// Activate custom EIPs: 0000, 0001, 0002, etc
evmosActivators = map[int]func(*vm.JumpTable){
	"evmos_0": eips.Enable0000,
	"evmos_1": eips.Enable0001,
	"evmos_2": eips.Enable0002,
}
```

It should be noted that the value of each key in the example is the modifier defined in the `eips` package in the
example provided at the of the [Custom EIPs](#custom-eips) section.

### Default EIPs

Custom EIPs defined in the `activators` map are not enabled by default. This type is only used to define the list of
custom functionalities that can be activated. To specify which custom EIP activate, we should modify the
**evmOS** `x/evm` module params. The parameter orchestrating enabled custom EIPs is the `DefaultExtraEIPs` and
**evmOS** provide an easy and safe way to customize it.

To specify which activator enable in the chain, a new variable containing a slice of keys of the custom activators
should be defined. An example is reported below:

```go
evmosEnabledEIPs = []int64{
    "evmos_0",
}
```

In this way, even though the custom activators defined $3$ new EIPs, we are going to activate only the number `evmos_0`

### EVM Configurator

The EVM configuration is the type used to modify the EVM configuration before starting a node. The type is defined as:

```go
type EVMConfigurator struct {
	extendedEIPs             map[int]func(*vm.JumpTable)
	extendedDefaultExtraEIPs []int64
	sealed                   bool
}
```

Currently, only 2 customizations are possible:

- `WithExtendedEips`: extended the default available EIPs.

- `WithExtendedDefaultExtraEIPs`: extended the default active EIPs.

It is important to notice that the configurator will only allow to append new entries to the default ones defined by
**evmOS**. The reason behind this choice is to ensure the correct and safe execution of the virtual machine but still
allowing partners to customize their implementation.

The `EVMConfigurator` type should be constructed using the builder pattern inside the `init()` function of the file so
that it is run during the creation of the application.

An example of the usage of the configurator is reported below:

```go
configurator := evmconfig.NewEVMConfigurator().
    WithExtendedEips(customActivators).
    WithExtendedDefaultExtraEIPs(defaultEnabledEIPs...).
    Configure()

err := configurator.Configure()
```

Errors are raised when the configurator tries to append an item with the same name of one of the default one. Since
this type is used to configure the EVM before starting the node, it is safe, and suggested, to panic:

```go
if err != nil {
    panic(err)
}
```

## Custom EIPs Deep Dive

When the chain receives an EVM transaction, it is handled by the `MsgServer` of the `x/evm` within the method
`EthereumTx`. The method then calls `ApplyTransaction` where the EVM configuration is created:

```go
cfg, err := k.EVMConfig(ctx, sdk.ConsAddress(ctx.BlockHeader().ProposerAddress), k.eip155ChainID)
```

During the creation of this type, a query is made to retrieve the `x/evm` params. After this step, the request is
passed inside the `ApplyMessageWithConfig` where a new instance of the EVM is created:

```go
evm := k.NewEVM(ctx, msg, cfg, tracer, stateDB)
```

The `NewEVM` method calls the `NewEVMWithHooks` where a new instance of the virtual machine interpreter is created:

```go
evm.interpreter = NewEVMInterpreter(evm, config)
```

The management of activators is handled in this function:

```go
func NewEVMInterpreter(evm *EVM, cfg Config) *EVMInterpreter {
	// If jump table was not initialised we set the default one.
	if cfg.JumpTable == nil {
		cfg.JumpTable = DefaultJumpTable(evm.chainRules)
		for i, eip := range cfg.ExtraEips {
			// Deep-copy jumptable to prevent modification of opcodes in other tables
			copy := CopyJumpTable(cfg.JumpTable)
			if err := EnableEIP(eip, copy); err != nil {
				// Disable it, so caller can check if it's activated or not
				cfg.ExtraEips = append(cfg.ExtraEips[:i], cfg.ExtraEips[i+1:]...)
				log.Error("EIP activation failed", "eip", eip, "error", err)
			}
			cfg.JumpTable = copy
		}
	}

	return &EVMInterpreter{
		evm: evm,
		cfg: cfg,
	}
}
```

As we can see, a new `JumpTable` is created if it is not received from previous evm executions in the same transaction.
After that, the function iterate over the `ExtraEips` defined in the configuration. Then, it is checked if the EIP is
associated with an activator. If yes, the activator function is execute, otherwise an error is returned and the EIP is
removed from the VM configuration. At this point, all the opcodes are ready to be executed.

## How to Use It

In previous sections has been described required structures and files to use the EVM configurator to enable custom
EIPs. In this the general procedure is taken into considerations. Two different scenarios are described:

- New chain.

- Running chain.

### New Chain

For a new chain starting from block genesis, the procedure described in the sections above is enough. To summarize it:

- Create the eip file with custom activators.

- Create the config file with custom activators, default EIPs, and the configurator.

After starting the chain, the genesis validation will perform all the required checks and the chain will be ready using
the new custom EIPs.

### Running Chain

The proper approach to include and enable new EIPs, with the current state of the development, is via coordinate chain
upgrade. During the chain upgrade it is important to define the custom activators since they are not stored in the
chain. To enable them there are two possibilities:

- Write a migration to add the new enabled EIPsm during the upgrade.

- After the upgrade, create a governance proposal to modify the `x/evm` params.
