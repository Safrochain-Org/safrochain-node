# CosmWasm Examples — Safrochain

This directory contains self-contained CosmWasm smart contract examples that
demonstrate the full lifecycle of developing, deploying, and interacting with
contracts on Safrochain.

## Prerequisites

- [Rust](https://rustup.rs/) with `wasm32-unknown-unknown` target:
  ```bash
  rustup target add wasm32-unknown-unknown
  ```
- [safrochaind](https://github.com/Safrochain-Org/safrochain-node/releases) binary in `$PATH`
- A running local Safrochain node (see [Localnet setup](#localnet-setup) below)
- Optional: [Docker](https://docker.com/) for the CosmWasm optimizer

## Contract Examples

| Contract | Description | Key Concepts |
|---|---|---|
| [counter/](./counter/) | Hello-world: increment, reset, get count | State, messages, queries, cw-multi-test |
| [cw20-mintable/](./cw20-mintable/) | Minimal mintable CW20 token | CW20 standard, mint/burn/transfer, queries |
| [clock-counter/](./clock-counter/) | Auto-increment on every block via x/clock | Sudo message, EndBlock execution, x/clock registration |
| [cw-hooks-listener/](./cw-hooks-listener/) | Listens for staking events via x/cw-hooks | Sudo hooks, staking event handling |

## Localnet Setup

Start a local Safrochain devnet:

```bash
# Build and start
make build
./build/safrochaind init safro-testnet --chain-id safro-testnet-1
./build/safrochaind start
```

Or using the provided Docker compose:

```bash
docker compose up -d
```

## Quick Start

Build all contracts:

```bash
cd examples/cosmwasm

# Build each contract
cd counter && cargo wasm && cd ..
cd cw20-mintable && cargo wasm && cd ..
cd clock-counter && cargo wasm && cd ..
cd cw-hooks-listener && cargo wasm && cd ..

# Or use the optimizer for production-ready small WASM
docker run --rm -v "$(pwd)":/code \
  --mount type=volume,source="$(basename "$(pwd)")_cache",target=/code/target \
  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
  cosmwasm/optimizer:0.16.0
```

### Step-by-step: Deploy Counter

```bash
# Set variables
CHAIN_ID="safro-testnet-1"
NODE="http://localhost:26657"
KEY="my-key"

# Create key if needed
safrochaind keys add $KEY --keyring-backend test

# Fund it (on a devnet, tokens are in the genesis)
# Or use the faucet account:
# safrochaind tx bank send faucet $(safrochaind keys show $KEY -a --keyring-backend test) 1000000usaf --chain-id $CHAIN_ID --keyring-backend test

# Store the WASM
RESP=$(safrochaind tx wasm store ./counter/target/wasm32-unknown-unknown/release/counter.wasm \
  --from $KEY --chain-id $CHAIN_ID --node $NODE --gas auto --gas-adjustment 1.4 --fees 5000usaf \
  -y --output json --keyring-backend test)
CODE_ID=$(echo $RESP | jq -r '.logs[0].events[] | select(.type == "store_code") | .attributes[] | select(.key == "code_id") | .value')
echo "Code ID: $CODE_ID"

# Instantiate
INIT='{"count":0}'
safrochaind tx wasm instantiate $CODE_ID "$INIT" --label "counter" --admin $(safrochaind keys show $KEY -a --keyring-backend test) \
  --from $KEY --chain-id $CHAIN_ID --node $NODE --gas auto --gas-adjustment 1.4 --fees 5000usaf \
  -y --output json --keyring-backend test
CONTRACT_ADDR=$(safrochaind query wasm list-contract-by-code $CODE_ID --node $NODE --output json | jq -r '.contracts[-1]')
echo "Contract: $CONTRACT_ADDR"

# Execute increment
safrochaind tx wasm execute $CONTRACT_ADDR '{"increment":{}}' \
  --from $KEY --chain-id $CHAIN_ID --node $NODE --gas auto --gas-adjustment 1.4 --fees 5000usaf \
  -y --keyring-backend test

# Query count
safrochaind query wasm contract-state smart $CONTRACT_ADDR '{"get_count":{}}' --node $NODE
# Expected: {"data":{"count":1}}
```

## Testing

Each contract has unit tests:

```bash
cd counter && cargo test && cd ..
cd cw20-mintable && cargo test && cd ..
cd clock-counter && cargo test && cd ..
cd cw-hooks-listener && cargo test && cd ..
```

## Troubleshooting

| Problem | Likely Cause | Fix |
|---|---|---|
| `out of gas` | Insufficient gas for WASM store | Increase `--gas-adjustment` to 1.5–2.0 |
| `wasm bytecode too large` | Contract exceeds 800KB limit | Use the Docker optimizer |
| `contract jailed` | Clock contract exceeded gas gas limit | Optimize the sudo handler |
| `not found: contract` | Wrong bech32 prefix | Verify `safrochaind` config uses `safro` prefix |
