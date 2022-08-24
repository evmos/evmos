<!--
order: 2
-->

# Running the Server

Learn how to run and setup the JSON-RPC server on Evmos. {synopsis}

:::tip
**Important**: You cannot use all JSON RPC methods unless your node stores the entire copy of the blockchain locally. Do you need archives/snapshots of our networks? Go to [this section](https://docs.evmos.org/validators/snapshots_archives.html).
:::

## Introduction

JSON-RPC is provided on multiple transports. Evmos supports JSON-RPC over HTTP and WebSocket.

## Requirements

We recommend to use the server with minimum 8-core CPU and 64gb of RAM.
You must have ports 8545 and 8546 open on your firewall.

## Enable Server

To enable RPC server use the following flag (set to true by default).

```bash
evmosd start --json-rpc.enable
```

## Defining Namespaces

`Eth`,`Net` and `Web3` [namespaces](./namespaces.md) are enabled by default, but for the JSON-RPC you need to add more namespaces.
In order to enable other namespaces edit `app.toml` file.

```toml
# API defines a list of JSON-RPC namespaces that should be enabled
# Example: "eth,txpool,personal,net,debug,web3"
api = "eth,net,web3,txpool,debug,personal"
```

## Set a Gas Cap

`eth_call` and `eth_estimateGas` define a global gas cap over rpc for DoS protection. You can override the default gas cap value of 25,000,000 by passing a custom value in `app.toml`:

```toml
# GasCap sets a cap on gas that can be used in eth_call/estimateGas (0=infinite). Default: 25,000,000.
gas-cap = 25000000
```

## CORS

If accessing the RPC from a browser, CORS will need to be enabled with the appropriate domain set. Otherwise, JavaScript calls are limit by the same-origin policy and requests will fail.

The CORS setting can be updated from the `app.toml`

```toml
###############################################################################
###                           API Configuration                             ###
###############################################################################

[api]

# ...

# EnableUnsafeCORS defines if CORS should be enabled (unsafe - use it at your own risk).
enabled-unsafe-cors = true # default false
```

## Pruning

For all methods to work correctly, your node must be archival (store the entire copy of the blockchain locally). Pruning must be disabled.
The pruning settings can be updated from the `app.toml`

```toml
###############################################################################
###                           Base Configuration                            ###
###############################################################################

# The minimum gas prices a validator is willing to accept for processing a
# transaction. A transaction's fees must meet the minimum of any denomination
# specified in this config (e.g. 0.25token1;0.0001token2).

# ...

# default: the last 100 states are kept in addition to every 500th state; pruning at 10 block intervals
# nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
# everything: all saved states will be deleted, storing only the current state; pruning at 10 block intervals
# custom: allow pruning options to be manually specified through 'pruning-keep-recent', 'pruning-keep-every', >
pruning = "nothing"
pruning-keep-recent = "0"
pruning-keep-every = "0"
pruning-interval = "0"
```

## WebSocket Server

Websocket is a bidirectional transport protocol. A Websocket connection is maintained by client and server until it is explicitly terminated by one. Most modern browsers support Websocket which means it has good tooling.

Because Websocket is bidirectional, servers can push events to clients. That makes Websocket a good choice for use-cases involving event subscription.
Another benefit of Websocket is that after the handshake procedure, the overhead of individual messages is low, making it good for sending high number of requests.
The WebSocket Server can be enabled from the `app.toml`

```toml
# Address defines the EVM WebSocket server address to bind to.
ws-address = "0.0.0.0:8546"
```
