#!/bin/bash
set -eux

CONSUMER_HOME="$HOME/.consumer-evmosd"
CONSUMER_HOME1="$HOME/.consumer-evmosd1"
PROVIDER_CHAIN_ID="provider"
CONSUMER_CHAIN_ID="evmos_9000-1"
MONIKER="evmos"
VALIDATOR="validator"
VALIDATOR1="validator1"
KEYRING="--keyring-backend test"
TX_FLAGS="--gas-adjustment 100 --gas auto"
PROVIDER_BINARY="interchain-security-pd"
CONSUMER_BINARY="evmosd"
NODE_IP="localhost"
PROVIDER_RPC_LADDR="$NODE_IP:26658"
PROVIDER_GRPC_ADDR="$NODE_IP:9091"
PROVIDER_RPC_LADDR1="$NODE_IP:26668"
PROVIDER_GRPC_ADDR1="$NODE_IP:9101"
CONSUMER_RPC_LADDR="$NODE_IP:26648"
CONSUMER_GRPC_ADDR="$NODE_IP:9081"
CONSUMER_RPC_LADDR1="$NODE_IP:26638"
CONSUMER_GRPC_ADDR1="$NODE_IP:9071"
CONSUMER_USER="consumer"
PROVIDER_HOME="$HOME/.provider"
PROVIDER_HOME1="$HOME/.provider1"
PROVIDER_NODE_ADDRESS="tcp://localhost:26658"

# Clean start
killall $CONSUMER_BINARY &> /dev/null || true
rm -rf $CONSUMER_HOME
rm -rf $CONSUMER_HOME1

################CONSUMER############################


# Build genesis file and node directory structure
./$CONSUMER_BINARY init --chain-id $CONSUMER_CHAIN_ID $MONIKER --home $CONSUMER_HOME
sleep 1

# Add ccv section
if ! ./$PROVIDER_BINARY q provider consumer-genesis "$CONSUMER_CHAIN_ID" --node "$PROVIDER_NODE_ADDRESS" --output json > "$CONSUMER_HOME"/consumer_section.json;
then
       echo "Failed to get consumer genesis for the chain-id '$CONSUMER_CHAIN_ID'! Finalize genesis failed. For more details please check the log file in output directory."
       exit 1
fi

jq -s '.[0].app_state.ccvconsumer = .[1] | .[0]' "$CONSUMER_HOME"/config/genesis.json "$CONSUMER_HOME"/consumer_section.json > "$CONSUMER_HOME"/genesis_consumer.json && \
	mv "$CONSUMER_HOME"/genesis_consumer.json "$CONSUMER_HOME"/config/genesis.json

# Modify genesis params
jq ".app_state.ccvconsumer.params.blocks_per_distribution_transmission = \"70\" | .app_state.tokenfactory.paused = { \"paused\": false }" \
  $CONSUMER_HOME/config/genesis.json > \
   $CONSUMER_HOME/edited_genesis.json && mv $CONSUMER_HOME/edited_genesis.json $CONSUMER_HOME/config/genesis.json
sleep 1

# Create user account keypair
./$CONSUMER_BINARY keys add $CONSUMER_USER $KEYRING --home $CONSUMER_HOME --output json > $CONSUMER_HOME/consumer_keypair.json 2>&1

# Add account in genesis (required by Hermes)
./$CONSUMER_BINARY add-genesis-account $(jq -r .address $CONSUMER_HOME/consumer_keypair.json) 1000000000stake --home $CONSUMER_HOME

# Copy validator key files
cp $PROVIDER_HOME/config/priv_validator_key.json $CONSUMER_HOME/config/priv_validator_key.json
cp $PROVIDER_HOME/config/node_key.json $CONSUMER_HOME/config/node_key.json

#######CHAIN2#######
./$CONSUMER_BINARY init --chain-id $CONSUMER_CHAIN_ID $MONIKER --home $CONSUMER_HOME1
sleep 1
#copy genesis
cp $CONSUMER_HOME/config/genesis.json $CONSUMER_HOME1/config/genesis.json

cp $PROVIDER_HOME1/config/priv_validator_key.json $CONSUMER_HOME1/config/priv_validator_key.json
cp $PROVIDER_HOME1/config/node_key.json $CONSUMER_HOME1/config/node_key.json

##########SET CONFIG.TOML#####################
# Set default client port
sed -i -r "/node =/ s/= .*/= \"tcp:\/\/${CONSUMER_RPC_LADDR1}\"/" $CONSUMER_HOME1/config/client.toml
sed -i -r "/node =/ s/= .*/= \"tcp:\/\/${CONSUMER_RPC_LADDR}\"/" $CONSUMER_HOME/config/client.toml
node=$(./$CONSUMER_BINARY tendermint show-node-id --home $CONSUMER_HOME)
node1=$(./$CONSUMER_BINARY tendermint show-node-id --home $CONSUMER_HOME1)
sed -i -r "/^persistent_peers =/ s/= .*/= \"$node@localhost:26656\"/" "$PROVIDER_HOME1"/config/config.toml
sed -i -r "/^persistent_peers =/ s/= .*/= \"$node1@localhost:26666\"/" "$PROVIDER_HOME"/config/config.toml

