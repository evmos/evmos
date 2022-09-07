# Point Validator FAQS

## Table of Contents

* [My transaction failed because of sequence error](#my-transaction-failed-because-of-sequence-error)
* [Check if node is synced](#check-if-node-is-synced)
* [Verify transactions](#verify-transactions)
* [Check if I am a validator](#check-if-I-am-a-validator)
* [Check my voting power](#check-my-voting-power)
* [Kind of keys](#kind-of-keys)
* [Get point address for a given key name](#get-point-address-for-a-given-key-name)
* [Get pointvaloper address for a given key name](#get-pointvaloper-address-for-a-given-key-name)
* [Check balance for a given point formatted address](#check-balance-for-a-given-point-formatted-address)
* [Convert between point formatted and ethereum formatted addresses](#convert-between-point-formatted-and-ethereum-formatted-addresses)
* [Get pointvalcons address](#get-pointvalcons-address)
* [Get information for all validators](#get-information-for-all-validators)
* [Get information for you validator providing you pointvaloper address](#get-information-for-you-validator-providing-you-pointvaloper-address)
* [Check if validator is jailed](#check-if-validator-is-jailed)
* [Check if jail has expired and I can unjail](#check-if-jail-has-expired-and-I-can-unjail)
* [How to unjail using the key](#how-to-unjail-using-the-key)
* [Unjail is not working](#unjail-is-not-working)
* [How to delegate more tokens](#how-to-delegate-more-tokens)
* [How to delegate tokens from one validator to another one](#how-to-delegate-tokens-from-one-validator-to-another-one)
* [Do you have a explorer?](#do-you-have-a-explorer?)
* [What do I need to backup for migrating my node to other vps](#what-do-I-need-to-backup-for-migrating-my-node-to-other-vps)
* [How to export my private key to import in metamask](#how-to-export-my-private-key-to-import-in-metamask)
* [How to recover a key using seeds](#how-to-recover-a-key-using-seeds)

## My transaction failed because of sequence error

If in the output of your transaction raw log field you see an error like this

raw_log: 'account sequence mismatch, expected 1, got 0: incorrect account sequence'

you need to add --sequence flag to your command.

for the error above it's waiting for sequence 1 so you need to add flag

```
--sequence 1
```

## Check if node is synced

```
pointd status 2>&1 | jq .SyncInfo | grep catching_up
```

If response says "catching_up": false it means you are synced.

## Verify transactions

```
pointd query tx <tx-id>
```

## Check if I am a validator

```
pointd query slashing signing-info $(pointd tendermint show-validator)
```

Also you should find your address here:

```
pointd query tendermint-validator-set
```

you can use grep to see if your pointvalcons address is part of active validators:

```
pointd query tendermint-validator-set | grep "$(pointd tendermint show-address)"
```

## Check my voting power

```
pointd status | jq .ValidatorInfo.VotingPower
```

## Kind of keys

### Tendermint Key

This is a unique key used to sign block hashes. It is associated with a public key pointvalconspub when you create your validator.
This key is saved in file ~/.pointd/config/priv_validator_key.json (backup this file if you plan to move the node to other vps)
You can see information for this key using.

To see public validator key
```
pointd tendermint show-validator
```

and also to see validator address (pointvalcons format)

```
pointd tendermint show-address
```

When you run the command to see current validators you will see a list with each validator

- address: pointvalcons1wqkgnazus8m7r6jcjkqshcmxq2qq9hcxjt437c
  proposer_priority: "23635"
  pub_key:
    type: tendermint/PubKeyEd25519
    value: RmLNBEb6FhvUVkuAqBg8B7qufHc2uhxbxqcLyu8gNPc=
  voting_power: "197"

As you can see here validators are uniquely identify using this address and public key.

When you run the command to see validators config

```
pointd query staking validators
```

the output returns an array with commisions per validator
```
- commission:
    commission_rates:
      max_change_rate: "0.010000000000000000"
      max_rate: "0.200000000000000000"
      rate: "0.100000000000000000"
    update_time: "2022-08-19T02:44:54.634555449Z"
  consensus_pubkey:
    '@type': /cosmos.crypto.ed25519.PubKey
    key: RmLNBEb6FhvUVkuAqBg8B7qufHc2uhxbxqcLyu8gNPc=
  delegator_shares: "203060714212098013253.632897042151911779"
  description:
    details: ""
    identity: ""
    moniker: supervalidator
    security_contact: ""
    website: ""
  jailed: false
  min_self_delegation: "1"
  operator_address: pointvaloper1arflh3r9cm8amy3tdlhtu79ywhq7gpwdnqlgfg
  status: BOND_STATUS_BONDED
  tokens: "197029810999998702260"
  unbonding_height: "211800"
  unbonding_time: "2022-09-09T17:02:49.643431493Z"
```
  Here you can see there are a relationship between tendermint public key and the operation_address (pointvaloper) used for staking the funds.
  That operation address is related to your application key.

### Application keys
These keys are created from the application and used to sign transactions. As a validator, you will probably use one key to sign staking-related transactions, and another key to sign oracle-related transactions. Application keys are associated with a public key pointpub- and an address point-. Both are derived from account keys generated by pointd keys add.

When you create a validator you associate this key with your validator public key. See the command below and pay attention to flags --pubkey and --from

```
pointd tx staking create-validator  \
--amount=100000000000000000000apoint \
--pubkey=$(pointd tendermint show-validator) \
--moniker="yourmoniker" \
--chain-id=point_10687-1 \
--commission-rate="0.10" \
--commission-max-rate="0.20" \
--commission-max-change-rate="0.01" \
--min-self-delegation="1" \
--gas="400000" \
--gas-prices="0.025apoint" \
--from=<key-name> \
--keyring-backend file
```

For each key, you have a pointvaloper that you can get using: [Get pointvaloper address](#get-pointvaloper-address-for-a-given-key-name)
Also you can get point address using: [Get point address](#get-point-address-for-a-given-key-name)

## Get point address for a given key name

```
pointd keys show <key-name>
```

## Get pointvaloper address for a given key name

```
pointd keys show <key-name> -a --bech val
```


## Check balance for a given point formatted address

```
pointd query bank balances <point formated address>
```

also you can see balances for a given key

```
pointd query bank balances $(pointd keys show <key-name> -a)
```


## Convert between point formatted and ethereum formatted addresses

Use this online tool:

https://pointnetwork.io/converter.html


## Get pointvalcons address

```
pointd tendermint show-address
```

## Get information for all validators
```
pointd query staking validators
```

## Get information for you validator providing you pointvaloper address

```
pointd query staking validator <pointvaloperaddress>
```

or you can try providing your key name

```
pointd query staking validator  $(pointd keys show <key-name> -a --bech val)

```

## Check if validator is jailed

```
pointd query staking validator  $(pointd keys show <key-name> -a --bech val) | grep jailed
```

## Check if jail has expired and I can unjail

Run this to see when you can unjail:

```
pointd query slashing signing-info $(pointd tendermint show-validator) | grep jailed_until
```

And run this to see current utc time:

```
date -u +"%Y-%m-%dT%H:%M:%SZ"
```

## How to unjail using the key

```
pointd tx slashing unjail \
--from=<key-name> \
--chain-id=point_10687-1 \
--keyring-backend file \
--gas="400000" \
--gas-prices="0.025apoint"
```

## Unjail is not working

Check if unjail period has expired: [Check if jail has expired and I can unjail](#check-if-jail-has-expired-and-I-can-unjail)
If it's ok check if you have enough balance to unjail yourself: [Get information for you validator providing you pointvaloper address](#get-information-for-you-validator-providing-you-pointvaloper-address)
Output will be something like this:

```
- commission:
    commission_rates:
      max_change_rate: "0.010000000000000000"
  ...
  jailed: true
  min_self_delegation: "10000000000000000"
  ...
  tokens: "9000000000000000"
```

If tokens amount is smaller than min_self_delegation then you cannot unajail.
You need to delegate more tokens: [How to delegate more tokens](#how-to-delegate-more-tokens)

Once you've delegated more tokens check again, if tokens amount is bigger than min_self_delegation amount then run the unjail command again: [How to unjail using the key](#how-to-unjail-using-the-key)

Check jailing status: [Check if validator is jailed](#check-if-validator-is-jailed)

## How to delegate more tokens

First check your available balance for the key

```
pointd query bank balances $(pointd keys show <key-name> -a)
```

this is command supposing you have 100000000apoint to delegate, adjust for you use case and replace <key-name> with your key name.

```
pointd tx staking delegate <pointvaloperaddress> "100000000apoint" \
--from <key-name> \
--keyring-backend file \
--chain-id=point_10687-1 \
--gas="400000" \
--gas-prices="0.025apoint"
```

## How to delegate tokens from one validator to another one

Here we are delegating 900000000000000000000apoint.
To see how much you can delegate run command [See validator info](#get-information-for-you-validator-providing-you-pointvaloper-address)
In tokens section you will see the max amount you can delegate to other validator

```
pointd tx staking redelegate <pointvaloper-source> <pointvaloper-dest> "900000000000000000000apoint" --gas="400000" --gas-prices="0.025apoint" --from=<key-related-to-source-validator> --keyring-backend file
```

## Do you have a explorer?

Yes, go to https://explorer-xnet-triton.point.space


## What do I need to backup for migrating my node to other vps

In the folder ~/.pointd/config you will find a file called priv_validator_key.json generated when the node is created with pointd init, this is the : [tendermint key](#tendermint-key)





## How to export my private key to import in metamask

Don't share the output of this command, it's your private key

```
pointd keys unsafe-export-eth-key <key-name> --keyring-backend file
```

## How to recover a key using seeds

```
pointd keys add <key-name> --keyring-backend file --recover
```
