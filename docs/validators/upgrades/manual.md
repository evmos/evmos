
<!--
order: 3
-->

# Manual Upgrades

Learn how to manually upgrade your node. {synopsis}

## Pre-requisites

- [Install Evoblock](./../quickstart/installation.md) {prereq}

## 1. Upgrade the Evoblock version

Before upgrading the Evoblock version. Stop your instance of `evoblockd` using `Ctrl/Cmd+C`.

Next, upgrade the software to the desired release version. Check the Evoblock [releases page](https://github.com/evoblockchain/evoblock/releases) for details on each release.

::: warning
Ensure that the version installed matches the one needed for the network you are running (mainnet or testnet).
:::

```bash
cd evoblock
git fetch --all && git checkout <new_version>
make install
```

::: tip
If you have issues at this step, please check that you have the latest stable version of [Golang](https://golang.org/dl/) installed.
:::

Verify that you've successfully installed Evoblock on your system by using the `version` command:

```bash
$ evoblockd version --long

name: evoblock
server_name: evoblockd
version: 3.0.0
commit: fe9df43332800a74a163c014c69e62765d8206e3
build_tags: netgo,ledger
go: go version go1.18 darwin/amd64
...
```

::: tip
If the software version does not match, then please check your `$PATH` to ensure the correct `evoblockd` is running.
:::

## 2. Replace Genesis file

::: tip
You can find the latest `genesis.json` file for mainnet or testnet in the following repositories:

- **Mainnet**: [github.com/tharsis/mainnet](https://github.com/tharsis/mainnet)
- **Testnet**: [github.com/tharsis/testnets](https://github.com/tharsis/testnets)
:::

Save the new genesis as `new_genesis.json`. Then, replace the old `genesis.json` located in your `config/` directory with `new_genesis.json`:

```bash
cd $HOME/.evoblockd/config
cp -f genesis.json new_genesis.json
mv new_genesis.json genesis.json
```

::: tip
We recommend using `sha256sum` to check the hash of the downloaded genesis against the expected genesis.

```bash
cd ~/.evoblockd/config
echo "<expected_hash>  genesis.json" | sha256sum -c
```

:::

## 3. Data Reset

::: danger
Check [here](./upgrades.md) if the version you are upgrading require a data reset (hard fork). If this is not the case, you can skip to [Restart](#restart-node).
:::

Remove the outdated files and reset the data:

```bash
rm $HOME/.evoblockd/config/addrbook.json
evoblockd tendermint unsafe-reset-all --home $HOME/.evoblockd
```

Your node is now in a pristine state while keeping the original `priv_validator.json` and `config.toml`. If you had any sentry nodes or full nodes setup before,
your node will still try to connect to them, but may fail if they haven't also
been upgraded.

::: danger
🚨 **IMPORTANT** 🚨

Make sure that every node has a unique `priv_validator.json`. **DO NOT** copy the `priv_validator.json` from an old node to multiple new nodes. Running two nodes with the same `priv_validator.json` will cause you to [double sign](https://docs.tendermint.com/master/spec/consensus/signing.html#double-signing).
:::

## 4. Restart Node

To restart your node once the new genesis has been updated, use the `start` command:

```bash
evoblockd start
```
