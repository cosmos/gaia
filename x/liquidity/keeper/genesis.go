package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// InitGenesis initializes the liquidity module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := k.ValidateGenesis(ctx, genState); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)

	for _, record := range genState.PoolRecords {
		k.SetPoolRecord(ctx, record)
	}
}

// ExportGenesis returns the liquidity module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)

	var poolRecords []types.PoolRecord

	pools := k.GetAllPools(ctx)

	for _, pool := range pools {
		record, found := k.GetPoolRecord(ctx, pool)
		if found {
			poolRecords = append(poolRecords, record)
		}
	}

	if len(poolRecords) == 0 {
		poolRecords = []types.PoolRecord{}
	}

	return types.NewGenesisState(params, poolRecords)
}

// ValidateGenesis validates the liquidity module's genesis state.
func (k Keeper) ValidateGenesis(ctx sdk.Context, genState types.GenesisState) error {
	if err := genState.Params.Validate(); err != nil {
		return err
	}

	cc, _ := ctx.CacheContext()
	k.SetParams(cc, genState.Params)

	for _, record := range genState.PoolRecords {
		record = k.SetPoolRecord(cc, record)
		if err := k.ValidatePoolRecord(cc, record); err != nil {
			return err
		}
	}

	return nil
}
