<!--
order: 7
-->

# Parameters

The fees module contains the following parameters:

| Key                        | Type          | Default Value                 |
| :------------------------- | :------------ | :---------------------------- |
| `EnableFees`               | bool          | `true`                        |
| `DeveloperShares`          | sdk.Dec       | `50%`                         |
| `ValidatorShares`          | sdk.Dec       | `50%`                         |
| `AddrDerivationCostCreate` | uint64        | `50`                          |
| `MinGasPrice`              | sdk.Dec       | `0`                           |

## Enable Fee Module

The `EnableFees` parameter toggles all state transitions in the module. When the parameter is disabled, it will prevent any transaction fees from being distributed to contract deployers and it will disallow contract registrations, updates or cancellations.

### Developer Shares Amount

The `DeveloperShares` parameter is the percentage of transaction fees that is sent to the contract deployers.

### Validator Shares Amount

The `ValidatorShares` parameter is the percentage of transaction fees that is sent to the block proposer.

### Address Derivation Cost with CREATE opcode

The `AddrDerivationCostCreate` parameter is the gas value charged for performing an address derivation in the contract registration process. A flat gas fee is charged for each address derivation iteration. We allow a maximum number of 20 iterations, and therefore a maximum number of 20 nonces can be given for deriving the smart contract address from the deployer's address.

### Minimum Gas Price

The `MinGasPrice` parameter is the minimum gas price that needs to be paid to include a Cosmos or EVM transaction in a block. In case the `feemarket` module gas price calculations result in a gas price lower than `MinGasPrice`, or the value that can be set by each validator individually is lower than `MinGasPrice`, `MinGasPrice` has priority and it will set the lower bound.
