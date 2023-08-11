#!/bin/bash

CHAINID="evmos_9000-1"
MONIKER="localtestnet"
CHAIN_DIR="$HOME/.test-evmosd"
KEYALGO="eth_secp256k1" #gitleaks:allow

GENESIS="$CHAIN_DIR/config/genesis.json"
TMP_GENESIS="$CHAIN_DIR/config/tmp_genesis.json"

# localKey address 0x7cb61d4117ae31a12e393a1cfa3bac666481d02e
VAL_KEY="localkey"
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

# remove existing daemon and client
rm -rf "$CHAIN_DIR"

# Import keys from mnemonics
echo "$VAL_MNEMONIC" | evmosd keys add "$VAL_KEY" --recover --keyring-backend test --algo "$KEYALGO" --home "$CHAIN_DIR"
echo "$USER1_MNEMONIC" | evmosd keys add "$USER1_KEY" --recover --keyring-backend test --algo "$KEYALGO" --home "$CHAIN_DIR"
echo "$USER2_MNEMONIC" | evmosd keys add "$USER2_KEY" --recover --keyring-backend test --algo "$KEYALGO" --home "$CHAIN_DIR"
echo "$USER3_MNEMONIC" | evmosd keys add "$USER3_KEY" --recover --keyring-backend test --algo "$KEYALGO" --home "$CHAIN_DIR"
echo "$USER4_MNEMONIC" | evmosd keys add "$USER4_KEY" --recover --keyring-backend test --algo "$KEYALGO" --home "$CHAIN_DIR"

evmosd init "$MONIKER" --chain-id "$CHAINID" --home "$CHAIN_DIR"

# Set gas limit in genesis
jq '.consensus_params["block"]["max_gas"]="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Change parameter token denominations to aevmos
jq '.app_state.staking.params.bond_denom="aevmos"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.crisis.constant_fee.denom="aevmos"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.gov.deposit_params.min_deposit[0].denom="aevmos"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.evm.params.evm_denom="aevmos"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.inflation.params.mint_denom="aevmos"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# modified default configs
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' 's/create_empty_blocks = true/create_empty_blocks = false/g' "$CHAIN_DIR"/config/config.toml
    sed -i '' 's/prometheus-retention-time = 0/prometheus-retention-time  = 1000000000000/g' "$CHAIN_DIR"/config/app.toml
    sed -i '' 's/enabled = false/enabled = true/g' "$CHAIN_DIR"/config/app.toml
    sed -i '' 's/prometheus = false/prometheus = true/' "$CHAIN_DIR"/config/config.toml
    sed -i '' 's/timeout_commit = "5s"/timeout_commit = "1s"/g' "$CHAIN_DIR"/config/config.toml
else
    sed -i 's/create_empty_blocks = true/create_empty_blocks = false/g' "$CHAIN_DIR"/config/config.toml
    sed -i 's/prometheus-retention-time  = "0"/prometheus-retention-time  = "1000000000000"/g' "$CHAIN_DIR"/config/app.toml
    sed -i 's/enabled = false/enabled = true/g' "$CHAIN_DIR"/config/app.toml
    sed -i 's/prometheus = false/prometheus = true/' "$CHAIN_DIR"/config/config.toml
    sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' "$CHAIN_DIR"/config/config.toml
fi

# Allocate genesis accounts (cosmos formatted addresses)
evmosd add-genesis-account "$(evmosd keys show "$VAL_KEY" -a --keyring-backend test --home "$CHAIN_DIR")" 1000000000000000000000aevmos,1000000000000000000stake --keyring-backend test --home "$CHAIN_DIR"
evmosd add-genesis-account "$(evmosd keys show "$USER1_KEY" -a --keyring-backend test --home "$CHAIN_DIR")" 1000000000000000000000aevmos,1000000000000000000stake --keyring-backend test --home "$CHAIN_DIR"
evmosd add-genesis-account "$(evmosd keys show "$USER2_KEY" -a --keyring-backend test --home "$CHAIN_DIR")" 1000000000000000000000aevmos,1000000000000000000stake --keyring-backend test --home "$CHAIN_DIR"
evmosd add-genesis-account "$(evmosd keys show "$USER3_KEY" -a --keyring-backend test --home "$CHAIN_DIR")" 1000000000000000000000aevmos,1000000000000000000stake --keyring-backend test --home "$CHAIN_DIR"
evmosd add-genesis-account "$(evmosd keys show "$USER4_KEY" -a --keyring-backend test --home "$CHAIN_DIR")" 1000000000000000000000aevmos,1000000000000000000stake --keyring-backend test --home "$CHAIN_DIR"

# Sign genesis transaction
evmosd gentx "$VAL_KEY" 1000000000000000000aevmos --amount=1000000000000000000000aevmos --chain-id "$CHAINID" --keyring-backend test --home "$CHAIN_DIR"

# Collect genesis tx
evmosd collect-gentxs --home "$CHAIN_DIR"

# Run this to ensure everything worked and that the genesis file is setup correctly
evmosd validate-genesis --home "$CHAIN_DIR"

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
evmosd start --metrics --pruning=nothing --rpc.unsafe --keyring-backend test --log_level info --json-rpc.enable true --json-rpc.api eth,txpool,personal,net,debug,web3 --api.enable --home "$CHAIN_DIR" --chain-id "$CHAINID"
