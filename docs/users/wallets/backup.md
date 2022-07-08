<!--
order: 4
-->

# Backup

Learn how to backup your wallet's mnemonic and private key. {synopsis}

## Mnemonics

When you create a new key, you'll recieve a mnemonic phrase that can be used to restore that key. Backup the mnemonic phrase:

```bash
evmosd keys add mykey
{
  "name": "mykey",
  "type": "local",
  "address": "evmos1n253dl2tgyhxjm592p580c38r4dn8023ctv28d",
  "pubkey": '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"ArJhve4v5HkLm+F7ViASU/rAGx7YrwU4+XKV2MNJt+Cq"}',
  "mnemonic": ""
}

**Important** write this mnemonic phrase in a safe place.
It is the only way to recover your account if you ever forget your password.

# <24 word mnemonic phrase>
```

To restore the key:

```bash
$ evmosd keys add mykey-restored --recover
> Enter your bip39 mnemonic
banner genuine height east ghost oak toward reflect asset marble else explain foster car nest make van divide twice culture announce shuffle net peanut
{
  "name": "mykey-restored",
  "type": "local",
  "address": "evmos1n253dl2tgyhxjm592p580c38r4dn8023ctv28d",
  "pubkey": '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"ArJhve4v5HkLm+F7ViASU/rAGx7YrwU4+XKV2MNJt+Cq"}'
}
```

## Export Key

### Tendermint-Formatted Private Keys

To backup this type of key without the mnemonic phrase, do the following:

```bash
evmosd keys export mykey
Enter passphrase to decrypt your key:
Enter passphrase to encrypt the exported key:
-----BEGIN TENDERMINT PRIVATE KEY-----
kdf: bcrypt
salt: 14559BB13D881A86E0F4D3872B8B2C82
type: secp256k1

# <Tendermint private key>
-----END TENDERMINT PRIVATE KEY-----

$ echo "\
-----BEGIN TENDERMINT PRIVATE KEY-----
kdf: bcrypt
salt: 14559BB13D881A86E0F4D3872B8B2C82
type: secp256k1

# <Tendermint private key>
-----END TENDERMINT PRIVATE KEY-----" > mykey.export
```

### Ethereum-Formatted Private Keys

:::tip
**Note**: These types of keys are MetaMask-compatible.
:::

To backup this type of key without the mnemonic phrase, do the following:

```bash
evmosd keys unsafe-export-eth-key mykey > mykey.export
**WARNING** this is an unsafe way to export your unencrypted private key, are you sure? [y/N]: y
Enter keyring passphrase:
```

## Import Key

### Tendermint-Formatted Private Keys

```bash
$ evmosd keys import mykey-imported ./mykey.export
Enter passphrase to decrypt your key:
```

### Ethereum-Formatted Private Keys

```
$ evmosd keys unsafe-import-eth-key mykey-imported ./mykey.export
Enter passphrase to encrypt your key:
```

### Verification

Verify that your key has been restored using the following command:

```bash
$ evmosd keys list
[
  {
    "name": "mykey-imported",
    "type": "local",
    "address": "evmos1n253dl2tgyhxjm592p580c38r4dn8023ctv28d",
    "pubkey": '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"ArJhve4v5HkLm+F7ViASU/rAGx7YrwU4+XKV2MNJt+Cq"}'
  },
  {
    "name": "mykey-restored",
    "type": "local",
    "address": "evmos1n253dl2tgyhxjm592p580c38r4dn8023ctv28d",
    "pubkey": '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"ArJhve4v5HkLm+F7ViASU/rAGx7YrwU4+XKV2MNJt+Cq"}'
  },
  {
    "name": "mykey",
    "type": "local",
    "address": "evmos1n253dl2tgyhxjm592p580c38r4dn8023ctv28d",
    "pubkey": '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"ArJhve4v5HkLm+F7ViASU/rAGx7YrwU4+XKV2MNJt+Cq"}'
  }
]
```
