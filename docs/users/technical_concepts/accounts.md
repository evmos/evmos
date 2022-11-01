<!--
order: 3
-->

# Accounts

This document describes the in-built accounts system of Evoblock. {synopsis}

## Pre-requisite Readings

- [Cosmos SDK Accounts](https://docs.cosmos.network/main/basics/accounts.html) {prereq}
- [Ethereum Accounts](https://ethereum.org/en/whitepaper/#ethereum-accounts) {prereq}

## Evoblock Accounts

Evoblock defines its own custom `Account` type that uses Ethereum's ECDSA secp256k1 curve for keys. This
satisfies the [EIP84](https://github.com/ethereum/EIPs/issues/84) for full [BIP44](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki) paths.
The root HD path for Evoblock-based accounts is `m/44'/60'/0'/0`.

+++ https://github.com/evoblockchain/ethermint/blob/main/types/account.pb.go#L28-L33

## Addresses and Public Keys

[BIP-0173](https://github.com/satoshilabs/slips/blob/master/slip-0173.md) defines a new format for segregated witness output addresses that contains a human-readable part that identifies the Bech32 usage. Evoblock uses the following HRP (human readable prefix) as the base HRP:

| Network   | Mainnet | Testnet |
|-----------|---------|---------|
| Evoblock     | `evoblock` | `evoblock` |

There are 3 main types of HRP for the `Addresses`/`PubKeys` available by default on Evoblock:

- Addresses and Keys for **accounts**, which identify users (e.g. the sender of a `message`). They are derived using the **`eth_secp256k1`** curve.
- Addresses and Keys for **validator operators**, which identify the operators of validators. They are derived using the **`eth_secp256k1`** curve.
- Addresses and Keys for **consensus nodes**, which identify the validator nodes participating in consensus. They are derived using the **`ed25519`** curve.

|                    | Address bech32 Prefix | Pubkey bech32 Prefix | Curve           | Address byte length | Pubkey byte length |
|--------------------|-----------------------|----------------------|-----------------|---------------------|--------------------|
| Accounts           | `evoblock`               | `evopub`           | `eth_secp256k1` | `20`                | `33` (compressed)  |
| Validator Operator | `evovaloper`        | `evovaloperpub`    | `eth_secp256k1` | `20`                | `33` (compressed)  |
| Consensus Nodes    | `evovalcons`        | `evovalconspub`    | `ed25519`       | `20`                | `32`               |

## Address formats for clients

`EthAccount` can be represented in both [Bech32](https://en.bitcoin.it/wiki/Bech32) (`evo1...`) and hex (`0x...`) formats for Ethereum's Web3 tooling compatibility.

The Bech32 format is the default format for Cosmos-SDK queries and transactions through CLI and REST
clients. The hex format on the other hand, is the Ethereum `common.Address` representation of a
Cosmos `sdk.AccAddress`.

- **Address (Bech32)**: `evo1z3t55m0l9h0eupuz3dp5t5cypyv674jj7mz2jw`
- **Address ([EIP55](https://eips.ethereum.org/EIPS/eip-55) Hex)**: `0x91defC7fE5603DFA8CC9B655cF5772459BF10c6f`
- **Compressed Public Key**: `{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"AsV5oddeB+hkByIJo/4lZiVUgXTzNfBPKC73cZ4K1YD2"}`

### Address conversion

The `evoblockd debug addr <address>` can be used to convert an address between hex and bech32 formats. For example:

:::: tabs
::: tab Bech32

```bash
evoblockd debug addr evo1z3t55m0l9h0eupuz3dp5t5cypyv674jj7mz2jw
  Address: [20 87 74 109 255 45 223 158 7 130 139 67 69 211 4 9 25 175 86 82]
  Address (hex): 14574A6DFF2DDF9E07828B4345D3040919AF5652
  Bech32 Acc: evo1z3t55m0l9h0eupuz3dp5t5cypyv674jj7mz2jw
  Bech32 Val: evovaloper1z3t55m0l9h0eupuz3dp5t5cypyv674jjn4d6nn
```

:::
::: tab Hex

```bash
evoblockd debug addr 14574A6DFF2DDF9E07828B4345D3040919AF5652
  Address: [20 87 74 109 255 45 223 158 7 130 139 67 69 211 4 9 25 175 86 82]
  Address (hex): 14574A6DFF2DDF9E07828B4345D3040919AF5652
  Bech32 Acc: evo1z3t55m0l9h0eupuz3dp5t5cypyv674jj7mz2jw
  Bech32 Val: evovaloper1z3t55m0l9h0eupuz3dp5t5cypyv674jjn4d6nn
```

:::
::::

### Key output

::: tip
The Cosmos SDK Keyring output (i.e `evoblockd keys`) only supports addresses and public keys in Bech32 format.
:::

We can use the `keys show` command of `evoblockd` with the flag `--bech <type> (acc|val|cons)` to
obtain the addresses and keys as mentioned above,

:::: tabs
::: tab Account

```bash
evoblockd keys show mykey --bech acc
- name: mykey
  type: local
  address: evo1z3t55m0l9h0eupuz3dp5t5cypyv674jj7mz2jw
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"AsV5oddeB+hkByIJo/4lZiVUgXTzNfBPKC73cZ4K1YD2"}'
  mnemonic: ""
```

:::
::: tab Validator

```bash
evoblockd keys show mykey --bech val
- name: mykey
  type: local
  address: evovaloper1z3t55m0l9h0eupuz3dp5t5cypyv674jjn4d6nn
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"AsV5oddeB+hkByIJo/4lZiVUgXTzNfBPKC73cZ4K1YD2"}'
  mnemonic: ""
```

:::
::: tab Consensus

```bash
evoblockd keys show mykey --bech cons
- name: mykey
  type: local
  address: evovalcons1rllqa5d97n6zyjhy6cnscc7zu30zjn3f7wyj2n
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"A/fVLgIqiLykFQxum96JkSOoTemrXD0tFaFQ1B0cpB2c"}'
  mnemonic: ""
```

:::
::::

## Querying an Account

You can query an account address using the CLI, gRPC or

### Command Line Interface

```bash
# NOTE: the --output (-o) flag will define the output format in JSON or YAML (text)
evoblockd q auth account $(evoblockd keys show mykey -a) -o text
|
  '@type': /ethermint.types.v1.EthAccount
  base_account:
    account_number: "0"
    address: evo1z3t55m0l9h0eupuz3dp5t5cypyv674jj7mz2jw
    pub_key:
      '@type': /ethermint.crypto.v1.ethsecp256k1.PubKey
      key: AsV5oddeB+hkByIJo/4lZiVUgXTzNfBPKC73cZ4K1YD2
    sequence: "1"
  code_hash: 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
```

### Cosmos gRPC and REST

``` bash
# GET /cosmos/auth/v1beta1/accounts/{address}
curl -X GET "http://localhost:10337/cosmos/auth/v1beta1/accounts/evo14au322k9munkmx5wrchz9q30juf5wjgz2cfqku" -H "accept: application/json"
```

### JSON-RPC

To retrieve the Ethereum hex address using Web3, use the JSON-RPC [`eth_accounts`](./../../developers/json-rpc/endpoints.md#eth-accounts) or [`personal_listAccounts`](./../../developers/json-rpc/endpoints.md#personal-listAccounts) endpoints:

```bash
# query against a local node
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

curl -X POST --data '{"jsonrpc":"2.0","method":"personal_listAccounts","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545
```
