<!--
order: 1
-->

# Software Upgrades

Learn how to upgrade your full nodes and validator nodes to the latest software version {synopsis}

With every new software release, we strongly recommend validators to perform a software upgrade, in order to prevent [double signing or halting the chain during consensus](https://docs.tendermint.com/master/spec/consensus/signing.html#double-signing).

You can upgrade your node by 1) upgrading your software version and 2) upgrading your node to that version. In this guide, you can find out how to automatically upgrade your node with Cosmovisor or perform the update manually.

## Coordinating upgrades

## Upgrading a Node

We highly recommend validators use Cosmovisor to run their nodes. This will make low-downtime upgrades smoother, as validators don't have to manually upgrade binaries during the upgrade. Instead users can preinstall new binaries, and Cosmovisor will [automatically update](automated.md) them based on on-chain Software Upgrade proposals.

::: tip
For more info about Cosmovisor, check their official [documentation](https://docs.cosmos.network/main/run-node/cosmovisor.html)
:::

If you choose to use Cosmovisor, please continue to the [automated upgrade guide](./automated.md). If you choose to upgrade your node manually instead, skip to the [the instructions without Cosmovisor](./manual.md)
