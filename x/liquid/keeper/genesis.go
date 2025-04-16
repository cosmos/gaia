package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v23/x/liquid/types"
)

// InitGenesis sets liquid information for genesis
func (k Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) {
	//
	// Set the total liquid staked tokens
	k.SetTotalLiquidStakedTokens(ctx, data.TotalLiquidStakedTokens)

	// Set each tokenize share record, as well as the last tokenize share record ID
	latestID := uint64(0)
	for _, tokenizeShareRecord := range data.TokenizeShareRecords {
		if err := k.AddTokenizeShareRecord(ctx, tokenizeShareRecord); err != nil {
			panic(err)
		}
		if tokenizeShareRecord.Id > latestID {
			latestID = tokenizeShareRecord.Id
		}
	}
	if data.LastTokenizeShareRecordId < latestID {
		panic("Tokenize share record specified with ID greater than the latest ID")
	}
	k.SetLastTokenizeShareRecordID(ctx, data.LastTokenizeShareRecordId)

	// Set the tokenize shares locks for accounts that have disabled tokenizing shares
	// The lock can either be in status LOCKED or LOCK_EXPIRING
	// If it is in status LOCK_EXPIRING, the unlocking must also be queued
	k.SetTotalLiquidStakedTokens(ctx, data.TotalLiquidStakedTokens)
}

func (k Keeper) SetTokenizeShareLocks(ctx sdk.Context, tokenizeShareLocks []types.TokenizeShareLock) {
	// Set the tokenize shares locks for accounts that have disabled tokenizing shares
	// The lock can either be in status LOCKED or LOCK_EXPIRING
	// If it is in status LOCK_EXPIRING, the unlocking must also be queued
	for _, tokenizeShareLock := range tokenizeShareLocks {
		address, err := k.authKeeper.AddressCodec().StringToBytes(tokenizeShareLock.Address)
		if err != nil {
			panic(err)
		}

		switch tokenizeShareLock.Status {
		case types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED.String():
			k.AddTokenizeSharesLock(ctx, address)

		case types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING.String():
			completionTime := tokenizeShareLock.CompletionTime

			authorizations := k.GetPendingTokenizeShareAuthorizations(ctx, completionTime)
			authorizations.Addresses = append(authorizations.Addresses, sdk.AccAddress(address).String())

			k.SetPendingTokenizeShareAuthorizations(ctx, completionTime, authorizations)
			k.SetTokenizeSharesUnlockTime(ctx, address, completionTime)

		default:
			panic(fmt.Sprintf("Unsupported tokenize share lock status %s", tokenizeShareLock.Status))
		}
	}
}

func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params:                    params,
		TokenizeShareRecords:      k.GetAllTokenizeShareRecords(ctx),
		LastTokenizeShareRecordId: k.GetLastTokenizeShareRecordID(ctx),
		TotalLiquidStakedTokens:   k.GetTotalLiquidStakedTokens(ctx),
		TokenizeShareLocks:        k.GetAllTokenizeSharesLocks(ctx),
	}
}
