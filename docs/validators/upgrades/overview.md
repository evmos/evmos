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

## Planned and Unplanned Upgrades

Planned upgrades are coordinated scheduled upgrades that use the [upgrade module](https://docs.cosmos.network/main/modules/upgrade/) logic. This facilitates smoothly upgrading Evmos to a new (breaking) software version as it automatically handles the state migration for the

Unplanned upgrades are upgrades where all the validators need to gracefully halt and shut down their nodes at exactly the same point in the process. This can be done by setting the `--halt-height` flag when running the `evmosd start` command.

If there are breaking changes during an unplanned upgrade (see below), validators will need to migrate the state and genesis before restarting their nodes.

### Breaking and Non-Breaking Upgrades

Upgrades can be categorized as breaking or non-breaking according to the Semantic versioning ([Semver](https://semver.org/)) of the corresponding software [release version](https://github.com/tharsis/evmos/releases) (*i.e* `vX.Y.Z`):

- **Major version (`X`)**: backward incompatible API and state machine breaking changes.
- **Minor version (`Y`)**: new backward compatible features. These can be also be state machine breaking.
- **Patch version (`Z`)**: backwards compatible bug fixes, small refactors and improvements.

#### Major Upgrades

If the new version you are upgrading to has breaking changes, you will have to:

1. Migrate genesis JSON
2. Migrate application state
3. Restart node

This needs to be done to prevent [double signing or halting the chain during consensus](https://docs.tendermint.com/master/spec/consensus/signing.html#double-signing).

To upgrade the genesis file, you can either fetch it from a trusted source or export it locally using the `evmosd export` command.

#### Minor Upgrades

If the new version you are upgrading to has breaking changes, you will have to:

1. Migrate the state (if applicable)
2. Restart node

#### Patch Upgrades

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
