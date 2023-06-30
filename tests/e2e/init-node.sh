#!/bin/bash

KEY="mykey"
CHAINID="${CHAIN_ID:-evmos_9000-1}"
MONIKER="localtestnet"
KEYRING="test" # remember to change to other types of keyring like 'file' in-case exposing to outside world, otherwise your balance will be wiped quickly. The keyring test does not require private key to steal tokens from you
KEYALGO="eth_secp256k1" #gitleaks:allow
LOGLEVEL="info"
# to trace evm
#TRACE="--trace"
TRACE=""
PRUNING="default"
#PRUNING="custom"

CHAINDIR="$HOME/.evmosd"
GENESIS="$CHAINDIR/config/genesis.json"
TMP_GENESIS="$CHAINDIR/config/tmp_genesis.json"
APP_TOML="$CHAINDIR/config/app.toml"
CONFIG_TOML="$CHAINDIR/config/config.toml"

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

# used to exit on first error (any non-zero exit code)
set -e

# Set client config
evmosd config keyring-backend "$KEYRING"
evmosd config chain-id "$CHAINID"

# if $KEY exists it should be deleted
evmosd keys add "$KEY" --keyring-backend $KEYRING --algo "$KEYALGO"

# Set moniker and chain-id for Evmos (Moniker can be anything, chain-id must be an integer)
evmosd init "$MONIKER" --chain-id "$CHAINID"

# Change parameter token denominations to aevmos
jq '.app_state.staking.params.bond_denom="aevmos"' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.crisis.constant_fee.denom="aevmos"' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.gov.deposit_params.min_deposit[0].denom="aevmos"' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.evm.params.evm_denom="aevmos"' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.inflation.params.mint_denom="aevmos"' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# set gov proposing && voting period
jq '.app_state.gov.deposit_params.max_deposit_period="30s"' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.gov.voting_params.voting_period="30s"' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set gas limit in genesis
jq '.consensus_params.block.max_gas="10000000"' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set claims start time
node_address=$(evmosd keys list | grep  "address: " | cut -c12-)
current_date=$(date -u +"%Y-%m-%dT%TZ")
jq -r --arg current_date "$current_date" '.app_state.claims.params.airdrop_start_time=$current_date' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set claims records for validator account
amount_to_claim=10000
jq -r --arg node_address "$node_address" --arg amount_to_claim "$amount_to_claim" '.app_state.claims.claims_records=[{"initial_claimable_amount":$amount_to_claim, "actions_completed":[false, false, false, false],"address":$node_address}]' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set claims decay
jq '.app_state.claims.params.duration_of_decay="1000000s"' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.claims.params.duration_until_decay="100000s"' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Claim module account:
# 0xA61808Fe40fEb8B3433778BBC2ecECCAA47c8c47 || evmos15cvq3ljql6utxseh0zau9m8ve2j8erz89m5wkz
jq -r --arg amount_to_claim "$amount_to_claim" '.app_state.bank.balances += [{"address":"evmos15cvq3ljql6utxseh0zau9m8ve2j8erz89m5wkz","coins":[{"denom":"aevmos", "amount":$amount_to_claim}]}]' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# disable produce empty block
sed -i 's/create_empty_blocks = true/create_empty_blocks = false/g' "$CONFIG_TOML"

# Allocate genesis accounts (cosmos formatted addresses)
evmosd add-genesis-account $KEY 100000000000000000000000000aevmos --keyring-backend $KEYRING

# Update total supply with claim values
# Bc is required to add this big numbers
# total_supply=$(bc <<< "$amount_to_claim+$validators_supply")
total_supply=100000000000000000000010000
jq -r --arg total_supply "$total_supply" '.app_state.bank.supply[0].amount=$total_supply' "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# set custom pruning settings
if [ "$PRUNING" = "custom" ]; then
  sed -i 's/pruning = "default"/pruning = "custom"/g' "$APP_TOML"
  sed -i 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
  sed -i 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"
fi

# make sure the localhost IP is 0.0.0.0
sed -i 's/pprof_laddr = "localhost:6060"/pprof_laddr = "0.0.0.0:6060"/g' "$CONFIG_TOML"
sed -i 's/127.0.0.1/0.0.0.0/g' "$APP_TOML"

# Sign genesis transaction
evmosd gentx $KEY 1000000000000000000000aevmos --keyring-backend $KEYRING --chain-id "$CHAINID"
## In case you want to create multiple validators at genesis
## 1. Back to `evmosd keys add` step, init more keys
## 2. Back to `evmosd add-genesis-account` step, add balance for those
## 3. Clone this ~/.evmosd home directory into some others, let's say `~/.clonedEvmosd`
## 4. Run `gentx` in each of those folders
## 5. Copy the `gentx-*` folders under `~/.clonedEvmosd/config/gentx/` folders into the original `~/.evmosd/config/gentx`

# Collect genesis tx
evmosd collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
evmosd validate-genesis

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
evmosd start "$TRACE" --log_level $LOGLEVEL --minimum-gas-prices=0.0001aevmos --json-rpc.enable true --json-rpc.api eth,txpool,personal,net,debug,web3 --api.enable
