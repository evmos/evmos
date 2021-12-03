<!--
order: 1
-->

# Upgrade Node

Learn how to upgrade your full node to the latest software version {synopsis}

With every new software release, we strongly recommend validators to perform a software upgrade.

<!-- `TODO: WHY? Can you get slashed otherwise?` -->

You can upgrade your node by 1) upgrading your software version and 2) upgrading your node to that version. In this guide, you can find out how to automatically upgrade your node with Cosmovisor or perform the update manually.

## Software Upgrade

These instructions are for full nodes that have ran on previous versions of and would like to upgrade to the latest testnet.

First, stop your instance of `evmosd`. Next, upgrade the software:

```bash
cd evmos
git fetch --all && git checkout <new_version>
make install
```

::: tip
If you have issues at this step, please check that you have the latest stable version of GO installed.
:::

You will need to ensure that the version installed matches the one needed for th testnet. Check the Evmos [release page](https://github.com/tharsis/evmos/releases) for details on each release.

Verify that everything is OK. If you get something like the following, you've successfully installed Evmos on your system.

```bash
$ evmosd version --long

name: evmos
server_name: evmosd
version: 0.4.0
commit: 070b668f2cbbf52548c46e96b236e09884483dd4
build_tags: netgo,ledger
go: go version go1.17 darwin/amd64
...
```

If the software version does not match, then please check your $PATH to ensure the correct evmosd is running.

## Upgrade Node

We highly recommend validators use Cosmovisor to run their nodes. This will make low-downtime upgrades smoother, as validators don't have to manually upgrade binaries during the upgrade. Instead users can preinstall new binaries, and cosmovisor will automatically update them based on on-chain Software Upgrade proposals.

You should review the docs for Cosmovisor located [here](https://docs.cosmos.network/master/run-node/cosmovisor.html)

If you choose to use Cosmovisor, please continue with these instructions. If you choose to upgrade your node manually instead, skip to the [the instructions without Cosmovisor](#upgrade-manually)

### Upgrade with Cosmovisor

> `cosmovisor` is a small process manager for Cosmos SDK application binaries that monitors the governance module for incoming chain upgrade proposals. If it sees a proposal that gets approved, cosmovisor can automatically download the new binary, stop the current binary, switch from the old binary to the new one, and finally restart the node with the new binary.

#### Install and Setup

To get started with [Cosmovisor](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor) first download it

```bash
go get github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor
```

Set up the Cosmovisor environment variables. We recommend setting these in your `.profile` so it is automatically set in every session.

```bash
echo "# Setup Cosmovisor" >> ~/.profile
echo "export DAEMON_NAME=evmosd" >> ~/.profile
echo "export DAEMON_HOME=$HOME/.evmosd" >> ~/.profile
echo 'export PATH="$DAEMON_HOME/cosmovisor/current/bin:$PATH"' >> ~/.profile
source ~/.profile
```


After this, you must make the necessary folders for cosmosvisor in your daemon home directory (~/.evmosd).

```bash
mkdir -p ~/.evmosd/cosmovisor/upgrades
mkdir -p ~/.evmosd/cosmovisor/genesis/bin
cp $(which evmosd) ~/.evmosd/cosmovisor/genesis/bin/

# Verify the setup
# It should return the same version as evmosd
cosmovisor version
```

#### Preparing an Upgrade

Cosmovisor will continually poll the `$DAEMON_HOME/data/upgrade-info.json` for new upgrade instructions. When an upgrade is ready, node operators can download the new binary and place it under `$DAEMON_HOME/cosmovisor/upgrades/<name>/bin` where `<name>` is the URI-encoded name of the upgrade as specified in the upgrade module plan.

It is possible to have Cosmovisor automatically download the new binary. To do this set the following environment variable.

```bash
export DAEMON_ALLOW_DOWNLOAD_BINARIES=true
```

#### Download Genesis File

You can now download the "genesis" file for the chain. It is pre-filled with the entire genesis state and gentxs.

```bash
$ curl https://raw.githubusercontent.com/tharsis/testnets/main/olympus_mons/genesis.json > ~/.evmosd/config/genesis.json
```

We recommend using `sha256sum` to check the hash of the genesis.

```bash
cd ~/.evmosd/config
echo "2b5164f4bab00263cb424c3d0aa5c47a707184c6ff288322acc4c7e0c5f6f36f  genesis.json" | sha256sum -c
```

#### Reset Chain Database

There shouldn't be any chain database yet, but in case there is for some reason, you should reset it. This is a good idea especially if you ran `evmosd start` on an old, broken genesis file.

```bash
evmosd unsafe-reset-all
```

#### Ensure that you have set peers

In `~/.evmosd/config/config.toml` you can set your peers. See the [peers.txt](https://github.com/tharsis/testnets/blob/main/olympus_mons/peers.txt) file for a list of up to date peers.

See the [Add persistent peers section](https://evmos.dev/testnet/join.html#add-persistent-peers) in our docs for an automated method, but field should look something like a comma separated string of peers (do not copy this, just an example):

```bash
persistent_peers = "5576b0160761fe81ccdf88e06031a01bc8643d51@195.201.108.97:24656,13e850d14610f966de38fc2f925f6dc35c7f4bf4@176.9.60.27:26656,38eb4984f89899a5d8d1f04a79b356f15681bb78@18.169.155.159:26656,59c4351009223b3652674bd5ee4324926a5a11aa@51.15.133.26:26656,3a5a9022c8aa2214a7af26ebbfac49b77e34e5c5@65.108.1.46:26656,4fc0bea2044c9fd1ea8cc987119bb8bdff91aaf3@65.21.246.124:26656,6624238168de05893ca74c2b0270553189810aa7@95.216.100.80:26656,9d247286cd407dc8d07502240245f836e18c0517@149.248.32.208:26656,37d59371f7578101dee74d5a26c86128a229b8bf@194.163.172.168:26656,b607050b4e5b06e52c12fcf2db6930fd0937ef3b@95.217.107.96:26656,7a6bbbb6f6146cb11aebf77039089cd038003964@94.130.54.247:26656"
```

You can share your peer with

```bash
evmosd tendermint show-node-id
```

**Peer Format**: `node-id@ip:port`

**Example**: `3d892cfa787c164aca6723e689176207c1a42025@143.198.224.124:26656`

If you are relying on just seed node and no persistent peers or a low amount of them, please increase the following params in `config.toml`:

```bash
# Maximum number of inbound peers
max_num_inbound_peers = 200

# Maximum number of outbound peers to connect to, excluding persistent peers
max_num_outbound_peers = 100
```

#### Start your node

Now that everything is setup and ready to go, you can start your node.

```bash
cosmovisor start
```

You will need some way to keep the process always running. If you're on linux, you can do this by creating a service.

```bash
sudo tee /etc/systemd/system/evmosd.service > /dev/null <<EOF
[Unit]
Description=Evmos Daemon
After=network-online.target

[Service]
User=$USER
ExecStart=$(which cosmovisor) start
Restart=always
RestartSec=3
LimitNOFILE=infinity

Environment="DAEMON_HOME=$HOME/.evmosd"
Environment="DAEMON_NAME=evmosd"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"

[Install]
WantedBy=multi-user.target
EOF
```

Then update and start the node

```bash
sudo -S systemctl daemon-reload
sudo -S systemctl enable evmosd
sudo -S systemctl start evmosd
```

You can check the status with:

```bash
systemctl status evmosd
```

### Upgrade Manually

#### Upgrade Genesis File

:::warning
If the new version you are upgrading to has breaking changes, you will have to [export](#export-state) the state  and [restart](#restart-node) your node.

If it is **not** breaking (eg. from `v0.1.x` to `v0.1.<x+1>`), you can skip to [Restart](#restart-node) after installing the new version.
:::

To upgrade the genesis file, you can either fetch it from a trusted source or export it locally using the `evmosd export` command.

#### Fetch from a Trusted Source

If you are joining an existing testnet, you can fetch the genesis from the appropriate testnet source/repository where the genesis file is hosted.

Save the new genesis as `new_genesis.json`. Then, replace the old `genesis.json` with `new_genesis.json`.

```bash
cd $HOME/.evmosd/config
cp -f genesis.json new_genesis.json
mv new_genesis.json genesis.json
```


#### Export State

Evmos can dump the entire application state to a JSON file. This, besides upgrades, can be
useful for manual analysis of the state at a given height.

Export state with:

```bash
evmosd export > new_genesis.json
```

You can also export state from a particular height (at the end of processing the block of that height):

```bash
evmosd export --height [height] > new_genesis.json
```

If you plan to start a new network for 0 height (i.e genesis) from the exported state, export with the `--for-zero-height` flag:

```bash
evmosd export --height [height] --for-zero-height > new_genesis.json
```

Then, replace the old `genesis.json` with `new_genesis.json`.

```bash
cp -f genesis.json new_genesis.json
mv new_genesis.json genesis.json
```

At this point, you might want to run a script to update the exported genesis into a genesis state that is compatible with your new version.

You can use the `migrate` command to migrate from a given version to the next one (eg: `v0.X.X` to `v1.X.X`):

```bash
evmosd migrate [target-version] [/path/to/genesis.json] --chain-id=<new_chain_id> --genesis-time=<yyyy-mm-ddThh:mm:ssZ>
```

#### Restart Node

To restart your node once the new genesis has been updated, use the `start` command:

```bash
evmosd start
```
