use cosmwasm_std::{
    entry_point, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo,
    Response, StdResult,
};
use cw_storage_plus::Map;

use crate::error::ContractError;
use crate::msg::{
    CwHooksSudoMsg, ExecuteMsg, InstantiateMsg, QueryMsg, StakingEventResponse,
};

/// Stores staking events that the contract receives via x/cw-hooks sudo messages.
/// Maps event ID -> event description. In production, IDs would be block-height + index.
pub const EVENTS: Map<u64, String> = Map::new("events");
pub const EVENT_COUNTER: cw_storage_plus::Item<u64> = cw_storage_plus::Item::new("event_counter");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> StdResult<Response> {
    EVENT_COUNTER.save(deps.storage, &0)?;
    Ok(Response::new()
        .add_attribute("method", "instantiate"))
}

/// Called by the x/cw-hooks module when staking or governance events occur.
/// The contract must register with: safrochaind tx cw-hooks register <event_type> <contract_addr>
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(
    deps: DepsMut,
    _env: Env,
    msg: CwHooksSudoMsg,
) -> Result<Response, ContractError> {
    match msg {
        CwHooksSudoMsg::AfterValidatorCreated {
            validator_address, ..
        } => log_event(deps, format!("validator_created:{}", validator_address)),

        CwHooksSudoMsg::AfterValidatorRemoved {
            validator_address, ..
        } => log_event(deps, format!("validator_removed:{}", validator_address)),

        CwHooksSudoMsg::AfterValidatorBonded {
            validator_address, ..
        } => log_event(deps, format!("validator_bonded:{}", validator_address)),

        CwHooksSudoMsg::AfterValidatorBeginUnbonding {
            validator_address, ..
        } => log_event(deps, format!("validator_unbonding:{}", validator_address)),
    }
}

fn log_event(deps: DepsMut, event: String) -> Result<Response, ContractError> {
    let count = EVENT_COUNTER.load(deps.storage)?;
    EVENTS.save(deps.storage, &count, &event)?;
    EVENT_COUNTER.save(deps.storage, &(count + 1))?;
    Ok(Response::new()
        .add_attribute("method", "log_event")
        .add_attribute("event_index", count.to_string())
        .add_attribute("event", &event))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::ClearEvents {} => Ok(Response::new().add_attribute("method", "clear_events")),
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetEvents {} => {
            let count = EVENT_COUNTER.load(deps.storage)?;
            let mut events = vec![];
            for i in 0..count {
                if let Ok(evt) = EVENTS.load(deps.storage, &i) {
                    events.push(StakingEventResponse {
                        id: i,
                        description: evt,
                    });
                }
            }
            to_json_binary(&events)
        }
        QueryMsg::GetEventCount {} => {
            let count = EVENT_COUNTER.load(deps.storage)?;
            to_json_binary(&count)
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
        let info = message_info(&deps.api.addr_make("creator"), &coins(1000, "usaf"));
        instantiate(deps.as_mut(), mock_env(), info, InstantiateMsg {}).unwrap();

        let res: u64 = from_json(
            &query(deps.as_ref(), mock_env(), QueryMsg::GetEventCount {}).unwrap(),
        )
        .unwrap();
        assert_eq!(res, 0);
    }

    #[test]
    fn receive_validator_created_event() {
        let mut deps = mock_dependencies();
        let info = message_info(&deps.api.addr_make("creator"), &coins(1000, "usaf"));
        instantiate(deps.as_mut(), mock_env(), info, InstantiateMsg {}).unwrap();

        // Simulate x/cw-hooks sending a sudo message
        let sudo_msg = CwHooksSudoMsg::AfterValidatorCreated {
            moniker: "test-validator".to_string(),
            validator_address: "safrovaloper1test".to_string(),
            commission: "0.1".to_string(),
            validator_tokens: "1000000".to_string(),
            bonded_tokens: "1000000".to_string(),
            bond_status: "BOND_STATUS_BONDED".to_string(),
        };
        let res = sudo(deps.as_mut(), mock_env(), sudo_msg).unwrap();
        assert_eq!(res.attributes[1].value, "0");

        // Check event count
        let count: u64 = from_json(
            &query(deps.as_ref(), mock_env(), QueryMsg::GetEventCount {}).unwrap(),
        )
        .unwrap();
        assert_eq!(count, 1);
    }

    #[test]
    fn receive_multiple_events() {
        let mut deps = mock_dependencies();
        let info = message_info(&deps.api.addr_make("creator"), &coins(1000, "usaf"));
        instantiate(deps.as_mut(), mock_env(), info, InstantiateMsg {}).unwrap();

        // Send 3 events
        for i in 0..3 {
            let sudo_msg = CwHooksSudoMsg::AfterValidatorBonded {
                validator_address: format!("safrovaloper{}", i),
                moniker: format!("val-{}", i),
                commission: "0.1".to_string(),
                validator_tokens: "1000000".to_string(),
                bonded_tokens: "1000000".to_string(),
                bond_status: "BOND_STATUS_BONDED".to_string(),
            };
            sudo(deps.as_mut(), mock_env(), sudo_msg).unwrap();
        }

        let count: u64 = from_json(
            &query(deps.as_ref(), mock_env(), QueryMsg::GetEventCount {}).unwrap(),
        )
        .unwrap();
        assert_eq!(count, 3);
    }
}
