<!--
order: 5
-->

# Hooks

The erc20 module implements transaction hooks from the EVM in order to trigger token pair conversion.

## EVM Hooks

The EVM hooks allows users to convert ERC20s to Cosmos Coins
by sending an Ethereum tx transfer to the module account address.
This enables native conversion of tokens via Metamask and EVM-enabled wallets for both token pairs
that have been registered through a native Cosmos coin or an ERC20 token.
Note that additional coin/token balance checks for sender and receiver to prevent malicious contract behaviour
(as performed in the [`ConvertERC20` msg](03_state_transitions.md#21-erc20-to-coin)) cannot be done here,
as the balance prior to the transaction is not available in the hook.

### Registered Coin: ERC20 to Coin

1. User transfers ERC20 tokens to the `ModuleAccount` address to escrow them
2. Check if the ERC20 Token that was transferred from the sender is a native ERC20
or a native Cosmos Coin by looking at the
[Ethereum event logs](https://medium.com/mycrypto/understanding-event-logs-on-the-ethereum-blockchain-f4ae7ba50378#:~:text=A%20log%20record%20can%20be,or%20a%20change%20of%20ownership.&text=Each%20log%20record%20consists%20of,going%20on%20in%20an%20event)
3. If the token contract address corresponds to the ERC20 representation of a native Cosmos Coin
    1. Call `burn()` ERC20 method from the  `ModuleAccount`.
       Note that this is the same as 1.2, but since the tokens are already on the ModuleAccount balance,
       we burn the tokens from the module address instead of calling `burnFrom()`.
       Also note that we don't need to mint
       because [1.1 coin to erc20](03_state_transitions.md#11-coin-to-erc20) escrows the coin
    2. Transfer Cosmos Coin to the bech32 account address of the sender hex address

### Registered ERC20: ERC20 to Coin

1. User transfers coins to the`ModuleAccount` to escrow them
2. Check if the ERC20 Token that was transferred is a native ERC20 or a native cosmos coin
3. If the token contract address is a native ERC20 token
    1. Mint Cosmos Coin
    2. Transfer Cosmos Coin to the bech32 account address of the sender hex
