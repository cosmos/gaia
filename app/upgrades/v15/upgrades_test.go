package v15_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v15/app/helpers"
	v15 "github.com/cosmos/gaia/v15/app/upgrades/v15"
)

func TestUpgradeSigningInfos(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	slashingKeeper := gaiaApp.SlashingKeeper

	signingInfosNum := 8
	emptyAddrSigningInfo := make(map[string]struct{})

	// create some dummy signing infos, half of which with an empty address field
	for i := 0; i < signingInfosNum; i++ {
		pubKey, err := mock.NewPV().GetPubKey()
		require.NoError(t, err)

		consAddr := sdk.ConsAddress(pubKey.Address())
		info := slashingtypes.NewValidatorSigningInfo(
			consAddr,
			0,
			0,
			time.Unix(0, 0),
			false,
			0,
		)

		if i < signingInfosNum/2 {
			info.Address = ""
			emptyAddrSigningInfo[consAddr.String()] = struct{}{}
		}

		slashingKeeper.SetValidatorSigningInfo(ctx, consAddr, info)
		require.NoError(t, err)
	}

	require.Equal(t, signingInfosNum/2, len(emptyAddrSigningInfo))

	// check that signing info are correctly set before migration
	slashingKeeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		if _, ok := emptyAddrSigningInfo[address.String()]; ok {
			require.Empty(t, info.Address)
		} else {
			require.NotEmpty(t, info.Address)
		}

		return false
	})

	// upgrade signing infos
	v15.UpgradeSigningInfos(ctx, slashingKeeper)

	// check that all signing info are updated as expected after migration
	slashingKeeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		require.NotEmpty(t, info.Address)

		return false
	})
}

func TestUpgradeMinCommissionRate(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	// set min commission rate to 0
	stakingParams := gaiaApp.StakingKeeper.GetParams(ctx)
	stakingParams.MinCommissionRate = sdk.ZeroDec()
	err := gaiaApp.StakingKeeper.SetParams(ctx, stakingParams)
	require.NoError(t, err)

	stakingKeeper := gaiaApp.StakingKeeper
	valNum := len(stakingKeeper.GetAllValidators(ctx))

	// create 3 new validators
	for i := 0; i < 3; i++ {
		pk := ed25519.GenPrivKeyFromSecret([]byte{uint8(i)}).PubKey()
		val, err := stakingtypes.NewValidator(
			sdk.ValAddress(pk.Address()),
			pk,
			stakingtypes.Description{},
		)
		require.NoError(t, err)
		// set random commission rate
		val.Commission.CommissionRates.Rate = sdk.NewDecWithPrec(tmrand.Int63n(100), 2)
		stakingKeeper.SetValidator(ctx, val)
		valNum++
	}

	validators := stakingKeeper.GetAllValidators(ctx)
	require.Equal(t, valNum, len(validators))

	// pre-test min commission rate is 0
	require.Equal(t, stakingKeeper.GetParams(ctx).MinCommissionRate, sdk.ZeroDec(), "non-zero previous min commission rate")

	// run the test and confirm the values have been updated
	require.NoError(t, v15.UpgradeMinCommissionRate(ctx, *stakingKeeper))

	newStakingParams := stakingKeeper.GetParams(ctx)
	require.NotEqual(t, newStakingParams.MinCommissionRate, sdk.ZeroDec(), "failed to update min commission rate")
	require.Equal(t, newStakingParams.MinCommissionRate, sdk.NewDecWithPrec(5, 2), "failed to update min commission rate")

	for _, val := range stakingKeeper.GetAllValidators(ctx) {
		require.True(t, val.Commission.CommissionRates.Rate.GTE(newStakingParams.MinCommissionRate), "failed to update update commission rate for validator %s", val.GetOperator())
	}
}

