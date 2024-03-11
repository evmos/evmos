#!/bin/bash

CHAINID="${CHAIN_ID:-evmos_9000-1}"
MONIKER="localtestnet"
KEYRING="test"          # remember to change to other types of keyring like 'file' in-case exposing to outside world, otherwise your balance will be wiped quickly. The keyring test does not require private key to steal tokens from you
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

# feemarket params basefee
BASEFEE=1000000000

# myKey address 0x7cb61d4117ae31a12e393a1cfa3bac666481d02e
VAL_KEY="mykey"
VAL_MNEMONIC="gesture inject test cycle original hollow east ridge hen combine junk child bacon zero hope comfort vacuum milk pitch cage oppose unhappy lunar seat"

# user1 address 0xc6fe5d33615a1c52c08018c47e8bc53646a0e101
USER1_KEY="user1"
USER1_MNEMONIC="copper push brief egg scan entry inform record adjust fossil boss egg comic alien upon aspect dry avoid interest fury window hint race symptom"

# user2 address 0x963ebdf2e1f8db8707d05fc75bfeffba1b5bac17
USER2_KEY="user2"
USER2_MNEMONIC="maximum display century economy unlock van census kite error heart snow filter midnight usage egg venture cash kick motor survey drastic edge muffin visual"

# user3 address 0x40a0cb1C63e026A81B55EE1308586E21eec1eFa9
USER3_KEY="user3"
USER3_MNEMONIC="will wear settle write dance topic tape sea glory hotel oppose rebel client problem era video gossip glide during yard balance cancel file rose"

# user4 address 0x498B5AeC5D439b733dC2F58AB489783A23FB26dA
USER4_KEY="user4"
USER4_MNEMONIC="doll midnight silk carpet brush boring pluck office gown inquiry duck chief aim exit gain never tennis crime fragile ship cloud surface exotic patch"

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
  echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
  exit 1
}

# used to exit on first error (any non-zero exit code)
set -e

# Set client config
evmosd config keyring-backend "$KEYRING"
evmosd config chain-id "$CHAINID"

# Import keys from mnemonics
echo "$VAL_MNEMONIC" | evmosd keys add "$VAL_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO"

# Store the validator address in a variable to use it later
node_address=$(evmosd keys show -a "$VAL_KEY")

echo "$USER1_MNEMONIC" | evmosd keys add "$USER1_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO"
echo "$USER2_MNEMONIC" | evmosd keys add "$USER2_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO"
echo "$USER3_MNEMONIC" | evmosd keys add "$USER3_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO"
echo "$USER4_MNEMONIC" | evmosd keys add "$USER4_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO"

# Set moniker and chain-id for Evmos (Moniker can be anything, chain-id must be an integer)
evmosd init "$MONIKER" --chain-id "$CHAINID"

# Change parameter token denominations to aevmos
jq '.app_state.staking.params.bond_denom="aevmos"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.crisis.constant_fee.denom="aevmos"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.gov.deposit_params.min_deposit[0].denom="aevmos"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.evm.params.evm_denom="aevmos"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.inflation.params.mint_denom="aevmos"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# set gov proposing && voting period
jq '.app_state.gov.deposit_params.max_deposit_period="10s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.gov.voting_params.voting_period="10s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# When upgrade to cosmos-sdk v0.47, use gov.params to edit the deposit params
# check if the 'params' field exists in the genesis file
if jq '.app_state.gov.params != null' "$GENESIS" | grep -q "true"; then
  jq '.app_state.gov.params.min_deposit[0].denom="aevmos"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  jq '.app_state.gov.params.max_deposit_period="10s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  jq '.app_state.gov.params.voting_period="10s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
fi

