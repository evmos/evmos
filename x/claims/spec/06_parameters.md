<!--
order: 6
-->

# Parameters

The `x/claims` module contains the parameters described below. All parameters can be modified via governance.

| Key                  | Type            | Default Value                  |
|----------------------|-----------------|--------------------------------|
| `EnableClaim`        | `bool`          | `true`                         |
| `AirdropStartTime`   | `time.Time`     | `time.Time{}` // empty         |
| `DurationUntilDecay` | `time.Duration` | `2629800000000000` // 1 month  |
| `DurationOfDecay`    | `time.Duration` | `5259600000000000` // 2 months |
| `ClaimDenom`         | `string`        | `"aevmos"`                     |

## Enable claim

The `EnableClaim` parameter toggles all state transitions in the module. When the parameter is disabled, it will disable all the allocation of airdropped tokens to users.

## Airdrop Start Time

The `AirdropStartTime` refers to the time when user can start to claim the airdrop tokens.

## Duration Until Decay

`DurationUntilDecay` defines the duration from airdrop start time to decay start time.

## Duration Of Decay

Refers to the duration from decay start time to claim end time. Users are not able to claim airdrop after this duration has ended.

## Claim Denom

Defines the coin denomination that users will receive as part of their airdrop allocation.
