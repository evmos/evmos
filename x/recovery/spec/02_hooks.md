<!--
order: 2
-->

# Hooks

The `x/recovery` module allows for state transitions that return IBC tokens that were previously transferred to EVMOS back to the source chains into the source accounts with the `Keeper.OnRecvPacket` callback. The source chain must be authorized.

## Withdraw

A user performs an IBC transfer to return the tokens that they previously transferred to their Cosmos `secp256k1` address instead of the Ethereum `ethsecp256k1` address. The behavior is implemented using an IBC`OnRecvPacket` callback.

1. A user performs an IBC transfer to their own account by sending tokens from their address on an authorized chain (e.g. `cosmos1...`) to their evmos `secp2561` address (i.e. `evmos1`)  which holds the stuck tokens.  This is done using a [`FungibleTokenPacket`](https://github.com/cosmos/ibc/blob/master/spec/app/ics-020-fungible-token-transfer/README.md) IBC packet.
2. Check that the withdrawal conditions are met and skip to the next middleware if any condition is not satisfied:
    1. recovery is enabled globally
    2. channel is authorized
    3. channel is not an EVM channel (as an EVM supports `eth_secp256k1` keys and tokens are not stuck)
    4. sender and receiver address belong to the same account as recovery is only possible for transfers to a sender's own account on Evmos. Both sender and recipient addresses are therefore converted from `bech32` to `sdk.AccAddress`.
    5. the sender/recipient account is a not vesting or module account
    6. recipient pubkey is not a supported key (`eth_secp256k1`, `amino multisig`, `ed25519`), as in this case tokens are not stuck and don’t require recovery
3. Check if sender/recipient address is blocked by the `x/bank` module and throw an acknowledgment error to prevent further execution along with the IBC middleware stack
4. Perform recovery to transfer the recipient’s balance back to the sender address with the IBC `OnRecvPacket` callback. There are two cases:
    1. First transfer from authorized source chain:
        1. sends back IBC tokens that originated from the source chain
        2. sends over all Evmos native tokens
    2. Second and further transfers from a different authorized source chain
        1. only sends back IBC tokens that originated from the source chain
5. If the recipient does not have any balance, return without recovering tokens
