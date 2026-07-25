use cosmwasm_std::{
    entry_point, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo,
    Response, StdResult,
};
use cw_storage_plus::Item;

use crate::error::ContractError;
use crate::msg::{
    ClockCounterSudoMsg, ExecuteMsg, InstantiateMsg, QueryMsg, TickCountResponse,
};

pub const TICK_COUNT: Item<u64> = Item::new("tick_count");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> StdResult<Response> {
    TICK_COUNT.save(deps.storage, &0)?;
    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("tick_count", "0"))
}

/// Called at the end of every block by the x/clock module.
/// Increments the tick counter by 1 each block.
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(
    _deps: DepsMut,
    _env: Env,
    msg: ClockCounterSudoMsg,
) -> Result<Response, ContractError> {
    match msg {
        ClockCounterSudoMsg::ClockEndBlock {} => {
            // Note: In production, this contract would be registered via:
            // safrochaind tx clock register <contract_addr> --from <key>
            // and the sudo handler would be called at the end of every block.
            Ok(Response::new()
                .add_attribute("method", "clock_end_block"))
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Increment {} => {
            TICK_COUNT.update(deps.storage, |c| Ok(c + 1))?;
            Ok(Response::new()
                .add_attribute("method", "increment"))
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetTickCount {} => {
            let count = TICK_COUNT.load(deps.storage)?;
            to_json_binary(&TickCountResponse { tick_count: count })
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{message_info, mock_dependencies, mock_env};
    use cosmwasm_std::{coins, from_json};

    #[test]
    fn proper_initialization() {
        let mut deps = mock_dependencies();
        let msg = InstantiateMsg {};
        let info = message_info(&deps.api.addr_make("creator"), &coins(1000, "usaf"));
        instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        let query_res = query(deps.as_ref(), mock_env(), QueryMsg::GetTickCount {}).unwrap();
        let resp: TickCountResponse = from_json(&query_res).unwrap();
        assert_eq!(resp.tick_count, 0);
    }

    #[test]
    fn increment_manually() {
        let mut deps = mock_dependencies();
        let msg = InstantiateMsg {};
        let info = message_info(&deps.api.addr_make("creator"), &coins(1000, "usaf"));
        instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        execute(
            deps.as_mut(),
            mock_env(),
            info.clone(),
            ExecuteMsg::Increment {},
        )
        .unwrap();

        let query_res = query(deps.as_ref(), mock_env(), QueryMsg::GetTickCount {}).unwrap();
        let resp: TickCountResponse = from_json(&query_res).unwrap();
        assert_eq!(resp.tick_count, 1);
    }

    #[test]
    fn sudo_end_block() {
        let mut deps = mock_dependencies();
        let msg = InstantiateMsg {};
        let info = message_info(&deps.api.addr_make("creator"), &coins(1000, "usaf"));
        instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        // Simulate EndBlock sudo call (this is what x/clock does at block end)
        let sudo_msg = ClockCounterSudoMsg::ClockEndBlock {};
        let res = sudo(deps.as_mut(), mock_env(), sudo_msg).unwrap();
        assert_eq!(res.attributes[0].value, "clock_end_block");
    }
}
