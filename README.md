# Safrochain Node

Safrochain is a community-first blockchain built with the Cosmos SDK. This repository hosts the
implementation of the Safrochain full node, `safrochaind`: consensus, application state and
network logic.

> ⚙️ Run a node. 🛠️ Build with us. 🧩 Connect Africa to the world.

Documentation: **[docs.safrochain.com](https://docs.safrochain.com)**

---

## 🚀 Features

- [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) v0.50 with [CometBFT](https://github.com/cometbft/cometbft) v0.38
- CosmWasm smart contracts ([wasmd](https://github.com/CosmWasm/wasmd) v0.54)
- IBC v8: ICS-20 transfers, interchain accounts, packet forwarding, IBC hooks
- Governance, staking, fee grants, authz
- Safrochain modules: `x/tokenfactory`, `x/globalfee`, `x/feepay`, `x/feeshare`, `x/drip`,
  `x/clock`, `x/cw-hooks`, `x/wrappers`

---

## 🌐 Networks

| Network | Chain ID | Version | Go |
| --- | --- | --- | --- |
| Mainnet | `safrochain-1` | `v0.2.2`, then **`v0.3.0`** from block 4,557,100 | 1.25.8 (`v0.2.2`), 1.26.11 (`v0.3.0`) |
| Testnet | `safro-testnet-1` | `release/v0.1.0` | 1.23.9 |

Mainnet upgrades from `v0.2.2` to `v0.3.0` at block **4,557,100** (about Wednesday 18 November
2026, 14:00 UTC). Validators and node operators: follow the
[v0.3.0 upgrade guide](https://docs.safrochain.com/governance/mainnet-upgrade-v0.3.0).

- **Binary:** `safrochaind`
- **Denom:** `usaf` (1 SAF = 10^6 `usaf`)
- **Bech32 prefix:** `addr_safro`

---

## 🛠️ Installation

### Build from source

Prerequisites: Git, Make, a C toolchain (`build-essential` on Debian/Ubuntu, Xcode command line
tools on macOS) and Go 1.21 or newer. `go.mod` pins Go **1.26.11**; Go downloads that toolchain
by itself on the first build.

```bash
git clone https://github.com/Safrochain-Org/safrochain-node.git
cd safrochain-node
git checkout v0.3.0        # always build a release tag, never main
make install               # installs safrochaind into $(go env GOPATH)/bin
safrochaind version --long | grep -E '^(version|commit|go):'
```

Use `make build` instead of `make install` to write the binary to `./bin/safrochaind` without
replacing the one your node runs (for example to stage an upgrade).

### Release binaries

Statically linked `linux/amd64` and `linux/arm64` binaries, built by CI from each version tag, are
attached to the [GitHub releases](https://github.com/Safrochain-Org/safrochain-node/releases) with
a `SHA256SUMS` file, and mirrored on `https://rpc1.safrochain.network/binaries/<version>/`:

```bash
curl -fLO https://github.com/Safrochain-Org/safrochain-node/releases/download/v0.3.0/safrochaind-v0.3.0-linux-amd64
curl -fLO https://github.com/Safrochain-Org/safrochain-node/releases/download/v0.3.0/SHA256SUMS
sha256sum -c --ignore-missing SHA256SUMS
```

---

## ⛓️ Run a node

- [Join mainnet](https://docs.safrochain.com/run-a-node/join-mainnet) (`safrochain-1`)
- [Join testnet](https://docs.safrochain.com/run-a-node/join-testnet) (`safro-testnet-1`)
- [Upgrades with cosmovisor](https://docs.safrochain.com/run-a-node/upgrades)
- [Become a validator](https://docs.safrochain.com/validators/become-a-validator)

### Local single-node network

For development only (the chain ID must not be one of the public networks):

```bash
safrochaind init localnode --chain-id safro-local-1
safrochaind keys add validator --keyring-backend test
safrochaind genesis add-genesis-account validator 1000000000usaf --keyring-backend test
safrochaind genesis gentx validator 700000000usaf --chain-id safro-local-1 --keyring-backend test
safrochaind genesis collect-gentxs
safrochaind start
```

Default paths: configuration in `~/.safrochain/config/`, data in `~/.safrochain/data/`.

---

## 📡 Public endpoints (mainnet)

| Service | Endpoint |
| --- | --- |
| RPC | https://rpc.safrochain.network |
| REST | https://api.safrochain.network |
| gRPC | https://grpc.safrochain.network |
| Explorer | https://explorer.safrochain.com |

All endpoints, including the testnet ones: [Networks](https://docs.safrochain.com/networks/mainnet-endpoints).

---

## 🧪 Testing

```bash
make test                  # unit and integration tests
make ictest-basic          # end-to-end tests (interchaintest, requires Docker)
```

If you change protocol-level code, start from `app/app.go` and `app/keepers/`.

---

## 🔐 Security

Please report vulnerabilities privately as described in [SECURITY.md](./SECURITY.md). Do not open
public issues for security problems.

---

## 🧑‍🤝‍🧑 Community

- **Discord:** https://discord.gg/fe2XAm6ENQ
- **X:** [@safrochain](https://x.com/safrochain)
- **Explorer:** https://explorer.safrochain.com
- **Documentation:** https://docs.safrochain.com

---

## 📄 License

This project is licensed under [Apache 2.0](./LICENSE).

---

## 🤝 Contributing

Pull requests are welcome. For major changes, please open an issue first. See
[CONTRIBUTING.md](./CONTRIBUTING.md).

---

## 🙏 Credits

Safrochain is maintained by the [Safrochain team](https://github.com/Safrochain-Org).

We thank the [Juno](https://github.com/CosmosContracts/juno) team for their open-source work.
**Safrochain is a fork of Juno** and benefits greatly from the Cosmos and Juno ecosystems.
