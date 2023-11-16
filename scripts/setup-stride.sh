#!/bin/bash

# This scripts applies the corresponding changes
# to run a Stride node for the nix e2e tests.
# This is needed because Stride is a Cosmos Hub consumer chain.
# So we need to update the setup to make it work properly.

source .env
set -eu

STRIDE_HOME="$BASE_DIR/node1"
STRIDED="strided --home ${STRIDE_HOME}"
CHAIN_ID=stride-1
DENOM=ustrd

STRIDE_DAY_EPOCH_DURATION="140s"
STRIDE_EPOCH_EPOCH_DURATION="35s"
MIN_DEPOSIT_AMT=1
MAX_DEPOSIT_PERIOD="10s"
VOTING_PERIOD="10s"
UNBONDING_TIME="1814400s"

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
mv $NODE0_DIR/config/config.toml ${tmp_config_toml}
mv $NODE0_DIR/config/app.toml ${tmp_app_toml}
mv $NODE0_DIR/config/client.toml ${tmp_client_toml}

# remove node0 - only use node1
# but keep node0 configs (ports mostly)
rm -rf ${STRIDE_HOME} ${NODE0_DIR}

$STRIDED init stride-local --chain-id $CHAIN_ID --overwrite

# use the generated config files
mv ${tmp_config_toml} ${config_toml}
mv ${tmp_app_toml} ${app_toml}
mv ${tmp_client_toml} ${client_toml}

jq '(.app_state.epochs.epochs[] | select(.identifier=="day") ).duration = $epochLen' --arg epochLen $STRIDE_DAY_EPOCH_DURATION $genesis_json >json.tmp && mv json.tmp $genesis_json
jq '(.app_state.epochs.epochs[] | select(.identifier=="stride_epoch") ).duration = $epochLen' --arg epochLen $STRIDE_EPOCH_EPOCH_DURATION $genesis_json >json.tmp && mv json.tmp $genesis_json
jq '.app_state.gov.params.max_deposit_period = $newVal' --arg newVal "$MAX_DEPOSIT_PERIOD" $genesis_json >json.tmp && mv json.tmp $genesis_json
jq '.app_state.gov.params.voting_period = $newVal' --arg newVal "$VOTING_PERIOD" $genesis_json >json.tmp && mv json.tmp $genesis_json
jq '.app_state.gov.params.min_deposit = [{"denom": $denom, "amount": $newVal}]' --arg newVal "$MIN_DEPOSIT_AMT" --arg denom "$DENOM" $genesis_json >json.tmp && mv json.tmp $genesis_json

$STRIDED add-consumer-section 1
jq '.app_state.ccvconsumer.params.unbonding_period = $newVal' --arg newVal "$UNBONDING_TIME" $genesis_json >json.tmp && mv json.tmp $genesis_json

echo "$VALIDATOR1_MNEMONIC" | $STRIDED keys add validator1 --recover --keyring-backend=test
$STRIDED add-genesis-account $($STRIDED keys show validator1 -a) 10000000000000${DENOM}

echo "$VALIDATOR2_MNEMONIC" | $STRIDED keys add validator2 --recover --keyring-backend=test
$STRIDED add-genesis-account $($STRIDED keys show validator2 -a) 10000000000000${DENOM}

echo "$COMMUNITY_MNEMONIC" | $STRIDED keys add community --recover --keyring-backend=test
$STRIDED add-genesis-account $($STRIDED keys show community -a) 10000000000000${DENOM}

echo "$SIGNER1_MNEMONIC" | $STRIDED keys add signer1 --recover --keyring-backend=test
$STRIDED add-genesis-account $($STRIDED keys show signer1 -a) 10000000000000${DENOM}

echo "$SIGNER2_MNEMONIC" | $STRIDED keys add signer2 --recover --keyring-backend=test
$STRIDED add-genesis-account $($STRIDED keys show signer2 -a) 10000000000000${DENOM}

# Update the task.ini file to run only one node
new_content="[program:stride-1-node]
autostart = true
autorestart = true
redirect_stderr = true
startsecs = 3
directory = %(here)s/node1
command = strided start --home . 
stdout_logfile = %(here)s/node.log"

# Replace the content of task.ini with the new content
echo "$new_content" >"$BASE_DIR/tasks.ini"
