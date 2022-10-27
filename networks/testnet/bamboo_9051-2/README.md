# Gentxs submission process

We introduce step by step guidance to submit gentxs for acrechain testnet `bamboo_9051-2`.

## How to create gentx

Install the binary:

```shell
git clone https://github.com/ArableProtocol/acrechain
cd acrechain
git checkout v1.0.0
make install
```

Verify the installation:

```shell
acred version
# v1.0.0
```

Init the chain:

```shell
acred init <moniker> --chain-id=bamboo_9051-2
```

Add your validator key:

```shell
acred keys add <YOUR_KEY>
```

Add genesis account:

```shell
# add 110 ACRE to the genesis
acred add-genesis-account <YOUR_KEY> 110000000000000000000aacre
```

Create the gentx:

```shell
# gentx to put 100 ACRE on the account
acred gentx <YOUR_KEY> 100000000000000000000aacre --moniker="" --min-self-delegation="1000000000000000000" --commission-max-change-rate="0.01" --commission-max-rate="0.20"  --commission-rate=0.05 --website="" --identity="" --security-contact="" --details="" --chain-id=bamboo_9051-2
```

## Note:

1. Save `<YOUR_KEY>` seed phrase and `priv_validator_key.json` from the `.acred/config` folder, in a secure place.
2. Do not add more than 110 ACRE on genesis account.

## Push the GenTx generated to the repository

1. Fork `acrechain` repo and clone
2. Copy `$HOME/.acred/config/gentx/gentx-<xxxxx>.json` to `<repo>/networks/testnet/bamboo_9051-2/gentx/<moniker>.json`
3. Create PR into the repo
