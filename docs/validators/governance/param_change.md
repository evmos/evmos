<!--
order: 6
-->

# Chain Parameters

::: tip
**Note:** If users are attempting to write governance proposals concerned with changing parameters (such as those of type `ParameterChangeProposal`), refer to [this document](../../validators/governance/best_practices.md#parameter-change-proposal).
:::

If a parameter-change proposal is successful, the change takes effect immediately upon completion of the voting period.

## List of Parameters

For a comprehensive list of available module parameters see the table below:

| Module         | Codebase     | Parameters                                                                                      |
| -------------- | ------------ | ----------------------------------------------------------------------------------------------- |
| `auth`         | `cosmos-sdk` | [reference](https://docs.cosmos.network/main/modules/auth/06_params.html)                     |
| `bank`         | `cosmos-sdk` | [reference](https://docs.cosmos.network/main/modules/bank/05_params.html)                     |
| `crisis`       | `cosmos-sdk` | [reference](https://docs.cosmos.network/main/modules/crisis/04_params.html)                   |
| `distribution` | `cosmos-sdk` | [reference](https://docs.cosmos.network/main/modules/distribution/06_events.html)             |
| `governance`   | `cosmos-sdk` | [reference](https://docs.cosmos.network/main/modules/gov/06_params.html)                      |
| `slashing`     | `cosmos-sdk` | [reference](https://docs.cosmos.network/main/modules/slashing/08_params.html)                 |
| `staking`      | `cosmos-sdk` | [reference](https://docs.cosmos.network/main/modules/staking/08_params.html)                  |
| `transfer`     | `ibc-go`     | [reference](https://github.com/cosmos/ibc-go/blob/main/modules/apps/transfer/spec/07_params.md) |
| `evm`          | `ethermint`  | [reference](https://evmos.dev/modules/evm/08_params.html)                                       |
| `feemarket`    | `ethermint`  | [reference](https://evmos.dev/modules/feemarket/07_params.html)                                 |
| `claims`       | `evmos`      | [reference](https://evmos.dev/modules/claims/06_parameters.html)                                |
| `erc20`        | `evmos`      | [reference](https://evmos.dev/modules/erc20/07_parameters.html)                                 |
| `incentives`   | `evmos`      | [reference](https://evmos.dev/modules/incentives/07_parameters.html)                            |
| `inflation`    | `evmos`      | [reference](https://evmos.dev/modules/inflation/05_parameters.html)                             |
