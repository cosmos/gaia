package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker removes expired tokenize share locks
func (k *Keeper) BeginBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	_, err := k.RemoveExpiredTokenizeShareLocks(ctx, sdkCtx.BlockTime())
	return err
}
