
# upload the smart contract, then create a validator. Confirm it works

export safrochaind_NODE="tcp://localhost:26657"
FLAGS="--from=safrochain1 --gas=2500000 --fees=50000usaf --node=http://localhost:26657 --yes --keyring-backend=test --home $HOME/.safrochain1 --chain-id=local-1 --output=json"

safrochaind tx wasm store ./keeper/contract/safrochain_staking_hooks_example.wasm $FLAGS

sleep 5

txhash=$(safrochaind tx wasm instantiate 1 '{}' --label=safrochain_staking --no-admin $FLAGS | jq -r .txhash)
sleep 5
addr=$(safrochaind q tx $txhash --output=json --node=http://localhost:26657 | jq -r .logs[0].events[2].attributes[0].value) && echo $addr

# register addr to staking
safrochaind tx cw-hooks register staking $addr $FLAGS
safrochaind q cw-hooks staking-contracts

# safrochaind tx cw-hooks unregister staking $addr $FLAGS
# safrochaind q cw-hooks staking-contracts

# get config
safrochaind q wasm contract-state smart $addr '{"get_config":{}}' --node=http://localhost:26657

# get last validator
safrochaind q wasm contract-state smart $addr '{"last_val_change":{}}' --node=http://localhost:26657
safrochaind q wasm contract-state smart $addr '{"last_delegation_change":{}}' --node=http://localhost:26657

# create validator
safrochaind tx staking create-validator --amount 1usaf --commission-rate="0.05" --commission-max-rate="1.0" --commission-max-change-rate="1.0" --moniker="test123" --from=safrochain2 --pubkey=$(safrochaind tendermint show-validator --home $HOME/.safrochain) --min-self-delegation="1" --gas=1000000 --fees=50000usaf --node=http://localhost:26657 --yes --keyring-backend=test --home $HOME/.safrochain1 --chain-id=local-1 --output=json

# safrochaind export --output-document=$HOME/Desktop/export.json --home=$HOME/.safrochain1