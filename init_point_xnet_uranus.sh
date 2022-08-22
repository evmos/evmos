#!/usr/bin/env bash

# apt update && apt dist-upgrade -y
# apt install jq make gcc -y
#
# wget https://go.dev/dl/go1.19.linux-amd64.tar.gz
# rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.linux-amd64.tar.gz && rm go1.19.linux-amd64.tar.gz
# echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.profile
# source ~/.profile
#
# # go mod tidy # - needed?

read -p "WARNING! This script will DELETE EVERYTHING IN ~/.pointd and configure new chain environment. Do you want to proceed? (yes/no) " yn

case $yn in
	yes ) echo "";;
	no ) echo "Exiting...";
		exit;;
	* ) echo "Invalid response";
		exit 1;;
esac

KEY="mykey"
CHAINID="point_10731-1"
MONIKER="point-xnet-uranus"
#KEYRING="test" # remember to change to other types of keyring like 'file' in-case exposing to outside world, otherwise your balance will be wiped quickly. The keyring test does not require private key to steal tokens from you
KEYRING="test"
KEYALGO="eth_secp256k1"
LOGLEVEL="info"
# to trace evm
#TRACE="--trace"
TRACE=""

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

# used to exit on first error (any non-zero exit code)
set -e

# Clear everything of previous installation
rm -rf ~/.pointd

# Reinstall daemon
make install

# Set client config
pointd config keyring-backend $KEYRING
pointd config chain-id $CHAINID

# if $KEY exists it should be deleted
pointd keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO

# Set moniker and chain-id for Evmos (Moniker can be anything, chain-id must be an integer)
pointd init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to axpoint
cat $HOME/.pointd/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="apoint"' > $HOME/.pointd/config/tmp_genesis.json && mv $HOME/.pointd/config/tmp_genesis.json $HOME/.pointd/config/genesis.json
cat $HOME/.pointd/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="apoint"' > $HOME/.pointd/config/tmp_genesis.json && mv $HOME/.pointd/config/tmp_genesis.json $HOME/.pointd/config/genesis.json
cat $HOME/.pointd/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="apoint"' > $HOME/.pointd/config/tmp_genesis.json && mv $HOME/.pointd/config/tmp_genesis.json $HOME/.pointd/config/genesis.json
cat $HOME/.pointd/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="apoint"' > $HOME/.pointd/config/tmp_genesis.json && mv $HOME/.pointd/config/tmp_genesis.json $HOME/.pointd/config/genesis.json
cat $HOME/.pointd/config/genesis.json | jq '.app_state["inflation"]["params"]["mint_denom"]="apoint"' > $HOME/.pointd/config/tmp_genesis.json && mv $HOME/.pointd/config/tmp_genesis.json $HOME/.pointd/config/genesis.json

# Set gas limit in genesis
cat $HOME/.pointd/config/genesis.json | jq '.consensus_params["block"]["max_gas"]="10000000"' > $HOME/.pointd/config/tmp_genesis.json && mv $HOME/.pointd/config/tmp_genesis.json $HOME/.pointd/config/genesis.json

# Set claims unavailable
node_address=$(pointd keys list | grep  "address: " | cut -c12-)
cat $HOME/.pointd/config/genesis.json | jq -r '.app_state["claims"]["params"]["enable_claims"]=false' > $HOME/.pointd/config/tmp_genesis.json && mv $HOME/.pointd/config/tmp_genesis.json $HOME/.pointd/config/genesis.json

# Allocate genesis accounts (cosmos formatted addresses)
pointd add-genesis-account $KEY 10000000000000000000000000apoint --keyring-backend $KEYRING

pointd add-genesis-account evmos1ev3575lx5q7dd0jg0p5rh49pvp0lffgu0f6cad 100000000000000000000000000apoint --keyring-backend $KEYRING

# Update total supply with claim values
validators_supply=$(cat $HOME/.pointd/config/genesis.json | jq -r '.app_state["bank"]["supply"][0]["amount"]')
# Bc is required to add this big numbers
# total_supply=$(bc <<< "$amount_to_claim+$validators_supply")
total_supply=110000000000000000000000000
cat $HOME/.pointd/config/genesis.json | jq -r --arg total_supply "$total_supply" '.app_state["bank"]["supply"][0]["amount"]=$total_supply' > $HOME/.pointd/config/tmp_genesis.json && mv $HOME/.pointd/config/tmp_genesis.json $HOME/.pointd/config/genesis.json

# Sign genesis transaction
pointd gentx $KEY 1000000000000000000000apoint --keyring-backend $KEYRING --chain-id $CHAINID
## In case you want to create multiple validators at genesis
## 1. Back to `pointd keys add` step, init more keys
## 2. Back to `pointd add-genesis-account` step, add balance for those
## 3. Clone this ~/.pointd home directory into some others, let's say `~/.clonedpointd`
## 4. Run `gentx` in each of those folders
## 5. Copy the `gentx-*` folders under `~/.clonedpointd/config/gentx/` folders into the original `~/.pointd/config/gentx`

# Collect genesis tx
pointd collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
pointd validate-genesis

if [[ $1 == "pending" ]]; then
  echo "pending mode is on, please wait for the first block committed."
fi

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
pointd start --pruning=nothing $TRACE --log_level $LOGLEVEL --minimum-gas-prices=0.0001apoint --json-rpc.api eth,txpool,personal,net,debug,web3
