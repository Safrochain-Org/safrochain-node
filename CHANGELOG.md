# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- (none yet)

### Changed

- (none yet)

### Fixed

- (none yet)

## [v0.2.2] - 2026-05-18

### Added

- Default `minimum-gas-prices = "0.05usaf"` calibrated for mainnet during `safrochaind init`
- Deterministic release binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 with SHA-256 checksums

### Changed

- CI hardened: `x/cw-hooks` keeper tests no longer break on fresh wasm rebuilds
- CodeQL config scoped to ignore vendored crypto paths (cosmossdk.io, cometbft) to silence false positives
- Release dispatch workflow rewritten with explicit matrix and deterministic `-trimpath` flags

### Fixed

- (none)

## [v0.2.1] - 2026-05-07

### Added

- (none)

### Changed

- Bumped aws-sdk-go-v2 dependency stack
- Deterministic map iteration ordering
- Scoped CodeQL configuration for reduced noise
- Polished build and install UX

### Fixed

- (none)

## [v0.2.0-rc.1] - Unreleased Pre-release

### Added

- Pre-release candidate for v0.2.0 line

### Changed

- (none)

### Fixed

- (none)

## [v0.1.0] - 2025-07-02

### Added

- Full CosmWasm smart contract support (cw-hooks, clock, fee modules)
- Governance module with on-chain proposal and voting
- IBC (Inter-Blockchain Communication) for cross-chain asset transfers
- Bank & staking modules for secure token management and rewards
- Custom modules: `x/clock`, `x/cw-hooks`, `x/drip`, `x/feepay`, `x/feeshare`, `x/globalfee`, `x/tokenfactory`, `x/wrappers`
- Professional CI/CD pipeline with automated build and release
- Per-module interchain tests covering all custom modules
- Upgrade and state-sync test coverage

### Changed

- (none)

### Fixed

- (none)

---

## Versioning Notes

- Tags: [v0.1.0](https://github.com/Safrochain-Org/safrochain-node/releases/tag/v0.1.0),
  [v0.2.1](https://github.com/Safrochain-Org/safrochain-node/releases/tag/v0.2.1),
  [v0.2.2](https://github.com/Safrochain-Org/safrochain-node/releases/tag/v0.2.2)
