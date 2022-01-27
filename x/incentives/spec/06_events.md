<!--
order: 6
-->

# Events

The `x/incentives` module emits the following events:

## Register Incentive Proposal

| Type                 | Attibute Key | Attibute Value                                |
| -------------------- | ------------ | --------------------------------------------- |
| `register_incentive` | `"contract"` | `{erc20_address}`                             |
| `register_incentive` | `"epochs"`   | `{strconv.FormatUint(uint64(in.Epochs), 10)}` |

## Cancel Incentive Proposal

| Type               | Attibute Key | Attibute Value    |
| ------------------ | ------------ | ----------------- |
| `cancel_incentive` | `"contract"` | `{erc20_address}` |

## Incentive Distribution

| Type                    | Attibute Key | Attibute Value                                |
| ----------------------- | ------------ | --------------------------------------------- |
| `distribute_incentives` | `"contract"` | `{erc20_address}`                             |
| `distribute_incentives` | `"epochs"`   | `{strconv.FormatUint(uint64(in.Epochs), 10)}` |
