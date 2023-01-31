<!--
order: 1
-->

# Single Node

## Pre-requisite Readings

- [Install Binary](./../../validators/quickstart/installation.md)  {prereq}

## Automated Localnet (script)

You can customize the local testnet script by changing values for convenience for example:

```bash
# customize the name of your key, the chain-id, moniker of the node, keyring backend, and log level
KEY="dev0"
CHAINID="evmos_9000-4"
MONIKER="localtestnet"
KEYRING="test"
LOGLEVEL="info"


# Allocate genesis accounts (cosmos formatted addresses)
evmosd add-genesis-account $KEY 100000000000000000000000000aevmos --keyring-backend $KEYRING

# Sign genesis transaction
evmosd gentx $KEY 1000000000000000000000aevmos --keyring-backend $KEYRING --chain-id $CHAINID
```

The default configuration will generate a single validator localnet with the chain-id
`evmosd-1` and one predefined account (`dev0`) with some allocated funds at the genesis.

You can start the local chain using:

```bash
 $ local_node.sh
...
```

:::tip
To avoid overwriting any data for a real node used in production,
it was decided to store the automatically generated testing configuration at `~/.tmp-evmosd`
instead of the default `~/.evmosd`.
:::

When working with the `local_node.sh` script, it is necessary to extend all `evmosd` commands,
that target the local test node, with the `--home ~/.tmp-evmosd` flag.
This is mandatory, because the `home` directory cannot be stored in the `evmosd` configuration,
which can be seen in the output below.
For ease of use, it might be sensible to export this directory path as an environment variable:

```
 $ export TMP=$HOME/.tmp-evmosd`
 $ evmosd config --home $TMP
{
	"chain-id": "evmos_9000-1",
	"keyring-backend": "test",
	"output": "text",
	"node": "tcp://localhost:26657",
	"broadcast-mode": "sync"
}
```

## Manual Localnet

This guide helps you create a single validator node that runs a network locally for testing
and other development related uses.

### Initialize the chain

Before actually running the node, we need to initialize the chain, and most importantly its genesis file.
This is done with the `init` subcommand:

```bash
$MONIKER=testing
$KEY=dev0
$CHAINID="evmos_9000-4"

# The argument $MONIKER is the custom username of your node, it should be human-readable.
evmosd init $MONIKER --chain-id=$CHAINID
```

::: tip
You can [edit](./../../validators/quickstart/binary.md#configuring-the-node) this `moniker` later
by updating the `config.toml` file.
:::

The command above creates all the configuration files needed for your node and validator to run,
as well as a default genesis file, which defines the initial state of the network.
All these [configuration files](./../../validators/quickstart/binary.md#configuring-the-node)
are in `~/.evmosd` by default, but you can overwrite the location of this folder by passing the `--home` flag.

### Genesis Procedure

### Adding Genesis Accounts

Before starting the chain, you need to populate the state with at least one account
using the [keyring](./../../users/keys/keyring.md#add-keys):

```bash
evmosd keys add my_validator
```

Once you have created a local account, go ahead and grant it some `aevmos` tokens in your chain's genesis file.
Doing so will also make sure your chain is aware of this account's existence:

```bash
evmosd add-genesis-account my_validator 10000000000aevmos
```

Now that your account has some tokens, you need to add a validator to your chain.

For this guide, you will add your local node (created via the `init` command above) as a validator of your chain.
Validators can be declared before a chain is first started
via a special transaction included in the genesis file called a `gentx`:

```bash
# Create a gentx
# NOTE: this command lets you set the number of coins. 
# Make sure this account has some coins with the genesis.app_state.staking.params.bond_denom denom
evmosd add-genesis-account my_validator 1000000000stake,10000000000aevmos
```

A `gentx` does three things:

1. Registers the `validator` account you created as a validator operator account
   (i.e. the account that controls the validator).
2. Self-delegates the provided `amount` of staking tokens.
3. Link the operator account with a Tendermint node pubkey that will be used for signing blocks.
If no `--pubkey` flag is provided, it defaults to the local node pubkey created via the `evmosd init` command above.

For more information on `gentx`, use the following command:

```bash
evmosd gentx --help
```

### Collecting `gentx`

By default, the genesis file do not contain any `gentxs`. A `gentx` is a transaction that bonds
staking token present in the genesis file under `accounts` to a validator, essentially creating a
validator at genesis. The chain will start as soon as more than 2/3rds of the validators (weighted
by voting power) that are the recipient of a valid `gentx` come online after `genesis_time`.

A `gentx` can be added manually to the genesis file, or via the following command:

```bash
# Add the gentx to the genesis file
evmosd collect-gentxs
```

This command will add all the `gentxs` stored in `~/.evmosd/config/gentx` to the genesis file.

### Run Testnet

Finally, check the correctness of the `genesis.json` file:

```bash
evmosd validate-genesis
```

Now that everything is set up, you can finally start your node:

```bash
evmosd start
```

:::tip
To check all the available customizable options when running the node, use the `--help` flag.
:::

You should see blocks come in.

The previous command allow you to run a single node.
This is enough for the next section on interacting with this node,
but you may wish to run multiple nodes at the same time, and see how consensus happens between them.

You can then stop the node using `Ctrl+C`.
