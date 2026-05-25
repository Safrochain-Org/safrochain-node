use cosmwasm_schema::cw_serde;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    ClearEvents {},
}

#[cw_serde]
pub enum QueryMsg {
    GetEvents {},
    GetEventCount {},
}

#[cw_serde]
pub struct StakingEventResponse {
    pub id: u64,
    pub description: String,
}

/// Sudo messages from x/cw-hooks module.
/// Register with: safrochaind tx cw-hooks register staking <contract_addr>
#[cw_serde]
pub enum CwHooksSudoMsg {
    AfterValidatorCreated {
        moniker: String,
        validator_address: String,
        commission: String,
        validator_tokens: String,
        bonded_tokens: String,
        bond_status: String,
    },
    AfterValidatorRemoved {
        moniker: String,
        validator_address: String,
        commission: String,
        validator_tokens: String,
        bonded_tokens: String,
        bond_status: String,
    },
    AfterValidatorBonded {
        moniker: String,
        validator_address: String,
        commission: String,
        validator_tokens: String,
        bonded_tokens: String,
        bond_status: String,
    },
    AfterValidatorBeginUnbonding {
        moniker: String,
        validator_address: String,
        commission: String,
        validator_tokens: String,
        bonded_tokens: String,
        bond_status: String,
    },
}
