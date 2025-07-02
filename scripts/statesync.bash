#!/bin/bash
# microtick and bitcanna contributed significantly here.
# Pebbledb state sync script.
set -uxe

# Set Golang environment variables.
export GOPATH=~/go
export PATH=$PATH:~/go/bin

# Initialize chain.
safrochaind init test

# Get Genesis
wget https://download.dimi.sh/safrochain-phoenix2-genesis.tar.gz
tar -xvf safrochain-phoenix2-genesis.tar.gz
mv safrochain-phoenix2-genesis.json "$HOME/.safrochain/config/genesis.json"




# Get "trust_hash" and "trust_height".
INTERVAL=1000
LATEST_HEIGHT="$(curl -s https://safrochain-rpc.polkachu.com/block | jq -r .result.block.header.height)"
BLOCK_HEIGHT="$((LATEST_HEIGHT-INTERVAL))"
TRUST_HASH="$(curl -s "https://safrochain-rpc.polkachu.com/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)"

# Print out block and transaction hash from which to sync state.
echo "trust_height: $BLOCK_HEIGHT"
echo "trust_hash: $TRUST_HASH"

# Export state sync variables.
export safrochaind_STATESYNC_ENABLE=true
export safrochaind_P2P_MAX_NUM_OUTBOUND_PEERS=200
export safrochaind_STATESYNC_RPC_SERVERS="https://rpc-safrochain-ia.notional.ventures:443,https://safrochain-rpc.polkachu.com:443"
export safrochaind_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export safrochaind_STATESYNC_TRUST_HASH=$TRUST_HASH

# Fetch and set list of seeds from chain registry.
safrochaind_P2P_SEEDS="$(curl -s https://raw.githubusercontent.com/cosmos/chain-registry/master/safrochain/chain.json | jq -r '[foreach .peers.seeds[] as $item (""; "\($item.id)@\($item.address)")] | join(",")')"
export safrochaind_P2P_SEEDS

# Start chain.
safrochaind start --x-crisis-skip-assert-invariants 
