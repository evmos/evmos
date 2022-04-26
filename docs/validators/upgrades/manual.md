
<!--
order: 3
-->

# Manual Upgrades

## Update Genesis file

If you are joining an existing testnet, you can fetch the genesis from the appropriate testnet or mainnet repository where the genesis file is hosted.

::: tip
You can find the latest `genesis.json` file for mainnet or testnet in the following repositories:

- **Mainnet**: [github.com/tharsis/mainnet](https://github.com/tharsis/mainnet)
- **Testnet**: [github.com/tharsis/testnets](https://github.com/tharsis/testnets)

:::

Save the new genesis as `new_genesis.json`. Then, replace the old `genesis.json` located in your `config/` directory with `new_genesis.json`:

```bash
cd $HOME/.evmosd/config
cp -f genesis.json new_genesis.json
mv new_genesis.json genesis.json
```

## Restart Node

To restart your node once the new genesis has been updated, use the `start` command:

```bash
evmosd start
```

## Updating the `evmosd` binary

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

You will need to ensure that the version installed matches the one needed for th testnet. Check the Evmos [releases page](https://github.com/tharsis/evmos/releases) for details on each release.

Verify that everything is OK. If you get something like the following, you've successfully installed Evmos on your system.

```bash
$ evmosd version --long

name: evmos
server_name: evmosd
version: 3.0.0
commit: fe9df43332800a74a163c014c69e62765d8206e3
build_tags: netgo,ledger
go: go version go1.18 darwin/amd64
...
```

If the software version does not match, then please check your `$PATH` to ensure the correct `evmosd` is running.
