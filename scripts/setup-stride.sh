#!/bin/bash

source .env
set -eu

# remove node0 - only use node1
rm -rf "$BASE_DIR/node0"

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

# need to update the ports based on the base port
P2P_ADDR_PORT=$((BASE_PORT + 100))
GRPC_ADDR_PORT=$((BASE_PORT + 103))
COSMOS_API_ADDR_PORT=$((BASE_PORT + 104))
PPROF_LADDR_PORT=$((BASE_PORT + 105))
GRPC_LADDR_PORT=$((BASE_PORT + 106))
RPC_LADDR_PORT=$((BASE_PORT + 107))
GRPC_WEB_ADDR_PORT=$((BASE_PORT + 108))

config_toml="${STRIDE_HOME}/config/config.toml"
client_toml="${STRIDE_HOME}/config/client.toml"
app_toml="${STRIDE_HOME}/config/app.toml"
genesis_json="${STRIDE_HOME}/config/genesis.json"

rm -rf ${STRIDE_HOME}

$STRIDED init stride-local --chain-id $CHAIN_ID --overwrite

app_config=$(
    cat <<EOF
minimum-gas-prices = "0ustrd"
pruning = "nothing"
pruning-keep-recent = "0"
pruning-interval = "0"
halt-height = 0
halt-time = 0
min-retain-blocks = 0
inter-block-cache = true
index-events = []
iavl-cache-size = 781250
iavl-disable-fastnode = false
iavl-lazy-loading = false
app-db-backend = ""

[telemetry]
enabled = false
enable-hostname = false
enable-hostname-label = false
enable-service-label = false
prometheus-retention-time = 0
global-labels = []

[api]
enable = true
swagger = true
address = "tcp://127.0.0.1:${COSMOS_API_ADDR_PORT}"
max-open-connections = 1000
rpc-read-timeout = 10
rpc-write-timeout = 0
enable-unsafe-cors = true
rpc-max-body-bytes = 1000000
enabled-unsafe-cors = true

[rosetta]
enable = false
address = ":8080"
blockchain = "app"
network = "network"
retries = 3
offline = false
enable-fee-suggestion = false
gas-to-suggest = 200000
denom-to-suggest = "uatom"

[grpc]
enable = true
address = "127.0.0.1:${GRPC_ADDR_PORT}"
max-recv-msg-size = "10485760"
max-send-msg-size = "2147483647"

[grpc-web]
address = "127.0.0.1:${GRPC_WEB_ADDR_PORT}"

[state-sync]
snapshot-interval = 5
snapshot-keep-recent = 10

[store]
streamers = []

[mempool]
max-txs = "5000"

[wasm]
query_gas_limit = 300000
lru_size = 0
EOF
)
echo "$app_config" >"$app_toml"

config=$(
    cat <<EOF
proxy_app = "tcp://127.0.0.1:26658"
moniker = "node0"
block_sync = true
db_backend = "goleveldb"
db_dir = "data"
log_level = "info"
log_format = "plain"
genesis_file = "config/genesis.json"
priv_validator_key_file = "config/priv_validator_key.json"
priv_validator_state_file = "data/priv_validator_state.json"
mode = "validator"
priv_validator_laddr = ""
node_key_file = "config/node_key.json"
abci = "socket"
filter_peers = false

[lite]
lite_enabled = false

[rpc]
laddr = "tcp://127.0.0.1:${RPC_LADDR_PORT}"
cors_allowed_origins = []
cors_allowed_methods = ["HEAD", "GET", "POST"]
cors_allowed_headers = ["Origin", "Accept", "Content-Type", "X-Requested-With", "X-Server-Time"]
grpc_laddr = "tcp://127.0.0.1:${GRPC_LADDR_PORT}"
grpc_max_open_connections = 900
unsafe = false
max_open_connections = 900
max_subscription_clients = 100
max_subscriptions_per_client = 5
experimental_subscription_buffer_size = 200
experimental_websocket_write_buffer_size = 200
experimental_close_on_slow_client = false
timeout_broadcast_tx_commit = "30s"
max_body_bytes = 1000000
max_header_bytes = 1048576
timeout-broadcast-tx-commit = "30s"
tls_cert_file = ""
tls_key_file = ""
pprof_laddr = "127.0.0.1:${PPROF_LADDR_PORT}"

[p2p]
laddr = "tcp://127.0.0.1:${P2P_ADDR_PORT}"
external_address = ""
seeds = ""
persistent_peers = ""
upnp = false
addr_book_file = "config/addrbook.json"
addr_book_strict = false
max_num_inbound_peers = 40
max_num_outbound_peers = 10
unconditional_peer_ids = ""
persistent_peers_max_dial_period = "0s"
flush_throttle_timeout = "100ms"
max_packet_msg_payload_size = 1024
send_rate = 5120000
recv_rate = 5120000
persistent-peers = ""
addr-book-strict = false
allow-duplicate-ip = true
pex = true
seed_mode = false
private_peer_ids = ""
allow_duplicate_ip = true
handshake_timeout = "20s"
dial_timeout = "3s"

[mempool]
version = "v0"
recheck = true
broadcast = true
wal_dir = ""
size = 5000
max_txs_bytes = 1073741824
cache_size = 10000
keep-invalid-txs-in-cache = false
max_tx_bytes = 1048576
max_batch_bytes = 0
ttl-duration = "0s"
ttl-num-blocks = 0

[statesync]
enable = false
rpc_servers = ""
trust_height = 0
trust_hash = ""
trust_period = "168h0m0s"
discovery_time = "15s"
temp_dir = ""
chunk_request_timeout = "10s"
chunk_fetchers = "4"

[blocksync]
version = "v0"

[consensus]
wal_file = "data/cs.wal/wal"
timeout_propose = "3s"
timeout_propose_delta = "500ms"
timeout_prevote = "1s"
timeout_prevote_delta = "500ms"
timeout_precommit = "1s"
timeout_precommit_delta = "500ms"
timeout_commit = "1s"
double_sign_check_height = 0
skip_timeout_commit = false
create_empty_blocks = true
create_empty_blocks_interval = "0s"
timeout-commit = "1s"
peer_gossip_sleep_duration = "100ms"
peer_query_maj23_sleep_duration = "2s"

[storage]
discard_abci_responses = false

[tx_index]
indexer = "kv"
psql-conn = ""

[instrumentation]
prometheus = false
prometheus_listen_addr = ":26660"
max_open_connections = 3
namespace = "cometbft"
EOF
)
echo "$config" >"$config_toml"

sed -i -E "s|chain-id = \"\"|chain-id = \"${CHAIN_ID}\"|g" $client_toml
sed -i -E "s|keyring-backend = \"os\"|keyring-backend = \"test\"|g" $client_toml
sed -i -E "s|node = \".*\"|node = \"tcp://localhost:${RPC_LADDR_PORT}\"|g" $client_toml

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
