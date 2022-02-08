<!--
order: 6
-->

# Events


The `x/inflation` module emits the following events:

[inflation ](https://www.notion.so/2f398a8d417d42cb8da307cdedade91b)


## Inflation

| Type                 | Attibute Key | Attibute Value                                |
| -------------------- | ------------ | --------------------------------------------- |
| `register_incentive` | `"contract"` | `{erc20_address}`                             |
| `register_incentive` | `"epochs"`   | `{strconv.FormatUint(uint64(in.Epochs), 10)}` |