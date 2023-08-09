package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	v2 "github.com/cosmos/gaia/v12/x/globalfee/migrations/v2"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	globalfeeSubspace paramtypes.Subspace
}

// NewMigrator returns a new Migrator.
func NewMigrator(globalfeeSubspace paramtypes.Subspace) Migrator {
	return Migrator{globalfeeSubspace: globalfeeSubspace}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStore(ctx, m.globalfeeSubspace)
}
