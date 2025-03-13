package keeper

import (
	"context"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v23/x/liquid/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the staking MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateParams defines a method to perform updating of params for the x/liquid module.
func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	// store params
	if err := k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// Tokenizes shares associated with a delegation by creating a tokenize share record
// and returning tokens with a denom of the format {validatorAddress}/{recordId}
func (k msgServer) TokenizeShares(goCtx context.Context, msg *types.MsgTokenizeShares) (*types.MsgTokenizeSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, valErr := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if valErr != nil {
		return nil, valErr
	}
	validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	_, err = k.authKeeper.AddressCodec().StringToBytes(msg.TokenizedShareOwner)
	if err != nil {
		return nil, err
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid shares amount")
	}

	// Check if the delegator has disabled tokenization
	lockStatus, unlockTime := k.GetTokenizeSharesLock(ctx, delegatorAddress)
	if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED {
		return nil, types.ErrTokenizeSharesDisabledForAccount
	}
	if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING {
		return nil, types.ErrTokenizeSharesDisabledForAccount.Wrapf("tokenization will be allowed at %s", unlockTime)
	}

	bondDenom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, types.ErrOnlyBondDenomAllowdForTokenize
	}

	acc := k.authKeeper.GetAccount(ctx, delegatorAddress)
	if acc != nil {
		acc, ok := acc.(vesting.VestingAccount)
		if ok {
			// if account is a vesting account, it checks if free delegation (non-vesting delegation) is not exceeding
			// the tokenize share amount and execute further tokenize share process
			// tokenize share is reducing unlocked tokens delegation from the vesting account and further process
			// is not causing issues
			if !CheckVestedDelegationInVestingAccount(acc, ctx.BlockTime(), msg.Amount) {
				return nil, types.ErrExceedingFreeVestingDelegations
			}
		}
	}

	shares, err := k.stakingKeeper.ValidateUnbondAmount(
		ctx, delegatorAddress, valAddr, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
	}

	// sanity check to avoid creating a tokenized share record with zero shares
	if shares.IsZero() {
		return nil, errorsmod.Wrap(types.ErrInsufficientShares, "cannot tokenize zero shares")
	}

	// Check that the delegator has no ongoing redelegations to the validator
	found, err := k.stakingKeeper.HasReceivingRedelegation(ctx, delegatorAddress, valAddr)
	if err != nil {
		return nil, err
	}
	if found {
		return nil, types.ErrRedelegationInProgress
	}

	if err := k.SafelyIncreaseTotalLiquidStakedTokens(ctx, msg.Amount.Amount, true); err != nil {
		return nil, err
	}
	_, err = k.SafelyIncreaseValidatorLiquidShares(ctx, valAddr, shares, true)
	if err != nil {
		return nil, err
	}

	recordID := k.GetLastTokenizeShareRecordID(ctx) + 1
	k.SetLastTokenizeShareRecordID(ctx, recordID)

	record := types.TokenizeShareRecord{
		Id:            recordID,
		Owner:         msg.TokenizedShareOwner,
		ModuleAccount: fmt.Sprintf("%s%d", types.TokenizeShareModuleAccountPrefix, recordID),
		Validator:     msg.ValidatorAddress,
	}

	// note: this returnAmount can be slightly off from the original delegation amount if there
	// is a decimal to int precision error
	returnAmount, err := k.stakingKeeper.Unbond(ctx, delegatorAddress, valAddr, shares)
	if err != nil {
		return nil, err
	}

	if validator.IsBonded() {
		coins := sdk.NewCoins(sdk.NewCoin(bondDenom, returnAmount))
		err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, coins)
		if err != nil {
			return nil, err
		}
	}

	// Note: UndelegateCoinsFromModuleToAccount is internally calling TrackUndelegation for vesting account
	returnCoin := sdk.NewCoin(bondDenom, returnAmount)
	err = k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, stakingtypes.NotBondedPoolName, delegatorAddress,
		sdk.Coins{returnCoin})
	if err != nil {
		return nil, err
	}

	// Re-calculate the shares in case there was rounding precision during the undelegation
	newShares, err := validator.SharesFromTokens(returnAmount)
	if err != nil {
		return nil, err
	}

	// The share tokens returned maps 1:1 with shares
	shareToken := sdk.NewCoin(record.GetShareTokenDenom(), newShares.TruncateInt())

	err = k.bankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.Coins{shareToken})
	if err != nil {
		return nil, err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delegatorAddress, sdk.Coins{shareToken})
	if err != nil {
		return nil, err
	}

	// create reward ownership record
	err = k.AddTokenizeShareRecord(ctx, record)
	if err != nil {
		return nil, err
	}
	// send coins to module account
	err = k.bankKeeper.SendCoins(ctx, delegatorAddress, record.GetModuleAddress(), sdk.Coins{returnCoin})
	if err != nil {
		return nil, err
	}

	// Note: it is needed to get latest validator object to get Keeper.Delegate function work properly
	validator, err = k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	// delegate from module account
	_, err = k.stakingKeeper.Delegate(ctx, record.GetModuleAddress(), returnAmount, stakingtypes.Unbonded, validator,
		true)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenizeShares,
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyShareOwner, msg.TokenizedShareOwner),
			sdk.NewAttribute(types.AttributeKeyShareRecordID, fmt.Sprintf("%d", record.Id)),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyTokenizedShares, shareToken.String()),
		),
	)

	return &types.MsgTokenizeSharesResponse{
		Amount: shareToken,
	}, nil
}

