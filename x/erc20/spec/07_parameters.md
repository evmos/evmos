<!--
order: 7
-->

# Parameters

The erc20 module contains the following parameters:

| Key                     | Type          | Default Value                 |
| ----------------------- | ------------- | ----------------------------- |
| `EnableErc20`    | bool          | `true`                        |
| `EnableEVMHook`         | bool          | `true`                        |

## Enable ERC20

The `EnableErc20` parameter toggles all state transitions in the module. When the parameter is disabled, it will prevent all token pair registration and conversion functionality.

### Enable EVM Hook

The `EnableEVMHook` parameter enables the EVM hook to convert an ERC20 token to a Cosmos Coin by transferring the Tokens through a `MsgEthereumTx`  to the `ModuleAddress` Ethereum address.
