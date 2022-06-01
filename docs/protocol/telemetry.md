<!--
order: 3
-->

# Telemetry

Gather relevant insights about the Evmos application and modules with custom metrics and telemetry. {synopsis}

To understand how to use the metrics below, please refer to the [Cosmos SDK telemetry documentation](https://docs.cosmos.network/master/core/telemetry.html).

## Supported Metrics

| Metric                                | Description                                                             | Unit       | Type    |
| :------------------------------------ | :---------------------------------------------------------------------- | :--------- | :------ |
| `ethereum_tx`                         | Total number of txs processed via the EVM                               | tx         | counter |
| `tx_msg_convert_coin_total`           | Total number of txs with a `ConvertCoin` msg                            | tx         | counter |
| `tx_msg_convert_erc20_total`          | Total number of txs with a `ConvertERC20` msg                           | tx         | counter |
| `tx_msg_ethereum_tx`                  | Total number of gas used by an etheruem tx                              | token      | gauge   |
| `tx_msg_ethereum_tx_incentives_total` | Total number of txs with an incentivized contract processed via the EVM | tx         | counter |
| `incentives_distribute_total`         | Total number of distributions to all incentives' participants           | distribute | counter |
| `inflation_hook_allocate_total`       | Total number of allocations through inflation                           | allocate   | counter |
| `recovery_ibc_on_recv_total`          | Total number of recoveries using the ibc `onRecvPacket` callback        | recover    | counter |