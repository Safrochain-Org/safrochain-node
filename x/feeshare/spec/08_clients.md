<!--
order: 8
-->

# Clients

## Command Line Interface

Find below a list of `safrochaind` commands added with the  `x/feeshare` module. You can obtain the full list by using the `safrochaind -h` command. A CLI command can look like this:

```bash
safrochaind query feeshare params
```

### Queries

| Command            | Subcommand             | Description                              |
| :----------------- | :--------------------- | :--------------------------------------- |
| `query` `feeshare` | `params`               | Get feeshare params                      |
| `query` `feeshare` | `contract`             | Get the feeshare for a given contract    |
| `query` `feeshare` | `contracts`            | Get all feeshares                        |
| `query` `feeshare` | `deployer-contracts`   | Get all feeshares of a given deployer    |
| `query` `feeshare` | `withdrawer-contracts` | Get all feeshares of a given withdrawer  |

### Transactions

| Command         | Subcommand | Description                                |
| :-------------- | :--------- | :----------------------------------------- |
| `tx` `feeshare` | `register` | Register a contract for receiving feeshare |
| `tx` `feeshare` | `update`   | Update the withdraw address for a contract |
| `tx` `feeshare` | `cancel`   | Remove the feeshare for a contract         |

## gRPC Queries

| Verb   | Method                                            | Description                              |
| :----- | :------------------------------------------------ | :--------------------------------------- |
| `gRPC` | `safrochain.feeshare.v1.Query/Params`                   | Get feeshare params                      |
| `gRPC` | `safrochain.feeshare.v1.Query/FeeShare`                  | Get the feeshare for a given contract    |
| `gRPC` | `safrochain.feeshare.v1.Query/FeeShares`                 | Get all feeshares                        |
| `gRPC` | `safrochain.feeshare.v1.Query/DeployerFeeShares`         | Get all feeshares of a given deployer    |
| `gRPC` | `safrochain.feeshare.v1.Query/WithdrawerFeeShares`       | Get all feeshares of a given withdrawer  |
| `GET`  | `/safrochain/feeshare/v1/params`                        | Get feeshare params                      |
| `GET`  | `/safrochain/feeshare/v1/feeshares/{contract_address}`  | Get the feeshare for a given contract    |
| `GET`  | `/safrochain/feeshare/v1/feeshares`                     | Get all feeshares                        |
| `GET`  | `/safrochain/feeshare/v1/feeshares/{deployer_address}`  | Get all feeshares of a given deployer    |
| `GET`  | `/safrochain/feeshare/v1/feeshares/{withdraw_address}`  | Get all feeshares of a given withdrawer  |

### gRPC Transactions

| Verb   | Method                                     | Description                                |
| :----- | :----------------------------------------- | :----------------------------------------- |
| `gRPC` | `safrochain.feeshare.v1.Msg/RegisterFeeShare`   | Register a contract for receiving feeshare   |
| `gRPC` | `safrochain.feeshare.v1.Msg/UpdateFeeShare`     | Update the withdraw address for a contract   |
| `gRPC` | `safrochain.feeshare.v1.Msg/CancelFeeShare`     | Remove the feeshare for a contract           |
| `POST` | `/safrochain/feeshare/v1/tx/register_feeshare` | Register a contract for receiving feeshare   |
| `POST` | `/safrochain/feeshare/v1/tx/update_feeshare`   | Update the withdraw address for a contract   |
| `POST` | `/safrochain/feeshare/v1/tx/cancel_feeshare`   | Remove the feeshare for a contract           |
