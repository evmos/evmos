<!--
order: 5
-->

# Prioritized Mempool

Learn about using the prioritized Tendermint mempool.{synopsis}

The mempool holds uncommitted transactions, which are not yet included in a block.
The default mempool for Tendermint blockchains follows a first-in-first-out (FIFO) principle,
which means the ordering of transactions depends solely on the order in which they arrived at the node.

Starting with [Tendermint v0.35](https://github.com/tendermint/tendermint/blob/v0.35.0/CHANGELOG.md)
(has also been backported to [v0.34.20](https://github.com/tendermint/tendermint/blob/17c94bb0dcb354c57f49cdcd1e62f4742752c803/UPGRADING.md?plain=1#L54))
it is possible to use a prioritized mempool implementation.
This allows validators to choose transactions based on the associated fees or other incentive mechanisms.
It is achieved by passing a `priority` field with each [`CheckTx` response](https://github.com/tendermint/tendermint/blob/17c94bb0dcb354c57f49cdcd1e62f4742752c803/proto/tendermint/abci/types.proto#L234),
which is run on any transaction trying to enter the mempool.
The current Cosmos SDK implementation allows the application layer to define a function of type [`TxFeeChecker`](https://github.com/cosmos/cosmos-sdk/blob/37a9bc3bb67bd82d4493d2d86f8cd31c0e768880/x/auth/ante/fee.go#L13),
which can be set as a [field](https://github.com/evmos/evmos/blob/main/app/ante/handler_options.go#L36) on the [`ante.HandlerOptions`](https://github.com/evmos/evmos/blob/main/app/app.go#L785-L798) in `app.go`.

The highest-priority transactions will be chosen for the creation of the next block.
When the mempool is full, the prioritized implementation allows to iterate over the stored transactions
and remove those with the lowest priority until enough disk space is available for
an incoming, higher-priority transaction ([v1/mempool.go](https://github.com/tendermint/tendermint/blob/17c94bb0dcb354c57f49cdcd1e62f4742752c803/mempool/v1/mempool.go#L505C2-L576)).

To use the a prioritized mempool, adjust `version = "v1"` inside of the node configuration at `~/.evmosd/config/config.toml`.
The default value `v0` indicates the traditional FIFO mempool.

See the relevant excerpt from `config.toml` here:

```
#######################################################
###          Mempool Configuration Option          ###
#######################################################
[mempool]

# Mempool version to use:
#   1) "v0" - (default) FIFO mempool.
#   2) "v1" - prioritized mempool.
version = "v1"
```

::: tip
Even though the transaction processing can be ordered by priority, the gossiping of transactions will always be according to FIFO.
:::

## Resources

More detailed information can be found in [Tendermint ADR-067 - Mempool Refactor](https://github.com/tendermint/tendermint/blob/main/docs/architecture/adr-067-mempool-refactor.md).
