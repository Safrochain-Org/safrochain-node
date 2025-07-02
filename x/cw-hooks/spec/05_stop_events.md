# Remove your contract from receiving events

Contracts can be removed from events at any time with the following commands:

## Staking

> `safrochaind tx cw-hooks unregister staking [contract_bech32] --from [admin|creator]`

## Governance

> `safrochaind tx cw-hooks unregister governance [contract_bech32] --from [admin|creator]`
