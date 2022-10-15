# Don't use this on a validator!  Please!  You'd double sign!
# If you must use on a validator, comment out evmosd unsafe-reset-all
# You have been warned. 

# Initialize chain.
evmosd init test --chain-id evmos_9000-1
#
# # Get Genesis
wget https://archive.evmos.org/mainnet/genesis.json
mv genesis.json ~/.evmosd/config/

# Get State

wget -nc -O evmos_5757874.tar.lz4 https://snapshots.polkachu.com/snapshots/evmos/evmos_5757874.tar.lz4 --inet4-only
evmosd tendermint unsafe-reset-all
lz4 -c -d evmos_5757874.tar.lz4  | tar -x -C ~/.evmosd
cp scripts/polkachu/* ~/.evmosd/config

# Get seeds
export EVMOSD_P2P_SEEDS=$(curl -s https://raw.githubusercontent.com/cosmos/chain-registry/master/evmos/chain.json | jq -r '[foreach .peers.seeds[] as $item (""; "\($item.id)@\($item.address)")] | join(",")')


evmosd start --x-crisis-skip-assert-invariants 

