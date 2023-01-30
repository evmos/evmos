<!--
order: 6
-->

# Parameters

The `x/claims` module contains the parameters described below. All parameters can be modified via governance.

::: danger
🚨 **IMPORTANT**: `time.Duration` store value is in nanoseconds but the JSON / `String` value is in seconds!
:::

| Key                  | Type            | Default Value                                               |
| -------------------- | --------------- | ----------------------------------------------------------- |
| `EnableClaim`        | `bool`          | `true`                                                      |
| `ClaimsDenom`        | `string`        | `"aevmos"`                                                  |
| `AirdropStartTime`   | `time.Time`     | `time.Time{}` // empty                                      |
| `DurationUntilDecay` | `time.Duration` | `2629800000000000` (nanoseconds) // 1 month                 |
| `DurationOfDecay`    | `time.Duration` | `5259600000000000` (nanoseconds) // 2 months                |
| `AuthorizedChannels` | `[]string`      | `[]string{"channel-0", "channel-3"}` // Osmosis, Cosmos Hub |
| `EVMChannels`        | `[]string`      | `[]string{"channel-2"}` // Injective                        |

## Enable claim

The `EnableClaim` parameter toggles all state transitions in the module.
When the parameter is disabled, it will disable all the allocation of airdropped tokens to users.

## Claims Denom

The `ClaimsDenom` parameter defines the coin denomination that users will receive as part of their airdrop allocation.

## Airdrop Start Time

The `AirdropStartTime` refers to the time when user can start to claim the airdrop tokens.

## Duration Until Decay

The `DurationUntilDecay` parameter defines the duration from airdrop start time to decay start time.

## Duration Of Decay

The `DurationOfDecay` parameter refers to the duration from decay start time to claim end time.
Users are not able to claim airdrop after this duration has ended.

## Authorized Channels

The `AuthorizedChannels` parameter describes the set of channels
that users can perform the ibc callback with to claim coins for the ibc action.

## EVM Channels

The `EVMChannels` parameter describes the list of Evmos channels
that connected to EVM compatible chains and can be used during the ibc callback action.
