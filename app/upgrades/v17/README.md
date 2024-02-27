# Testing Procedure

## Migration Logic

The files in this directory have only been added in order to test
how long the migration is going to take using mainnet data.

In order to do so, only the files containing the actual migration logic
(not the upgrade handler or tests) have been ported.
The source commit is this one:
https://github.com/evmos/evmos/blob/fee60bbe9f1dff12f480c3f33b3706458a5c0604/

The basic procedure has been described by Tom previously in this document:
https://www.notion.so/altiplanic/Test-Upgrade-Logic-with-mainnet-DB-502d842799174935afc4a9b7dc1b49c7#70237a9b76a54de782880bdfe49e3b40

**Note:** Not all steps in this document have to be followed here,
because the snapshot is already available on the used machine,
since the Devops team provided a synced mainnet node.

## Applying In The BeginBlocker

The migration logic has been added to the `BeginBlocker` of the `app.go` file.
This will allow us to run the migration logic without adjusting the validator set.

## Running The Tests

The testing environment is the _dev machine_ provided by the Devops team.
To connect to it, you have to provide your SSH keys to the Devops team
and then run the following command on your machine:

```bash
ssh altiplanic@65.109.156.253
```

The _dev machine_ has an `evmosd` instance running with that is synced to mainnet.
In order to run the tests, you have to stop the running node:

```bash
sudo systemctl stop evmosd
```

Then, you have to navigate to the `evmos` directory
and build the binary from this branch, which has the migration logic included in the `BeginBlocker`.

```bash
cd ~/evmos
git checkout malte/test-strv2-migration-beginblocker
make install
```

Afterwards, you can restart the node and listen to the logs:

```bash
sudo systemctl restart evmosd && journalctl -fu evmosd -o cat
```

## Results