sed -i -r "114s/.*/address = \"tcp:\/\/0.0.0.0:1318\"/" "$CONSUMER_HOME1"/config/app.toml

# Start the chain
./$CONSUMER_BINARY start \
       --home $CONSUMER_HOME \
       --rpc.laddr http://${CONSUMER_RPC_LADDR} \
       --grpc.address ${CONSUMER_GRPC_ADDR} \
       --address tcp://${NODE_IP}:26645 \
       --p2p.laddr tcp://${NODE_IP}:26646 \
       --grpc-web.enable=false \
       --log_level trace \
       --trace \
       &> $CONSUMER_HOME/logs &

./$CONSUMER_BINARY start \
       --home $CONSUMER_HOME1 \
       --rpc.laddr tcp://${CONSUMER_RPC_LADDR1} \
       --grpc.address ${CONSUMER_GRPC_ADDR1} \
       --address tcp://${NODE_IP}:26635 \
       --p2p.laddr tcp://${NODE_IP}:26636 \
       --grpc-web.enable=false \
       --log_level trace \
       --trace \
       &> $CONSUMER_HOME1/logs &
sleep 10

######################################HERMES###################################

# Setup Hermes in packet relayer mode
killall hermes 2> /dev/null || true

tee ~/.hermes/config.toml<<EOF
[global]
log_level = "trace"

[[chains]]
id = "$CONSUMER_CHAIN_ID"
type = "CosmosSdk"
rpc_addr = "http://${CONSUMER_RPC_LADDR}"
grpc_addr = "http://${CONSUMER_GRPC_ADDR}"
event_source = { mode = "push", url = "ws://${CONSUMER_RPC_LADDR}/websocket", batch_delay = "500ms" }
rpc_timeout = "10s"
account_prefix = "evmos"
key_name = "relayer"
store_prefix = "ibc"
default_gas = 2000000
max_gas = 2000000
gas_price = { price = 0.00, denom = "stake" }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = "5s"
trusting_period = "599s"
trust_threshold = { numerator = "1", denominator = "3" }

[[chains]]
id = "$PROVIDER_CHAIN_ID"
type = "CosmosSdk"
rpc_addr = "http://${PROVIDER_RPC_LADDR}"
grpc_addr = "http://${PROVIDER_GRPC_ADDR}"
event_source = { mode = "push", url = "ws://${PROVIDER_RPC_LADDR}/websocket", batch_delay = "500ms" }
rpc_timeout = "10s"
account_prefix = "cosmos"
key_name = "relayer"
store_prefix = "ibc"
default_gas = 2000000
max_gas = 2000000
gas_price = { price = 0.00, denom = "stake" }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = "5s"
trusting_period = "599s"
trust_threshold = { numerator = "1", denominator = "3" }
EOF

# Delete all previous keys in relayer
hermes keys delete --chain $CONSUMER_CHAIN_ID --all
hermes keys delete --chain $PROVIDER_CHAIN_ID --all

# Restore keys to hermes relayer
#hermes keys add --mnemonic "$(jq -r .mnemonic $CONSUMER_HOME/consumer_keypair.json)" $CONSUMER_CHAIN_ID
## temp_start_provider.sh creates key pair and stores it in keypair.json
#hermes keys add --mnemonic "$(jq -r .mnemonic $PROVIDER_HOME/keypair.json)" $PROVIDER_CHAIN_ID

# Restore keys to hermes relayer
hermes keys add --chain $CONSUMER_CHAIN_ID --mnemonic-file <(echo "$(jq -r .mnemonic $CONSUMER_HOME/consumer_keypair.json)")
hermes keys add --chain $PROVIDER_CHAIN_ID --mnemonic-file <(echo "$(jq -r .mnemonic $PROVIDER_HOME/keypair.json)")

sleep 5

# Create a connection
hermes create connection --a-chain $CONSUMER_CHAIN_ID --b-chain $PROVIDER_CHAIN_ID

# Wait for the connection to be established
sleep 10

# Create a channel
hermes create channel --a-chain $CONSUMER_CHAIN_ID --b-chain $PROVIDER_CHAIN_ID --a-port consumer --b-port provider --new-client-connection --channel-version 1

hermes -j start &> ~/.hermes/logs &

############################################################

PROVIDER_VALIDATOR_ADDRESS=$(jq -r .address $PROVIDER_HOME/keypair.json)
DELEGATIONS=$($PROVIDER_BINARY q staking delegations $PROVIDER_VALIDATOR_ADDRESS --home $PROVIDER_HOME --node tcp://${PROVIDER_RPC_LADDR} -o json)
OPERATOR_ADDR=$(echo $DELEGATIONS | jq -r .delegation_responses[0].delegation.validator_address)

./$PROVIDER_BINARY tx staking delegate $OPERATOR_ADDR 50000000stake \
       --from $VALIDATOR \
       $KEYRING \
       --home $PROVIDER_HOME \
       --node tcp://${PROVIDER_RPC_LADDR} \
       --chain-id $PROVIDER_CHAIN_ID -y -b block
sleep 1