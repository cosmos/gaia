package keeper

import (
	"context"
	goerrors "errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v24/x/liquid/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	*Keeper
}

var _ types.QueryServer = Querier{}

func NewQuerier(keeper *Keeper) Querier {
	return Querier{Keeper: keeper}
}

// Params queries the staking parameters
func (k Querier) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

// LiquidValidator queries for a LiquidValidator record by validator address
func (k Querier) LiquidValidator(c context.Context, req *types.QueryLiquidValidatorRequest) (*types.QueryLiquidValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(req.ValidatorAddr)
	if err != nil {
		return nil, err
	}
	lv, err := k.GetLiquidValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}
	return &types.QueryLiquidValidatorResponse{LiquidValidator: lv}, nil
}

// TokenizeShareRecordById queries for individual tokenize share record information by share by id
func (k Querier) TokenizeShareRecordById(c context.Context, req *types.QueryTokenizeShareRecordByIdRequest) (*types.QueryTokenizeShareRecordByIdResponse, error) { //nolint:revive // fixing this would require changing the .proto files, so we might as well leave it alone
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	record, err := k.GetTokenizeShareRecord(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.QueryTokenizeShareRecordByIdResponse{
		Record: record,
	}, nil
}

// TokenizeShareRecordByDenom queries for individual tokenize share record information by share denom
func (k Querier) TokenizeShareRecordByDenom(c context.Context, req *types.QueryTokenizeShareRecordByDenomRequest) (*types.QueryTokenizeShareRecordByDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	record, err := k.GetTokenizeShareRecordByDenom(ctx, req.Denom)
	if err != nil {
		return nil, err
	}

	return &types.QueryTokenizeShareRecordByDenomResponse{
		Record: record,
	}, nil
}

// TokenizeShareRecordsOwned queries tokenize share records by address
func (k Querier) TokenizeShareRecordsOwned(c context.Context, req *types.QueryTokenizeShareRecordsOwnedRequest) (*types.QueryTokenizeShareRecordsOwnedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	owner, err := k.authKeeper.AddressCodec().StringToBytes(req.Owner)
	if err != nil {
		return nil, err
	}
	records := k.GetTokenizeShareRecordsByOwner(ctx, owner)

	return &types.QueryTokenizeShareRecordsOwnedResponse{
		Records: records,
	}, nil
}

// AllTokenizeShareRecords queries for all tokenize share records
func (k Querier) AllTokenizeShareRecords(c context.Context, req *types.QueryAllTokenizeShareRecordsRequest) (*types.QueryAllTokenizeShareRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	var records []types.TokenizeShareRecord

	store := k.storeService.OpenKVStore(ctx)
	valStore := prefix.NewStore(runtime.KVStoreAdapter(store), types.TokenizeShareRecordPrefix)
	pageRes, err := query.FilteredPaginate(valStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		var tokenizeShareRecord types.TokenizeShareRecord
		if err := k.cdc.Unmarshal(value, &tokenizeShareRecord); err != nil {
			return false, err
		}

		if accumulate {
			records = append(records, tokenizeShareRecord)
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllTokenizeShareRecordsResponse{
		Records:    records,
		Pagination: pageRes,
	}, nil
}

// LastTokenizeShareRecordId queries for last tokenize share record id
func (k Querier) LastTokenizeShareRecordId(c context.Context, req *types.QueryLastTokenizeShareRecordIdRequest) (*types.QueryLastTokenizeShareRecordIdResponse, error) { //nolint:revive // fixing this would require changing the .proto files, so we might as well leave it alone
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryLastTokenizeShareRecordIdResponse{
		Id: k.GetLastTokenizeShareRecordID(ctx),
	}, nil
}

// TotalTokenizeSharedAssets queries for total tokenized staked assets
func (k Querier) TotalTokenizeSharedAssets(c context.Context, req *types.QueryTotalTokenizeSharedAssetsRequest) (*types.QueryTotalTokenizeSharedAssetsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	records := k.GetAllTokenizeShareRecords(ctx)
	totalTokenizeShared := math.ZeroInt()

	for _, record := range records {
		moduleAcc := record.GetModuleAddress()
		valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(record.Validator)
		if err != nil {
			return nil, err
		}

		validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			return nil, err
		}

		delegation, err := k.stakingKeeper.GetDelegation(ctx, moduleAcc, valAddr)
		if err != nil {
			return nil, err
		}

		tokens := validator.TokensFromShares(delegation.Shares)
		totalTokenizeShared = totalTokenizeShared.Add(tokens.RoundInt())
	}

	bondDenom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryTotalTokenizeSharedAssetsResponse{
		Value: sdk.NewCoin(bondDenom, totalTokenizeShared),
	}, nil
}

// TotalLiquidStaked queries for total tokenized staked tokens
// Liquid staked tokens are either tokenized delegations or delegations
// owned by a module account
func (k Querier) TotalLiquidStaked(c context.Context, req *types.QueryTotalLiquidStaked) (*types.QueryTotalLiquidStakedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	totalLiquidStaked := k.GetTotalLiquidStakedTokens(ctx).String()
	return &types.QueryTotalLiquidStakedResponse{
		Tokens: totalLiquidStaked,
	}, nil
}

// TokenizeShareLockInfo queries status of an account's tokenize share lock
func (k Querier) TokenizeShareLockInfo(c context.Context, req *types.QueryTokenizeShareLockInfo) (*types.QueryTokenizeShareLockInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	address, err := k.authKeeper.AddressCodec().StringToBytes(req.Address)
	if err != nil {
		panic(err)
	}

	lockStatus, completionTime := k.GetTokenizeSharesLock(ctx, address)

	timeString := ""
	if !completionTime.IsZero() {
		timeString = completionTime.String()
	}

	return &types.QueryTokenizeShareLockInfoResponse{
		Status:         lockStatus.String(),
		ExpirationTime: timeString,
	}, nil
}

// TokenizeShareRecordReward returns estimated amount of reward from tokenize share record ownership
func (k Keeper) TokenizeShareRecordReward(c context.Context, req *types.QueryTokenizeShareRecordRewardRequest) (*types.QueryTokenizeShareRecordRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	totalRewards := sdk.DecCoins{}
	rewards := []types.TokenizeShareRecordReward{}

	ownerAddr, err := k.authKeeper.AddressCodec().StringToBytes(req.OwnerAddress)
	if err != nil {
		return nil, err
	}
	records := k.GetTokenizeShareRecordsByOwner(ctx, ownerAddr)
	for _, record := range records {
		valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(record.Validator)
		if err != nil {
			return nil, err
		}

		moduleAddr := record.GetModuleAddress()
		moduleBalance := k.bankKeeper.GetAllBalances(ctx, moduleAddr)
		moduleBalanceDecCoins := sdk.NewDecCoinsFromCoins(moduleBalance...)

		validatorFound := true
		val, err := k.stakingKeeper.Validator(ctx, valAddr)
		if err != nil {
			if !goerrors.Is(err, stakingtypes.ErrNoValidatorFound) {
				return nil, err
			}

			validatorFound = false
		}

		delegationFound := true
		del, err := k.stakingKeeper.Delegation(ctx, moduleAddr, valAddr)
		if err != nil {
			if !goerrors.Is(err, stakingtypes.ErrNoDelegation) {
				return nil, err
			}

			delegationFound = false
		}

		if validatorFound && delegationFound {
			// withdraw rewards
			endingPeriod, err := k.distKeeper.IncrementValidatorPeriod(ctx, val)
			if err != nil {
				return nil, err
			}

			recordReward, err := k.distKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
			if err != nil {
				return nil, err
			}

			rewards = append(rewards, types.TokenizeShareRecordReward{
				RecordId: record.Id,
				Reward:   recordReward.Add(moduleBalanceDecCoins...),
			})
			totalRewards = totalRewards.Add(recordReward...).Add(moduleBalanceDecCoins...)
		} else if !moduleBalance.IsZero() {
			rewards = append(rewards, types.TokenizeShareRecordReward{
				RecordId: record.Id,
				Reward:   moduleBalanceDecCoins,
			})
			totalRewards = totalRewards.Add(moduleBalanceDecCoins...)
		}
	}

	return &types.QueryTokenizeShareRecordRewardResponse{
		Rewards: rewards,
		Total:   totalRewards,
	}, nil
}
