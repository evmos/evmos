<!--
order: 6
-->

# Snapshots & Archive Nodes

Quickly sync your node with Evmos using a snapshot or serve queries for prev versions using archive nodes {synopsis}

## List of Snapshots and Archives

Below is a list of publicly available snapshots that you can use to sync with the Evmos mainnet and
archived [9001-1 mainnet](https://github.com/evmos/mainnet/tree/main/evmos_9001-1):

<!-- markdown-link-check-disable -->

### Snapshots

| Name        | URL                                                                     |
| -------------|------------------------------------------------------------------------ |
| `Staketab`   | [github.com/staketab/nginx-cosmos-snap](https://github.com/staketab/nginx-cosmos-snap/blob/main/docs/evmos.md) |
| `Polkachu`   | [polkachu.com](https://www.polkachu.com/tendermint_snapshots/evmos)                   |
| `Nodes Guru` | [snapshots.nodes.guru/evmos_9001-2/](snapshots.nodes.guru/evmos_9001-2/)                   |
| `Notional`   | [mainnet/pruned/evmos_9001-2(pebbledb)](https://snapshot.notional.ventures/evmos/) <br> [mainnet/archive/evmos_9001-2(pebbledb)](https://snapshot.notional.ventures/evmos-archive/) <br> [testnet/archive/evmos_9000-4(pebbledb)](https://snapshot.notional.ventures/evmos-testnet-archive/)                   |

### Archives
<!-- markdown-link-check-disable -->

| Name           | URL                                                                             |
| ---------------|---------------------------------------------------------------------------------|
| `Nodes Guru`   | [snapshots.nodes.guru/evmos_9001-1](https://snapshots.nodes.guru/evmos_9001-1/)                                    |
| `Polkachu`     | [polkachu.com/tendermint_snapshots/evmos](https://www.polkachu.com/tendermint_snapshots/evmos)                           |
| `Forbole`      | [bigdipper.live/evmos_9001-1](https://s3.bigdipper.live.eu-central-1.linodeobjects.com/evmos_9001-1.tar.lz4) |

To access snapshots and archives, follow the process below (this code snippet is to access a snapshot of the current network, `evmos_9001-2`, from Nodes Guru):

```bash
cd $HOME/.evmosd/data
wget https://snapshots.nodes.guru/evmos_9001-2/evmos_9001-2-410819.tar
tar xf evmos_9001-2-410819.tar
```

### PebbleDB

To use PebbleDB instead of GoLevelDB when using snapshots from Notional:

Build:

```bash
go mod edit -replace github.com/tendermint/tm-db=github.com/baabeetaa/tm-db@pebble
go mod tidy
go install -tags pebbledb -ldflags "-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=pebbledb" ./...
```

Download snapshot:

```bash
cd $HOME/.evmosd/
URL_SNAPSHOT="https://snapshot.notional.ventures/evmos/data_20221024_193254.tar.gz"
wget -O - "$URL_SNAPSHOT" |tar -xzf -
```

Start:

Set `db_backend = "pebbledb"` in `config.toml` or start with `--db_backend=pebbledb`

```bash
evmosd start --db_backend=pebbledb
```

**Note**: use this [workaround](https://github.com/notional-labs/cosmosia/blob/main/docs/pebbledb.md) when upgrading a node running PebbleDB.
