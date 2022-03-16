<!--
order: 6
-->

# Events

The `x/vesting` module emits the following events:

## Create Clawback Vesting Account

| Type                              | Attibute Key   | Attibute Value                    |
| --------------------------------- | -------------- | --------------------------------- |
| `create_clawback_vesting_account` | `"from"`       | `{msg.FromAddress}`               |
| `create_clawback_vesting_account` | `"coins"`      | `{vestingCoins.String()}`         |
| `create_clawback_vesting_account` | `"start_time"` | `{msg.StartTime.String()}`        |
| `create_clawback_vesting_account` | `"merge"`      | `{strconv.FormatBool(msg.Merge)}` |
| `create_clawback_vesting_account` | `"amount"`     | `{msg.ToAddress}`                 |

## Clawback

| Type       | Attibute Key    | Attibute Value         |
| ---------- | --------------- | ---------------------- |
| `clawback` | `"funder"`      | `{msg.FromAddress}`    |
| `clawback` | `"acciyubt"`    | `{msg.AccountAddress}` |
| `clawback` | `"destination"` | `{msg.DestAddress}`    |
