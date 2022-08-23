### Sending a transaction for validator but having a problem?

Send us:

1) what your problem is (which command you're running and what it the error/problem)

AND

2) the output of these commands for debugging:

_(if your output message gets deleted from Discord due to bots being stupid - happens sometimes - send as a screenshot instead, or with https://paste.ofcode.org/)_

```
set -x
git rev-list HEAD | head -n 1
pointd status
pointd tendermint show-validator
pointd query tendermint-validator-set | grep "$(pointd tendermint show-address)"
pointd query slashing signing-info $(pointd tendermint show-validator)
pointd query staking validator $(pointd keys show validatorkey -a --bech val)
pointd query bank balances $(pointd keys show validatorkey | grep address: | cut -d ':' --complement -f 1)
set +x
```
