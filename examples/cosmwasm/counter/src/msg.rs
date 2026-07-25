use cosmwasm_schema::cw_serde;

#[cw_serde]
pub struct InstantiateMsg {
    pub count: u64,
}

#[cw_serde]
pub enum ExecuteMsg {
    Increment {},
    Decrement {},
    Reset { count: u64 },
}

#[cw_serde]
pub enum QueryMsg {
    GetCount {},
}

#[cw_serde]
pub struct CountResponse {
    pub count: u64,
}
