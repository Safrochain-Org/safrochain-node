#!/usr/bin/env bash
# -------------------------------------------------------------------
# examples/tokenfactory/run.sh
#
# End-to-end Token Factory lifecycle demonstration.
# Run this against a local safro-testnet-1 node that already has
# 'creator' and 'recipient' keys funded with usaft.
#
# Prerequisites:
#   1. safrochaind installed (make install from repo root)
#   2. Local node running: safrochaind start
#   3. Keys exist:
#        safrochaind keys add creator --keyring-backend test
#        safrochaind keys add recipient --keyring-backend test
#      (or run the setup section below)
#
# Usage:
#   chmod +x examples/tokenfactory/run.sh
#   ./examples/tokenfactory/run.sh
# -------------------------------------------------------------------
set -euo pipefail

# --------------- Configuration ----------------
CHAIN_ID="${CHAIN_ID:-safro-testnet-1}"
KEYRING="${KEYRING:---keyring-backend test}"
NODE="${NODE:-http://localhost:26657}"
FEES="${FEES:-5000usaft}"
GAS="${GAS:-auto}"
CREATOR_KEY="${CREATOR_KEY:-creator}"
RECIPIENT_KEY="${RECIPIENT_KEY:-recipient}"
SUB_DENOM="${SUB_DENOM:-mytoken}"
# ------------------------------------------------

C_RESET="\033[0m"
C_BOLD="\033[1m"
C_GREEN="\033[38;5;82m"
C_CYAN="\033[38;5;51m"
C_YELLOW="\033[38;5;221m"
C_RED="\033[38;5;196m"

info()  { printf "${C_CYAN}[INFO]${C_RESET}  %s\n" "$*"; }
ok()    { printf "${C_GREEN}[OK]${C_RESET}    %s\n" "$*"; }
warn()  { printf "${C_YELLOW}[WARN]${C_RESET}  %s\n" "$*"; }
fail()  { printf "${C_RED}[FAIL]${C_RESET}  %s\n" "$*"; exit 1; }
step()  { printf "\n${C_BOLD}[STEP %s]${C_RESET} %s\n" "$1" "$2"; }

run_tx() {
    safrochaind tx "$@" \
        --chain-id "$CHAIN_ID" \
        --node "$NODE" \
        --fees "$FEES" \
        --gas "$GAS" \
        $KEYRING \
        -y 2>&1
}

# ---- 0. Setup ----
step "0" "Resolve addresses"

CREATOR=$(safrochaind keys show "$CREATOR_KEY" -a $KEYRING 2>/dev/null) ||     fail "Key '$CREATOR_KEY' not found. Create it with: safrochaind keys add $CREATOR_KEY $KEYRING"
RECIPIENT=$(safrochaind keys show "$RECIPIENT_KEY" -a $KEYRING 2>/dev/null) ||     fail "Key '$RECIPIENT_KEY' not found. Create it with: safrochaind keys add $RECIPIENT_KEY $KEYRING"

info "Creator:   $CREATOR"
info "Recipient: $RECIPIENT"

DENOM="factory/${CREATOR}/${SUB_DENOM}"

# ---- 1. Check Balances ----
step "1" "Check initial balances"

safrochaind query bank balances "$CREATOR" --node "$NODE" --denom usaft 2>/dev/null
safrochaind query bank balances "$RECIPIENT" --node "$NODE" --denom usaft 2>/dev/null

# ---- 2. Create Denom ----
step "2" "Create denom: $DENOM"

run_tx tokenfactory create-denom "$CREATOR" "$SUB_DENOM" --from "$CREATOR"
info "Denom created: $DENOM"

# Verify creation
safrochaind query tokenfactory denoms-from-creator "$CREATOR" --node "$NODE"
ok "Denom created successfully"

# ---- 3. Set Denom Metadata ----
step "3" "Set denom metadata"

METADATA=$(cat <<EOF
{
  "description": "My first token created via Safrochain Token Factory",
  "denom_units": [
    {"denom": "$DENOM", "exponent": 0, "aliases": ["u$SUB_DENOM"]},
    {"denom": "$SUB_DENOM", "exponent": 6, "aliases": []}
  ],
  "base": "$DENOM",
  "display": "$SUB_DENOM",
  "name": "MyToken",
  "symbol": "MYT"
}
EOF
)

run_tx tokenfactory modify-metadata "$CREATOR" "$METADATA" --from "$CREATOR"

# Verify metadata
safrochaind query bank denom-metadata --denom "$DENOM" --node "$NODE" 2>/dev/null
ok "Denom metadata set"

# ---- 4. Mint Tokens to Creator ----
step "4" "Mint 1000 $SUB_DENOM to creator"

run_tx tokenfactory mint "$CREATOR" "1000$DENOM" "$CREATOR" --from "$CREATOR"

