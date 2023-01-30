<!--
order: 5
-->

# Run an IBC Relayer

Learn how to run an IBC Relayer for Evmos. {synopsis}

## Minimum Requirements

- 8 core (4 physical core), x86_64 architecture processor
- 32 GB RAM (or equivalent swap file set up)
- 1 TB+ nVME drives

If running many nodes on a single VM, [ensure your open files limit is increased](https://tecadmin.net/increase-open-files-limit-ubuntu/).

## Prerequisites
<!-- textlint-disable -->
Before beginning, ensure you have an Evmos node running in the background of the same machine that you intend to relay on.
Follow [this guide](quickstart/run_node.md) to set up an Evmos node if you have not already.
<!-- textlint-enable -->

In this guide, we will be relaying between [Evmos (channel-3) and Cosmos Hub (channel-292)](https://www.mintscan.io/evmos/relayers).
When setting up your Evmos and Cosmos full nodes,
be sure to offset the ports being used in both the `app.toml` and `config.toml` files of the respective chains
(this process will be shown below).

<!-- textlint-disable -->
In this example, the default ports for Evmos will be used, and the ports of the Cosmos Hub node will be manually changed.
<!-- textlint-enable -->

## Evmos Daemon Settings

First, set `grpc server` on port `9090` in the `app.toml` file from the `$HOME/.evmosd/config` directory:

```bash
vim $HOME/.evmosd/config/app.toml
```

```bash
[grpc]

# Enable defines if the gRPC server should be enabled.
enable = true

# Address defines the gRPC server address to bind to.
address = "0.0.0.0:9090"
```

Then, set the `pprof_laddr` to port `6060`, `rpc laddr` to port `26657`, and `prp laddr` to `26656` in the `config.toml`
file from the `$HOME/.evmosd/config` directory:

```bash
vim $HOME/.evmosd/config/config.toml
```

```bash
# pprof listen address (https://golang.org/pkg/net/http/pprof)
pprof_laddr = "localhost:6060"
```

```bash
[rpc]

# TCP or UNIX socket address for the RPC server to listen on
laddr = "tcp://127.0.0.1:26657"
```

```bash
[p2p]

# Address to listen for incoming connections
laddr = "tcp://0.0.0.0:26656"
```

## Cosmos Daemon Settings

First, set `grpc server` to port `9090` in the `app.toml` file from the `$HOME/.gaiad/config` directory:

```bash
vim $HOME/.gaiad/config/app.toml
```

```bash
[grpc]

# Enable defines if the gRPC server should be enabled.
enable = true

# Address defines the gRPC server address to bind to.
address = "0.0.0.0:9092"
```

Then, set the `pprof_laddr` to port `6062`, `rpc laddr` to port `26757`, and `prp laddr` to `26756` in the `config.toml`
file from the `$HOME/.gaiad/config` directory:

```bash
vim $HOME/.gaiad/config/app.toml
```

```bash
# pprof listen address (https://golang.org/pkg/net/http/pprof)
pprof_laddr = "localhost:6062"
```

```bash
[rpc]

# TCP or UNIX socket address for the RPC server to listen on
laddr = "tcp://127.0.0.1:26757"
```

```bash
[p2p]

# Address to listen for incoming connections
laddr = "tcp://0.0.0.0:26756"
```

## Install Rust Dependencies

Install the following rust dependencies:

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

```bash
source $HOME/.cargo/env
sudo apt-get install pkg-config libssl-dev
```

```bash
sudo apt install librust-openssl-dev build-essential git
```

## Build & Setup Hermes

Create the directory where the binary will be placed,
clone the hermes source repository,
and build it using the latest release.

```bash
mkdir -p $HOME/hermes
git clone https://github.com/informalsystems/ibc-rs.git hermes
cd hermes
git checkout v0.12.0
cargo install ibc-relayer-cli --bin hermes --locked
```

Make the hermes `config` and `keys` directory, and copy `config.toml` to the config directory:

```bash
mkdir -p $HOME/.hermes
mkdir -p $HOME/.hermes/keys
cp config.toml $HOME/.hermes
```

Check the hermes version and configuration directory setup:

```bash
$ hermes version
INFO ThreadId(01) using default configuration from '/home/relay/.hermes/config.toml'
hermes 0.12.0
```

Edit the hermes configuration (use ports according the port configuration set above,
adding only chains that will be relayed):

```bash
vim $HOME/.hermes/config/config.toml
```

```bash
# In this example, we will set channel-292 on the cosmoshub-4 chain settings and channel-3 on the evmos_9001-2 chain settings:
[[chains]]
id = 'cosmoshub-4'
rpc_addr = 'http://127.0.0.1:26757'
grpc_addr = 'http://127.0.0.1:9092'
websocket_addr = 'ws://127.0.0.1:26757/websocket'
...
[chains.packet_filter]
policy = 'allow'
list = [
   ['transfer', 'channel-292'], # evmos_9001-2
]

[[chains]]
id = 'evmos_9001-2'
rpc_addr = 'http://127.0.0.1:26657'
grpc_addr = 'http://127.0.0.1:9090'
websocket_addr = 'ws://127.0.0.1:26657/websocket'
...
address_type = { derivation = 'ethermint', proto_type = { pk_type = '/ethermint.crypto.v1.ethsecp256k1.PubKey' } }
[chains.packet_filter]
policy = 'allow'
list = [
  ['transfer', 'channel-3'], # cosmoshub-4
]
```

Add your relayer wallet to Hermes' keyring (located in `$HOME/.hermes/keys`)

The best practice is to use the same mnemonic over all networks.
Do not use your relaying-addresses for anything else, because it will lead to account sequence errors.

```bash
hermes keys restore cosmoshub-4 -m "24-word mnemonic seed"
hermes keys restore evmos_9001-2 -m "24-word mnemonic seed"
```

Ensure this wallet has funds in both EVMOS and ATOM in order to pay the fees required to relay.

## Final Checks

Validate your hermes configuration file:

```bash
$ hermes config validate
INFO ThreadId(01) using default configuration from '/home/relay/.hermes/config.toml'
Success: "validation passed successfully"
```

Perform the hermes `health-check` to see if all connected nodes are up and synced:

```bash
$ hermes health-check
INFO ThreadId(01) using default configuration from '/home/relay/.hermes/config.toml'
INFO ThreadId(01) telemetry service running, exposing metrics at http://0.0.0.0:3001/metrics
INFO ThreadId(01) starting REST API server listening at http://127.0.0.1:3000
INFO ThreadId(01) [cosmoshub-4] chain is healthy
INFO ThreadId(01) [evmos_9001-2] chain is healthy
```

When your nodes are fully synced, you can start the hermes daemon:

```bash
hermes start
```

Watch hermes' output for successfully relayed packets, or any errors.
It will try and clear any unrecieved packets after startup has completed.

## Helpful Commands

Query hermes for unrecieved packets and acknowledgements (ie. check if channels are "clear") with the following:

```bash
hermes query packet unreceived-packets cosmoshub-4 transfer channel-292
hermes query packet unreceived-acks cosmoshub-4 transfer channel-292
```

```bash
hermes query packet unreceived-packets evmos_9001-2 transfer channel-3
hermes query packet unreceived-acks evmos_9001-2 transfer channel-3
```

Query hermes for packet commitments with the following:

```bash
hermes query packet commitments cosmoshub-4 transfer channel-292
hermes query packet commitments evmos_9001-2 transfer channel-3
```

Clear the channel (only works on hermes `v0.12.0` and higher) with the following:

```bash
hermes clear packets cosmoshub-4 transfer channel-292
hermes clear packets evmos_9001-2 transfer channel-3
```

Clear unrecieved packets manually
(experimental, will need to stop hermes daemon to prevent confusion with account sequences)
with the following:

```bash
hermes tx raw packet-recv evmos_9001-2 cosmoshub-4 transfer channel-292
hermes tx raw packet-ack evmos_9001-2 cosmoshub-4 transfer channel-292
hermes tx raw packet-recv cosmoshub-4 evmos_9001-2 transfer channel-3
hermes tx raw packet-ack cosmoshub-4 evmos_9001-2 transfer channel-3
```