// Converts tokenized shares back into a native delegation
func (k msgServer) RedeemTokensForShares(goCtx context.Context, msg *types.MsgRedeemTokensForShares) (*types.MsgRedeemTokensForSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid shares amount")
	}

	shareToken := msg.Amount
	balance := k.bankKeeper.GetBalance(ctx, delegatorAddress, shareToken.Denom)
	if balance.Amount.LT(shareToken.Amount) {
		return nil, types.ErrNotEnoughBalance
	}

	record, err := k.GetTokenizeShareRecordByDenom(ctx, shareToken.Denom)
	if err != nil {
		return nil, err
	}

	valAddr, valErr := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(record.Validator)
	if valErr != nil {
		return nil, valErr
	}

	validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	delegation, err := k.stakingKeeper.GetDelegation(ctx, record.GetModuleAddress(), valAddr)
	if err != nil {
		return nil, err
	}

	// Similar to undelegations, if the account is attempting to tokenize the full delegation,
	// but there's a precision error due to the decimal to int conversion, round up to the
	// full decimal amount before modifying the delegation
	shares := math.LegacyNewDecFromInt(shareToken.Amount)
	if shareToken.Amount.Equal(delegation.Shares.TruncateInt()) {
		shares = delegation.Shares
	}
	tokens := validator.TokensFromShares(shares).TruncateInt()

	// prevent redemption that returns a 0 amount
	if tokens.IsZero() {
		return nil, types.ErrTinyRedemptionAmount
	}

	// If this redemption is NOT from a liquid staking provider, decrement the total liquid staked
	// If the redemption was from a liquid staking provider, the shares are still considered
	// liquid, even in their non-tokenized form (since they are owned by a liquid staking provider)
	if !k.DelegatorIsLiquidStaker(delegatorAddress) {
		if err := k.DecreaseTotalLiquidStakedTokens(ctx, tokens); err != nil {
			return nil, err
		}
		_, err = k.DecreaseValidatorLiquidShares(ctx, valAddr, shares)
		if err != nil {
			return nil, err
		}
	}

	returnAmount, err := k.stakingKeeper.Unbond(ctx, record.GetModuleAddress(), valAddr, shares)
	if err != nil {
		return nil, err
	}

	if validator.IsBonded() {
		bondDenom, err := k.stakingKeeper.BondDenom(ctx)
		if err != nil {
			return nil, err
		}

		coins := sdk.NewCoins(sdk.NewCoin(bondDenom, returnAmount))
		err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, coins)
		if err != nil {
			return nil, err
		}
	}

	// Note: since delegation object has been changed from unbond call, it gets latest delegation
	_, err = k.stakingKeeper.GetDelegation(ctx, record.GetModuleAddress(), valAddr)
	if err != nil && !errors.Is(err, stakingtypes.ErrNoDelegation) {
		return nil, err
	}

	// this err will be ErrNoDelegation
	if err != nil {
		if err := k.WithdrawSingleShareRecordReward(ctx, record.Id); err != nil {
			return nil, err
		}
		err = k.DeleteTokenizeShareRecord(ctx, record.Id)
		if err != nil {
			return nil, err
		}
	}

	// send share tokens to NotBondedPool and burn
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegatorAddress, stakingtypes.NotBondedPoolName,
		sdk.Coins{shareToken})
	if err != nil {
		return nil, err
	}
	err = k.bankKeeper.BurnCoins(ctx, stakingtypes.NotBondedPoolName, sdk.Coins{shareToken})
	if err != nil {
		return nil, err
	}

	bondDenom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}
	// send equivalent amount of tokens to the delegator
	returnCoin := sdk.NewCoin(bondDenom, returnAmount)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, stakingtypes.NotBondedPoolName, delegatorAddress,
		sdk.Coins{returnCoin})
	if err != nil {
		return nil, err
	}

	// Note: it is needed to get latest validator object to get Keeper.Delegate function work properly
	validator, err = k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	// convert the share tokens to delegated status
	// Note: Delegate(substractAccount => true) -> DelegateCoinsFromAccountToModule -> TrackDelegation for vesting account
	_, err = k.stakingKeeper.Delegate(ctx, delegatorAddress, returnAmount, stakingtypes.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRedeemShares,
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(types.AttributeKeyValidator, validator.OperatorAddress),
			sdk.NewAttribute(types.AttributeKeyAmount, shareToken.String()),
		),
	)

	return &types.MsgRedeemTokensForSharesResponse{
		Amount: returnCoin,
	}, nil
}

