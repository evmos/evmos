# Unjailing tutorial
Being a validator you can be jailed for some reason. 

to check if your validator is jailed or not run the command below:
```
pointd query slashing signing-info $(pointd tendermint show-validator)
```
If your are jailed or was jailed earlier, it will return something like this.
```aidl
address: evmosvalcons1qae8r355w5cryez6klj2d47nlddlk0k6n4nt8m
index_offset: "2"
jailed_until: "2022-08-19T17:02:30.543786843Z"
missed_blocks_counter: "2"
start_height: "203923"
tombstoned: false
```
There is a `jailed_until` property in UTC time zone. So recalculate it according to your local timezone to check if jailing time ended alread.
If jailing time ended already, than you can unjail, if you did not unjail your self earlier.
Remember, the jailing data is just a historical record about your last jail, so once you unjailed, you still will see that jail info you had when you where jailed last time. it is not a way to check if you are jailed now or not.
To check if you are jailed not or not it is neede to check if your validator is in acve validator list or not using this command:
```aidl
pointd query tendermint-validator-set | grep "$(pointd tendermint show-address)"
```
if it's output is something like output below that means you are in active validator set already and no need to unjail.
```aidl
address: evmosvalcons1qae8r355w5cryez6klj2d47nlddlk0k6n4nt8m
index_offset: "2"
jailed_until: "2022-08-19T17:02:30.543786843Z"
missed_blocks_counter: "2"
start_height: "203923"
tombstoned: false
```

If it show empty output it may mean that you need to unjail your validator and check again.

Run the unjail command below replacing text placeholders in brackets with your values:
```aidl
pointd tx slashing unjail --from=<validatorkey> --chain-id=point_10721-1 --gas-prices=0.025apoint
```

The check again if you are in active validator set or not using the command below:
```aidl
pointd query tendermint-validator-set | grep "$(pointd tendermint show-address)"
```
  
than run pointd query tendermint-validator-set | grep "$(pointd tendermint show-address)" and if it shows correct data then everything is fine and validator is working again. But if you run pointd query slashing signing-info $(pointd tendermint show-validator) command again it will always show your previous jail even if it is ended already. It is just a historical record.