# Unstake money:
At first you can see all the info for contracts doing this:
```pointd query staking validators```.

Once you find your validator (by moniker or using some id)
```pointd tendermint show-validator```, you can have details for yours, 
you need your key evmosvaloper format.

Next command gives you a key starting with `evmosvaloper`:
```pointd keys show mykey -a --bech val```

Which you need for the command:
```pointd query staking validator <evmosvaloperkey>```

You will get response:
```
commission:
  commission_rates:
    max_change_rate: “0.010000000000000000”
    max_rate: “0.200000000000000000"
    rate: “0.100000000000000000”
  update_time: “2022-08-18T23:22:38.373880888Z”
consensus_pubkey:
  ‘@type’: /cosmos.crypto.ed25519.PubKey
  key: +iAqX3tMVKEfh9S7o6lYKHSIYyrQsUacqeST2vZNxHA=
delegator_shares: “0.000000000000000000”
description:
  details: “”
  identity: “”
  moniker: mymonikkerr
  security_contact: “”
  website: “”
jailed: true
min_self_delegation: “100000000000000000000"
operator_address: evmosvaloper1uzwfry3nlrsc36j88zlk0un6nfyn6rrzkp86vr
status: BOND_STATUS_UNBONDING
tokens: “0”
unbonding_height: “174822"
unbonding_time: “2022-09-08T23:37:07.821587227Z”
```
In tokens you can see the amount of staked tokens.

Then you can run this command
```
pointd tx staking unbond evmosvaloper1uzwfry3nlrsc36j88zlk0un6nfyn6rrzkp86vr 98898998998000000000apoint \
--chain-id=point_10721-1 \
--from=pugliese \
--keyring-backend file \
--gas="400000" \
--gas-prices="0.025apoint"
```

If didn’t let you use the same wallet address and validator address, 
you can try changing wallet address but it is notenough, so you need
delete the file: ```~/.pointd/config/priv_validator_key.json```

Then restarted the node, and run the query again:
```
pointd tx staking create-validator  \
--amount=100000000000000000000apoint \
--pubkey=$(pointd tendermint show-validator) \
--moniker="brianvalidator" \
--chain-id=point_10721-1 \
--commission-rate="0.10" \
--commission-max-rate="0.20" \
--commission-max-change-rate="0.01" \
--min-self-delegation="1" \
--gas="400000" \
--gas-prices="0.025apoint" \
--from=pugliese2 \
--keyring-backend file  
```

Once you check the tx was successful you can run:
```
pointd query staking validator <evmosvaloperkey>
```
and get the info from the blockchain (