<!--
order: 2
-->

# Evmos APIs

Learn about all the available services for clients {synopsis}

::: tip
🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧

This documentation page is currently under work in progress.

🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧 🚧
:::

The Evmos client supports

- Cosmos [gRPC](#cosmos-grpc):
- Cosmos REST ([gRPC-Gateway](#cosmos-grpc-gateway)):
- Ethereum [JSON-RPC](#ethereum-json-rpc):
- Ethereum [Websocket](#ethereum-websocket):
- Tendermint [RPC](#tendermint-rpc):
- Tendermint [Websocket](#tendermint-websocket):

<!-- TODO: default port and address -->

## Ethereum JSON-RPC

<!-- TODO: Link JSON-RPC docs -->

Evmos also supports most of the standard [JSON-RPC APIs](./json-rpc/running_server) to connect with existing Ethereum-compatible web3 tooling.

::: tip
Check out the list of supported JSON-RPC API [endpoints](./json-rpc/endpoints) and [namespaces](./../api/json-rpc/namespaces).
:::

## Ethereum Websocket

<!-- TODO: Link WSS docs -->

## Cosmos gRPC

Evmos exposes gRPC endpoints for all the integrated Cosmos SDK modules. This makes it easier for
wallets and block explorers to interact with the Proof-of-Stake logic and native Cosmos transactions and queries.

### Cosmos gRPC-Gateway (HTTP REST)

[gRPC-Gateway](https://grpc-ecosystem.github.io/grpc-gateway/) reads a gRPC service definition and
generates a reverse-proxy server which translates RESTful JSON API into gRPC. With gRPC-Gateway,
users can use REST to interact the Cosmos gRPC service.

See the list of supported gRPC-Gateway API endpoints for the Evmos testnet [here](https://api.evmos.dev/).

## Tendermint Websocket

## Command Line Interface (CLI)