BALANCE=$(safrochaind query bank balances "$CREATOR" --denom "$DENOM" --node "$NODE" -o json 2>/dev/null |     python3 -c "import sys,json; print(json.load(sys.stdin).get('amount','?'))" 2>/dev/null || echo "?")
info "Creator balance: $BALANCE $DENOM"

# ---- 5. Mint Tokens to Recipient ----
step "5" "Mint 500 $SUB_DENOM directly to recipient"

run_tx tokenfactory mint "$CREATOR" "500$DENOM" "$RECIPIENT" --from "$CREATOR"

RECIP_BAL=$(safrochaind query bank balances "$RECIPIENT" --denom "$DENOM" --node "$NODE" -o json 2>/dev/null |     python3 -c "import sys,json; print(json.load(sys.stdin).get('amount','?'))" 2>/dev/null || echo "?")
info "Recipient balance: $RECIP_BAL $DENOM"

# ---- 6. Bank Send (plain transfer) ----
step "6" "Transfer 200 $SUB_DENOM from creator to recipient via bank send"

run_tx bank send "$CREATOR" "$RECIPIENT" "200$DENOM" --from "$CREATOR"

RECIP_BAL2=$(safrochaind query bank balances "$RECIPIENT" --denom "$DENOM" --node "$NODE" -o json 2>/dev/null |     python3 -c "import sys,json; print(json.load(sys.stdin).get('amount','?'))" 2>/dev/null || echo "?")
info "Recipient balance after transfer: $RECIP_BAL2 $DENOM"

# ---- 7. Burn Tokens ----
step "7" "Burn 100 $SUB_DENOM from creator"

run_tx tokenfactory burn "$CREATOR" "100$DENOM" "$CREATOR" --from "$CREATOR"

BALANCE2=$(safrochaind query bank balances "$CREATOR" --denom "$DENOM" --node "$NODE" -o json 2>/dev/null |     python3 -c "import sys,json; print(json.load(sys.stdin).get('amount','?'))" 2>/dev/null || echo "?")
info "Creator balance after burn: $BALANCE2 $DENOM"

# ---- 8. Force Transfer (admin power) ----
step "8" "Force-transfer 50 $SUB_DENOM from recipient back to creator"

run_tx tokenfactory force-transfer "$CREATOR" "50$DENOM" "$RECIPIENT" "$CREATOR" --from "$CREATOR"

RECIP_BAL3=$(safrochaind query bank balances "$RECIPIENT" --denom "$DENOM" --node "$NODE" -o json 2>/dev/null |     python3 -c "import sys,json; print(json.load(sys.stdin).get('amount','?'))" 2>/dev/null || echo "?")
info "Recipient after force-transfer: $RECIP_BAL3 $DENOM"

# ---- 9. Change Admin ----
step "9" "Change admin from creator to recipient"

run_tx tokenfactory change-admin "$CREATOR" "$DENOM" "$RECIPIENT" --from "$CREATOR"

# Verify new admin
safrochaind query tokenfactory denom-authority-metadata "$DENOM" --node "$NODE"
ok "Admin changed to $RECIPIENT"

# Recipient can now mint (as the new admin)
step "9b" "Recipient (new admin) mints 200 more $SUB_DENOM"

run_tx tokenfactory mint "$RECIPIENT" "200$DENOM" "$RECIPIENT" --from "$RECIPIENT"

RECIP_BAL4=$(safrochaind query bank balances "$RECIPIENT" --denom "$DENOM" --node "$NODE" -o json 2>/dev/null |     python3 -c "import sys,json; print(json.load(sys.stdin).get('amount','?'))" 2>/dev/null || echo "?")
info "Recipient final balance: $RECIP_BAL4 $DENOM"

# ---- 10. Summary ----
step "10" "Summary"

echo ""
echo "  Token Factory demonstration complete!                    "
echo "                                                           "
echo "  Denom:        $DENOM"
echo "  Creator:      $CREATOR"
echo "  Recipient:    $RECIPIENT"
echo "                                                           "
echo "  Commands demonstrated:                                   "
echo "    ✅ create-denom     Create a new denom                 "
echo "    ✅ modify-metadata  Set display metadata               "
echo "    ✅ mint             Mint tokens to any address         "
echo "    ✅ bank send        Standard Cosmos transfer            "
echo "    ✅ burn             Burn tokens from any address       "
echo "    ✅ force-transfer   Admin-powered transfer              "
echo "    ✅ change-admin     Transfer admin authority            "
echo "                                                           "
echo "  To query at any point:                                    "
echo "    safrochaind query bank balances <addr>                  "
echo "    safrochaind query tokenfactory params                   "
echo "    safrochaind query tokenfactory denom-authority-metadata "
echo "    safrochaind query tokenfactory denoms-from-creator      "
echo "                                                           "
