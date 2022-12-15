<!--
order: 6
-->

# Events

The `x/inflation` module emits the following events:

## Inflation

| Type        | Attribute Key        | Attribute Value                               |
| ----------- |----------------------|-----------------------------------------------|
| `inflation` | `"epoch_provisions"` | `{fmt.Sprintf("%d", epochNumber)}`            |
| `inflation` | `"epoch_number"`     | `{strconv.FormatUint(uint64(in.Epochs), 10)}` |
| `inflation` | `"amount"`           | `{mintedCoin.Amount.String()}`                |
