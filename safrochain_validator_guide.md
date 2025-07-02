# Safrochain Validator Setup and Maintenance Guide

This guide provides a comprehensive and professional walkthrough to setting up and maintaining a validator node on the Safrochain testnet. It includes secure wallet creation, key management, validator registration, profile updates, recovery processes, unjailing procedures, redelegation, and reward handling.

---

## 1. Create a Wallet Using the `file` Keyring

A validator wallet is necessary to hold funds, sign transactions, and interact with the chain.

```bash
echo "Enter a name for your validator wallet (e.g., validator):"
read WALLET_NAME
export WALLET_NAME=$WALLET_NAME
safrochaind keys add "$WALLET_NAME" --keyring-backend file
```

> 🔐 **Important**: Copy and store the mnemonic phrase securely offline. This is your only recovery method.

To verify the wallet was created:

```bash
safrochaind keys list --keyring-backend file
```

Get the wallet address:

```bash
safrochaind keys show "$WALLET_NAME" -a --keyring-backend file
```

---

## 2. Secure Your Wallet with Encryption

To protect your key from unauthorized access, export and encrypt it:

```bash
safrochaind keys export "$WALLET_NAME" --keyring-backend file > "$WALLET_NAME"_key.json
```

Encrypt the exported key:

```bash
gpg -c "$WALLET_NAME"_key.json
```

> 💾 Store the encrypted `.gpg` file on a secure USB drive, external offline storage, or an encrypted cloud storage location.

---

## 3. Request SAF Test Tokens from the Faucet

