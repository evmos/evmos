<!--
order: 5
-->

# Hard Fork Upgrades

Learn how to manually upgrade your node. {synopsis}

One of the significant limitations of the normal upgrade procedure [via
governance](overview.md#governance-proposal) is that it requires waiting for the
entire duration of the voting period. This duration makes it unsuitable for
automated upgrades that involve patches for security vulnerabilities or other
critical components.

A faster alternative to using governance is to create a Hard Fork procedure.
This procedure [automatically](automated.md) applies the changes from an upgrade plan, allowing
them to be executed at a given block height without the need of having to create
a governance proposal.

The high-level strategy for coordinating an upgrade is as follows:

1. The vulnerability is fixed on a private branch that contains breaking
   changes.
2. A new patch release (e.g. `v8.0.0` -> `v8.0.1`) needs to be created that
   contains a hard fork logic and performs an upgrade to the next breaking
   version (e.g. `v9.0.0`) at a predefined block height.
3. Validators upgrade their nodes to the patch release (e.g. `v8.0.1`). In order to perform the
   hard fork successfully, it’s important that enough validators upgrade to the
   patch release so that they make up at least 2/3 of the total validator voting
   power.
4. One hour before the upgrade time (corresponding to the upgrade block height),
   the new major release (e.g. `v9.0.0`) including the vulnerability fix is
   published.

::: danger
**Important**: The release needs to be created with 1hr anticipation because the
release binaries take ~30min to be created and validators need a buffer time to
download them and update their
[cosmovisor](/docs/validators/upgrades/automated.md#using-cosmovisor) settings.
:::
