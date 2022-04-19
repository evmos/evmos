<!--
order: 3
-->

# Multisig

Learn how to generate, sign and broadcast a transaction using the keyring multisig {synopsis}

A **multisig account** is an Evmos account with a special key that can require more than one signature to sign transactions. This can be useful for increasing the security of the account or for requiring the consent of multiple parties to make transactions. Multisig accounts can be created by specifying:

- threshold number of signatures required
- the public keys involved in signing

To sign with a multisig account, the transaction must be signed individually by the different keys specified for the account. Then, the signatures will be combined into a multisignature which can be used to sign the transaction. If fewer than the threshold number of signatures needed are present, the resultant multisignature is considered invalid.

## Generate a Multisig key

```bash
evmosd keys add --multisig=name1,name2,name3[...] --multisig-threshold=K new_key_name
```

`K` is the minimum number of private keys that must have signed the transactions that carry the public key's address as signer.

The `--multisig` flag must contain the name of public keys that will be combined into a public key that will be generated and stored as `new_key_name` in the local database. All names supplied through `--multisig` must already exist in the local database.

Unless the flag `--nosort` is set, the order in which the keys are supplied on the command line does not matter, i.e. the following commands generate two identical keys:

```bash
evmosd keys add --multisig=p1,p2,p3 --multisig-threshold=2 multisig_address
evmosd keys add --multisig=p2,p3,p1 --multisig-threshold=2 multisig_address
```

Multisig addresses can also be generated on-the-fly and printed through the which command:

```bash
evmosd keys show --multisig-threshold=K name1 name2 name3 [...]
```

## Signing a transaction

### Step 1: Create the multisig key

Let's assume that you have `test1` and `test2` want to make a multisig account with `test3`.

First import the public keys of `test3` into your keyring.

```sh
evmosd keys add \
    test3 \
    --pubkey=evmospub1addwnpepqgcxazmq6wgt2j4rdfumsfwla0zfk8e5sws3p3zg5dkm9007hmfysxas0u2
```

Generate the multisig key with 2/3 threshold.

```sh
evmosd keys add \
    multi \
    --multisig=test1,test2,test3 \
    --multisig-threshold=2
```

You can see its address and details:

```sh
evmosd keys show multi

- name: multi
  type: multi
  address: evmos1e0fx0q9meawrcq7fmma9x60gk35lpr4xk3884m
  pubkey: evmospub1ytql0csgqgfzd666axrjzq3mxw59ys6yqcd3ydjvhgs0uzs6kdk5fp4t73gmkl8t6y02yfq7tvfzd666axrjzq3sd69kp5usk492x6nehqjal67ynv0nfqapzrzy3gmdk27la0kjfqfzd666axrjzq6utqt639ka2j3xkncgk65dup06t297ccljmxhvhu3rmk92u3afjuyz9dg9
  mnemonic: ""
  threshold: 0
  pubkeys: []
```

Let's add 10 EVMOS to the multisig wallet:

```bash
evmosd tx bank send \
    test1 \
    evmos1e0fx0q9meawrcq7fmma9x60gk35lpr4xk3884m \
    10000000000000000000aevmos \
    --chain-id=evmos_9000-4 \
    --gas=auto \
    --fees=1000000aevmos \
    --broadcast-mode=block
```

### Step 2: Create the multisig transaction

We want to send 5 EVMOS from our multisig account to `evmos1rgjxswhuxhcrhmyxlval0qa70vxwvqn2e0srft`.

```bash
evmosd tx bank send \
    evmos1rgjxswhuxhcrhmyxlval0qa70vxwvqn2e0srft \
    evmos157g6rn6t6k5rl0dl57zha2wx72t633axqyvvwq \
    5000000000000000000aevmos \
    --gas=200000 \
    --fees=1000000aevmos \
    --chain-id=evmos_9000-4 \
    --generate-only > unsignedTx.json
```

The file `unsignedTx.json` contains the unsigned transaction encoded in JSON.

