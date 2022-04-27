<!--
order: 2
-->

# Evmos Clients

Learn about all the available services for clients {synopsis}

The Evmos supports different clients in order to support Cosmos and Ethereum transactions
and queries:

|                                                        | Description                                                                          | Default Port |
| ------------------------------------------------------ | ------------------------------------------------------------------------------------ | ------------ |
| **Cosmos [gRPC](#cosmos-grpc)**                        | Query or send Evmos transactions using gRPC                                          | `9090`       |
| **Cosmos REST ([gRPC-Gateway](#cosmos-grpc-gateway))** | Query or send Evmos transactions using an HTTP RESTful API                           | `9091`       |
| **Ethereum [JSON-RPC](#ethereum-json-rpc)**            | Query Ethereum-formatted transactions and blocks or send Ethereum txs using JSON-RPC | `8545`       |
| **Ethereum [Websocket](#ethereum-websocket)**          | Subscribe to Ethereum logs and events emitted in smart contracts.                    | `8586`       |
| **Tendermint [RPC](#tendermint-rpc)**                  | Subscribe to Ethereum logs and events emitted in smart contracts.                    | `26657`      |
| **Tendermint [Websocket](#tendermint-websocket)**      | Query transactions, blocks, consensus state, broadcast transactions, etc.            | `26657`      |
| **Command Line Interface ([CLI](#cli))**               | Query or send Evmos transactions using your Terminal or Console.                     | N/A          |

## Cosmos gRPC

Evmos exposes gRPC endpoints for all the integrated Cosmos SDK modules. This makes it easier for
wallets and block explorers to interact with the Proof-of-Stake logic and native Cosmos transactions and queries.

### Cosmos gRPC-Gateway (HTTP REST)

[gRPC-Gateway](https://grpc-ecosystem.github.io/grpc-gateway/) reads a gRPC service definition and
generates a reverse-proxy server which translates RESTful JSON API into gRPC. With gRPC-Gateway,
users can use REST to interact the Cosmos gRPC service.

See the list of supported gRPC-Gateway API endpoints for the Evmos testnet [here](https://api.evmos.dev/).

## Ethereum JSON-RPC

<!-- TODO: Link JSON-RPC docs -->

Evmos supports most of the standard [JSON-RPC APIs](./json-rpc/server) to connect with existing Ethereum-compatible web3 tooling.

::: tip
Check out the list of supported JSON-RPC API [endpoints](./json-rpc/endpoints) and [namespaces](./../api/json-rpc/namespaces).
:::

## Ethereum Websocket

<!-- TODO: Link WSS docs -->

Then, start a websocket subscription with [`ws`](https://github.com/hashrocket/ws)

```bash
# connect to tendermint websocet at port 8546 as defined above
ws ws://localhost:8546/

# subscribe to new Ethereum-formatted block Headers
> {"id": 1, "method": "eth_subscribe", "params": ["newHeads", {}]}
< {"jsonrpc":"2.0","result":"0x44e010cb2c3161e9c02207ff172166ef","id":1}
```

## Tendermint Websocket

Tendermint Core provides a Websocket connection to subscribe or unsubscribe to Tendermint ABCI events.

::: tip
For more info about the how to subscribe to events, please refer to the official [Tendermint documentation](https://docs.tendermint.com/v0.34/tendermint-core/subscription.html).
:::

```json
{
    "jsonrpc": "2.0",
    "method": "subscribe",
    "id": "0",
    "params": {
        "query": "tm.event='<event_value>' AND eventType.eventAttribute='<attribute_value>'"
    }
}
```

### List of Tendermint Events

The main events you can subscribe to are:

- `NewBlock`: Contains `events` triggered during `BeginBlock` and `EndBlock`.
- `Tx`: Contains `events` triggered during `DeliverTx` (i.e. transaction processing).
- `ValidatorSetUpdates`: Contains validator set updates for the block.

::: tip
👉 The list of events types and values for each Cosmos SDK module can be found in the [Modules Specification](./modules) section.
Check the `Events` page to obtain the event list of each supported module on Evmos.
:::

List of all Tendermint event keys:

|                                                      | Event Type       | Categories  |
| ---------------------------------------------------- | ---------------- | ----------- |
| Subscribe to a specific event                        | `"tm.event"`     | `block`     |
| Subscribe to a specific transaction                  | `"tx.hash"`      | `block`     |
| Subscribe to transactions at a specific block height | `"tx.height"`    | `block`     |
| Index `BeginBlock` and `Endblock` events             | `"block.height"` | `block`     |
| Subscribe to ABCI `BeginBlock` events                | `"begin_block"`  | `block`     |
| Subscribe to ABCI `EndBlock` events                  | `"end_block"`    | `consensus` |

Below is a list of values that you can use to subscribe for the `tm.event` type:

|                        | Event Value             | Categories  |
| ---------------------- | ----------------------- | ----------- |
| New block              | `"NewBlock"`            | `block`     |
| New block header       | `"NewBlockHeader"`      | `block`     |
| New Byzantine Evidence | `"NewEvidence"`         | `block`     |
| New transaction        | `"Tx"`                  | `block`     |
| Validator set updated  | `"ValidatorSetUpdates"` | `block`     |
| Block sync status      | `"BlockSyncStatus"`     | `consensus` |
| lock                   | `"Lock"`                | `consensus` |
| New consensus round    | `"NewRound"`            | `consensus` |
| Polka                  | `"Polka"`               | `consensus` |
| Relock                 | `"Relock"`              | `consensus` |
| State sync status      | `"StateSyncStatus"`     | `consensus` |
| Timeout propose        | `"TimeoutPropose"`      | `consensus` |
| Timeout wait           | `"TimeoutWait"`         | `consensus` |
| Unlock                 | `"Unlock"`              | `consensus` |
| Block is valid         | `"ValidBlock"`          | `consensus` |
| Consensus vote         | `"Vote"`                | `consensus` |

### Example

```bash
ws ws://localhost:26657/websocket
> { "jsonrpc": "2.0", "method": "subscribe", "params": ["tm.event='ValidatorSetUpdates'"], "id": 1 }
```

Example response:

```json
{
    "jsonrpc": "2.0",
    "id": 0,
    "result": {
        "query": "tm.event='ValidatorSetUpdates'",
        "data": {
            "type": "tendermint/event/ValidatorSetUpdates",
            "value": {
              "validator_updates": [
                {
                  "address": "09EAD022FD25DE3A02E64B0FE9610B1417183EE4",
                  "pub_key": {
                    "type": "tendermint/PubKeyEd25519",
                    "value": "ww0z4WaZ0Xg+YI10w43wTWbBmM3dpVza4mmSQYsd0ck="
                  },
                  "voting_power": "10",
                  "proposer_priority": "0"
                }
              ]
            }
        }
    }
}
```

## CLI

Users can use the `{{ $themeConfig.project.binary }}` binary to interact directly with an Evmos node
though the CLI.

::: tip
👉 To use the CLI, you will need to provide a Tendermint RPC address for the `--node` flag.
Look for a publicly available addresses for testnet and mainnet in the [Quick Connect](./connect) page.
:::

- **Transactions**: `{{ $themeConfig.project.binary }} tx`

    The list of available commands, as of `v3.0.0`, are:

    ```bash
    Available Commands:
      authz               Authorization transactions subcommands
      bank                Bank transaction subcommands
      broadcast           Broadcast transactions generated offline
      crisis              Crisis transactions subcommands
      decode              Decode a binary encoded transaction string
      distribution        Distribution transactions subcommands
      encode              Encode transactions generated offline
      erc20               erc20 subcommands
      evidence            Evidence transaction subcommands
      evm                 evm transactions subcommands
      feegrant            Feegrant transactions subcommands
      gov                 Governance transactions subcommands
      ibc                 IBC transaction subcommands
      ibc-transfer        IBC fungible token transfer transaction subcommands
      multisign           Generate multisig signatures for transactions generated offline
      multisign-batch     Assemble multisig transactions in batch from batch signatures
      sign                Sign a transaction generated offline
      sign-batch          Sign transaction batch files
      slashing            Slashing transaction subcommands
      staking             Staking transaction subcommands
      validate-signatures validate transactions signatures
      vesting             Vesting transaction subcommands
    ```

- **Queries**: `{{ $themeConfig.project.binary }} query`

  The list of available commands, as of `v3.0.0`, are:

    ```bash
    Available Commands:
      account                  Query for account by address
      auth                     Querying commands for the auth module
      authz                    Querying commands for the authz module
      bank                     Querying commands for the bank module
      block                    Get verified data for a the block at given height
      claims                   Querying commands for the claims module
      distribution             Querying commands for the distribution module
      epochs                   Querying commands for the epochs module
      erc20                    Querying commands for the erc20 module
      evidence                 Query for evidence by hash or for all (paginated) submitted evidence
      evm                      Querying commands for the evm module
      feegrant                 Querying commands for the feegrant module
      feemarket                Querying commands for the fee market module
      gov                      Querying commands for the governance module
      ibc                      Querying commands for the IBC module
      ibc-transfer             IBC fungible token transfer query subcommands
      incentives               Querying commands for the incentives module
      inflation                Querying commands for the inflation module
      params                   Querying commands for the params module
      recovery                 Querying commands for the recovery module
      slashing                 Querying commands for the slashing module
      staking                  Querying commands for the staking module
      tendermint-validator-set Get the full tendermint validator set at given height
      tx                       Query for a transaction by hash, "<addr>/<seq>" combination or comma-separated signatures in a committed block
      txs                      Query for paginated transactions that match a set of events
      upgrade                  Querying commands for the upgrade module
      vesting                  Querying commands for the vesting module
    ```
