package keeper

import (
	"context"
	goerrors "errors"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v22/x/lsm/types"
)

func (k Keeper) WithdrawSingleShareRecordReward(ctx context.Context, recordID uint64) error {
	record, err := k.GetTokenizeShareRecord(ctx, recordID)
	if err != nil {
		return err
	}

	ownerAddr, err := k.authKeeper.AddressCodec().StringToBytes(record.Owner)
	if err != nil {
		return err
	}
	owner := sdk.AccAddress(ownerAddr)

	// This check is necessary to prevent sending rewards to a blacklisted address
	if k.bankKeeper.BlockedAddr(owner) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", owner.String())
	}

	valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(record.Validator)
	if err != nil {
		return err
	}

	validatorFound := true
	_, err = k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		if !goerrors.Is(err, stakingtypes.ErrNoValidatorFound) {
			return err
		}

		validatorFound = false
	}

	delegationFound := true
	_, err = k.stakingKeeper.Delegation(ctx, record.GetModuleAddress(), valAddr)
	if err != nil {
		if !goerrors.Is(err, stakingtypes.ErrNoDelegation) {
			return err
		}

		delegationFound = false
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if validatorFound && delegationFound {
		// withdraw rewards into reward module account and send it to reward owner
		cacheCtx, write := sdkCtx.CacheContext()
		_, err = k.distKeeper.WithdrawDelegationRewards(cacheCtx, record.GetModuleAddress(), valAddr)
		if err != nil {
			return err
		}
		write()
	}

	// apply changes when the module account has positive balance
	balances := k.bankKeeper.GetAllBalances(ctx, record.GetModuleAddress())
	if !balances.Empty() {
		err = k.bankKeeper.SendCoins(ctx, record.GetModuleAddress(), owner, balances)
		if err != nil {
			return err
		}

		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeWithdrawTokenizeShareReward,
				sdk.NewAttribute(types.AttributeKeyWithdrawAddress, owner.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
			),
		)
	}
	return nil
}

// WithdrawTokenizeShareRecordReward withdraws rewards for owning a TokenizeShareRecord
func (k Keeper) WithdrawTokenizeShareRecordReward(ctx context.Context, ownerAddr sdk.AccAddress,
	recordID uint64,
) (sdk.Coins, error) {
	record, err := k.GetTokenizeShareRecord(ctx, recordID)
	if err != nil {
		return nil, err
	}

	// This check is necessary to prevent sending rewards to a blacklisted address
	if k.bankKeeper.BlockedAddr(ownerAddr) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", ownerAddr)
	}

	if record.Owner != ownerAddr.String() {
		return nil, types.ErrNotTokenizeShareRecordOwner
	}

	valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(record.Validator)
	if err != nil {
		return nil, err
	}

	_, err = k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	_, err = k.stakingKeeper.Delegation(ctx, record.GetModuleAddress(), valAddr)
	if err != nil {
		return nil, err
	}

	// withdraw rewards into reward module account and send it to reward owner
	_, err = k.distKeeper.WithdrawDelegationRewards(ctx, record.GetModuleAddress(), valAddr)
	if err != nil {
		return nil, err
	}

	// apply changes when the module account has positive balance
	rewards := k.bankKeeper.GetAllBalances(ctx, record.GetModuleAddress())
	if !rewards.Empty() {
		err = k.bankKeeper.SendCoins(ctx, record.GetModuleAddress(), ownerAddr, rewards)
		if err != nil {
			return nil, err
		}
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawTokenizeShareReward,
			sdk.NewAttribute(types.AttributeKeyWithdrawAddress, ownerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, rewards.String()),
		),
	)

	return rewards, nil
}
