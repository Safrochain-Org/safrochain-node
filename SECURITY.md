# Security Policy

## Overview

The Safrochain team takes the security of the Safrochain node software seriously. We appreciate the efforts of security researchers and the community in helping us maintain a secure blockchain network. This document outlines our security policy, including how to report vulnerabilities, which versions are supported, and how we handle disclosed issues.

---

## Supported Versions

Only the versions listed below actively receive security patches and updates. Please ensure you are running a supported version before reporting a vulnerability.

| Version | Supported |
|---------|-----------|
| Latest (`main` branch) | ✅ Yes |
| Previous minor release | ✅ Yes |
| Older releases | ❌ No — please upgrade |

> **Recommendation:** Always run the latest release to benefit from all security patches.

---

## Reporting a Vulnerability

**Please do NOT open a public GitHub issue for security vulnerabilities.** Public disclosure before a fix is available can put the network and its participants at risk.

### How to Report

Send a detailed report to our security team via **private disclosure**:

- **Email:** [security@safrochain.com](mailto:security@safrochain.com)
- **Subject line:** `[SECURITY] <Brief description of the issue>`
- **PGP encryption:** Recommended for sensitive reports *(public key available on request)*

### What to Include

To help us triage and respond quickly, please provide:

1. **Description** — A clear explanation of the vulnerability and its potential impact.
2. **Affected component** — Module, file, or endpoint (e.g., `x/staking`, `api/`, `cmd/`).
3. **Steps to reproduce** — Minimal, reliable steps or a proof-of-concept (PoC).
4. **Version affected** — The node version or Git commit hash you tested against.
5. **Suggested fix** *(optional)* — Any remediation ideas you may have.
6. **Your contact info** — So we can follow up and credit you appropriately.

---

## Response Timeline

We are committed to responding promptly to all valid security reports.

| Milestone | Target Timeframe |
|-----------|-----------------|
| Acknowledgement of report | Within **48 hours** |
| Initial assessment & severity triage | Within **5 business days** |
| Patch development begins | Within **7 days** of confirmation |
| Security fix released | Within **30 days** (critical issues: within **7 days**) |
| Public disclosure (CVE / advisory) | After patch is released and deployed |

If we cannot meet a deadline, we will notify you and provide a revised estimate.

---

## Severity Classification

We use the following severity levels to prioritize responses:

| Severity | Description | Examples |
|----------|-------------|---------|
| **Critical** | Immediate network-wide risk or fund loss | Consensus bypass, double-spend, remote code execution |
| **High** | Significant impact on node operation or validator security | DoS attacks, private key exposure vectors |
| **Medium** | Limited impact, requires special conditions | Information leaks, minor consensus deviations |
| **Low** | Minimal impact, theoretical risk | Configuration edge cases, minor info disclosure |

---

## Disclosure Policy

We follow **coordinated responsible disclosure**:

1. You report the vulnerability privately.
2. We acknowledge, investigate, and develop a fix.
3. We release the patch and notify node operators.
4. We publish a public security advisory (GitHub Security Advisory / CVE) after the patch is widely deployed.
5. We credit the reporter (unless anonymity is requested).

We ask that you:
- Give us a reasonable time to fix the issue before any public disclosure.
- Avoid accessing, modifying, or deleting data on live network nodes.
- Do not perform denial-of-service testing against the Safrochain testnet or mainnet.

---

## Security Best Practices for Node Operators

If you are running a Safrochain validator or full node, follow these security guidelines:

### System Hardening
- Run the node under a dedicated non-root user account.
- Keep your OS and all packages up to date.
- Enable a firewall (e.g., `ufw`) and expose only required ports:
  - `26656` — P2P
  - `26657` — RPC *(restrict to trusted IPs only)*
  - `1317` — REST API *(restrict or disable if not needed)*
  - `9090` — gRPC *(restrict or disable if not needed)*

### Validator Key Security
- Store your **validator consensus key** (`priv_validator_key.json`) securely — preferably using a Hardware Security Module (HSM) or remote signer (e.g., `tmkms`).
- **Never** share or commit your validator private key or mnemonic to any repository.
- Back up your keyring securely and store backups offline.

### Node Configuration
- Set `pex = true` carefully and restrict `persistent_peers` to trusted nodes.
- Enable `addr_book_strict = true` to prevent IP spoofing.
- Use TLS for RPC and REST endpoints when exposed externally.
- Configure rate-limiting on public-facing APIs.

### Monitoring & Alerting
- Monitor node uptime, block signing rate, and peer count continuously.
- Set alerts for missed blocks, unexpected restarts, or disk/CPU spikes.
- Review node logs regularly for anomalies.

### Software
- Only download Safrochain binaries from official sources.
- Verify binary checksums before deployment.
- Keep the Go toolchain and all dependencies up to date.

---

## Known Security Mitigations in This Repository

The following mitigations are already applied in this codebase:

- **GHSA-h395-qcrw-5vmq** — Patched via `github.com/gin-gonic/gin v1.7.0` (see `go.mod`).
- **Broken goleveldb** — Replaced with `github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7`.

---

## Bug Bounty

We do not currently operate a formal paid bug bounty program. However, we recognize and publicly credit all responsible disclosures in our release notes and security advisories. We are evaluating a formal bounty program for future releases.

---

## Contact

| Purpose | Contact |
|---------|---------|
| Security vulnerabilities | [security@safrochain.com](mailto:security@safrochain.com) |
| General development | [GitHub Issues](https://github.com/safrochain/safrochain-node/issues) |
| Community & validators | Safrochain Discord / Telegram |

---

*This policy is maintained by the Safrochain core team and is subject to updates. Last revised: March 2026.*
