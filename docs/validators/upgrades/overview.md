<!--
order: 1
-->

# Software Upgrades

Learn how to manage chain upgrades for your full and validator nodes {synopsis}

There are different categories for upgrades:

- **Automatic or Manual Upgrades**: Validators and full node can choose if they want Cosmovisor process to perform the upgrade for them or just
- **Planned or Unplanned Upgrades**: Chain upgrades can be scheduled through an upgrade proposal or, in the case of a critical vulnerability, be coordinated internally between the validators and core teams to halt the chain at a given height and perform the upgrade
- **Breaking or Non-breaking Upgrades**: 

## Planned and Unplanned Upgrades

With every new software release, we strongly recommend full nodes and validator operators to perform a software upgrade.

You can upgrade your node by:

1) automatically, and
2) upgrading your node to that version

In this guide, you can find out how to automatically upgrade your node with Cosmovisor or perform the update manually.

## Breaking and Non-Breaking Upgrades

Upgrades can be categorized according to the Semantic versioning ([Semver](https://semver.org/)) of the corresponding software [release version](https://github.com/tharsis/evmos/releases) (*i.e* `vX.Y.Z`):

- Major version (`X`): backward incompatible API and state machine breaking changes
- Minor version (`Y`): new backward compatible features. These can be state machine breaking
- Patch version (`Z`): backwards compatible bug fixes, small refactors and improvements.

### Major Upgrades

If the new version you are upgrading to has breaking changes, you will have to [export](#export-state) the state  and [restart](#restart-node) your node.

in order to prevent [double signing or halting the chain during consensus](https://docs.tendermint.com/master/spec/consensus/signing.html#double-signing)

To upgrade the genesis file, you can either fetch it from a trusted source or export it locally using the `evmosd export` command.

### Minor Upgrades



### Patch Upgrades

In order to update a pat

you can skip to [Restart](#restart-node) after installing the new version.



## Upgrading a Node

We highly recommend validators use Cosmovisor to run their nodes. This will make low-downtime upgrades smoother, as validators don't have to manually upgrade binaries during the upgrade. Instead users can preinstall new binaries, and Cosmovisor will [automatically update](automated.md) them based on on-chain Software Upgrade proposals.

::: tip
For more info about Cosmovisor, check their official [documentation](https://docs.cosmos.network/main/run-node/cosmovisor.html)
:::

If you choose to use Cosmovisor, please continue to the [automated upgrade guide](./automated.md). If you choose to upgrade your node manually instead, skip to the [the instructions without Cosmovisor](./manual.md)
