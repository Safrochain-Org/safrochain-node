use cosmwasm_std::{
    entry_point, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
};
use cw20_base::contract::{
    execute as cw20_base_execute, instantiate as cw20_base_instantiate,
    query as cw20_base_query,
};
use cw20_base::msg::{
    ExecuteMsg as Cw20ExecuteMsg, InstantiateMsg as Cw20InstantiateMsg,
    QueryMsg as Cw20QueryMsg,
};

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> StdResult<Response> {
    match msg {
        InstantiateMsg::Cw20(inner) => {
            cw20_base_instantiate(deps, env, info, inner)
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Cw20(inner) => Ok(cw20_base_execute(deps, env, info, inner)?),
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Cw20(inner) => cw20_base_query(deps, env, inner),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{message_info, mock_dependencies, mock_env};
    use cosmwasm_std::{coins, from_json, Addr, Uint128};
    use cw20::{BalanceResponse, Cw20QueryMsg as _, TokenInfoResponse};

    #[test]
    fn proper_initialization() {
        let mut deps = mock_dependencies();
        let msg = InstantiateMsg::Cw20(Cw20InstantiateMsg {
            name: "Safro Token".to_string(),
            symbol: "SAFRO".to_string(),
            decimals: 6,
            initial_balances: vec![],
            mint: Some(cw20::MinterResponse {
                minter: deps.api.addr_make("minter").to_string(),
                cap: None,
            }),
            marketing: None,
        });
        let info = message_info(&deps.api.addr_make("creator"), &coins(1000, "usaf"));
        instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        let query_res = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::Cw20(Cw20QueryMsg::TokenInfo {}),
        )
        .unwrap();
        let resp: TokenInfoResponse = from_json(&query_res).unwrap();
        assert_eq!(resp.name, "Safro Token");
        assert_eq!(resp.symbol, "SAFRO");
        assert_eq!(resp.decimals, 6);
    }

    #[test]
    fn mint_and_transfer() {
        let mut deps = mock_dependencies();
        let minter = deps.api.addr_make("minter");
        let recipient = deps.api.addr_make("recipient");

        let msg = InstantiateMsg::Cw20(Cw20InstantiateMsg {
            name: "Safro Token".to_string(),
            symbol: "SAFRO".to_string(),
            decimals: 6,
            initial_balances: vec![],
            mint: Some(cw20::MinterResponse {
                minter: minter.to_string(),
                cap: None,
            }),
            marketing: None,
        });
        let info = message_info(&deps.api.addr_make("creator"), &coins(1000, "usaf"));
        instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        // Mint 1000 tokens to minter
        let mint_msg = ExecuteMsg::Cw20(Cw20ExecuteMsg::Mint {
            recipient: minter.to_string(),
            amount: Uint128::new(1000),
        });
        execute(
            deps.as_mut(),
            mock_env(),
            message_info(&minter, &[]),
            mint_msg,
        )
        .unwrap();

        // Check balance
        let query_res = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::Cw20(Cw20QueryMsg::Balance {
                address: minter.to_string(),
            }),
        )
        .unwrap();
        let balance: BalanceResponse = from_json(&query_res).unwrap();
        assert_eq!(balance.balance.u128(), 1000);

        // Transfer 500 to recipient
        let transfer_msg = ExecuteMsg::Cw20(Cw20ExecuteMsg::Transfer {
            recipient: recipient.to_string(),
            amount: Uint128::new(500),
        });
        execute(
            deps.as_mut(),
            mock_env(),
            message_info(&minter, &[]),
            transfer_msg,
        )
        .unwrap();

        let query_res = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::Cw20(Cw20QueryMsg::Balance {
                address: recipient.to_string(),
            }),
        )
        .unwrap();
        let balance: BalanceResponse = from_json(&query_res).unwrap();
        assert_eq!(balance.balance.u128(), 500);
    }
}
