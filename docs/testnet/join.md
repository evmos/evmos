<!--
order: 1
-->

# Joining a Testnet

This document outlines the steps to join an existing testnet {synopsis}

## Install `evmosd`

Follow the [installation](./../quickstart/installation) document to install the Evmos binary `evmosd`.

:::warning
Make sure you have the right version of `evmosd` installed
:::

## Initialize Node

We need to initialize the node to create all the necessary validator and node configuration files:

```bash
evmosd init <your_custom_moniker>
```

::: danger
Monikers can contain only ASCII characters. Using Unicode characters will render your node unreachable.
:::

By default, the `init` command creates your `~/.evmosd` directory with subfolders `config/` and `data/`.
In the `config` directory, the most important files for configuration are `app.toml` and `config.toml`.

## Genesis & Seeds

### Copy the Genesis File

Check the genesis file from the [`testnets`](https://github.com/tharsis/testnets) repository and copy it over to the `config` directory: `~/.evmosd/config/genesis.json`.

Then verify the correctness of the genesis configuration file:

```bash
evmosd validate-genesis
```

### Add Seed Nodes

Your node needs to know how to find peers. You'll need to add healthy seed nodes to `$HOME/.evmosd/config/config.toml`. The [`testnets`](https://github.com/tharsis/testnets) repo contains links to some seed nodes.

Edit the file located in `~/.evmosd/config/config.toml` and the `seeds` to the following:

```toml
#######################################################
###           P2P Configuration Options             ###
#######################################################
[p2p]

# ...

# Comma separated list of seed nodes to connect to
seeds = ""
```

:::tip
For more information on seeds and peers, you can the Tendermint [P2P documentation](https://docs.tendermint.com/master/spec/p2p/peer.html).
:::

### Start testnet

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
