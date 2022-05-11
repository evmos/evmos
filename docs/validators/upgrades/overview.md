<!--
order: 1
-->

# Overview

Learn how to manage chain upgrades for your full and validator nodes. {synopsis}

## Upgrade Categories

There are 3 different categories for upgrades:

- **Planned or Unplanned Upgrades**: Chain upgrades can be scheduled at a given height through an upgrade proposal plan.
- **Breaking or Non-breaking Upgrades**: Upgrades can be API or State Machine breaking, which affects backwards compatibility. To address this, the application state or genesis file would need to be migrated in preparation for the upgrade.
- **Data Reset Upgrades**: Some upgrades will need a full data reset in order to clean the state. This can sometimes occur in the case of a rollback or hard fork.

Additionally, validators can choose how to manage the upgrade according to their preferred option:

- **Automatic or Manual Upgrades**: Validator can run the `cosmovisor` process to automatically perform the upgrade or do it manually.

## Planned and Forks Upgrades

### Planned Upgrades

Planned upgrades are coordinated scheduled upgrades that use the [upgrade module](https://docs.evmos.org/modules/upgrade/) logic. This facilitates smoothly upgrading Evmos to a new (breaking) software version as it automatically handles the state migration for the new release.

#### Governance Proposal

Governance Proposals are a mechanism for coordinating an upgrade at a given height or time using an [`SoftwareProposal`](https://docs.evmos.org/modules/upgrade/01_concepts.html#proposal).

::: tip
All governance proposals, including software upgrades, need to wait for the voting period to conclude before the upgrade can be executed. Consider this duration when submitting a software upgrade proposal.
:::

If the proposal passes, the upgrade `Plan`, which targets a specific upgrade logic to migrate the state, is persisted to the blockchain state and scheduled at the given upgrade height. The upgrade can be delayed or expedited by updating the `Plan.Height` in a new proposal.

#### Hard Forks

A special type of planned upgrades are hard forks. Hard Forks, as opposed to [Governance Proposal}(#governance-proposal), don't require waiting for the full voting
period. This makes them ideal for coordinating security vulnerabilities and patches.

The upgrade (fork) block height is set in the `BeginBlock` of the application (i.e before the transactions are processed for the block). Once the blockchain reaches that height, it automatically schedules an upgrade `Plan` for the same height and then triggers the upgrade process. After upgrading, the block operations (`BeginBlock`, transaction processing and state `Commit`) continue normally.

::: tip
In order to execute an upgrade hard fork, a [patch version](#patch-versions) needs to first be released with the `BeginBlock` upgrade scheduling logic. After a +2/3 of the validators upgrade to the new patch version, their nodes will automatically halt and upgrade the binary.
:::

### Unplanned Upgrades

Unplanned upgrades are upgrades where all the validators need to gracefully halt and shut down their nodes at exactly the same point in the process. This can be done by setting the `--halt-height` flag when running the `evmosd start` command.

If there are breaking changes during an unplanned upgrade (see below), validators will need to migrate the state and genesis before restarting their nodes.

::: tip
The main consideration with unplanned upgrades is that the genesis state needs to be exported and the blockchain data needs to be [reset](#data-reset-upgrades). This mainly affects infrastructure providers, tools and clients like block explorers and clients, which have to use archival nodes to serve queries for the pre-upgrade heights.
:::

### Breaking and Non-Breaking Upgrades

Upgrades can be categorized as breaking or non-breaking according to the Semantic versioning ([Semver](https://semver.org/)) of the corresponding software [release version](https://github.com/tharsis/evmos/releases) (*i.e* `vX.Y.Z`):

- **Major version (`X`)**: backward incompatible API and state machine breaking changes.
- **Minor version (`Y`)**: new backward compatible features. These can be also be state machine breaking.
- **Patch version (`Z`)**: backwards compatible bug fixes, small refactors and improvements.

#### Major Versions

If the new version you are upgrading to has breaking changes, you will have to:

1. Migrate genesis JSON
2. Migrate application state
3. Restart node

This needs to be done to prevent [double signing or halting the chain during consensus](https://docs.tendermint.com/master/spec/consensus/signing.html#double-signing).

To upgrade the genesis file, you can either fetch it from a trusted source or export it locally using the `evmosd export` command.

#### Minor Versions

If the new version you are upgrading to has breaking changes, you will have to:

1. Migrate the state (if applicable)
2. Restart node

#### Patch Versions

In order to update a patch:

1. Stop Node
2. Download new release binary manually
3. Restart node

### Data Reset Upgrades

Data Reset upgrades require node operators to fully reset the blockchain state and restart their nodes from a clean
state, but using the same validator keys.

### Automatic or Manual Upgrades

With every new software release, we strongly recommend full nodes and validator operators to perform a software upgrade.

You can upgrade your node by either:

- [automatically](./automated.md) bumping the software version and restart the node once the upgrade occurs, or
- download the new binary and perform a [manual upgrade](./manual.md)

Follow the links in the options above to learn how to upgrade your node according to your preferred option.
