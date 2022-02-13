<!--
order: 3
-->

# Clients

Learn about the client supported by your Evmos node. {synopsis}

## Client Servers

The Evmos client supports both Cosmos [gRPC endpoints](https://api.evmos.org/) and Ethereum's [JSON-RPC](https://eth.wiki/json-rpc/API).

### Cosmos gRPC and Tendermint RPC

Evmos exposes gRPC endpoints (and REST) for all the integrated Cosmos-SDK modules. This makes it easier for
wallets and block explorers to interact with the proof-of-stake logic and native Cosmos transactions and queries

::: tip
See the list of supported gRPC Gatewat API [endpoints](https://api.evmos.org/).
:::

### Ethereum JSON-RPC server

Evmos also supports most of the standard web3 [JSON-RPC APIs](./../api/json-rpc/running_server) to connect with existing web3 tooling.

::: tip
See the list of supported JSON-RPC API [endpoints](./../api/json-rpc/endpoints) and [namespaces](./../api/json-rpc/namespaces).
:::

To connect to the JSON-PRC server, start the node with the `--json-rpc.enable=true` flag and define the namespaces that you would like to run using the `--evm.rpc.api` flag (e.g. `"txpool,eth,web3,net,personal"`. Then, you can point any Ethereum development tooling to `http://localhost:8545` or whatever port you choose with the listen address flag (`--json-rpc.address`).

<!-- TODO: add Rosetta -->