```json
{
  "body": {
    "messages": [
      {
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from_address": "evmos1rgjxswhuxhcrhmyxlval0qa70vxwvqn2e0srft",
        "to_address": "evmos157g6rn6t6k5rl0dl57zha2wx72t633axqyvvwq",
        "amount": [
          {
            "denom": "aevmos",
            "amount": "5000000000000000000"
          }
        ]
      }
    ],
    "memo": "",
    "timeout_height": "0",
    "extension_options": [],
    "non_critical_extension_options": []
  },
  "auth_info": {
    "signer_infos": [],
    "fee": {
      "amount": [
        {
          "denom": "aevmos",
          "amount": "1000000"
        }
      ],
      "gas_limit": "200000",
      "payer": "",
      "granter": ""
    }
  },
  "signatures": []
}
```

### Step 3: Sign individually

Sign with `test1` and `test2` and create individual signatures.

```sh
evmosd tx sign \
    unsignedTx.json \
    --multisig=evmos1e0fx0q9meawrcq7fmma9x60gk35lpr4xk3884m \
    --from=test1 \
    --output-document=test1sig.json \
    --chain-id=evmos_9000-4
```

```sh
evmosd tx sign \
    unsignedTx.json \
    --multisig=evmos1e0fx0q9meawrcq7fmma9x60gk35lpr4xk3884m \
    --from=test2 \
    --output-document=test2sig.json \
    --chain-id=evmos_9000-4
```

### Step 4: Create multisignature

Combine signatures to sign transaction.

```sh
evmosd tx multisign \
    unsignedTx.json \
    multi \
    test1sig.json test2sig.json \
    --output-document=signedTx.json \
    --chain-id=evmos_9000-4
```

The TX is now signed:

```json
{
  "body": {
    "messages": [
      {
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from_address": "evmos1rgjxswhuxhcrhmyxlval0qa70vxwvqn2e0srft",
        "to_address": "evmos157g6rn6t6k5rl0dl57zha2wx72t633axqyvvwq",
        "amount": [
          {
            "denom": "aevmos",
            "amount": "5000000000000000000"
          }
        ]
      }
    ],
    "memo": "",
    "timeout_height": "0",
    "extension_options": [],
    "non_critical_extension_options": []
  },
  "auth_info": {
    "signer_infos": [
      {
        "public_key": {
          "@type": "/cosmos.crypto.multisig.LegacyAminoPubKey",
          "threshold": 2,
          "public_keys": [
            {
              "@type": "/cosmos.crypto.secp256k1.PubKey",
              "key": "ApCzSG8k7Tr4aM6e4OJRExN7cNtvH21L9azbh+uRrvt4"
            },
            {
              "@type": "/cosmos.crypto.secp256k1.PubKey",
              "key": "Ah91erz8ChNanqLe9ea948rvAiXMCRlR5Ka7EE/c0xUK"
            },
            {
              "@type": "/cosmos.crypto.secp256k1.PubKey",
              "key": "A0OjtIUCFJM3AobJ9HJTWKP9RZV2+WPcwVjLgsAidrZ/"
            }
          ]
        },
        "mode_info": {
          "multi": {
            "bitarray": {
              "extra_bits_stored": 3,
              "elems": "wA=="
            },
            "mode_infos": [
              {
                "single": {
                  "mode": "SIGN_MODE_LEGACY_AMINO_JSON"
                }
              },
              {
                "single": {
                  "mode": "SIGN_MODE_LEGACY_AMINO_JSON"
                }
              }
            ]
          }
        },
        "sequence": "1"
      }
    ],
    "fee": {
      "amount": [
        {
          "denom": "aevmos",
          "amount": "1000000"
        }
      ],
      "gas_limit": "200000",
      "payer": "",
      "granter": ""
    }
  },
  "signatures": [
    "CkCEeIbeGc+I1ipZuhp/0KhVNnWAv2tTlvgo5x61lzk1KHmLPV38m/YFurrFt5cm5+fqIXrn+FlOjrJuzBhw8ogYCkCawm9mpXsBHk0CFsE5618fVnvScEkfrzW0c2jCcjqV8EPuj3ut74UWzZyQkwtJGxUWtro9EgnGsB7Di1Gzizst"
  ]
}
```

### Step 5: Broadcast transaction

```sh
evmosd tx broadcast signedTx.json \
    --chain-id=evmos_9000-4 \
    --broadcast-mode=block
```
