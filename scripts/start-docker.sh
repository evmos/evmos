#!/bin/bash

KEY="mykey"
CHAINID="evoblock_9000-1"
MONIKER="mymoniker"
DATA_DIR=$(mktemp -d -t evoblock-datadir.XXXXX)

echo "create and add new keys"
./evoblockd keys add $KEY --home $DATA_DIR --no-backup --chain-id $CHAINID --algo "eth_secp256k1" --keyring-backend test
echo "init Evoblock with moniker=$MONIKER and chain-id=$CHAINID"
./evoblockd init $MONIKER --chain-id $CHAINID --home $DATA_DIR
echo "prepare genesis: Allocate genesis accounts"
./evoblockd add-genesis-account \
"$(./evoblockd keys show $KEY -a --home $DATA_DIR --keyring-backend test)" 1000000000000000000aEVO,1000000000000000000stake \
--home $DATA_DIR --keyring-backend test
echo "prepare genesis: Sign genesis transaction"
./evoblockd gentx $KEY 1000000000000000000stake --keyring-backend test --home $DATA_DIR --keyring-backend test --chain-id $CHAINID
echo "prepare genesis: Collect genesis tx"
./evoblockd collect-gentxs --home $DATA_DIR
echo "prepare genesis: Run validate-genesis to ensure everything worked and that the genesis file is setup correctly"
./evoblockd validate-genesis --home $DATA_DIR

echo "starting evoblock node $i in background ..."
./evoblockd start --pruning=nothing --rpc.unsafe \
--keyring-backend test --home $DATA_DIR \
>$DATA_DIR/node.log 2>&1 & disown

echo "started evoblock node"
tail -f /dev/null