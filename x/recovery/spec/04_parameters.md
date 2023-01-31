<!--
order: 4
-->

# Parameters

The `x/recovery` module contains the following parameters:

| Key                     |      Type       |             Default Value |
| :---------------------- | :-------------- | :------------------------ |
| `EnableRecovery`        |     `bool`      |                    `true` |
| `PacketTimeoutDuration` | `time.Duration` | `14400000000000`  // 4hrs |

## Enable Recovery

The `EnableRecovery` parameter toggles Recovery IBC middleware.
When the parameter is disabled, it will disable the recovery of stuck tokens to users.

## Packet Timeout Duration

The `PacketTimeoutDuration` parameter is the duration before the IBC packet timeouts
and the transaction is reverted on the counter party chain.