// Transfers the ownership of rewards associated with a tokenize share record
func (k msgServer) TransferTokenizeShareRecord(goCtx context.Context, msg *types.MsgTransferTokenizeShareRecord) (*types.MsgTransferTokenizeShareRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	record, err := k.GetTokenizeShareRecord(ctx, msg.TokenizeShareRecordId)
	if err != nil {
		return nil, types.ErrTokenizeShareRecordNotExists
	}

	if record.Owner != msg.Sender {
		return nil, types.ErrNotTokenizeShareRecordOwner
	}

	// Remove old account reference
	oldOwner, err := k.authKeeper.AddressCodec().StringToBytes(record.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}
	k.deleteTokenizeShareRecordWithOwner(ctx, oldOwner, record.Id)

	record.Owner = msg.NewOwner
	k.setTokenizeShareRecord(ctx, record)

	// Set new account reference
	newOwner, err := k.authKeeper.AddressCodec().StringToBytes(record.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}
	k.setTokenizeShareRecordWithOwner(ctx, newOwner, record.Id)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTransferTokenizeShareRecord,
			sdk.NewAttribute(types.AttributeKeyShareRecordID, fmt.Sprintf("%d", msg.TokenizeShareRecordId)),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyShareOwner, msg.NewOwner),
		),
	)

	return &types.MsgTransferTokenizeShareRecordResponse{}, nil
}

// DisableTokenizeShares prevents an address from tokenizing any of their delegations
func (k msgServer) DisableTokenizeShares(ctx context.Context, msg *types.MsgDisableTokenizeShares) (*types.MsgDisableTokenizeSharesResponse, error) {
	delegator, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}

	// If tokenized shares is already disabled, alert the user
	lockStatus, completionTime := k.GetTokenizeSharesLock(ctx, delegator)
	if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED {
		return nil, types.ErrTokenizeSharesAlreadyDisabledForAccount
	}

	// If the tokenized shares lock is expiring, remove the pending unlock from the queue
	if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING {
		k.CancelTokenizeShareLockExpiration(ctx, delegator, completionTime)
	}

	// Create a new tokenization lock for the user
	// Note: if there is a lock expiration in progress, this will override the expiration
	k.AddTokenizeSharesLock(ctx, delegator)

	return &types.MsgDisableTokenizeSharesResponse{}, nil
}

// EnableTokenizeShares begins the countdown after which tokenizing shares by the
// sender address is re-allowed, which will complete after the unbonding period
func (k msgServer) EnableTokenizeShares(ctx context.Context, msg *types.MsgEnableTokenizeShares) (*types.MsgEnableTokenizeSharesResponse, error) {
	delegator, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}

	// If tokenized shares aren't current disabled, alert the user
	lockStatus, unlockTime := k.GetTokenizeSharesLock(ctx, delegator)
	if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_UNLOCKED {
		return nil, types.ErrTokenizeSharesAlreadyEnabledForAccount
	}
	if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING {
		return nil, types.ErrTokenizeSharesAlreadyEnabledForAccount.Wrapf(
			"tokenize shares re-enablement already in progress, ending at %s", unlockTime)
	}

	// Otherwise queue the unlock
	completionTime, err := k.QueueTokenizeSharesAuthorization(ctx, delegator)
	if err != nil {
		panic(err)
	}

	return &types.MsgEnableTokenizeSharesResponse{CompletionTime: completionTime}, nil
}

// WithdrawTokenizeShareRecordReward defines a method to withdraw reward for owning TokenizeShareRecord
func (k msgServer) WithdrawTokenizeShareRecordReward(goCtx context.Context, msg *types.MsgWithdrawTokenizeShareRecordReward) (*types.MsgWithdrawTokenizeShareRecordRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAddr, err := k.authKeeper.AddressCodec().StringToBytes(msg.OwnerAddress)
	if err != nil {
		return nil, err
	}

	_, err = k.Keeper.WithdrawTokenizeShareRecordReward(ctx, ownerAddr, msg.RecordId)
	if err != nil {
		return nil, err
	}

	return &types.MsgWithdrawTokenizeShareRecordRewardResponse{}, nil
}

// WithdrawAllTokenizeShareRecordReward defines a method to withdraw reward for owning TokenizeShareRecord
func (k msgServer) WithdrawAllTokenizeShareRecordReward(goCtx context.Context, msg *types.MsgWithdrawAllTokenizeShareRecordReward) (*types.MsgWithdrawAllTokenizeShareRecordRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAddr, err := k.authKeeper.AddressCodec().StringToBytes(msg.OwnerAddress)
	if err != nil {
		return nil, err
	}

	_, err = k.Keeper.WithdrawAllTokenizeShareRecordReward(ctx, ownerAddr)
	if err != nil {
		return nil, err
	}

	return &types.MsgWithdrawAllTokenizeShareRecordRewardResponse{}, nil
}
