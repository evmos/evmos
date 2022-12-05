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
| `clawback` | `"account"`     | `{msg.AccountAddress}` |
| `clawback` | `"destination"` | `{msg.DestAddress}`    |

## Update Clawback Vesting Account Funder

| Type                    | Attibute Key   | Attibute Value           |
| ----------------------- | -------------- | ------------------------ |
| `update_vesting_funder` | `"funder"`     | `{msg.FromAddress}`      |
| `update_vesting_funder` | `"account"`    | `{msg.VestingAddress}`   |
| `update_vesting_funder` | `"new_funder"` | `{msg.NewFunderAddress}` |
