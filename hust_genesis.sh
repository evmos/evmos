#!/bin/bash
KEYS[0]="hust0"

echo "CHAINID"
read -r CHAINID

echo "MONIKER"
read -r MONIKER

CHAINID="$CHAINID"
MONIKER="$MONIKER"
# Remember to change to other types of keyring like 'file' in-case exposing to outside world,
# otherwise your balance will be wiped quickly
# The keyring test does not require private key to steal tokens from you
KEYRING="file"
KEYALGO="eth_secp256k1"
LOGLEVEL="info"
# Set dedicated home directory for the hustd instance
HOMEDIR="$HOME/.tmp-hust-hustd"
# to trace evm
#TRACE="--trace"
TRACE=""

# Path variables
CONFIG=$HOMEDIR/config/config.toml
APP_TOML=$HOMEDIR/config/app.toml
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

# denominations
GOV_TOKEN="abkc" 
EVM_TOKEN="avnd"

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

# used to exit on first error (any non-zero exit code)
set -e

# Reinstall daemon
make install

# User prompt if an existing local node configuration is found.
if [ -d "$HOMEDIR" ]; then
	printf "\nAn existing folder at '%s' was found. You can choose to delete this folder and start a new local node with new keys from genesis. When declined, the existing local node is started. \n" "$HOMEDIR"
	echo "Overwrite the existing configuration and start a new local node? [y/n]"
	read -r overwrite
else
	overwrite="Y"
fi


