<!--
order: 2
-->

# Configuration

## Block Time

The timeout-commit value in the node config defines how long we wait after committing a block, before starting on the new height (this gives us a chance to receive some more pre-commits, even though we already have +2/3). The current default value is `"1s"`.

::: tip
**Note**: From v6, this is handled automatically by the server when initializing the node.
Validators will need to ensure their local node configurations in order to speed up the network to ~2s block times.
:::

```toml
# In ~/.evmosd/config/config.toml

#######################################################
###         Consensus Configuration Options         ###
#######################################################
[consensus]

### ... 

# How long we wait after committing a block, before starting on the new
# height (this gives us a chance to receive some more precommits, even
# though we already have +2/3).
timeout_commit = "1s"
```

## Peers

In `~/.evmosd/config/config.toml` you can set your peers.

See the [Add persistent peers section](../testnet.md#add-persistent-peers) in our docs for an automated method, but field should look something like a comma separated string of peers (do not copy this, just an example):

```bash
persistent_peers = "5576b0160761fe81ccdf88e06031a01bc8643d51@195.201.108.97:24656,13e850d14610f966de38fc2f925f6dc35c7f4bf4@176.9.60.27:26656,38eb4984f89899a5d8d1f04a79b356f15681bb78@18.169.155.159:26656,59c4351009223b3652674bd5ee4324926a5a11aa@51.15.133.26:26656,3a5a9022c8aa2214a7af26ebbfac49b77e34e5c5@65.108.1.46:26656,4fc0bea2044c9fd1ea8cc987119bb8bdff91aaf3@65.21.246.124:26656,6624238168de05893ca74c2b0270553189810aa7@95.216.100.80:26656,9d247286cd407dc8d07502240245f836e18c0517@149.248.32.208:26656,37d59371f7578101dee74d5a26c86128a229b8bf@194.163.172.168:26656,b607050b4e5b06e52c12fcf2db6930fd0937ef3b@95.217.107.96:26656,7a6bbbb6f6146cb11aebf77039089cd038003964@94.130.54.247:26656"
```

### Sharing your Peer

You can see and share your peer with the `tendermint show-node-id` command

```bash
evmosd tendermint show-node-id
ac29d21d0a6885465048a4481d16c12f59b2e58b
```

- **Peer Format**: `node-id@ip:port`
- **Example**: `ac29d21d0a6885465048a4481d16c12f59b2e58b@143.198.224.124:26656`

### Healthy peers

If you are relying on just seed node and no persistent peers or a low amount of them, please increase the following params in the `config.toml`:

```bash
# Maximum number of inbound peers
max_num_inbound_peers = 120

# Maximum number of outbound peers to connect to, excluding persistent peers
max_num_outbound_peers = 60
```
