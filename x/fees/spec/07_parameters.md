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

## Enable Fee Module

The `EnableFees` parameter toggles all state transitions in the module. When the parameter is disabled, it will prevent any transaction fees from being distributed to contract deployers and it will disallow contract registrations, updates or cancellations.

### Developer Shares Amount

The `DeveloperShares` parameter is the percentage of transaction fees that is sent to the contract deployers.

### Validator Shares Amount

The `ValidatorShares` parameter is the percentage of transaction fees that is sent to the block proposer.

### AddrDerivationCostCreate

The `AddrDerivationCostCreate` parameter is the gas value charged for performing an address derivation in the contract registration process. Because we allow an unlimited number of nonces to be given for deriving the smart contract address, a flat gas fee is charged for each address derivation iteration.
