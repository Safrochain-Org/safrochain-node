use cosmwasm_schema::cw_serde;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    Increment {},
}

#[cw_serde]
pub enum QueryMsg {
    GetTickCount {},
}

#[cw_serde]
pub struct TickCountResponse {
    pub tick_count: u64,
}

/// Sudo message received from the x/clock module at the end of every block.
/// Register this contract with: safrochaind tx clock register <contract_addr>
#[cw_serde]
pub enum ClockCounterSudoMsg {
    ClockEndBlock {},
}
