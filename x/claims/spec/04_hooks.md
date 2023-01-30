<!--
order: 4
-->

# Hooks

The `x/claims` module implements transaction hooks for each of the four actions
from the `x/staking`, `x/gov` and  `x/evm` modules.
It also implements an IBC Middleware in order to claim the IBC transfer action
and to claim the tokens for Cosmos Hub and Osmosis users by migrating the claims record to the recipient address.

## Governance Hook - Vote Action

The user votes on a Governance proposal using their Evmos account.
Once the vote is successfully included, the claimable amount corresponding
to the vote action is transferred to the user address:

1. The user submits a `MsgVote`.
2. Begin claiming process for the `ActionVote`.
3. Check if the claims is allowed:
    - global parameter is enabled
    - current block time is before the end of the claims period
    - user has a claims record (i.e allocation) for the airdrop
    - user hasn't already claimed the action
    - claimable amount is greater than zero
4. Transfer the claimable amount from the escrow account to the user balance
5. Mark the `ActionVote` as completed on the claims record.
6. Update the claims record and retain it, even if all the actions have been claimed.

## Staking Hook - Delegate Action

The user delegates their EVMOS tokens to a validator.
Once the tokens are staked, the claimable amount corresponding to the delegate action is transferred to the user address:

1. The user submits a `MsgDelegate`.
2. Begin claiming process for the `ActionDelegate`.
3. Check if the claims is allowed:
    - global parameter is enabled
    - current block time is before the end of the claims period
    - user has a claims record (i.e allocation) for the airdrop
    - user hasn't already claimed the action
    - claimable amount is greater than zero
4. Transfer the claimable amount from the escrow account to the user balance
5. Mark the `ActionDelegate` as completed on the claims record.
6. Update the claims record and retain it, even if all the actions have been claimed.

## EVM Hook - EVM Action

The user deploys or interacts with a smart contract using their Evmos account or send a transfer using their Web3 wallet.
Once the EVM state transition is successfully processed,
the claimable amount corresponding to the EVM action is transferred to the user address:

1. The user submits a `MsgEthereumTx`.
2. Begin claiming process for the `ActionEVM`.
3. Check if the claims is allowed:
    - global parameter is enabled
    - current block time is before the end of the claims period
    - user has a claims record (i.e allocation) for the airdrop
    - user hasn't already claimed the action
    - claimable amount is greater than zero
4. Transfer the claimable amount from the escrow account to the user balance
5. Mark the `ActionEVM` as completed on the claims record.
6. Update the claims record and retain it, even if all the actions have been claimed.

## IBC Middleware - IBC Transfer Action

### Send

The user submits an IBC transfer to a recipient in the destination chain.
Once the transfer acknowledgement package is received,
the claimable amount corresponding to the IBC transfer action is transferred to the user address:

1. The user submits a `MsgTransfer` to a recipient address in the destination chain.
2. The transfer packet is processed by the IBC ICS20 Transfer app module and relayed.
3. Once the packet acknowledgement is received, the IBC transfer module `OnAcknowledgementPacket` callback is executed.
   After which the claiming process for the `ActionIBCTransfer` begins.
4. Check if the claims is allowed:
    - global parameter is enabled
    - current block time is before the end of the claims period
    - user has a claims record (i.e allocation) for the airdrop
    - user hasn't already claimed the action
    - claimable amount is grater than zero
5. Transfer the claimable amount from the escrow account to the user balance
6. Mark the `ActionIBC` as completed on the claims record.
7. Update the claims record and retain it, even if all the actions have been claimed.

### Receive

The user receives an IBC transfer from a counterparty chain.
If the transfer is successful,
the claimable amount corresponding to the IBC transfer action is transferred to the user address.
Additionally, if the sender address is Cosmos Hub or Osmosis address with an airdrop allocation,
the `ClaimsRecord` is merged with the recipient's claims record.

1. The user receives an packet containing an IBC transfer data.
2. The transfer is processed by the IBC ICS20 Transfer app module
3. Check if the claims is allowed:
   - global parameter is enabled
   - current block time is before the end of the claims period
4. Check if package is from a sent NON EVM channel and sender and recipient
	address are the same. If a packet is sent from a non-EVM chain, the sender
	addresss is not an ethereum key (i.e. `ethsecp256k1`). Thus, if
	`sameAddress` is true, the recipient address must be a non-ethereum key as
	well, which is not supported on Evmos. To prevent funds getting stuck,
	return an error, unless the destination channel from a connection to a chain
	is EVM-compatible or supports ethereum keys (eg: Cronos, Injective).
6. Check if destination channel is authorized to perform the IBC claim.
   Without this authorization the claiming process is vulerable to attacks.
7. Handle one of four cases by comparing sender and recipient addresses with each other
   and checking if either addresses have a claims record (i.e allocation) for the airdrop.
   To compare both addresses, the sender address's bech32 human readable prefix (HRP) is replaced with `evmos`.

   1. both sender and recipient are distinct and have a claims record ->
      merge sender's record with the recipient's record and claim actions that have been completed by one or the other
   2. only the sender has a claims record -> migrate the sender record to the recipient address and claim IBC action
   3. only the recipient has a claims record ->
      only claim IBC transfer action and transfer the claimable amount from the escrow account to the user balance
   4. neither the sender or recipient have a claims record ->
      perform a no-op by returning the original success acknowledgement
