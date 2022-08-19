### Sending a transaction for validator but having a problem?

Send us the output of these commands for debugging:

```
git rev-list HEAD | head -n 1
evmosd status
evmosd tendermint show-validator
evmosd query tendermint-validator-set | grep "$(evmosd tendermint show-address)"
evmosd query slashing signing-info $(evmosd tendermint show-validator)
evmosd query staking validator $(evmosd keys show validatorkey -a --bech val)
evmosd query bank balances $(evmosd keys show validatorkey | grep address: | cut -d ':' --complement -f 1)
```
