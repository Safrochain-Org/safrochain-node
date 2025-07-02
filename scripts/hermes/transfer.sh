#!/bin/sh

export safrochaind_NODE="http://localhost:26657"
CHAIN_A_ARGS="--from safrochain1 --keyring-backend test --chain-id local-1 --home $HOME/.safrochain1/ --node http://localhost:26657 --yes"

# safrochaind q ibc channel channels

# Send from local-1 to local-2 via the relayer
safrochaind tx ibc-transfer transfer transfer channel-0 addr_safro1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl 9usaf $CHAIN_A_ARGS --packet-timeout-height 0-0

sleep 6

# check the query on the other chain to ensure it went through
safrochaind q bank balances addr_safro1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl --chain-id local-2 --node http://localhost:36657