#!/bin/bash

KEY="dev0"
CHAINID="eidon-chain_9000-1"
MONIKER="mymoniker"
DATA_DIR=$(mktemp -d -t eidon-chain-datadir.XXXXX)

echo "create and add new keys"
./eidond keys add $KEY --home "$DATA_DIR" --no-backup --chain-id $CHAINID --algo "eth_secp256k1" --keyring-backend test
echo "init Eidon-chain with moniker=$MONIKER and chain-id=$CHAINID"
./eidond init $MONIKER --chain-id "$CHAINID" --home "$DATA_DIR"
echo "prepare genesis: Allocate genesis accounts"
./eidond add-genesis-account \
	"$(./eidond keys show "$KEY" -a --home "$DATA_DIR" --keyring-backend test)" 1000000000000000000aeidon-chain,1000000000000000000stake \
	--home "$DATA_DIR" --keyring-backend test
echo "prepare genesis: Sign genesis transaction"
./eidond gentx "$KEY" 1000000000000000000stake --keyring-backend test --home "$DATA_DIR" --keyring-backend test --chain-id "$CHAINID"
echo "prepare genesis: Collect genesis tx"
./eidond collect-gentxs --home "$DATA_DIR"
echo "prepare genesis: Run validate-genesis to ensure everything worked and that the genesis file is setup correctly"
./eidond validate-genesis --home "$DATA_DIR"

echo "starting eidon-chain node in background ..."
./eidond start --pruning=nothing --rpc.unsafe \
	--keyring-backend test --home "$DATA_DIR" \
	>"$DATA_DIR"/node.log 2>&1 &
disown

echo "started eidon-chain node"
tail -f /dev/null
