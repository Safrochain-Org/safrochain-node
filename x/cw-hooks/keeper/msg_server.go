package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

var _ types.MsgServer = &msgServer{}

// msgServer is a wrapper of Keeper.
type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/cw-hooks MsgServer interface.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	if _, err := sdk.AccAddressFromBech32(req.Authority); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) RegisterStaking(goCtx context.Context, req *types.MsgRegisterStaking) (*types.MsgRegisterStakingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(req.RegisterAddress); err != nil {
		return nil, errorsmod.Wrap(err, "invalid register address")
	}

	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, errorsmod.Wrap(err, "invalid contract address")
	}

	if err := k.handleContractRegister(ctx, req.RegisterAddress, req.ContractAddress, types.KeyPrefixStaking, "staking"); err != nil {
		return nil, err
	}

	return &types.MsgRegisterStakingResponse{}, nil
}

func (k msgServer) RegisterGovernance(goCtx context.Context, req *types.MsgRegisterGovernance) (*types.MsgRegisterGovernanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(req.RegisterAddress); err != nil {
		return nil, errorsmod.Wrap(err, "invalid register address")
	}

	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, errorsmod.Wrap(err, "invalid contract address")
	}

	if err := k.handleContractRegister(ctx, req.RegisterAddress, req.ContractAddress, types.KeyPrefixGov, "governance"); err != nil {
		return nil, err
	}

	return &types.MsgRegisterGovernanceResponse{}, nil
}

func (k msgServer) UnregisterGovernance(goCtx context.Context, req *types.MsgUnregisterGovernance) (*types.MsgUnregisterGovernanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(req.RegisterAddress); err != nil {
		return nil, errorsmod.Wrap(err, "invalid register address")
	}

	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, errorsmod.Wrap(err, "invalid contract address")
	}

	if err := k.handleContractRemoval(ctx, req.RegisterAddress, req.ContractAddress, types.KeyPrefixGov, "governance"); err != nil {
		return nil, err
	}

	return &types.MsgUnregisterGovernanceResponse{}, nil
}

func (k msgServer) UnregisterStaking(goCtx context.Context, req *types.MsgUnregisterStaking) (*types.MsgUnregisterStakingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(req.RegisterAddress); err != nil {
		return nil, errorsmod.Wrap(err, "invalid register address")
	}

	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, errorsmod.Wrap(err, "invalid contract address")
	}

	if err := k.handleContractRemoval(ctx, req.RegisterAddress, req.ContractAddress, types.KeyPrefixStaking, "staking"); err != nil {
		return nil, err
	}

	return &types.MsgUnregisterStakingResponse{}, nil
}

// isContractManager reports whether sender manages contract: its admin if it
// has one, otherwise its creator. This matches x/clock and x/feepay and the
// cw-hooks spec ("admin of the contract, or the creator if there is no
// admin").
func (k msgServer) isContractManager(ctx sdk.Context, sender string, contract sdk.AccAddress) (bool, error) {
	if !k.GetWasmKeeper().HasContractInfo(ctx, contract) {
		return false, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "contract does not exist: %s", contract)
	}

	contractInfo := k.GetWasmKeeper().GetContractInfo(ctx, contract)
	if contractInfo.Admin != "" {
		return contractInfo.Admin == sender, nil
	}

	return contractInfo.Creator == sender, nil
}

// handleContractRegister registers a contract for a hook type.
//
// Only the governance authority can register a contract. Hooks run inside the
// staking and governance flows (including Begin/EndBlock) and every user pays
// for their gas, so each contract must be vetted by governance first.
func (k msgServer) handleContractRegister(ctx sdk.Context, sender, contractAddr string, keyPrefix []byte, prefixModuleName string) error {
	if sender != k.authority {
		return errorsmod.Wrapf(govtypes.ErrInvalidSigner, "only the governance authority (%s) can register %s hook contracts; submit a governance proposal", k.authority, prefixModuleName)
	}

	contract, err := sdk.AccAddressFromBech32(contractAddr)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", err)
	}

	if !k.GetWasmKeeper().HasContractInfo(ctx, contract) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "contract does not exist: %s", contract)
	}

	if k.IsContractRegistered(ctx, keyPrefix, contract) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "contract already registered for %s", prefixModuleName)
	}

	if k.CountContracts(ctx, keyPrefix) >= types.MaxContractsPerHook {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "at most %d contracts can be registered for %s hooks", types.MaxContractsPerHook, prefixModuleName)
	}

	k.SetContract(ctx, keyPrefix, contract)

	return nil
}

// handleContractRemoval unregisters a contract from a hook type. The
// governance authority can remove any contract (emergency path); the contract
// manager (admin, or creator if there is no admin) can remove its own.
func (k msgServer) handleContractRemoval(ctx sdk.Context, sender, contractAddr string, keyPrefix []byte, prefixModuleName string) error {
	contract, err := sdk.AccAddressFromBech32(contractAddr)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", err)
	}

	if !k.IsContractRegistered(ctx, keyPrefix, contract) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "contract is not registered for %s", prefixModuleName)
	}

	if sender != k.authority {
		// A registered contract may have been deleted from x/wasm since
		// registration; only governance can remove it then.
		isManager, err := k.isContractManager(ctx, sender, contract)
		if err != nil {
			return err
		}
		if !isManager {
			return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "only the contract admin (or creator if it has no admin) or the governance authority can unregister it")
		}
	}

	k.DeleteContract(ctx, keyPrefix, contract)

	return nil
}
