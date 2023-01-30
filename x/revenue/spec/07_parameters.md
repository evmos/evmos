<!--
order: 7
-->

# Parameters

The fees module contains the following parameters:

| Key                        | Type    | Default Value |
| :------------------------- | :------ | :------------ |
| `EnableRevenue`           | bool    | `true`        |
| `DeveloperShares`          | sdk.Dec | `50%`         |
| `AddrDerivationCostCreate` | uint64  | `50`          |

## Enable Revenue Module

The `EnableRevenue` parameter toggles all state transitions in the module.
When the parameter is disabled, it will prevent any transaction fees from being distributed to contract deployers
and it will disallow contract registrations, updates or cancellations.

### Developer Shares Amount

The `DeveloperShares` parameter is the percentage of transaction fees that is sent to the contract deployers.

### Address Derivation Cost with CREATE opcode

The `AddrDerivationCostCreate` parameter is the gas value charged
for performing an address derivation in the contract registration process.
A flat gas fee is charged for each address derivation iteration.
We allow a maximum number of 20 iterations, and therefore a maximum number of 20 nonces can be given
for deriving the smart contract address from the deployer's address.
