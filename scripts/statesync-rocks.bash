# microtick and bitcanna contributed significantly here.
# rocksdb works
# This defaults to rocksdb.  Read the script, some stuff is commented out based on your system. 


# PRINT EVERY COMMAND
set -ux

# uncomment the three lines below to build osmosis

export GOPATH=~/go
export PATH=$PATH:~/go/bin

# Use if building on a mac
# export CGO_CFLAGS="-I/opt/homebrew/Cellar/rocksdb/6.27.3/include"
# export CGO_LDFLAGS="-L/opt/homebrew/Cellar/rocksdb/6.27.3/lib -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd -L/opt/homebrew/Cellar/snappy/1.1.9/lib -L/opt/homebrew/Cellar/lz4/1.9.3/lib/ -L /opt/homebrew/Cellar/zstd/1.5.1/lib/"
go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb' -tags rocksdb ./...

# MAKE HOME FOLDER AND GET GENESIS
osmosisd init osmosis-rocks
wget -O ~/.osmosisd/config/genesis.json https://cloudflare-ipfs.com/ipfs/QmXRvBT3hgoXwwPqbK6a2sXUuArGM8wPyo1ybskyyUwUxs

# this will let tendermint know that we want rocks
sed -i 's/goleveldb/rocksdb/' ~/.osmosisd/config/config.toml


# Uncomment if resyncing a server
# osmosisd unsafe-reset-all --home /osmosis/osmosis



INTERVAL=1500

# GET TRUST HASH AND TRUST HEIGHT

LATEST_HEIGHT=$(curl -s https://osmosis.validator.network/block | jq -r .result.block.header.height);
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL))
TRUST_HASH=$(curl -s "https://osmosis.validator.network/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)


# TELL USER WHAT WE ARE DOING
echo "TRUST HEIGHT: $BLOCK_HEIGHT"
echo "TRUST HASH: $TRUST_HASH"


# export state sync vars
export OSMOSISD_P2P_MAX_NUM_OUTBOUND_PEERS=200
export OSMOSISD_STATESYNC_ENABLE=true
export OSMOSISD_STATESYNC_RPC_SERVERS="https://osmosis.validator.network:443,https://rpc.osmosis.notional.ventures:443,https://rpc-osmosis.ecostake.com:443"
export OSMOSISD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export OSMOSISD_STATESYNC_TRUST_HASH=$TRUST_HASH

# Rockdb won't make this folder, so we make it 
# we can remove this if my changes are accepted upstream
mkdir -p ~/.osmosisd/data/snapshots/metadata.db

# THIS WILL FAIL BECAUSE THE APP VERSION IS CORRECTLY SET IN OSMOSIS
evmosd start --db_backend rocksdb 
