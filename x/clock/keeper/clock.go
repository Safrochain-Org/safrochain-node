package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	globalerrors "github.com/Safrochain_Org/safrochain/app/helpers"
	"github.com/Safrochain_Org/safrochain/x/clock/types"
)

// Store Keys for clock contracts (both jailed and unjailed)
var (
	StoreKeyContracts = []byte("contracts")
)

// Get the store for the clock contracts.
func (k Keeper) getStore(ctx context.Context) prefix.Store {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(store, StoreKeyContracts)
}

// Set a clock contract address in the KV store.
func (k Keeper) SetClockContract(ctx context.Context, contract types.ClockContract) error {
	// Get store, marshal content
	store := k.getStore(ctx)
	bz, err := k.cdc.Marshal(&contract)
	if err != nil {
		return err
	}

	// Set the contract
	store.Set([]byte(contract.ContractAddress), bz)
	return nil
}

// Check if a clock contract address is in the KV store.
func (k Keeper) IsClockContract(ctx context.Context, contractAddress string) bool {
	store := k.getStore(ctx)
	return store.Has([]byte(contractAddress))
}

// Get a clock contract address from the KV store.
func (k Keeper) GetClockContract(ctx context.Context, contractAddress string) (*types.ClockContract, error) {
	// Check if the contract is registered
	if !k.IsClockContract(ctx, contractAddress) {
		return nil, globalerrors.ErrContractNotRegistered
	}

	// Get the KV store
	store := k.getStore(ctx)
	bz := store.Get([]byte(contractAddress))

	// Unmarshal the contract
	var contract types.ClockContract
	err := k.cdc.Unmarshal(bz, &contract)
	if err != nil {
		return nil, err
	}

	// Return the contract
	return &contract, nil
}

// Get all clock contract addresses from the KV store.
func (k Keeper) GetAllContracts(ctx context.Context) ([]types.ClockContract, error) {
	// Get the KV store
	store := k.getStore(ctx)

	// Create iterator for contracts
	iterator := storetypes.KVStorePrefixIterator(store, []byte(nil))
	defer iterator.Close() //nolint:errcheck

	// Iterate over all contracts
	contracts := []types.ClockContract{}
	for ; iterator.Valid(); iterator.Next() {
		// Unmarshal iterator
		var contract types.ClockContract
		err := k.cdc.Unmarshal(iterator.Value(), &contract)
		if err != nil {
			return nil, err
		}

		contracts = append(contracts, contract)
	}

	// Return array of contracts
	return contracts, nil
}

// Get all registered fee pay contracts
func (k Keeper) GetPaginatedContracts(ctx context.Context, pag *query.PageRequest) (*types.QueryClockContractsResponse, error) {
	store := k.getStore(ctx)

	// Filter and paginate all contracts
	results, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pag,
		func(_ []byte, value *types.ClockContract) (*types.ClockContract, error) {
			return value, nil
		},
		func() *types.ClockContract {
			return &types.ClockContract{}
		},
	)
	if err != nil {
		return nil, err
	}

	// Dereference pointer array of contracts
	var contracts []types.ClockContract
	for _, contract := range results {
		contracts = append(contracts, *contract)
	}

	// Return paginated contracts
	return &types.QueryClockContractsResponse{
		ClockContracts: contracts,
		Pagination:     pageRes,
	}, nil
}

// Remove a clock contract address from the KV store.
func (k Keeper) RemoveContract(ctx context.Context, contractAddress string) {
	store := k.getStore(ctx)
	key := []byte(contractAddress)

	if store.Has(key) {
		store.Delete(key)
	}
}

// Register a clock contract address in the KV store.
func (k Keeper) RegisterContract(ctx context.Context, senderAddress string, contractAddress string) error {
	// Check if the contract is already registered
	if k.IsClockContract(ctx, contractAddress) {
		return globalerrors.ErrContractAlreadyRegistered
	}

	// Ensure the sender is the contract admin or creator
	if ok, err := k.IsContractManager(ctx, senderAddress, contractAddress); !ok {
		return err
	}

	// Register contract
	return k.SetClockContract(ctx, types.ClockContract{
		ContractAddress: contractAddress,
		IsJailed:        false,
	})
}

// Unregister a clock contract from either the jailed or unjailed KV store.
func (k Keeper) UnregisterContract(ctx context.Context, senderAddress string, contractAddress string) error {
	// Check if the contract is registered in either store
	if !k.IsClockContract(ctx, contractAddress) {
		return globalerrors.ErrContractNotRegistered
	}

	// Ensure the sender is the contract admin or creator
	if ok, err := k.IsContractManager(ctx, senderAddress, contractAddress); !ok {
		return err
	}

	// Remove contract from both stores
	k.RemoveContract(ctx, contractAddress)
	return nil
}

// Set the jail status of a clock contract in the KV store.
func (k Keeper) SetJailStatus(ctx context.Context, contractAddress string, isJailed bool) error {
	// Get the contract
	contract, err := k.GetClockContract(ctx, contractAddress)
	if err != nil {
		return err
	}

	// Check if the contract is already jailed or unjailed
	if contract.IsJailed == isJailed {
		if isJailed {
			return types.ErrContractAlreadyJailed
		}

		return types.ErrContractNotJailed
	}

	// Set the jail status
	contract.IsJailed = isJailed

	// Set the contract
	return k.SetClockContract(ctx, *contract)
}

// Set the jail status of a clock contract by the sender address.
func (k Keeper) SetJailStatusBySender(ctx context.Context, senderAddress string, contractAddress string, jailStatus bool) error {
	// Ensure the sender is the contract admin or creator
	if ok, err := k.IsContractManager(ctx, senderAddress, contractAddress); !ok {
		return err
	}

	return k.SetJailStatus(ctx, contractAddress, jailStatus)
}

// Check if the sender is the designated contract manager for the FeePay contract. If
// an admin is present, they are considered the manager. If there is no admin, the
// contract creator is considered the manager.
func (k Keeper) IsContractManager(ctx context.Context, senderAddress string, contractAddress string) (bool, error) {
	contractAddr := sdk.MustAccAddressFromBech32(contractAddress)

	// Ensure the contract is a cosm wasm contract
	if ok := k.wasmKeeper.HasContractInfo(ctx, contractAddr); !ok {
		return false, globalerrors.ErrInvalidCWContract
	}

	// Get the contract info
	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddr)

	// Flags for admin existence & sender being admin/creator
	adminExists := len(contractInfo.Admin) > 0
	isSenderAdmin := contractInfo.Admin == senderAddress
	isSenderCreator := contractInfo.Creator == senderAddress

	// Check if the sender is the admin or creator
	if adminExists && !isSenderAdmin {
		return false, globalerrors.ErrContractNotAdmin
	} else if !adminExists && !isSenderCreator {
		return false, globalerrors.ErrContractNotCreator
	}

	return true, nil
}
