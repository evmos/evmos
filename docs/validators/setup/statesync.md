<!--
order: 4
-->

# State Sync

Learn about Tendermint Core state sync and support offered by the Cosmos SDK. {synopsis}

:::tip
**Note**: Only curious about how to sync a node with the network? Skip to [this section](#state-syncing-a-node).
:::

## Tendermint Core State Sync

State sync allows a new node to join a network by fetching a snapshot of the network state at a recent height, instead of fetching and replaying all historical blocks. Since application state is smaller than the combination of all blocks, and restoring state is faster than replaying blocks, this reduces the time to sync with the network from days to minutes.

This section of the document provides a brief overview of the Tendermint state sync protocol, and how to sync a node. For more details, refer to the [ABCI Application Guide](https://docs.tendermint.com/master/spec/abci/apps.html#state-sync) and the [ABCI Reference Documentation](https://docs.tendermint.com/master/spec/abci/abci.html).

### State Sync Snapshots

A guiding principle when designing Tendermint state sync was to give applications as much flexibility as possible. Therefore, Tendermint does not care what snapshots contain, how they are taken, or how they are restored. It is only concerned with discovering existing snapshots in the network, fetching them, and passing them to applications via ABCI. Tendermint uses light client verification to check the final app hash of a restored application against the chain app hash, but any further verification must be done by the application itself during restoration.

Snapshots consist of binary chunks in an arbitrary format. Chunks cannot be larger than 16 MB, but otherwise there are no restrictions. [Snapshot metadata](https://docs.tendermint.com/master/spec/abci/abci.html#snapshot), exchanged via ABCI and P2P, contains the following fields:

- `height` (`uint64`): height at which the snapshot was taken
- `format` (`uint32`): arbitrary application-specific format identifier (eg. version)
- `chunks` (`uint32`): number of binary chunks in the snapshot
- `hash` (`bytes`): arbitrary snapshot hash for comparing snapshots across nodes
- `metadata` (`bytes`): arbitrary binary snapshot metadata for use by applications

The `format` field allows applications to change their snapshot format in a backwards-compatible manner, by providing snapshots in multiple formats, and choosing which formats to accept during restoration. This is useful when, for example, changing serialization or compression formats: as nodes may be able to provide snapshots to peers running older verions, or make use of old snapshots when starting up with a newer version.

The `hash` field contains an arbitrary snapshot hash. Snapshots that have identical `metadata` fields (including `hash`) across nodes are considered identical, and `chunks` will be fetched from any of these nodes. The `hash` cannot be trusted, and is not verified by Tendermint itself, which guards against inadvertent nondeterminism in snapshot generation. The `hash` may be verified by the application instead.

The `metadata` field can contain any arbitrary metadata needed by the application. For example, the application may want to include chunk checksums to discard damaged `chunks`, or [Merkle proofs](https://ethereum.org/en/developers/tutorials/merkle-proofs-for-offline-data-integrity/) to verify each chunk individually against the chain app hash. In [Protobuf](https://developers.google.com/protocol-buffers/docs/overview)-encoded form, snapshot `metadata` messages cannot exceed 4 MB.

### Taking, Serving Snapshots

To enable state sync, some nodes in the network must take and serve snapshots. When a peer is attempting to state sync, an existing Tendermint node will call the following ABCI methods on the application to provide snapshot data to this peer:

- [`ListSnapshots`](https://docs.tendermint.com/master/spec/abci/abci.html#listsnapshots): returns a list of available snapshots, with metadata
- [`LoadSnapshotChunk`](https://docs.tendermint.com/master/spec/abci/abci.html#loadsnapshotchunk): returns binary chunk data

Snapshots should typically be generated at regular intervals rather than on-demand: this improves state sync performance, since snapshot generation can be slow, and avoids a denial-of-service vector where an adversary floods a node with such requests. Older snapshots can usually be removed, but it may be useful to keep at least the two most recent to avoid deleting the previous snapshot while a node is restoring it.

It is entirely up to the application to decide how to take snapshots, but it should strive to satisfy the following guarantees:

- **Asynchronous**: snapshotting should not halt block processing, and it should therefore happen asynchronously, eg. in a separate thread
- **Consistent**: snapshots should be taken at isolated heights, and should not be affected by concurrent writes, eg. due to block processing in the main thread
- **Deterministic**: snapshot `chunks` and `metadata` should be identical (at the byte level) across all nodes for a given `height` and `format`, to ensure good availability of `chunks`

As an example, this can be implemented as follows:

1. Use a data store that supports transactions with snapshot isolation, such as RocksDB or BadgerDB.
2. Start a read-only database transaction in the main thread after committing a block.
3. Pass the database transaction handle into a newly spawned thread.
4. Iterate over all data items in a deterministic order (eg. sorted by key)
5. Serialize data items (eg. using [Protobuf](https://developers.google.com/protocol-buffers/docs/overview)), and write them to a byte stream.
6. Hash the byte stream, and split it into fixed-size chunks (eg. of 10 MB)
7. Store the chunks in the file system as separate files.
8. Write the snapshot metadata to a database or file, including the byte stream hash.
9. Close the database transaction and exit the thread.

Applications may want to take additional steps as well, such as compressing the data, checksumming chunks, generating proofs for incremental verification, and removing old snapshots.

### Restoring Snapshots

When Tendermint starts, it will check whether the local node has any state (ie. whether `LastBlockHeight == 0`), and if it doesn't, it will begin discovering snapshots via the P2P network. These snapshots will be provided to the local application via the following ABCI calls:

- [`OfferSnapshot(snapshot, apphash)`](https://docs.tendermint.com/master/spec/abci/abci.html#offersnapshot): offers a discovered snapshot to the application
- [`ApplySnapshotChunk(index, chunk, sender)`](https://docs.tendermint.com/master/spec/abci/abci.html#applysnapshotchunk): applies a snapshot chunk

Discovered snapshots are offered to the application and it can respond by accepting the snapshot, rejecting it, rejecting the format, rejecting the senders, aborting state sync, and so on.

Once a snapshot is accepted, Tendermint will fetch chunks from across available peers, and apply them sequentially to the application, which can choose to accept the chunk, refetch it, reject the snapshot, reject the sender, abort state sync, and so on.

Once all chunks have been applied, Tendermint will call the [`Info` ABCI method](https://docs.tendermint.com/master/spec/abci/abci.html#info) on the application, and check that the app hash and height correspond to the trusted values from the chain. It will then switch to fast sync to fetch any remaining blocks (if enabled), before finally joining normal consensus operation.

How snapshots are actually restored is entirely up to the application, but will generally be the inverse of how they are generated. Note, however, that Tendermint only verifies snapshots after all chunks have been restored, and does not reject any P2P peers on its own. As long as the trusted hash and application code are correct, it is not possible for an adversary to cause a state synced node to have incorrect state when joining consensus, but it is up to the application to counteract state sync denial-of-service (eg. by implementing incremental verification, rejecting invalid peers).

Note that state synced nodes will have a truncated block history starting at the height of the restored snapshot, and there is currently no [backfill of all block data](https://github.com/tendermint/tendermint/issues/4629). Networks should consider broader implications of this, and may want to ensure at least a few archive nodes retain a complete block history, for both auditability and backup.

## Cosmos SDK State Sync

[Cosmos SDK](https://github.com/cosmos/cosmos-sdk) v0.40+ includes automatic support for state sync, so application developers only need to enable it to take advantage. They will not need to implement the state sync protocol described in the [above section on Tendermint](#tendermint-core-state-sync) themselves.

### State Sync Snapshots

Tendermint Core handles most of the grunt work of discovering, exchanging, and verifying state data for state sync, but the application must take snapshots of its state at regular intervals, and make these available to Tendermint via ABCI calls, and be able to restore these when syncing a new node.

The Cosmos SDK stores application state in a data store called [IAVL](https://github.com/cosmos/iavl), and each module can set up its own IAVL stores. At regular height intervals (which are configurable), the Cosmos SDK will export the contents of each store at that height, [Protobuf](https://developers.google.com/protocol-buffers/docs/overview)-encode and compress it, and save it to a snapshot store in the local filesystem. Since IAVL keeps historical versions of data, these snapshots can be generated simultaneously with new blocks being executed. These snapshots will then be fetched by Tendermint via ABCI when a new node is state syncing.

Note that only IAVL stores that are managed by the Cosmos SDK can be snapshotted. If the application stores additional data in external data stores, there is currently no mechanism to include these in state sync snapshots, so the application therefore cannot make use of automatic state sync via the SDK. However, it is free to implement the state sync protocol itself as described in the [ABCI Documentation](https://docs.tendermint.com/master/spec/abci/apps.html#state-sync).

When a new node is state synced, Tendermint will fetch a snapshot from peers in the network and provide it to the local (empty) application, which will import it into its IAVL stores. Tendermint then verifies the application's app hash against the main blockchain using light client verification, and proceeds to execute blocks as usual. Note that a state synced node will only restore the application state for the height the snapshot was taken at, and will not contain historical data nor historical blocks.

### Enabling State Sync Snapshots

To enable state sync snapshots, an application using the CosmosSDK `BaseApp` needs to set up a snapshot store (with a database and filesystem directory) and configure the snapshotting interval and the number of historical snapshots to keep. A minimal exmaple of this follows:

```bash
snapshotDir := filepath.Join(
  cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
snapshotDB, err := sdk.NewLevelDB("metadata", snapshotDir)
if err != nil {
  panic(err)
}
snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
if err != nil {
  panic(err)
}
app := baseapp.NewBaseApp(
  "app", logger, db, txDecoder,
  baseapp.SetSnapshotStore(snapshotStore),
  baseapp.SetSnapshotInterval(cast.ToUint64(appOpts.Get(
    server.FlagStateSyncSnapshotInterval))),
  baseapp.SetSnapshotKeepRecent(cast.ToUint32(appOpts.Get(
    server.FlagStateSyncSnapshotKeepRecent))),
)
```

When starting the application with the appropriate flags, (eg. `--state-sync.snapshot-interval 1000 --state-sync.snapshot-keep-recent 2`) it should generate snapshots and output log messages:

```bash
Creating state snapshot    module=main height=3000
Completed state snapshot   module=main height=3000 format=1
```

Note that the snapshot interval must currently be a multiple of the `pruning-keep-every` (defaults to 100), to prevent heights from being pruned while taking snapshots. It's also usually a good idea to keep at least 2 recent snapshots, such that the previous snapshot isn't removed while a node is attempting to state sync using it.

### State Syncing a Node

:::tip
Looking for snapshots or archive nodes to sync your node with? Check out [this page](../snapshots_archives.md).
:::

Once a few nodes in a network have taken state sync snapshots, new nodes can join the network using state sync. To do this, the node should first be configured as usual, and the following pieces of information must be obtained for light client verification:

- Two available RPC servers (at least)
- Trusted height
- Block ID hash of trusted height

The trusted hash must be obtained from a trusted source (eg. a block explorer), but the RPC servers do not need to be trusted. Tendermint will use the hash to obtain trusted app hashes from the blockchain in order to verify restored application snapshots. The app hash and corresponding height are the only pieces of information that can be trusted when restoring snapshots. Everything else can be forged by adversaries.

The required information can be obtained eg. via RPC.

```bash
$ curl -s http://foo.net:26657/block | \
  jq -r '.result.block.header.height + "\n" + .result.block_id.hash'
# <trusted height>
# <block ID hash of trusted height>
```

We can then configure Tendermint to use state sync in `config.toml`.

```bash
[statesync]
enable = true
rpc_servers = “rpc.a.com:26657,rpc.b.org:26657”
trust_height = # <trusted height>
trust_hash = # “<block ID hash of trusted height>”
trust_period = “336h” # 2/3 of the unbonding time
```

When the node is started it will then attempt to find a state sync snapshot in the network, and restore it:

```bash
Started node                   module=main nodeInfo="..."
Discovering snapshots for 20s
Discovered new snapshot        height=3000 format=1 hash=0F14A473
Discovered new snapshot        height=2000 format=1 hash=C6209AF7
Offering snapshot to ABCI app  height=3000 format=1 hash=0F14A473
Snapshot accepted, restoring   height=3000 format=1 hash=0F14A473
Fetching snapshot chunk        height=3000 format=1 chunk=0 total=3
Fetching snapshot chunk        height=3000 format=1 chunk=1 total=3
Fetching snapshot chunk        height=3000 format=1 chunk=2 total=3
Applied snapshot chunk         height=3000 format=1 chunk=0 total=3
Applied snapshot chunk         height=3000 format=1 chunk=1 total=3
Applied snapshot chunk         height=3000 format=1 chunk=2 total=3
Verified ABCI app              height=3000 appHash=F7D66BC9
Snapshot restored              height=3000 format=1 hash=0F14A473
Executed block                 height=3001 validTxs=16 invalidTxs=0
Committed state                height=3001 txs=16 appHash=0FDBB0D5F
Executed block                 height=3002 validTxs=25 invalidTxs=0
Committed state                height=3002 txs=25 appHash=40D12E4B3
```

The node is now state synced, having joined the network in seconds.

:::tip
**Note**: Information included in this document is sourced from [Erik Grinaker](https://medium.com/@erikgrinaker), specifically his state sync guides for [Tendermint Core](https://medium.com/tendermint/tendermint-core-state-sync-for-developers-70a96ba3ee35) and the [Cosmos SDK](https://medium.com/cosmos-blockchain/cosmos-sdk-state-sync-guide-99e4cf43be2f).
:::
