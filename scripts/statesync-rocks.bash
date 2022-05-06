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
evmosd init evmos-rocks
wget -O ~/.evmosd/config/genesis.json https://cloudflare-ipfs.com/ipfs/QmYmcmveFYMdA1NRAvAa3hKM8HQRdiwfmYtf3GgyJGTMEa

# this will let tendermint know that we want rocks
sed -i 's/goleveldb/rocksdb/' ~/.evmosd/config/config.toml


# Uncomment if resyncing a server
# evmosd unsafe-reset-all 



INTERVAL=1500

# GET TRUST HASH AND TRUST HEIGHT

LATEST_HEIGHT=$(curl -s https://tendermint.bd.evmos.org:26657/block | jq -r .result.block.header.height);
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL))
TRUST_HASH=$(curl -s "https://tendermint.bd.evmos.org:26657/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)


# TELL USER WHAT WE ARE DOING
echo "TRUST HEIGHT: $BLOCK_HEIGHT"
echo "TRUST HASH: $TRUST_HASH"


# export state sync vars
export EVMOSD_P2P_SEEDS="906840c2f447915f3d0e37bc68221f5494f541db@3.39.58.32:26656,7aa31684d201f8ebc0b1e832d90d7490345d77ee@52.10.99.253:26656,5740e4a36e646e80cc5648daf5e983e5b5d8f265@54.39.18.27:26656,de2c5e946e21360d4ffa3885579fa038a7d9776e@46.101.148.190:26656"
export EVMOSD_P2P_MAX_NUM_OUTBOUND_PEERS=200
export EVMOSD_STATESYNC_ENABLE=true
export EVMOSD_STATESYNC_RPC_SERVERS="https://tendermint.bd.evmos.org:26657"
export EVMOSD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export EVMOSD_STATESYNC_TRUST_HASH=$TRUST_HASH

# Rockdb won't make this folder, so we make it 
# we can remove this if my changes are accepted upstream
mkdir -p ~/.osmosisd/data/snapshots/metadata.db

# THIS WILL FAIL BECAUSE THE APP VERSION IS CORRECTLY SET IN OSMOSIS
evmosd start --db_backend rocksdb 