# Set gas limit in genesis
jq '.consensus_params.block.max_gas="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set claims start time
current_date=$(date -u +"%Y-%m-%dT%TZ")
jq -r --arg current_date "$current_date" '.app_state.claims.params.airdrop_start_time=$current_date' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set claims records for validator account
amount_to_claim=10000
jq -r --arg node_address "$node_address" --arg amount_to_claim "$amount_to_claim" '.app_state.claims.claims_records=[{"initial_claimable_amount":$amount_to_claim, "actions_completed":[false, false, false, false],"address":$node_address}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set claims decay
jq '.app_state.claims.params.duration_of_decay="1000000s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.claims.params.duration_until_decay="100000s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Claim module account:
# 0xA61808Fe40fEb8B3433778BBC2ecECCAA47c8c47 || evmos15cvq3ljql6utxseh0zau9m8ve2j8erz89m5wkz
jq -r --arg amount_to_claim "$amount_to_claim" '.app_state.bank.balances += [{"address":"evmos15cvq3ljql6utxseh0zau9m8ve2j8erz89m5wkz","coins":[{"denom":"aevmos", "amount":$amount_to_claim}]}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set base fee in genesis
jq '.app_state["feemarket"]["params"]["base_fee"]="'${BASEFEE}'"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# disable produce empty block
sed -i.bak 's/create_empty_blocks = true/create_empty_blocks = false/g' "$CONFIG_TOML"

# Allocate genesis accounts (cosmos formatted addresses)
evmosd add-genesis-account "$(evmosd keys show "$VAL_KEY" -a --keyring-backend "$KEYRING")" 100000000000000000000000000aevmos --keyring-backend "$KEYRING"
evmosd add-genesis-account "$(evmosd keys show "$USER1_KEY" -a --keyring-backend "$KEYRING")" 1000000000000000000000aevmos --keyring-backend "$KEYRING"
evmosd add-genesis-account "$(evmosd keys show "$USER2_KEY" -a --keyring-backend "$KEYRING")" 1000000000000000000000aevmos --keyring-backend "$KEYRING"
evmosd add-genesis-account "$(evmosd keys show "$USER3_KEY" -a --keyring-backend "$KEYRING")" 1000000000000000000000aevmos --keyring-backend "$KEYRING"
evmosd add-genesis-account "$(evmosd keys show "$USER4_KEY" -a --keyring-backend "$KEYRING")" 1000000000000000000000aevmos --keyring-backend "$KEYRING"

# Update total supply with claim values
# Bc is required to add this big numbers
# total_supply=$(bc <<< "$amount_to_claim+$validators_supply")
total_supply=100004000000000000000010000
jq -r --arg total_supply "$total_supply" '.app_state.bank.supply[0].amount=$total_supply' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# set custom pruning settings
if [ "$PRUNING" = "custom" ]; then
  sed -i.bak 's/pruning = "default"/pruning = "custom"/g' "$APP_TOML"
  sed -i.bak 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
  sed -i.bak 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"
fi

# make sure the localhost IP is 0.0.0.0
sed -i.bak 's/localhost/0.0.0.0/g' "$CONFIG_TOML"
sed -i.bak 's/127.0.0.1/0.0.0.0/g' "$APP_TOML"

# use timeout_commit 1s to make test faster
sed -i.bak 's/timeout_commit = "3s"/timeout_commit = "1s"/g' "$CONFIG_TOML"

# Sign genesis transaction
evmosd gentx "$VAL_KEY" 1000000000000000000000aevmos --gas-prices ${BASEFEE}aevmos --keyring-backend "$KEYRING" --chain-id "$CHAINID"
## In case you want to create multiple validators at genesis
## 1. Back to `evmosd keys add` step, init more keys
## 2. Back to `evmosd add-genesis-account` step, add balance for those
## 3. Clone this ~/.evmosd home directory into some others, let's say `~/.clonedEvmosd`
## 4. Run `gentx` in each of those folders
## 5. Copy the `gentx-*` folders under `~/.clonedEvmosd/config/gentx/` folders into the original `~/.evmosd/config/gentx`

# Enable the APIs for the tests to be successful
sed -i.bak 's/enable = false/enable = true/g' "$APP_TOML"

# Don't enable memiavl by default
grep -q -F '[memiavl]' "$APP_TOML" && sed -i.bak '/\[memiavl\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"

# Collect genesis tx
evmosd collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
evmosd validate-genesis

# Start the node
evmosd start "$TRACE" \
  --log_level $LOGLEVEL \
  --minimum-gas-prices=0.0001aevmos \
  --json-rpc.api eth,txpool,personal,net,debug,web3 \
  --chain-id "$CHAINID"
