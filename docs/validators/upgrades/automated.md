
<!--
order: 2
-->

# Automated Upgrades

## Using Cosmovisor

> `cosmovisor` is a small process manager for Cosmos SDK application binaries that monitors the governance module for incoming chain upgrade proposals. If it sees a proposal that gets approved, cosmovisor can automatically download the new binary, stop the current binary, switch from the old binary to the new one, and finally restart the node with the new binary.

### Install and Setup

To get started with [Cosmovisor](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor) first download it

```bash
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v1.0.0
```

Set up the Cosmovisor environment variables. We recommend setting these in your `.profile` so it is automatically set in every session.

```bash
echo "# Setup Cosmovisor" >> ~/.profile
echo "export DAEMON_NAME=evmosd" >> ~/.profile
echo "export DAEMON_HOME=$HOME/.evmosd" >> ~/.profile
source ~/.profile
```

After this, you must make the necessary folders for cosmosvisor in your daemon home directory (~/.evmosd) and copy over the current binary.

```bash
mkdir -p ~/.evmosd/cosmovisor
mkdir -p ~/.evmosd/cosmovisor/genesis
mkdir -p ~/.evmosd/cosmovisor/genesis/bin
mkdir -p ~/.evmosd/cosmovisor/upgrades

cp $GOPATH/bin/evmosd ~/.evmosd/cosmovisor/genesis/bin
```

To check that you did this correctly, ensure your versions of cosmovisor and evmosd are the same:

```
cosmovisor version
evmosd version
```

### Generally Preparing an Upgrade

Cosmovisor will continually poll the `$DAEMON_HOME/data/upgrade-info.json` for new upgrade instructions. When an upgrade is ready, node operators can download the new binary and place it under `$DAEMON_HOME/cosmovisor/upgrades/<name>/bin` where `<name>` is the URI-encoded name of the upgrade as specified in the upgrade module plan.

It is possible to have Cosmovisor automatically download the new binary. To do this set the following environment variable.

```bash
echo "export DAEMON_ALLOW_DOWNLOAD_BINARIES=true" >> ~/.profile
```

### Download Genesis File

You can now download the "genesis" file for the chain. It is pre-filled with the entire genesis state and gentxs.

```bash
curl https://raw.githubusercontent.com/tharsis/testnets/main/olympus_mons/genesis.json > ~/.evmosd/config/genesis.json
```

We recommend using `sha256sum` to check the hash of the genesis.

```bash
cd ~/.evmosd/config
echo "2b5164f4bab00263cb424c3d0aa5c47a707184c6ff288322acc4c7e0c5f6f36f  genesis.json" | sha256sum -c
```

### Reset Chain Database

There shouldn't be any chain database yet, but in case there is for some reason, you should reset it. This is a good idea especially if you ran `evmosd start` on an old, broken genesis file.

```bash
evmosd unsafe-reset-all
```

### Ensure that you have set peers

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

### Start your node

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

### update Cosmosvisor

If you're not yet on the latest V1 release (`v1.1.2`) please upgrade your current version first:

```bash
cd $HOME/evmos
git pull
git checkout v1.1.2
make build
systemctl stop evmosd.service
cp build/evmosd ~/.evmosd/cosmovisor/genesis/bin
systemctl start evmosd.service
cd $HOME
```

If you are on the latest V1 release (`v1.1.2`) and you want evmosd to upgrade automatically from V1 to V2, do the following steps prior to the upgrade height:

```bash
mkdir -p ~/.evmosd/cosmovisor/upgrades/v2/bin
cd $HOME/evmos
git pull
git checkout v2.0.0
make build
systemctl stop evmosd.service
cp build/evmosd ~/.evmosd/cosmovisor/upgrades/v2/bin
systemctl start evmosd.service
cd $HOME
```
