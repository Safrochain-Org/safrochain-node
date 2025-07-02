# Safrochain Node

Safrochain is a community-first blockchain built using the Cosmos SDK. This repository hosts the core implementation of the Safrochain full node — including consensus, application state, and network logic.

> ⚙️ Run a node. 🛠️ Build with us. 🧩 Connect Africa to the world.

---

## 🚀 Features

- Built with [Cosmos SDK v0.50+](https://github.com/cosmos/cosmos-sdk)
- IBC-enabled for cross-chain compatibility
- WASM Smart Contracts (via CosmWasm)
- Governance, Staking, Token module, Oracle integration
- Custom modules: Likelemba, XPoints, USAF Faucet

---

## 📦 Binary

- **Binary name:** `safrochaind`
- **Chain ID:** `safro-testnet-1` or `safro-mainnet-1`
- **Denom:** `usaf` (1 SAF = 10^6 micro `usaf`)
- **Token metadata:** Available in `config/genesis.json`

---

## 🛠️ Installation

### Prerequisites

- Go >= 1.22.x
- Git
- GCC (for protobuf & Wasm)

```bash
git clone https://github.com/danbaruka/safrochain-node.git
cd safrochain-node
make install
```

> This will install the binary `safrochaind` into your `$GOPATH/bin`.

---

## ⛓️ Running a Local Node (Testnet)

```bash
safrochaind init localnode --chain-id safro-testnet-1
safrochaind keys add validator
safrochaind add-genesis-account $(safrochaind keys show validator -a) 1000000000usaf
safrochaind gentx validator 700000000usaf --chain-id safro-testnet-1
safrochaind collect-gentxs
safrochaind start
```

---

## 🔧 Configuration

Default paths:

- Config files: `~/.safrochain/config/`
- Data: `~/.safrochain/data/`

To modify ports or peers, edit `~/.safrochain/config/config.toml`.

---

## 📡 Public Endpoints (Testnet)

| Service     | Endpoint                                     |
|-------------|----------------------------------------------|
| Chain ID    | `safro-testnet-1`                            |
| RPC         | https://rpc.testnet.safrochain.com           |
| REST API    | https://rest.testnet.safrochain.com          |
| gRPC        | https://grpc.testnet.safrochain.com/         |
| Faucet A    | `#testnet-faucet` on [Discord](https://discord.gg/YOUR_INVITE) |
| Faucet B    | https://faucet.safrochain.com                |

---

## 🧪 Testing

```bash
make test
```

---

## 🧱 Build From Source

```bash
make install
```

If you're making protocol-level changes, see `app/app.go` and `app/app_config.go`.

---

## 🧑‍🤝‍🧑 Community

- **Discord**: [Safrochain Discord](https://discord.gg/YOUR_INVITE)
- **Twitter**: [@Safrochain](https://twitter.com/safrochain)
- **Block Explorer**: Coming soon
- **Faucet**: `!request <address>` on Discord or [faucet.safrochain.com](https://faucet.safrochain.com)

---

## 📄 License

This project is licensed under [Apache 2.0](./LICENSE).

---

## 🤝 Contributing

Pull requests are welcome! For major changes, please open an issue first.

- Fork the repo
- Create your branch (`git checkout -b feature/xyz`)
- Commit your changes (`git commit -am 'feat: add xyz'`)
- Push (`git push origin feature/xyz`)
- Open a PR

---

## 🙏 Credits

Safrochain is proudly maintained by [Safrochain Team](https://github.com/safrochain).

We thank the [Juno](https://github.com/CosmosContracts/juno) team for their open-source contributions — **Safrochain is a fork of Juno** and benefits greatly from the Cosmos and Juno ecosystems.
