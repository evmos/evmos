<!--
order: 6
-->

# Smart Contract User Incentives

Learn how to submit contract incentive proposals. {synopsis}

Contract incentives are a part of Evmos tokenomics and aim to increase the growth of the network by distributing rewards to users who interact with incentivized smart contracts. The rewards drive users to interact with applications on Evmos, and reinvest their rewards in more services in the network.

The usage incentives are taken from block reward emission (inflation) and are pooled up in the incentives module account (escrow address). The incentives functionality is fully governed by native $EVMOS token holders who manage the registration of incentives, so that native $EVMOS token holders decide which application should be part of the usage incentives. This governance functionality is implemented using the Cosmos-SDK `gov` module with custom proposal types (`RegisterIncentiveProposal`) for registering the incentives.

Users participate in incentives by submitting transactions to an incentivized contract. The module keeps a record of how much gas the participants spent on their transactions and stores these in gas meters. Based on their gas meters, participants in the incentive are rewarded in regular intervals (epochs). 

:::tip
**Note**: Make sure all the [concepts of the incentives module](./../../../x/incentives/spec/01_concepts.md) are understood before submitting a proposal to incentivize a contract.
:::

## Workflow

![workflow](./../../img/incentives_workflow.png)

## Submit Contract Incentive Proposal

There are three steps in submitting contract incentive proposals:

1. [Configuring Node](#configure-node)
2. [Defining Proposal Content](#define-proposal-content)
3. [Submitting Proposal](#submit-proposal)

### Configure Node

::tip
**Note**: For information on the Evmos Daemon (`evmosd`), consider the following two documents:

- [Installation of Binary](../../validators/quickstart/installation.md)
- [Build & Configuration of Binary](../../validators/quickstart/binary.md)

:::

Set up the node's client config to match the network that the node is connected to:

- `$NODE` is the Tendermint RPC endpoint being used (eg. `https://tendermint.bd.evmos.org:26657`)
- `$CHAINID` is the chain ID of the chain connected to (eg. mainnet: `evmos_9001-2`)

```bash
evmosd config node $NODE
evmosd config chain-id evmos_9001-2
```

### Define Proposal Content

Define the title and description of your governance proposal.

***IMPORTANT***: Refer to [this document](../../users/governance/best_practices.md) to understand the formulation of acceptable governance proposals. If the guidelines set in the document are not adhered to, developers can expect the chances of their incentives proposals passing to be ***greatly reduced***. Proposals should be explanatory, detailed, understandable, and overall well-developed.

### Submit Proposal

Submit your proposal with the following CLI command, with the corresponding arguments:

- `$CONTRACTADDRESS`: Ethereum hex-formatted (`0x...`) address of the contract that users will interact with in your dApp.

    - **Note**: If you are using several external/internal contracts, make sure the contract is the correct one.

- `$ALLOCATION`: Denominations and percentage of the total rewards (25% of block distribution) to be allocated to users that interact and spend gas using the `$CONTRACTADDRESS`.

    - eg. `"0.005000000000000000aevmos"` will distribute 0.5% of out of the 25% tokens minted on each daily epoch rewards.

- `$NUMWEEKS`: Number of weeks (counted by epochs) that you want the `$CONTRACTADDRESS` to be incentivized for.

    - 6 months (`26` epochs): recommended for long-term incentives on apps that have a lot of traction

    - 3 months (`13` epochs): recommended for long-term incentives on apps that have some traction

    - 1 months (`4` epochs): recommended for short-term incentives on apps that don't have much traction

- `$DESCRIPTION`: Description of the proposal.

- `$PROPOSALTITLE`: Title of the proposal.

```bash
evmosd tx gov submit-proposal register-incentive $CONTRACTADDRESS $ALLOCATION $NUMWEEKS --description=$DESCRIPTION --title=$PROPOSALTITLE
```

See below for an example using [Diffusion Finance's](https://diffusion.fi/) router contract:

```bash
evmosd tx gov submit-proposal register-incentive 0xFCd2Ce20ef8ed3D43Ab4f8C2dA13bbF1C6d9512F 0.050000000000000000aevmos 13 --description=$DESCRIPTION --title=$PROPOSALTITLE
```

## Incentives Analysis through Telemetry

:::tip
**Note**: For more detailed information on telemetry, consider the following sources:

- [Cosmos SDK Telemetry Documentation](https://docs.cosmos.network/master/core/telemetry.html)
- [Evmos Supported Telemetry Metrics](https://docs.evmos.org/protocol/telemetry.html)
- [`telemetry` Package](https://docs.evmos.org/protocol/telemetry.html)
- [`go-metrics` Library](https://github.com/armon/go-metrics)

:::

### Telemetry Basics & Setup

The telemetry package of the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) allows operators and developers to gain insight into the performance and behavior of their applications.

To enable telemetrics, set `telemetry.enabled = true` in the `app.toml` config file of the node. The Cosmos SDK currently supports enabling in-memory and [Prometheus](https://prometheus.io/) telemetry sinks. The in-memory sink is always attached (when telemetry is enabled) with a ten second interval and one minute retention. This means that metrics will be aggregated over ten seconds, and metrics will be kept alive for one minute. To query active metrics, set `api.enabled = true` in the `app.toml`. This exposes a single API endpoint: `http://localhost:1317/metrics?format={text|prometheus}`, the default being `text`.

### Emitting & Collecting Metrics

If telemetry is enabled via configuration, a single global metrics collector is exposed via the `go-metrics` library. This allows for emitting and collecting metrics through a simple [API](https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/telemetry/wrapper.go). For example:

```go
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
  defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

  // ...
}
```

Developers may use the `telemetry` package directly, which provides wrappers around metric APIs that include adding useful labels, or they must use the `go-metrics` library directly. It is preferable to add as much context and adequate dimensionality to metrics as possible, so the `telemetry` package is advised.

Regardless of the package or method used, the Cosmos SDK supports the following metrics types:

- gauges
- summaries
- counters

### Incentive Metrics

Evmos supports the following metrics related to the `x/incentives` module, which can be collected for incentive analysis:

| Metric                                         | Description                                                                         | Unit        | Type    |
| :--------------------------------------------- | :---------------------------------------------------------------------------------- | :---------- | :------ |
| `tx_msg_ethereum_tx_incentives_total`          | Total number of txs with an incentivized contract processed via the EVM             | tx          | counter |
| `tx_msg_ethereum_tx_incentives_gas_used_total` | Total amount of gas used by txs with an incentivized contract processed via the EVM | token       | counter |
| `incentives_distribute_participant_total`      | Total number of participants who received rewards                                   | participant | counter |
| `incentives_distribute_reward_total`           | Total amount of rewards that are distributed to all incentives' participants        | token       | counter |

To calculate specific values, such as paid out incentives to a given smart contract user, custom metrics will have to be made following the [above section](#emitting--collecting-metrics).

In addition, gRPC queries related to the `x/incentives` module found [here](../../../x/incentives/spec/08_clients.md#clients) can produce useful analysis.




