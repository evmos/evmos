<!--
order: 5
-->

# Mempool

Learn about the available mempool options in Tendermint.{synopsis}

## FIFO Mempool

The mempool holds uncommitted transactions, which are not yet included in a block.
The default mempool implementation for Tendermint blockchains follows a first-in-first-out (FIFO) principle,
which means the ordering of transactions depends solely on the order in which they arrive at the node.
The first transaction to be received will be the first transaction to be processed.
This is true for gossiping the received transactions to the rest of the peers as well as including them in a block.

## Prioritized Mempool

Starting with [Tendermint v0.35](https://github.com/tendermint/tendermint/blob/v0.35.0/CHANGELOG.md)
(has also been backported to [v0.34.20](https://github.com/tendermint/tendermint/blob/17c94bb0dcb354c57f49cdcd1e62f4742752c803/UPGRADING.md?plain=1#L54))
it is possible to use a prioritized mempool implementation.
This allows validators to choose transactions based on the associated fees or other incentive mechanisms.
It is achieved by passing a `priority` field with each [`CheckTx` response](https://github.com/tendermint/tendermint/blob/17c94bb0dcb354c57f49cdcd1e62f4742752c803/proto/tendermint/abci/types.proto#L234),
which is run on any transaction trying to enter the mempool.

Evmos supports [EIP-1559](https://eips.ethereum.org/EIPS/eip-1559#simple-summary) EVM transactions through its
<!-- markdown-link-check-disable-next-line -->
[feemarket](../../modules/feemarket/01_concepts.md) module.
This transaction type uses a base fee and a selectable priority tip that add up to the total transaction fees.
The prioritized mempool presents an option to automatically make use of this mechanism regarding block generation.

When using the prioritized mempool, transactions for the next produced block are chosen
by order of their priority (i.e. their fees) from highest to lowest.
Should the mempool be full, the prioritized implementation allows
to remove the transactions with the lowest priority until enough disk space is available for
an incoming, higher-priority transaction (see [v1/mempool.go](https://github.com/tendermint/tendermint/blob/17c94bb0dcb354c57f49cdcd1e62f4742752c803/mempool/v1/mempool.go#L505C2-L576) implementation for more details).

::: tip
Even though the transaction processing can be ordered by priority, the gossiping of transactions will always be according to FIFO.
:::

## Configuration

To use the a prioritized mempool, adjust `version = "v1"` in the node configuration at `~/.evmosd/config/config.toml`.
The default value `"v0"` indicates the traditional FIFO mempool.

::: tip
Remember to **restart** the node for the changes to take effect.
:::

See the relevant excerpt from `config.toml` here:

```toml
#######################################################
###          Mempool Configuration Option          ###
#######################################################
[mempool]

# Mempool version to use:
#   1) "v0" - (default) FIFO mempool.
#   2) "v1" - prioritized mempool.
version = "v1"
```

## Resources

More detailed information can be found here:

- [Tendermint ADR-067 - Mempool Refactor](https://github.com/tendermint/tendermint/blob/main/docs/architecture/adr-067-mempool-refactor.md).
- [Blogpost: Tendermint v0.35 Announcement](https://medium.com/tendermint/tendermint-v0-35-introduces-prioritized-mempool-a-makeover-to-the-peer-to-peer-network-more-61eea6ec572d)
- [EIP-1559: Fee market change for ETH 1.0 chain](https://eips.ethereum.org/EIPS/eip-1559)
- [EIP-1559 FAQ](https://notes.ethereum.org/@vbuterin/eip-1559-faq)
- [Blogpost: What is EIP-1559? How will it change Ethereum?](https://consensys.net/blog/quorum/what-is-eip-1559-how-will-it-change-ethereum/)