# Setup local node if overwrite is set to Yes, otherwise skip setup
if [[ $overwrite == "y" || $overwrite == "Y" ]]; then
	# Remove the previous folder
	rm -rf "$HOMEDIR"

	# Set client config
	hustd config keyring-backend $KEYRING --home "$HOMEDIR"
	hustd config chain-id $CHAINID --home "$HOMEDIR"

	# If keys exist they should be deleted
	for KEY in "${KEYS[@]}"; do
		hustd keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --home "$HOMEDIR"
	done

	# Set moniker and chain-id for Evmos (Moniker can be anything, chain-id must be an integer)
	hustd init $MONIKER -o --chain-id $CHAINID --home "$HOMEDIR"

	# Change parameter token denominations to GOV_TOKEN 
	jq -r --arg GOV_TOKEN "$GOV_TOKEN" '.app_state["staking"]["params"]["bond_denom"]=$GOV_TOKEN' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq -r --arg GOV_TOKEN "$GOV_TOKEN" '.app_state["crisis"]["constant_fee"]["denom"]=$GOV_TOKEN' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq -r --arg GOV_TOKEN "$GOV_TOKEN" '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]=$GOV_TOKEN' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq -r --arg EVM_TOKEN "$EVM_TOKEN" '.app_state["evm"]["params"]["evm_denom"]=$EVM_TOKEN' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq -r --arg GOV_TOKEN "$GOV_TOKEN" '.app_state["inflation"]["params"]["mint_denom"]=$GOV_TOKEN' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Set gas limit in genesis
	jq '.consensus_params["block"]["max_gas"]="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Set claims start time
	current_date=$(date -u +"%Y-%m-%dT%TZ")
	jq -r --arg current_date "$current_date" '.app_state["claims"]["params"]["airdrop_start_time"]=$current_date' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Set claims records for validator account
	amount_to_claim=10000
	# claims_key=${KEYS[0]}
	# node_address=$(evmosd keys show $claims_key --keyring-backend $KEYRING --home "$HOMEDIR" | grep "address" | cut -c12-)
	# jq -r --arg node_address "$node_address" --arg amount_to_claim "$amount_to_claim" '.app_state["claims"]["claims_records"]=[{"initial_claimable_amount":$amount_to_claim, "actions_completed":[false, false, false, false],"address":$node_address}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Set claims decay
	jq '.app_state["claims"]["params"]["duration_of_decay"]="1000000s"' >"$TMP_GENESIS" "$GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state["claims"]["params"]["duration_until_decay"]="100000s"' >"$TMP_GENESIS" "$GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  
  # Set claim denom
	jq -r --arg GOV_TOKEN "$GOV_TOKEN" '.app_state["claims"]["params"]["claims_denom"]=$GOV_TOKEN' >"$TMP_GENESIS" "$GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  
	# # Claim module account:
	# # 0xA61808Fe40fEb8B3433778BBC2ecECCAA47c8c47 || evmos15cvq3ljql6utxseh0zau9m8ve2j8erz89m5wkz
	# jq -r --arg amount_to_claim "$amount_to_claim" --arg GOV_TOKEN "$GOV_TOKEN" '.app_state["bank"]["balances"] += [{"address":"evmos15cvq3ljql6utxseh0zau9m8ve2j8erz89m5wkz","coins":[{"denom":$GOV_TOKEN, "amount":$amount_to_claim}]}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	if [[ $1 == "pending" ]]; then
		if [[ "$OSTYPE" == "darwin"* ]]; then
			sed -i '' 's/timeout_propose = "3s"/timeout_propose = "30s"/g' "$CONFIG"
			sed -i '' 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' "$CONFIG"
			sed -i '' 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' "$CONFIG"
			sed -i '' 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' "$CONFIG"
			sed -i '' 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' "$CONFIG"
			sed -i '' 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' "$CONFIG"
			sed -i '' 's/timeout_commit = "5s"/timeout_commit = "150s"/g' "$CONFIG"
			sed -i '' 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' "$CONFIG"
		else
			sed -i 's/timeout_propose = "3s"/timeout_propose = "30s"/g' "$CONFIG"
			sed -i 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' "$CONFIG"
			sed -i 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' "$CONFIG"
			sed -i 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' "$CONFIG"
			sed -i 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' "$CONFIG"
			sed -i 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' "$CONFIG"
			sed -i 's/timeout_commit = "5s"/timeout_commit = "150s"/g' "$CONFIG"
			sed -i 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' "$CONFIG"
		fi
	fi

    # enable prometheus metrics
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' 's/prometheus = false/prometheus = true/' "$CONFIG"
        sed -i '' 's/prometheus-retention-time = 0/prometheus-retention-time  = 1000000000000/g' "$APP_TOML"
        sed -i '' 's/enabled = false/enabled = true/g' "$APP_TOML"
    else
        sed -i 's/prometheus = false/prometheus = true/' "$CONFIG"
        sed -i 's/prometheus-retention-time  = "0"/prometheus-retention-time  = "1000000000000"/g' "$APP_TOML"
        sed -i 's/enabled = false/enabled = true/g' "$APP_TOML"
    fi


    printf "\nAn existing folder at '%s' was found. You can choose to delete this folder and start a new local node with new keys from genesis. When declined, the existing local node is started. \n" "$HOMEDIR"
    echo "Overwrite the existing configuration and start a new local node? [[t]est/[p]roduct]"
    read -r type_network
  else
    type_network="P"
  fi

  if [[ $type_network == "P" || $type_network == "p" ]]; then 
    if [[ "$OSTYPE" == "darwin"* ]]; then 
      sed -i '' 's/127.0.0.1/0.0.0.0/g' "$CONFIG"
      sed -i '' 's/127.0.0.1/0.0.0.0/g' "$APP_TOML"
      sed -i '' 's/enabled-unsafe-cors *= *.*/enabled-unsafe-cors = true/g' "$APP_TOML" 
      sed -i '' 's/cors_allowed_origins *= *.*/cors_allowed_origins = \[\"*\"\]/g' "$CONFIG"
      sed -i '' '/\[api\]/,+3 s/enable = false/enable = true/' "$APP_TOML" 
      sed -i '' 's/aevmos/abkc/g' "$APP_TOML"
    else
      sed -i 's/127.0.0.1/0.0.0.0/g' "$CONFIG"
      sed -i 's/127.0.0.1/0.0.0.0/g' "$APP_TOML"
      sed -i 's/enabled-unsafe-cors *= *.*/enabled-unsafe-cors = true/g' "$APP_TOML" 
      sed -i 's/cors_allowed_origins *= *.*/cors_allowed_origins = \[\"*\"\]/g' "$CONFIG"
      sed -i '/\[api\]/,+3 s/enable = false/enable = true/' "$APP_TOML" 
      sed -i 's/aevmos/abkc' "$APP_TOML"
    fi
  fi
  

	# Allocate genesis accounts (cosmos formatted addresses)
	for KEY in "${KEYS[@]}"; do
		hustd add-genesis-account $KEY "100000000000000000000000000${GOV_TOKEN},100000000000000000000000000${EVM_TOKEN}"  --keyring-backend $KEYRING --home "$HOMEDIR"
	done

	# bc is required to add these big numbers
	total_supply=$(echo "${#KEYS[@]} * 100000000000000000000000000" | bc)
	total_supply_vnd=$(echo "${#KEYS[@]} * 100000000000000000000000000" | bc)
	jq -r --arg total_supply "$total_supply" '.app_state["bank"]["supply"][0]["amount"]=$total_supply' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq -r --arg total_supply "$total_supply_vnd" '.app_state["bank"]["supply"][1]["amount"]=$total_supply' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Sign genesis transaction
	hustd gentx ${KEYS[0]} "1000000000000000000000${GOV_TOKEN}" --keyring-backend $KEYRING --chain-id $CHAINID --home "$HOMEDIR"
	## In case you want to create multiple validators at genesis
	## 1. Back to `evmosd keys add` step, init more keys
	## 2. Back to `evmosd add-genesis-account` step, add balance for those
	## 3. Clone this ~/.evmosd home directory into some others, let's say `~/.clonedEvmosd`
	## 4. Run `gentx` in each of those folders
	## 5. Copy the `gentx-*` folders under `~/.clonedEvmosd/config/gentx/` folders into the original `~/.evmosd/config/gentx`

	# Collect genesis tx
	hustd collect-gentxs --home "$HOMEDIR"

	# Run this to ensure everything worked and that the genesis file is setup correctly
	hustd validate-genesis --home "$HOMEDIR"
