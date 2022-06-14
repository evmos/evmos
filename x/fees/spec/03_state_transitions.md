<!--
order: 3
-->

# State Transitions

The `x/fees` module allows for three types of state transitions: `RegisterDevFeeInfo`, `UpdateDevFeeInfo` and `CancelDevFeeInfo`. The logic for distributing transaction fees is handled through [Hooks](./05_hooks.md).

### Fee Info Registration

A developer registers a contract for receiving transaction fees, defining the contract address, an array of nonces for [address deriviation](01_concepts.md#address-derivation) and an optional withdraw address for receiving fees. If the withdraw address is not set, the fees are sent to the deployer address by default.

1. User submits a `RegisterDevFeeInfo` to register a contract address, along with a withdraw address that they would like to receive the fees to
2. Check if the following conditions pass:
    1. `x/fees` module is enabled
    2. the contract was not previously registered
    3. deployer has a valid account (it has done at least one transaction) and is not a smart contract
    4. an account corresponding to the contract address exists, with a non-empty bytecode
    5. contract address can be derived from the deployer’s address and provided nonces using the `CREATE` operation
3. Store an instance of the provided fee information.

All transactions sent to the registered contract occurring after registration will have their fees distributed to the developer, according to the global `DeveloperShares` parameter.

### Fee Info Update

A developer updates the withdraw address for a registered contract, defining the contract address and the new withdraw address.

1. User submits a `UpdateDevFeeInfo`
2. Check if the following conditions pass:
    1. `x/fees` module is enabled
    2. the contract is registered
    3. the signer of the transaction is the same as the contract deployer
3. Update the fee information with the new withdraw address

After this update, the developer receives the fees on the new withdraw address.

### Fee Info Cancel

A developer cancels receiving fees for a registered contract, defining the contract address.

1. User submits a `CancelDevFeeInfo`
2. Check if the following conditions pass:
    1. `x/fees` module is enabled
    2. the contract is registered
    3. the signer of the transaction is the same as the contract deployer
3. Remove fee information from storage

The developer no longer receives fees from transactions sent to this contract.
