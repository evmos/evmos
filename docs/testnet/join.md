<!--
order: 1
-->

# Joining a Testnet

This document outlines the steps to join an existing testnet {synopsis}

## Pick a Testnet

You specify the network you want to join by setting the **genesis file** and **seeds**. If you need more information about past networks, check our [testnets repo](https://github.com/tharsis/testnets).

| Network Chain ID | Description                       | Site                                                                       | Version                                               |
| ---------------- | --------------------------------- | -------------------------------------------------------------------------- | ----------------------------------------------------- |
| `evmos_9000-3`   | Evmos_9000-3 Testnet              | [Evmos 9000-3](https://github.com/tharsis/testnets/tree/main/evmos_9000-3) | [`v1.x.x`](https://github.com/tharsis/evmos/releases) |
| `evmos_9000-2`   | Olympus Mons Incentivized Testnet | [Olympus Mons](https://github.com/tharsis/testnets/tree/main/olympus_mons) | [`v0.3.x`](https://github.com/tharsis/evmos/releases) |
| `evmos_9000-1`   | Arsia Mons Testnet                | [Arsia Mons](https://github.com/tharsis/testnets/tree/main/arsia_mons)     | [`v0.1.x`](https://github.com/tharsis/evmos/releases) |

## Install `evmosd`

Follow the [installation](./../quickstart/installation) document to install the {{ $themeConfig.project.name }} binary `{{ $themeConfig.project.binary }}`.

:::warning
Make sure you have the right version of `{{ $themeConfig.project.binary }}` installed.
:::

### Save Chain ID

We recommend saving the testnet `chain-id` into your `{{ $themeConfig.project.binary }}`'s `client.toml`. This will make it so you do not have to manually pass in the `chain-id` flag for every CLI command.

::: tip
See the Official [Chain IDs](./../basics/chain_id.md#official-chain-ids) for reference.
:::

```bash
evmosd config chain-id evmos_9000-3
```

## Initialize Node

We need to initialize the node to create all the necessary validator and node configuration files:

```bash
evmosd init <your_custom_moniker> --chain-id evmos_9000-3
```

::: danger
Monikers can contain only ASCII characters. Using Unicode characters will render your node unreachable.
:::

By default, the `init` command creates your `~/.evmosd` (i.e `$HOME`) directory with subfolders `config/` and `data/`.
In the `config` directory, the most important files for configuration are `app.toml` and `config.toml`.

## Genesis & Seeds

### Copy the Genesis File

Check the `genesis.json` file from the [`testnets`](https://github.com/tharsis/testnets) repository and copy it over to the `config` directory: `~/.evmosd/config/genesis.json`. This is a genesis file with the chain-id and genesis accounts balances.

```bash
curl https://raw.githubusercontent.com/tharsis/testnets/main/evmos_9000-3/genesis.zip > ~/.evmosd/config/genesis.zip
sudo apt install -y unzip
cd ~/.evmosd/config
unzip genesis.zip
rm genesis.zip
```

Then verify the correctness of the genesis configuration file:

```bash
evmosd validate-genesis
```

### Add Seed Nodes

Your node needs to know how to find [peers](https://docs.tendermint.com/master/tendermint-core/using-tendermint.html#peers). You'll need to add healthy [seed nodes](https://docs.tendermint.com/master/tendermint-core/using-tendermint.html#seed) to `$HOME/.evmosd/config/config.toml`. The [`testnets`](https://github.com/tharsis/testnets) repo contains links to some seed nodes.

Edit the file located in `~/.evmosd/config/config.toml` and the `seeds` to the following:

```toml
#######################################################
###           P2P Configuration Options             ###
#######################################################
[p2p]

# ...

# Comma separated list of seed nodes to connect to
seeds = "<node-id>@<ip>:<p2p port>"
```

You can use the following code to get seeds from the repo and add it to your config:

```bash
SEEDS=`curl -sL https://raw.githubusercontent.com/tharsis/testnets/main/evmos_9000-3/seeds.txt | awk '{print $1}' | paste -s -d, -`
sed -i.bak -e "s/^seeds =.*/seeds = \"$SEEDS\"/" ~/.evmosd/config/config.toml
```

:::tip
For more information on seeds and peers, you can the Tendermint [P2P documentation](https://docs.tendermint.com/master/spec/p2p/peer.html).
:::

### Add Persistent Peers

We can set the [`persistent_peers`](https://docs.tendermint.com/master/tendermint-core/using-tendermint.html#persistent-peer) field in `~/.evmosd/config/config.toml` to specify peers that your node will maintain persistent connections with. You can retrieve them from the list of
available peers on the [`testnets`](https://github.com/tharsis/testnets) repo.

A list of available persistent peers is also available in the `#find-peers` channel in the [Evmos Discord](https://discord.gg/evmos). You can get a random 10 entries from the `peers.txt` file in the `PEERS` variable by running the following command:

```bash
PEERS=`curl -sL https://raw.githubusercontent.com/tharsis/testnets/main/evmos_9000-3/peers.txt | sort -R | head -n 10 | awk '{print $1}' | paste -s -d, -`
```

Use `sed` to include them into the configuration. You can also add them manually:

```bash
sed -i.bak -e "s/^persistent_peers *=.*/persistent_peers = \"$PEERS\"/" ~/.evmosd/config/config.toml
```

## Run a Testnet Validator

Claim your testnet {{ $themeConfig.project.testnet_denom }} on the [faucet](./faucet.md) using your validator account address and submit your validator account address:

::: tip
For more details on how to configure your validator, follow the validator [setup](./../guides/validators/setup.md) instructions.
:::

```bash
evmosd tx staking create-validator \
  --amount=1000000000000atevmos \
  --pubkey=$(evmosd tendermint show-validator) \
  --moniker="EvmosWhale" \
  --chain-id=<chain_id> \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1000000" \
  --gas="auto" \
  --gas-prices="0.025aevmos" \
  --from=<key_name>
```

## Start testnet

The final step is to [start the nodes](./../quickstart/run_node#start-node). Once enough voting power (+2/3) from the genesis validators is up-and-running, the testnet will start producing blocks.

```bash
evmosd start
```

## Upgrading Your Node

> NOTE: These instructions are for full nodes that have ran on previous versions of and would like to upgrade to the latest testnet.

### Reset Data

:::warning
If the version <new_version> you are upgrading to is not breaking from the previous one, you **should not** reset the data. If this is the case you can skip to [Restart](#restart)
:::

First, remove the outdated files and reset the data.

```bash
rm $HOME/.evmosd/config/addrbook.json $HOME/.evmosd/config/genesis.json
evmosd unsafe-reset-all
```

Your node is now in a pristine state while keeping the original `priv_validator.json` and `config.toml`. If you had any sentry nodes or full nodes setup before,
your node will still try to connect to them, but may fail if they haven't also
been upgraded.

::: danger Warning
Make sure that every node has a unique `priv_validator.json`. Do not copy the `priv_validator.json` from an old node to multiple new nodes. Running two nodes with the same `priv_validator.json` will cause you to double sign.
:::

### Restart

To restart your node, just type:

```bash
evmosd start
```

## Share your Peer

You can share your peer to posting it in the `#find-peers` channel in the [Evmos Discord](https://discord.gg/evmos).

::: tip
To get your Node ID use

```bash
evmosd tendermint show-node-id
```

:::
