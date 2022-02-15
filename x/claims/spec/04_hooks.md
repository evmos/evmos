<!--
order: 4
-->

# Hooks

The `x/claims` module implements transaction hooks for each of the four actions  from the `x/staking`, `x/gov` and  `x/evm` modules. It also implements an IBC Middleware in order to claim the IBC transfer action and to claim the tokens for Cosmos Hub and Osmosis users by migrating the claim record to the recipient address.

## Governance Hook - Vote Action

The user votes on a Governance proposal using their Evmos account. Once the vote is successfully included, the claimable amount corresponding to the vote action is transferred to the user address:

1. The user submits a `MsgVote`.
2. Begin claiming process for the `ActionVote`.
3. Check if the claims is allowed:
    - global parameter is enabled
    - current block time is before the end of the claims period
    - user has a claim record (i.e allocation) for the airdrop
    - user hasn't already claimed the action
    - claimable amount is greater than zero
3. Transfer the claimable amount from the escrow account to the user balance
4. Mark the `ActionVote` as completed on the claim record.
5. Update the claim record or delete it if all the actions have been claimed.

## Staking Hook - Delegate Action

The user delegates their EVMOS tokens to a validator. Once the tokens are staked, the claimable amount corresponding to the delegate action is transferred to the user address:

1. The user submits a `MsgDelegate`.
2. Begin claiming process for the `ActionDelegate`.
3. Check if the claims is allowed:
    - global parameter is enabled
    - current block time is before the end of the claims period
    - user has a claim record (i.e allocation) for the airdrop
    - user hasn't already claimed the action
    - claimable amount is greater than zero
3. Transfer the claimable amount from the escrow account to the user balance
4. Mark the `ActionDelegate` as completed on the claim record.
5. Update the claim record or delete it if all the actions have been claimed.

## EVM Hook - EVM Action

The user deploys or interacts with a smart contract using their Evmos account or send a transfer using their Web3 wallet. Once the EVM state transition is successfully processed, the claimable amount corresponding to the EVM action is transferred to the user address:

1. The user submits a `MsgEthereumTx`.
2. Begin claiming process for the `ActionEVM`.
3. Check if the claims is allowed:
    - global parameter is enabled
    - current block time is before the end of the claims period
    - user has a claim record (i.e allocation) for the airdrop
    - user hasn't already claimed the action
    - claimable amount is greater than zero
3. Transfer the claimable amount from the escrow account to the user balance
4. Mark the `ActionEVM` as completed on the claim record.
5. Update the claim record or delete it if all the actions have been claimed.

## IBC Middleware - IBC Transfer Action

### Send

The user submits an IBC transfer to a recipient in the destination chain. Once the transfer acknowledgement package is received, the claimable amount corresponding to the IBC transfer action is transferred to the user address:

1. The user submits a `MsgTransfer` to a recipient address in the destination chain.
2. The transfer packet is processed by the IBC ICS20 Transfer app module and relayed.
3. Once the packet acknowledgement is received, the IBC transfer module `OnAcknowledgementPacket` callback is executed. After which the claiming process for the `ActionIBCTransfer` begins.
5. Check if the claims is allowed:
    - global parameter is enabled
    - current block time is before the end of the claims period
    - user has a claim record (i.e allocation) for the airdrop
    - user hasn't already claimed the action
    - claimable amount is grater than zero
6. Transfer the claimable amount from the escrow account to the user balance
7. Mark the `ActionIBC` as completed on the claim record.
8. Update the claim record or delete it if all the actions have been claimed.

### Receive

The user receives an IBC transfer from a counterparty chain. If the transfer is successful, the claimable amount corresponding to the IBC transfer action is transferred to the user address. Additionally, if the sender address is Cosmos Hub or Osmosis address with an airdrop allocation, the `ClaimsRecord` is merged with the recipient's claims record.

1. The user receives an packet containing an IBC transfer data.
2. The transfer is processed by the IBC ICS20 Transfer app module
4. Check if the claims is allowed:
    - global parameter is enabled
    - current block time is before the end of the claims period
    - user has a claim record (i.e allocation) for the airdrop
    - user hasn't already claimed the action
    - claimable amount is grater than zero
5. Transfer the claimable amount from the escrow account to the user balance
6. Mark the `ActionIBCTransfer` as completed on the claim record.
7. Update the claim record or delete it if all the actions have been claimed.
