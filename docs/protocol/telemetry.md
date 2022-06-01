<!--
order: 3
-->

# Telemetry

Gather relevant insights about the Evmos application and modules with custom metrics and telemetry. {synopsis}

To understand how to use the metrics below, please refer to the [Cosmos SDK telemetry documentation](https://docs.cosmos.network/master/core/telemetry.html).

## Supported Metrics

| Metric                                         | Description                                                                  | Unit         | Type    |
| :--------------------------------------------- | :--------------------------------------------------------------------------- | :----------- | :------ |
| `ethereum_tx`                                  | Total number of txs processed via the EVM                                    | tx           | counter |
| `tx_msg_convert_coin_amount_total`             | Total amount of converted coins using a `ConvertCoin` msg                    | tokens       | counter |
| `tx_msg_convert_coin_total`                    | Total number of txs with a `ConvertCoin` msg                                 | tx           | counter |
| `tx_msg_convert_erc20_amount_total`            | Total amount of converted erc20 using a `ConvertERC20` msg                   | tokens       | counter |
| `tx_msg_convert_erc20_total`                   | Total number of txs with a `ConvertERC20` msg                                | tx           | counter |
| `tx_msg_ethereum_tx`                           | Total number of gas used by an etheruem tx                                   | token        | gauge   |
| `tx_msg_ethereum_tx_incentives_total`          | Total number of txs with an incentivized contract processed via the EVM      | tx           | counter |
| `incentives_distribute_participants_total`     | Total number of participants who received rewards                            | participants | counter |
| `incentives_distribute_rewards_total`          | Total amount of rewards that are distributed to all incentives' participants | rewards      | counter |
| `inflation_hook_allocate_total`                | Total amount of tokens allocated through inflation                           | tokens       | counter |
| `inflation_hook_allocate_staking_total`        | Total amount of tokens allocated through inflation to staking                | tokens       | counter |
| `inflation_hook_allocate_incentives_total`     | Total amount of tokens allocated through inflation to incentives             | tokens       | counter |
| `inflation_hook_allocate_community_pool_total` | Total amount of tokens allocated through inflation to community pool         | tokens       | counter |
| `recovery_ibc_on_recv_total`                   | Total number of recoveries using the ibc `onRecvPacket` callback             | recover      | counter |
| `recovery_ibc_on_recv_amount_total`            | Total number of recoveries using the ibc `onRecvPacket` callback             | recover      | counter |
