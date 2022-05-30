<!--
order: 3
-->

# Telemetry

Gather relevant insights about the Evmos application and modules with custom metrics and telemetry. {synopsis}

To understand how to use the metrics below, please refer to the [Cosmos SDK telemetry documentation](https://docs.cosmos.network/master/core/telemetry.html).

## Supported Metrics

| Metric                          | Description                                                        | Unit  | Type    |
| :------------------------------ | :----------------------------------------------------------------- | :---- | :------ |
| `ethereum_tx`                   | Total number of txs processed via the EVM                          | tx    | counter |
| `tx_msg_convert_coin`           | The total amount of coins converted with `MsgConvertCoin`          | token | gauge   |
| `tx_msg_convert_erc20`          | The total amount of erc20 converted with `MsgConvertErc20`         | token | gauge   |
| `tx_msg_ethereum_tx`            | The total amount of gas used by an etheruem tx                     | token | gauge   |
| `tx_msg_ethereum_tx_incentives` | The total amount of gas used by a tx with an incentivized contract | token | gauge   |
| `incentive_distribute_rewards`  | The total amount of tokens transfered to incentive participants    | token | gauge   |
| `inflation_allocate`            | The total amount of tokens minted through inflation                | token | gauge   |
| `recovery_on_receive`           | The total amount of tokens recovered through ibc `onRecvPacket`    | token | gauge   |