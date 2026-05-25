# Contributing to Safrochain Node

Thank you for your interest in contributing to Safrochain — a community-first blockchain connecting Africa to the world. Every contribution, no matter the size, makes a difference.

> **Before you start:** Please read [SECURITY.md](./SECURITY.md) for responsible disclosure of vulnerabilities. Do **not** open public issues for security bugs.

---

## Table of Contents

- [Ways to Contribute](#ways-to-contribute)
- [Development Environment](#development-environment)
- [Workflow](#workflow)
- [Commit Conventions](#commit-conventions)
- [Testing](#testing)
- [Code Style](#code-style)
- [Protobuf Development](#protobuf-development)
- [Module Guide](#module-guide)
- [Pull Request Checklist](#pull-request-checklist)
- [Review Process](#review-process)
- [Community](#community)

---

## Ways to Contribute

| Type | How |
|---|---|
| **Bug report** | [Open an issue](https://github.com/Safrochain-Org/safrochain-node/issues/new?template=bug_report.md) with steps to reproduce |
| **Feature request** | [Open an issue](https://github.com/Safrochain-Org/safrochain-node/issues/new?template=feature_request.md) before building |
| **Documentation** | Edit markdown files or inline godoc comments and open a PR |
| **Code** | Fork → branch → PR (details below) |
| **Testing** | Add or improve unit / interchain tests |
| **Security** | See [SECURITY.md](./SECURITY.md) — private disclosure only |

For significant changes, **open an issue first** so the approach can be agreed on before you invest coding time.

---

## Development Environment

### Prerequisites

| Tool | Minimum version | Notes |
|---|---|---|
| Go | `1.22` | Check with `go version` |
| Git | any recent | — |
| GCC | any | Required for CosmWasm & Ledger |
| Docker | any | Required for interchain tests & proto generation |
| `buf` | `1.x` | For protobuf work only |

### Fork & clone

```bash
# 1. Fork on GitHub, then:
git clone https://github.com/<your-username>/safrochain-node.git
cd safrochain-node
git remote add upstream https://github.com/Safrochain-Org/safrochain-node.git
```

### Build & install

```bash
make install          # build + install safrochaind into $GOPATH/bin
make build            # build only, binary at ./build/safrochaind
```

### Verify the installation

```bash
safrochaind version
make print-summary    # shows version, commit, SDK version, CometBFT version
```

### Run a local node for manual testing

```bash
safrochaind init localnode --chain-id safro-testnet-1
safrochaind keys add validator
safrochaind add-genesis-account $(safrochaind keys show validator -a) 1000000000usaf
safrochaind gentx validator 700000000usaf --chain-id safro-testnet-1
safrochaind collect-gentxs
safrochaind start
```

---

## Workflow

```
upstream/main
      │
      ▼
  your fork ──▶ feature/my-change ──▶ Pull Request ──▶ upstream/main
```

1. **Sync with upstream** before starting work:

   ```bash
   git fetch upstream
   git checkout main
   git merge upstream/main
   ```

2. **Create a branch** using the naming convention below.

3. **Make your changes**, commit early and often.

4. **Push and open a PR** against `upstream/main`.

### Branch naming

```
<type>/<short-description>
```

| Prefix | When to use |
|---|---|
| `feat/` | New feature or module |
| `fix/` | Bug fix |
| `chore/` | Build, CI, dependency, tooling |
| `docs/` | Documentation only |
| `test/` | Tests only |
| `refactor/` | Code restructuring, no behaviour change |
| `security/` | Security patches |

Examples: `feat/xpoints-module`, `fix/clock-epoch-overflow`, `docs/contributing`

---

## Commit Conventions

This repo follows **Conventional Commits**. The format is:

```
<type>(<optional scope>): <short summary>

[optional body]

[optional footer(s)]
```

**Types:** `feat` · `fix` · `chore` · `docs` · `test` · `refactor` · `security` · `perf`

**Scopes** (optional, use the affected area): `x/tokenfactory` · `x/feeshare` · `x/clock` · `app` · `cmd` · `proto` · `ci`

### Examples

```
feat(x/tokenfactory): add MsgBurnTokens handler

fix(app): prevent nil pointer in BeginBlock under empty block

chore(ci): switch dependabot schedule to monthly

docs: add top-level CONTRIBUTING guide
```

Rules:
- Summary line ≤ 72 characters, lowercase, no trailing period.
- Use the imperative mood ("add", "fix", "remove" — not "added").
- Reference issues in the footer: `Closes #42`, `Refs #17`.

---

## Testing

### Unit tests

```bash
make test
```

Runs `go test ./...` across the entire module. All new code must be accompanied by meaningful unit tests.

### Interchain tests (Docker required)

Interchain tests spin up real nodes inside Docker and cover end-to-end upgrade paths, IBC, and module behaviour.

```bash
make ictest-basic          # basic chain liveness
make ictest-ibc            # IBC packet relay
make ictest-tokenfactory   # x/tokenfactory end-to-end
make ictest-feeshare       # x/feeshare end-to-end
make ictest-feepay         # x/feepay end-to-end
make ictest-globalfee      # x/globalfee end-to-end
make ictest-cwhooks        # x/cw-hooks end-to-end
make ictest-clock          # x/clock end-to-end
make ictest-drip           # x/drip end-to-end
make ictest-upgrade        # chain upgrade path
make ictest-statesync      # state-sync bootstrap
make ictest-pfm            # packet-forward middleware
```

> Run `make rm-testcache` before any interchain test to guarantee a clean run.

### Writing tests

- Unit tests live alongside source files (`foo_test.go` next to `foo.go`).
- Interchain tests live in `interchaintest/`. See an existing test (e.g. `interchaintest/basic_test.go`) for the setup pattern.
- Do **not** use mocks for database or consensus layer interactions — use real in-process state or Docker nodes so failures surface early.

---

## Code Style

### Formatting

```bash
make format          # runs gofmt + goimports
```

All submitted code must be formatted. CI will reject unformatted diffs.

### Linting

```bash
make install-lint    # install golangci-lint (one-time)
make lint            # run the linter
```

The linter configuration lives in `.golangci.yml`. Fix all reported issues before opening a PR — do not suppress linter warnings without a comment explaining why.

### General guidelines

- Keep functions focused; prefer many small, named functions over large anonymous closures.
- Avoid unnecessary abstractions. Three similar lines are better than a premature helper.
- Do not add error handling for scenarios that cannot happen. Trust the SDK's guarantees.
- Comments should explain *why*, not *what*. Well-named identifiers carry the *what*.
- No half-finished implementations. If something is a stub, mark it with a `TODO(yourhandle):` comment and open a follow-up issue.

---

## Protobuf Development

Protobuf sources live in `proto/`. Generated code is committed to the repo — do **not** edit generated files by hand.

### Regenerate from `.proto` sources

```bash
make proto-all       # format + lint + generate Go + generate Swagger
```

Individual steps:

```bash
make proto-format    # clang-format all .proto files
make proto-lint      # buf lint
make proto-gen       # generate Go types
make proto-swagger-gen  # generate REST swagger docs
```

Requires Docker (the `ghcr.io/cosmos/proto-builder` image is pulled automatically).

---

## Module Guide

Safrochain ships several custom Cosmos SDK modules on top of the standard set. If your PR touches one of these, make sure the relevant interchain test still passes.

| Module | Path | Purpose |
|---|---|---|
| `x/clock` | `x/clock/` | Block-epoch hooks for smart contracts |
| `x/cw-hooks` | `x/cw-hooks/` | CosmWasm contract lifecycle hooks |
| `x/drip` | `x/drip/` | Token drip / distribution primitive |
| `x/feepay` | `x/feepay/` | Sponsored transaction fee payments |
| `x/feeshare` | `x/feeshare/` | Contract fee revenue sharing |
| `x/globalfee` | `x/globalfee/` | Chain-wide minimum fee floor |
| `x/tokenfactory` | `x/tokenfactory/` | Permissionless token minting |
| `x/wrappers` | `x/wrappers/` | SDK message wrappers / helpers |

The app wiring lives in [`app/app.go`](./app/app.go) and [`app/app_config.go`](./app/app_config.go).

---

## Pull Request Checklist

Before marking your PR ready for review:

- [ ] Branch is up to date with `upstream/main`
- [ ] `make install` succeeds with no errors
- [ ] `make lint` and `make format` pass with no outstanding issues
- [ ] `make test` passes
- [ ] Relevant interchain tests pass (see [Testing](#testing))
- [ ] Proto changes regenerated via `make proto-all` (if applicable)
- [ ] New public types / functions have godoc comments
- [ ] Commit messages follow [Conventional Commits](#commit-conventions)
- [ ] PR description explains **what** changed and **why**
- [ ] Issue referenced in the PR description or commit footer (if applicable)
- [ ] `CHANGELOG.md` entry added for user-facing changes (unreleased section)

---

## Review Process

1. A maintainer will be assigned within **2 business days** of opening the PR.
2. At least **one approving review** from a core maintainer is required to merge.
3. CI checks (build, lint, unit tests) must all be green before merge.
4. Interchain tests are run on the CI pipeline for PRs that touch module logic, `app/`, or `cmd/`.
5. After approval, a maintainer will **squash-merge** the PR. Your branch can be deleted afterward.

**Responding to review feedback:** Push new commits — do not force-push while a review is in progress, as it clears reviewer annotations.

---

## Community

| Channel | Purpose |
|---|---|
| [GitHub Issues](https://github.com/Safrochain-Org/safrochain-node/issues) | Bugs, feature requests, questions |
| [Discord](https://discord.gg/YOUR_INVITE) | General chat, validator support, testnet faucet |
| [Twitter / X](https://twitter.com/safrochain) | Announcements |
| [security@safrochain.com](mailto:security@safrochain.com) | Security disclosures only |

We follow the [Contributor Covenant](https://www.contributor-covenant.org/) code of conduct. Be respectful, constructive, and welcoming to everyone.

---

*Safrochain is a proud fork of [Juno](https://github.com/CosmosContracts/juno) and is built on the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk). We are grateful to those open-source communities.*