func TestClawbackVestingFunds(t *testing.T) {
	gaiaApp := helpers.Setup(t)

	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bankKeeper := gaiaApp.BankKeeper
	accountKeeper := gaiaApp.AccountKeeper
	distrKeeper := gaiaApp.DistrKeeper
	stakingKeeper := gaiaApp.StakingKeeper

	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{Height: 1})
	ctx = ctx.WithBlockHeader(tmproto.Header{Height: ctx.BlockHeight(), Time: now})

	validator := stakingKeeper.GetAllValidators(ctx)[0]
	bondDenom := stakingKeeper.GetParams(ctx).BondDenom

	// create continuous vesting account
	origCoins := sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 100))
	addr := sdk.AccAddress([]byte("cosmos145hytrc49m0hn6fphp8d5h4xspwkawcuzmx498"))

	vestingAccount := vesting.NewContinuousVestingAccount(
		authtypes.NewBaseAccountWithAddress(addr),
		origCoins,
		now.Unix(),
		endTime.Unix(),
	)

	require.True(t, vestingAccount.GetVestingCoins(now).IsEqual(origCoins))

	accountKeeper.SetAccount(ctx, vestingAccount)

	// check vesting account balance was set correctly
	require.NoError(t, bankKeeper.ValidateBalance(ctx, addr))
	require.Empty(t, bankKeeper.GetAllBalances(ctx, addr))

	// send original vesting coin amount
	require.NoError(t, banktestutil.FundAccount(bankKeeper, ctx, addr, origCoins))
	require.True(t, origCoins.IsEqual(bankKeeper.GetAllBalances(ctx, addr)))

	initBal := bankKeeper.GetAllBalances(ctx, vestingAccount.GetAddress())
	require.True(t, initBal.IsEqual(origCoins))

	// save validator tokens
	oldValTokens := validator.Tokens

	// delegate all vesting account tokens
	_, err := stakingKeeper.Delegate(
		ctx,
		vestingAccount.GetAddress(),
		origCoins.AmountOf(bondDenom),
		stakingtypes.Unbonded,
		validator,
		true)
	require.NoError(t, err)

	// check that the validator's tokens and shares increased
	validator = stakingKeeper.GetAllValidators(ctx)[0]
	del, found := stakingKeeper.GetDelegation(ctx, addr, validator.GetOperator())
	require.True(t, found)
	require.True(t, validator.Tokens.Equal(oldValTokens.Add(origCoins.AmountOf(bondDenom))))
	require.Equal(
		t,
		validator.TokensFromShares(del.Shares),
		math.LegacyNewDec(origCoins.AmountOf(bondDenom).Int64()),
	)

	// check vesting account delegations
	vestingAccount = accountKeeper.GetAccount(ctx, addr).(*vesting.ContinuousVestingAccount)
	require.Equal(t, vestingAccount.GetDelegatedVesting(), origCoins)
	require.Empty(t, vestingAccount.GetDelegatedFree())

	// check that migration succeeds when all coins are already vested
	require.NoError(t, v15.ClawbackVestingFunds(ctx.WithBlockTime(endTime), addr, &gaiaApp.AppKeepers))

	// vest half of the tokens
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	currVestingCoins := vestingAccount.GetVestingCoins(ctx.BlockTime())
	currVestedCoins := vestingAccount.GetVestedCoins(ctx.BlockTime())

	require.True(t, currVestingCoins.IsEqual(origCoins.QuoInt(math.NewInt(2))))
	require.True(t, currVestedCoins.IsEqual(origCoins.QuoInt(math.NewInt(2))))

	// execute migration script
	require.NoError(t, v15.ClawbackVestingFunds(ctx, addr, &gaiaApp.AppKeepers))

	// check that the validator's delegation is removed and that
	// their total tokens decreased
	validator = stakingKeeper.GetAllValidators(ctx)[0]
	_, found = stakingKeeper.GetDelegation(ctx, addr, validator.GetOperator())
	require.False(t, found)
	require.Equal(
		t,
		validator.TokensFromShares(validator.DelegatorShares),
		math.LegacyNewDec(oldValTokens.Int64()),
	)

	// verify that all modules can end/begin blocks
	gaiaApp.EndBlock(abci.RequestEndBlock{})
	gaiaApp.BeginBlock(
		abci.RequestBeginBlock{
			Header: tmproto.Header{
				ChainID: ctx.ChainID(),
				Height:  ctx.BlockHeight() + 1,
			},
		},
	)

	// check that the resulting account is of BaseAccount type now
	account, ok := accountKeeper.GetAccount(ctx, addr).(*authtypes.BaseAccount)
	require.True(t, ok)
	// check that the account values are still the same
	require.EqualValues(t, account, vestingAccount.BaseAccount)

	// check that the account's balance still has the vested tokens
	require.True(t, bankKeeper.GetAllBalances(ctx, addr).IsEqual(currVestedCoins))
	// check that the community pool balance received the vesting tokens
	require.True(
		t,
		distrKeeper.GetFeePoolCommunityCoins(ctx).
			IsEqual(sdk.NewDecCoinsFromCoins(currVestingCoins...)),
	)

	// verify that normal operations work in banking and staking
	_, err = stakingKeeper.Delegate(
		ctx, addr,
		sdk.NewInt(30),
		stakingtypes.Unbonded,
		validator,
		true)
	require.NoError(t, err)

	newAddr := sdk.AccAddress([]byte("cosmos1qqp9myctmh8mh2y7gynlsnw4y2wz3s3089dak6"))
	err = bankKeeper.SendCoins(
		ctx,
		addr,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(bondDenom, sdk.NewInt(10))),
	)
	require.NoError(t, err)
}

func TestSetMinInitialDepositRatio(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	err := v15.SetMinInitialDepositRatio(ctx, *gaiaApp.GovKeeper)
	require.NoError(t, err)

	minInitialDepositRatioStr := gaiaApp.GovKeeper.GetParams(ctx).MinInitialDepositRatio
	minInitialDepositRatio, err := math.LegacyNewDecFromStr(minInitialDepositRatioStr)
	require.NoError(t, err)
	require.True(t, minInitialDepositRatio.Equal(sdk.NewDecWithPrec(1, 1)))
}

func TestUpgradeEscrowAccounts(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	bankKeeper := gaiaApp.BankKeeper
	transferKeeper := gaiaApp.TransferKeeper

	escrowUpdates := v15.GetEscrowUpdates(ctx)

	// check escrow accounts are empty
	for addr, coins := range escrowUpdates {
		require.Empty(t, bankKeeper.GetAllBalances(ctx, sdk.MustAccAddressFromBech32(addr)))
		for _, coin := range coins {
			require.Equal(t, sdk.ZeroInt(), transferKeeper.GetTotalEscrowForDenom(ctx, coin.Denom).Amount)
		}
	}

	// execute the upgrade
	v15.UpgradeEscrowAccounts(ctx, bankKeeper, transferKeeper)

	// check that new assets are minted and transferred to the escrow accounts
	numUpdate := 0
	for addr, coins := range escrowUpdates {
		for _, coin := range coins {
			require.Equal(t, coin, bankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(addr), coin.Denom))
			// check that the total escrow amount for the denom is updated
			require.Equal(t, coin, transferKeeper.GetTotalEscrowForDenom(ctx, coin.Denom))
			numUpdate++
		}
	}

	// verify that all tree discrepancies are covered in the update
	require.Equal(t, 3, numUpdate)
}
