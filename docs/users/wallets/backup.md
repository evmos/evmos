<!--
order: 4
-->

# Backup

Below, it's detailed how to backup your wallet with [evmosd](../../validators/quickstart/binary.md).

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

$ evmosd keys list
[
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

## Export

To backup a local key without the mnemonic phrase, do the following:

```bash
evmosd keys export mykey
Enter passphrase to decrypt your key:
Enter passphrase to encrypt the exported key:
-----BEGIN TENDERMINT PRIVATE KEY-----
kdf: bcrypt
salt: 14559BB13D881A86E0F4D3872B8B2C82
type: secp256k1

3OkvaNgdxSfThr4VoEJMsa/znHmJYm0sDKyyZ+6WMfdzovDD2BVLUXToutY/6iw0
AOOu4v0/1+M6wXs3WUwkKDElHD4MOzSPrM3YYWc=
=JpKI
-----END TENDERMINT PRIVATE KEY-----

$ echo "\
-----BEGIN TENDERMINT PRIVATE KEY-----
kdf: bcrypt
salt: 14559BB13D881A86E0F4D3872B8B2C82
type: secp256k1

3OkvaNgdxSfThr4VoEJMsa/znHmJYm0sDKyyZ+6WMfdzovDD2BVLUXToutY/6iw0
AOOu4v0/1+M6wXs3WUwkKDElHD4MOzSPrM3YYWc=
=JpKI
-----END TENDERMINT PRIVATE KEY-----" > mykey.export
```

To restore the key:

```bash
$ evmosd keys import mykey-imported ./mykey.export
Enter passphrase to decrypt your key:
```

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