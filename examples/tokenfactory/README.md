# Token Factory End-to-End Example

This example demonstrates the complete lifecycle of a **Token Factory** denom on Safrochain —
from creation through minting, transfer, metadata update, and burn — all using `safrochaind` CLI commands.

## Prerequisites

- [Safrochain Node](https://github.com/Safrochain-Org/safrochain-node) built and installed (`make install`)
- `safrochaind` in your `$PATH`
- Two key pairs for testing: `creator` and `recipient`

## Quick Start

```bash
# Build safrochaind (from repo root)
make install

# Start a fresh local node
safrochaind init test-node --chain-id safro-testnet-1
safrochaind keys add creator --keyring-backend test
safrochaind keys add recipient --keyring-backend test
safrochaind genesis add-genesis-account $(safrochaind keys show creator -a --keyring-backend test) 1000000000000usaft
safrochaind genesis add-genesis-account $(safrochaind keys show recipient -a --keyring-backend test) 1000000000000usaft
safrochaind genesis gentx creator 700000000usaft --chain-id safro-testnet-1 --keyring-backend test
safrochaind genesis collect-gentxs
safrochaind start
```

Then run the example script:

```bash
chmod +x examples/tokenfactory/run.sh
./examples/tokenfactory/run.sh
```

> **Note:** The denom creation fee is `10,000,000 usaft` (default). Make sure the creator has enough balance.

## Step-by-Step Walkthrough

### 1. Setup

```bash
export CHAIN_ID="safro-testnet-1"
export KEYRING="--keyring-backend test"
export NODE="http://localhost:26657"
```

Prepare variables for the creator and recipient addresses:

```bash
CREATOR=$(safrochaind keys show creator -a $KEYRING)
RECIPIENT=$(safrochaind keys show recipient -a $KEYRING)
```

Check balances to confirm both accounts are funded:

```bash
safrochaind query bank balances $CREATOR --node $NODE
safrochaind query bank balances $RECIPIENT --node $NODE
```

### 2. Create a Denom

Create a new subdenom called `mytoken`:

```bash
safrochaind tx tokenfactory create-denom $CREATOR mytoken   --from $CREATOR $KEYRING   --chain-id $CHAIN_ID   --node $NODE   --fees 5000usaft   -y
```

The resulting denom will be `factory/{creator_address}/mytoken`. Query the creation fee first:

```bash
safrochaind query tokenfactory params --node $NODE
```

List all denoms created by your address:

```bash
safrochaind query tokenfactory denoms-from-creator $CREATOR --node $NODE
```

### 3. Set Denom Metadata

Set the display metadata so wallets and explorers render it nicely:

```bash
safrochaind tx tokenfactory modify-metadata $CREATOR '{
  "description": "My first token on Safrochain",
  "denom_units": [
    {"denom": "factory/'"$CREATOR"'/mytoken", "exponent": 0, "aliases": ["umytoken"]},
    {"denom": "mytoken", "exponent": 6, "aliases": []}
  ],
  "base": "factory/'"$CREATOR"'/mytoken",
  "display": "mytoken",
  "name": "MyToken",
  "symbol": "MYT"
}'   --from $CREATOR $KEYRING   --chain-id $CHAIN_ID   --node $NODE   --fees 5000usaft   -y
```

Verify the metadata:

```bash
safrochaind query bank denom-metadata --denom factory/$CREATOR/mytoken --node $NODE
```

### 4. Mint Tokens

Mint 1000 tokens to the creator's own address:

```bash
DENOM=factory/$CREATOR/mytoken
safrochaind tx tokenfactory mint $CREATOR 1000$DENOM $CREATOR   --from $CREATOR $KEYRING   --chain-id $CHAIN_ID   --node $NODE   --fees 5000usaft   -y
```

Check the balance:

```bash
safrochaind query bank balances $CREATOR --denom $DENOM --node $NODE
```

### 5. Mint to a Recipient

Mint 500 tokens directly to the recipient address:

```bash
safrochaind tx tokenfactory mint $CREATOR 500$DENOM $RECIPIENT   --from $CREATOR $KEYRING   --chain-id $CHAIN_ID   --node $NODE   --fees 5000usaft   -y
```

Verify:

```bash
safrochaind query bank balances $RECIPIENT --denom $DENOM --node $NODE
```

### 6. Transfer (plain `bank send`)

The new denom behaves like any native bank coin:

```bash
safrochaind tx bank send $CREATOR $RECIPIENT 200$DENOM   --from $CREATOR $KEYRING   --chain-id $CHAIN_ID   --node $NODE   --fees 5000usaft   -y
```

Check recipient now has 700 (500 minted + 200 transferred):

```bash
safrochaind query bank balances $RECIPIENT --denom $DENOM --node $NODE
```

### 7. Burn Tokens

Burn 100 tokens from the creator's address:

```bash
safrochaind tx tokenfactory burn $CREATOR 100$DENOM $CREATOR   --from $CREATOR $KEYRING   --chain-id $CHAIN_ID   --node $NODE   --fees 5000usaft   -y
```

Check the creator balance decreased:

```bash
safrochaind query bank balances $CREATOR --denom $DENOM --node $NODE
```

### 8. Change Admin

Transfer admin control to the recipient. The admin can mint, burn, and force-transfer.

```bash
safrochaind tx tokenfactory change-admin $CREATOR $DENOM $RECIPIENT   --from $CREATOR $KEYRING   --chain-id $CHAIN_ID   --node $NODE   --fees 5000usaft   -y
```

Verify the new admin:

```bash
safrochaind query tokenfactory denom-authority-metadata $DENOM --node $NODE
```

Now the recipient can mint on behalf of the denom:

```bash
safrochaind tx tokenfactory mint $RECIPIENT 300$DENOM $RECIPIENT   --from $RECIPIENT $KEYRING   --chain-id $CHAIN_ID   --node $NODE   --fees 5000usaft   -y
```

To **renounce admin** (irreversible), set the admin to a burn address:

```bash
# Only do this if you are certain!
# safrochaind tx tokenfactory change-admin $RECIPIENT $DENOM addr_safro1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqp3q6qq #   --from $RECIPIENT $KEYRING #   --chain-id $CHAIN_ID #   --node $NODE #   --fees 5000usaft #   -y
```

### 9. Query Options

Query all parameters:

```bash
safrochaind query tokenfactory params --node $NODE
```

Get denom authority metadata:

```bash
safrochaind query tokenfactory denom-authority-metadata $DENOM --node $NODE
```

List denoms for a creator:

```bash
safrochaind query tokenfactory denoms-from-creator $CREATOR --node $NODE
```

Query bank metadata:

```bash
safrochaind query bank denom-metadata --denom $DENOM --node $NODE
```

Query total supply:

```bash
safrochaind query bank total --denom $DENOM --node $NODE --count-total
```

### 10. Cleanup

Stop the local node (`Ctrl+C`). To fully reset:

```bash
pkill safrochaind 2>/dev/null || true
rm -rf ~/.safrochain
```

## Key Concepts

| Concept | Description |
|---------|-------------|
| **Denom format** | `factory/{creator_addr}/{subdenom}` — globally unique by namespace |
| **Admin** | The account that created the denom (or a successor via `change-admin`) |
| **Mint authority** | Only the current admin can mint |
| **Burn authority** | Only the current admin can burn from any address |
| **Force transfer** | Admin can move tokens between any two addresses |
| **Renunciation** | Set admin to a burn address to make the denom permanently immutable |
| **Creation fee** | Configurable via `Params.DenomCreationFee` (default: 10,000,000 `usaft`) |

## References

- [Token Factory Module Source](../../x/tokenfactory/)
- [Cosmos SDK Bank Module](https://docs.cosmos.network/main/modules/bank/)
- [Safrochain Documentation](https://docs.safrochain.com)
