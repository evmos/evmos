# Wrappers

The wrapper package defines wrappers around Cosmos SDK modules required to
handle coins with different representation of the decimals inside the x/evm module.

All wrapper implementations should be used only for transaction executions that
involves the EVM. When a keeper is required as a dependency for another Cosmos
SDK module, it should be used the original <MODULE>Keeper.

## BankWrapper

This package contains the `BankWrapper`, a wrapper around the Cosmos SDK bank keeper that is designed
to manage the EVM denomination with a custom decimal representation. The primary purpose of the
`BankWrapper` is to handle conversions between Cosmos SDK's default decimal system and the 18-decimal
representation commonly used in EVM-based systems.

## Features

- **Balance Conversion:** Automatically converts balances to 18 decimals, the standard for EVM coins.
- **Send and Receive Coins:** Handles sending coins between accounts and modules, ensuring proper conversion
  to and from the 18-decimal system.
- **Mint and Burn Coins:** Provides methods for minting and burning coins, with conversions applied
  as necessary.

## Conversion Logic

The wrapper uses helper functions to convert between Cosmos SDK's bank module decimal representation
and EVM's 18-decimal standard:

- `mustConvertEvmCoinTo18Decimals`: Converts a coin to 18 decimals.
- `convertCoinsFrom18Decimals`: Converts coins from 18 decimals to their original representation.

Both methods convert only the evm denom amount.
