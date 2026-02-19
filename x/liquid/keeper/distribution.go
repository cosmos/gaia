package keeper

import (
	"context"
	goerrors "errors"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v27/x/liquid/types"
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

// withdraw reward for all owning TokenizeShareRecord
func (k Keeper) WithdrawAllTokenizeShareRecordReward(ctx sdk.Context, ownerAddr sdk.AccAddress) (sdk.Coins, error) {
	// This check is necessary to prevent sending rewards to a blacklisted address
	if k.bankKeeper.BlockedAddr(ownerAddr) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive external funds", ownerAddr)
	}

	totalRewards := sdk.Coins{}

	records := k.GetTokenizeShareRecordsByOwner(ctx, ownerAddr)

	for _, record := range records {
		valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(record.Validator)
		if err != nil {
			return nil, err
		}

		_, err = k.stakingKeeper.Validator(ctx, valAddr)
		if err != nil && !goerrors.Is(err, stakingtypes.ErrNoValidatorFound) {
			return nil, err
		}

		// if the error is ErrNoValidatorFound
		if err != nil {
			continue
		}

		_, err = k.stakingKeeper.Delegation(ctx, record.GetModuleAddress(), valAddr)
		if err != nil && !goerrors.Is(err, stakingtypes.ErrNoDelegation) {
			return nil, err
		}

		// if the error is ErrNoDelegation
		if err != nil {
			continue
		}

		// withdraw rewards into reward module account and send it to reward owner
		cacheCtx, write := ctx.CacheContext()
		_, err = k.distKeeper.WithdrawDelegationRewards(cacheCtx, record.GetModuleAddress(), valAddr)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
			continue
		}

		// apply changes when the module account has positive balance
		balances := k.bankKeeper.GetAllBalances(cacheCtx, record.GetModuleAddress())
		if !balances.Empty() {
			err = k.bankKeeper.SendCoins(cacheCtx, record.GetModuleAddress(), ownerAddr, balances)
			if err != nil {
				k.Logger(ctx).Error(err.Error())
				continue
			}
			write()
			totalRewards = totalRewards.Add(balances...)
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawTokenizeShareReward,
			sdk.NewAttribute(types.AttributeKeyWithdrawAddress, ownerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, totalRewards.String()),
		),
	)

	return totalRewards, nil
}
