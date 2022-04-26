
<!--
order: 2
-->

# Automated Upgrades

Learn how to automate chain upgrades using Cosmovisor. {synopsis}

## Pre-requisites

- [Install Cosmovisor](https://docs.cosmos.network/main/run-node/cosmovisor.html#installation) {prereq}

## Using Cosmovisor

> `cosmovisor` is a small process manager for Cosmos SDK application binaries that monitors the governance module for incoming chain upgrade proposals. If it sees a proposal that gets approved, cosmovisor can automatically download the new binary, stop the current binary, switch from the old binary to the new one, and finally restart the node with the new binary.

::: tip
👉 For more info about Cosmovisor, please refer to the project official documentation [here](https://docs.cosmos.network/main/run-node/cosmovisor.html).
:::

We highly recommend validators use Cosmovisor to run their nodes. This will make low-downtime upgrades smoother, as validators don't have to [manually upgrade](./manual.md) binaries during the upgrade. Instead users can [pre-install](#manual-download) new binaries, and Cosmovisor will automatically update them based on on-chain Software Upgrade proposals.

### 1. Setup Cosmovisor

Set up the Cosmovisor environment variables. We recommend setting these in your `.profile` so it is automatically set in every session.

```bash
echo "# Setup Cosmovisor" >> ~/.profile
echo "export DAEMON_NAME=evmosd" >> ~/.profile
echo "export DAEMON_HOME=$HOME/.evmosd" >> ~/.profile
source ~/.profile
```

After this, you must make the necessary folders for `cosmosvisor` in your `DAEMON_HOME` directory (`~/.evmosd`) and copy over the current binary.

```bash
mkdir -p ~/.evmosd/cosmovisor
mkdir -p ~/.evmosd/cosmovisor/genesis
mkdir -p ~/.evmosd/cosmovisor/genesis/bin
mkdir -p ~/.evmosd/cosmovisor/upgrades

cp $GOPATH/bin/evmosd ~/.evmosd/cosmovisor/genesis/bin
```

To check that you did this correctly, ensure your versions of `cosmovisor` and `evmosd` are the same:

```bash
cosmovisor version
evmosd version
```

### 2. Download the Evmos release

#### 2.a) Manual Download

Cosmovisor will continually poll the `$DAEMON_HOME/data/upgrade-info.json` for new upgrade instructions. When an upgrade is [released](https://github.com/tharsis/evmos/releases), node operators need to:

1. Download (**NOT INSTALL**) the binary for the new release
2. Place it under `$DAEMON_HOME/cosmovisor/upgrades/<name>/bin`, where `<name>` is the URI-encoded name of the upgrade as specified in the Software Upgrade Plan.

**Example**: for a `Plan` with name `v3.0.0` with the following `upgrade-info.json`:

```json
{
    "binaries": {
        "darwin/arm64": "https://github.com/tharsis/evmos/releases/download/v3.0.0/evmos_3.0.0_Darwin_arm64.tar.gz",
        "darwin/x86_64": "https://github.com/tharsis/evmos/releases/download/v3.0.0/evmos_3.0.0_Darwin_x86_64.tar.gz",
        "linux/arm64": "https://github.com/tharsis/evmos/releases/download/v3.0.0/evmos_3.0.0_Linux_arm64.tar.gz",
        "linux/x86_64": "https://github.com/tharsis/evmos/releases/download/v3.0.0/evmos_3.0.0_Linux_x86_64.tar.gz",
        "windows/x86_64": "https://github.com/tharsis/evmos/releases/download/v3.0.0/evmos_3.0.0_Windows_x86_64.zip"
    }
}
```

Your `cosmovisor/` directory should look like this:

```shell
cosmovisor/
├── current/   # either genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── evmosd
└── upgrades
    └── v3.0.0
        ├── bin
        │   └── evmosd
        └── upgrade-info.json
```

#### 2.b) Automatic Download

::: warning
**NOTE**: Auto-download doesn't verify in advance if a binary is available. If there will be any issue with downloading a binary, `cosmovisor` will stop and won't restart an the chain (which could lead it to a halt).
:::

It is possible to have Cosmovisor [automatically download](https://docs.cosmos.network/main/run-node/cosmovisor.html#auto-download) the new binary. Validators can use the automatic download option to prevent unnecessary downtime during the upgrade process. This option will automatically restart the chain with the upgrade binary once the chain has halted at the proposed `upgrade-height`. The major benefit of this option is that validators can prepare the upgrade binary in advance and then relax at the time of the upgrade.

To set the auto-download use set the following environment variable:

```bash
echo "export DAEMON_ALLOW_DOWNLOAD_BINARIES=true" >> ~/.profile
```

### 3. Start your node

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
