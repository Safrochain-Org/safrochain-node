use cosmwasm_schema::cw_serde;
use cw20_base::msg::{
    ExecuteMsg as Cw20ExecuteMsg, InstantiateMsg as Cw20InstantiateMsg,
    QueryMsg as Cw20QueryMsg,
};

#[cw_serde]
pub enum InstantiateMsg {
    Cw20(Cw20InstantiateMsg),
}

#[cw_serde]
pub enum ExecuteMsg {
    Cw20(Cw20ExecuteMsg),
}

#[cw_serde]
pub enum QueryMsg {
    Cw20(Cw20QueryMsg),
}