Visit the faucet to fund your wallet with testnet tokens: [https://faucet.safrochain.com](https://faucet.safrochain.com)

Paste your wallet address and request funds. Then confirm the balance:

```bash
safrochaind query bank balances $(safrochaind keys show "$WALLET_NAME" -a --keyring-backend file)
```

---

## 4. Register Your Validator Node

### Ensure Your Node Is Fully Synced

```bash
curl http://localhost:26657/status | jq '.result.sync_info.catching_up'
```

Output must be `false` before proceeding.

### Get Validator Consensus Public Key

```bash
PUBKEY=$(safrochaind tendermint show-validator)
```

### Create Configuration File

```bash
cat > validator.json <<EOL
{
  "pubkey": $PUBKEY,
  "amount": "20000000saf",
  "moniker": "YourMoniker",
  "identity": "",
  "website": "",
  "security": "",
  "details": "My Safrochain testnet validator",
  "commission-rate": "0.1",
  "commission-max-rate": "0.2",
  "commission-max-change-rate": "0.01",
  "min-self-delegation": "1"
}
EOL
```

### Submit the Registration Transaction

```bash
safrochaind tx staking create-validator validator.json \
  --from "$WALLET_NAME" \
  --chain-id safrochain-testnet \
  --keyring-backend file \
  --fees 5000saf \
  --gas auto \
  --yes
```

---

## 5. Update Validator Profile (Metadata)

You can update your validator’s public metadata such as its name, website, description, and image by submitting a transaction using the `edit-validator` command. This does not affect your stake or validator status but improves how your validator appears in public explorers.

### Example Command:

```bash
safrochaind tx staking edit-validator \
  --moniker "NouveauNom" \
  --identity "your_keybase_id" \
  --website "https://yourwebsite.com" \
  --details "Professional validator on Safrochain" \
  --from "$WALLET_NAME" \
  --chain-id safrochain-testnet \
  --keyring-backend file \
  --fees 5000saf \
  --yes
```

### Explanation:

- `--moniker`: Public name of your validator node.
- `--identity`: Your **Keybase username** used to fetch your profile picture and proofs.
- `--website`: Link to your validator or team website.
- `--details`: Textual description of your validator's mission, goals, or experience.

### Setting Up Your Keybase ID

To show your profile image and build credibility:

1. Visit [https://keybase.io](https://keybase.io)
2. Create an account and choose a unique **username** — this becomes your `identity`.
3. Upload a profile picture and add verifications (GitHub, website, etc.).
4. Make sure your profile is public.
5. You can test your profile at:

```text
https://keybase.io/<your_username>
```

Once set, use that Keybase ID in the `--identity` flag when editing your validator.

> ✅ Explorers like Mintscan or Big Dipper may use this information to display an image, website, and social verification next to your validator.

To change your validator name, description, website, or image identity:

```bash
safrochaind tx staking edit-validator \
  --moniker "NewMoniker" \
  --identity "YourKeybaseID" \
  --website "https://yourwebsite.com" \
  --details "Professional validator operator." \
  --from "$WALLET_NAME" \
  --chain-id safrochain-testnet \
  --keyring-backend file \
  --fees 5000saf \
  --yes
```

> 🎨 The `identity` should be your [Keybase](https://keybase.io) profile to enable validator image display on explorers.

---

## 6. Unjail a Validator (After Downtime)

If your validator has been jailed due to inactivity or missed blocks:

```bash
safrochaind tx slashing unjail \
  --from "$WALLET_NAME" \
  --chain-id safrochain-testnet \
  --keyring-backend file \
  --fees 5000saf \
  --gas auto \
  --yes
```

Verify validator status:

```bash
export VALOPER_ADDRESS=$(safrochaind keys show "$WALLET_NAME" --bech val -a --keyring-backend file)
safrochaind query staking validator $VALOPER_ADDRESS
```

---

## 7. Restore Wallet from Mnemonic

If you lost access to your machine but have the 24-word mnemonic, recover your wallet:

```bash
safrochaind keys add "$WALLET_NAME" --recover --keyring-backend file
```

Paste your mnemonic phrase when prompted.

---

## 8. Redelegate to Reactivate an Unbonding Validator

If the validator is unbonding, you must redelegate to bring it back online:

```bash
safrochaind tx staking delegate $VALOPER_ADDRESS 20000000saf \
  --from "$WALLET_NAME" \
  --keyring-backend file \
  --chain-id safrochain-testnet \
  --fees 5000saf \
  --gas auto \
  --yes
```

---

## 9. Validator Monitoring: Status, Balance, Rewards

Set permanent environment variables:

```bash
export DELEGATOR_ADDRESS=$(safrochaind keys show "$WALLET_NAME" -a --keyring-backend file)
export VALOPER_ADDRESS=$(safrochaind keys show "$WALLET_NAME" --bech val -a --keyring-backend file)
```

### Check Validator Status

```bash
safrochaind query staking validator $VALOPER_ADDRESS
```

### Check Wallet Balance

```bash
safrochaind query bank balances $DELEGATOR_ADDRESS
```

### Query Delegator Rewards

```bash
safrochaind query distribution rewards $DELEGATOR_ADDRESS
```

### Withdraw Validator Rewards

```bash
safrochaind tx distribution withdraw-rewards $VALOPER_ADDRESS \
  --from "$WALLET_NAME" \
  --commission \
  --chain-id safrochain-testnet \
  --keyring-backend file \
  --fees 5000saf \
  --yes
```

---

## 10. Best Practices and Backup Strategy

- 🧠 **Always back up your mnemonic securely** (offline or in a fireproof safe).
- 🔒 **Encrypt all exported keys** and never leave them on an internet-connected machine.
- 🧪 **Test your recovery plan** at least once on a test machine.
- ⚙️ **Use monitoring tools** like Prometheus/Grafana for uptime and slash protection.
- 📁 **Organize key material** in a designated, protected folder with restricted permissions.

---

Need help? Join the Safrochain community:

- [Telegram](https://t.me/safrochain)
- [Discord](https://discord.gg/safrochain)
- [GitHub](https://github.com/safrochain)

Stay secure, stay active — and happy validating on Safrochain 🚀

