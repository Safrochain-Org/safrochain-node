# Clients

## Command Line Interface

Find below a list of `safrochaind` commands added with the `x/cw-hooks` module. You can obtain the full list by using the `safrochaind -h` command. A CLI command can look like this:

```bash
safrochaind query cw-hooks params
```

### Queries

| Command            | Subcommand             | Description                              |
| :----------------- | :--------------------- | :--------------------------------------- |
| `query` `cw-hooks` | `params`               | Get module params                        |
| `query` `cw-hooks` | `governance-contracts` | Get registered governance contracts      |
| `query` `cw-hooks` | `staking-contracts`    | Get registered staking contracts         |

### Transactions

| Command         | Subcommand   | Description                           |
| :-------------- | :----------- | :------------------------------------ |
| `tx` `cw-hooks` | `register`   | Register a contract for events        |
| `tx` `cw-hooks` | `unregister` | Unregister a contract from events     |

## gRPC Queries

| Verb   | Method                                            |
| :----- | :------------------------------------------------ |
| `gRPC` | `safrochain.cwhooks.v1.Query/Params`                    |
| `gRPC` | `safrochain.cwhooks.v1.Query/StakingContracts`          |
| `gRPC` | `safrochain.cwhooks.v1.Query/GovernanceContracts`       |
| `GET`  | `/safrochain/cwhooks/v1/params`                         |
| `GET`  | `/safrochain/cwhooks/v1/staking_contracts`              |
| `GET`  | `/safrochain/cwhooks/v1/governance_contracts`           |

### gRPC Transactions

| Verb   | Method                                      |
| :----- | :------------------------------------------ |
| `gRPC` | `safrochain.cwhooks.v1.Msg/RegisterStaking`       |
| `gRPC` | `safrochain.cwhooks.v1.Msg/UnregisterStaking`     |
| `gRPC` | `safrochain.cwhooks.v1.Msg/RegisterGovernance`    |
| `gRPC` | `safrochain.cwhooks.v1.Msg/UnregisterGovernance`  |
| `POST` | `/safrochain/cwhooks/v1/tx/register_staking`      |
| `POST` | `/safrochain/cwhooks/v1/tx/unregister_staking`    |
| `POST` | `/safrochain/cwhooks/v1/tx/register_governance`   |
| `POST` | `/safrochain/cwhooks/v1/tx/unregister_governance` |
