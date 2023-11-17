#!/bin/bash

# This scripts applies the corresponding changes
# to run a Stride node for the nix e2e tests.
# This is needed because Stride is a Cosmos Hub consumer chain.
# So we need to update the setup to make it work properly.

# shellcheck source=/dev/null
source .env

set -eu

STRIDE_HOME="$BASE_DIR/node1"
STRIDED="strided --home ${STRIDE_HOME}"

config_toml="${STRIDE_HOME}/config/config.toml"
client_toml="${STRIDE_HOME}/config/client.toml"
app_toml="${STRIDE_HOME}/config/app.toml"
genesis_json="${STRIDE_HOME}/config/genesis.json"

# before removing the node0 folder
# move the config files to reuse them
NODE0_DIR="$BASE_DIR/node0"
tmp_config_toml="/tmp/config.toml"
tmp_app_toml="/tmp/app.toml"
tmp_client_toml="/tmp/client.toml"
mv "$NODE0_DIR"/config/config.toml ${tmp_config_toml}
mv "$NODE0_DIR"/config/app.toml ${tmp_app_toml}
mv "$NODE0_DIR"/config/client.toml ${tmp_client_toml}

# remove node0 - only use node1
# but keep node0 configs (ports mostly)
rm -rf "${NODE0_DIR}"

# use the generated config files
mv ${tmp_config_toml} "${config_toml}"
mv ${tmp_app_toml} "${app_toml}"
mv ${tmp_client_toml} "${client_toml}"

# add consumer section to run a stand-alone node
$STRIDED add-consumer-section 1

# remove the genesis txs (no validators in consumer chain)
jq '.app_state.genutil.gen_txs = []' "$genesis_json" >json.tmp && mv json.tmp "$genesis_json"

# unbonding period should match staking unbonding time
unbonding_time=$(jq -r '.app_state.staking.params.unbonding_time' <"$genesis_json")
jq '.app_state.ccvconsumer.params.unbonding_period = $unbonding_time' --arg unbonding_time "$unbonding_time" "$genesis_json" >json.tmp && mv json.tmp "$genesis_json"

# add keys to the keyring
echo "$COMMUNITY_MNEMONIC" | $STRIDED keys add community --recover --keyring-backend=test
echo "$SIGNER1_MNEMONIC" | $STRIDED keys add signer1 --recover --keyring-backend=test
echo "$SIGNER2_MNEMONIC" | $STRIDED keys add signer2 --recover --keyring-backend=test

# Update the tasks.ini file to run only one node
new_content="[program:stride-1-node]
autostart = true
autorestart = true
redirect_stderr = true
startsecs = 3
directory = %(here)s/node1
command = strided start --home . 
stdout_logfile = %(here)s/node.log"

# Replace the content of tasks.ini with the new content
echo "$new_content" >"$BASE_DIR/tasks.ini"
