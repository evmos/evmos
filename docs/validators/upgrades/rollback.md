<!--
order: 5
-->

# Rollback

Learn how to rollback the chain version in the case of an unsuccessful chain upgrade. {synopsis}

In order to restore a previous chain version, the following data must be recovered by validators:

- the database that contains the state of the previous chain (in `~/.evoblockd/data` by default)
- the `priv_validator_state.json` file of the validator (also in `~/.evoblockd/data` by default)

If validators don't possess their database data, another validator should share a copy of the database. Validators will be able to download a copy of the data and verify it before starting their node. If validators don't have the backup `priv_validator_state.json` file, then those validators will not have double-sign protection on their first block.

## Restoring State Procedure

1. First, stop your node.

2. Then, copy the contents of your backup data directory back to the `$EVO_HOME/data` directory (which, by default, should be `~/.evoblockd/data`).

```bash
# Assumes backup is stored in "backup" directory
rm -rf ~/.evoblockd/data
mv backup/.evoblockd/data ~/.evoblockd/data
```

3. Next, install the previous version of Evoblock.

```bash
# from evoblock directory
git checkout <prev_version>
make install
## verify version
evoblockd version --long
```

4. Finally, start the node.

```bash
evoblockd start
```
