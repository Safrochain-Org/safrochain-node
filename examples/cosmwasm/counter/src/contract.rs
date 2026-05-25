use cosmwasm_std::{
    entry_point, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo,
    Response, StdResult,
};
use cw_storage_plus::Item;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg, CountResponse};

pub const COUNT: Item<u64> = Item::new("count");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> StdResult<Response> {
    COUNT.save(deps.storage, &msg.count)?;
    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("count", msg.count.to_string()))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Increment {} => increment(deps),
        ExecuteMsg::Decrement {} => decrement(deps),
        ExecuteMsg::Reset { count } => reset(deps, count),
    }
}

fn increment(deps: DepsMut) -> Result<Response, ContractError> {
    let val = COUNT.load(deps.storage)?;
    COUNT.save(deps.storage, &val.checked_add(1).unwrap())?;
    Ok(Response::new()
        .add_attribute("method", "increment")
        .add_attribute("new_count", (val + 1).to_string()))
}

fn decrement(deps: DepsMut) -> Result<Response, ContractError> {
    let val = COUNT.load(deps.storage)?;
    COUNT.save(deps.storage, &val.checked_sub(1).unwrap())?;
    Ok(Response::new()
        .add_attribute("method", "decrement")
        .add_attribute("new_count", (val - 1).to_string()))
}

fn reset(deps: DepsMut, count: u64) -> Result<Response, ContractError> {
    COUNT.save(deps.storage, &count)?;
    Ok(Response::new()
        .add_attribute("method", "reset")
        .add_attribute("new_count", count.to_string()))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetCount {} => to_json_binary(&query_count(deps)?),
    }
}

fn query_count(deps: Deps) -> StdResult<CountResponse> {
    let count = COUNT.load(deps.storage)?;
    Ok(CountResponse { count })
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{message_info, mock_dependencies, mock_env};
    use cosmwasm_std::{coins, from_json};

    #[test]
    fn proper_initialization() {
        let mut deps = mock_dependencies();
        let msg = InstantiateMsg { count: 42 };
        let info = message_info(&deps.api.addr_make("creator"), &coins(1000, "usaf"));
        let res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert_eq!(res.attributes[1].value, "42");
    }

    #[test]
    fn increment() {
        let mut deps = mock_dependencies();
        let msg = InstantiateMsg { count: 0 };
        let info = message_info(&deps.api.addr_make("creator"), &coins(1000, "usaf"));
        instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let res = execute(deps.as_mut(), mock_env(), info.clone(), ExecuteMsg::Increment {}).unwrap();
        assert_eq!(res.attributes[1].value, "1");

        let query_res = query(deps.as_ref(), mock_env(), QueryMsg::GetCount {}).unwrap();
        let resp: CountResponse = from_json(&query_res).unwrap();
        assert_eq!(resp.count, 1);
    }

    #[test]
    fn decrement() {
        let mut deps = mock_dependencies();
        let msg = InstantiateMsg { count: 5 };
        let info = message_info(&deps.api.addr_make("creator"), &coins(1000, "usaf"));
        instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        execute(deps.as_mut(), mock_env(), info.clone(), ExecuteMsg::Decrement {}).unwrap();
        let query_res = query(deps.as_ref(), mock_env(), QueryMsg::GetCount {}).unwrap();
        let resp: CountResponse = from_json(&query_res).unwrap();
        assert_eq!(resp.count, 4);
    }

    #[test]
    fn reset() {
        let mut deps = mock_dependencies();
        let msg = InstantiateMsg { count: 0 };
        let info = message_info(&deps.api.addr_make("creator"), &coins(1000, "usaf"));
        instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        execute(deps.as_mut(), mock_env(), info.clone(), ExecuteMsg::Reset { count: 100 }).unwrap();
        let query_res = query(deps.as_ref(), mock_env(), QueryMsg::GetCount {}).unwrap();
        let resp: CountResponse = from_json(&query_res).unwrap();
        assert_eq!(resp.count, 100);
    }
}
