package v15_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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

	"github.com/cosmos/gaia/v18/app/helpers"
	v15 "github.com/cosmos/gaia/v18/app/upgrades/v15"
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
	stakingParams, err := gaiaApp.StakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	stakingParams.MinCommissionRate = math.LegacyZeroDec()
	err = gaiaApp.StakingKeeper.SetParams(ctx, stakingParams)
	require.NoError(t, err)

	stakingKeeper := gaiaApp.StakingKeeper
	validators, err := stakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	valNum := len(validators)

	// create 3 new validators
	for i := 0; i < 3; i++ {
		pk := ed25519.GenPrivKeyFromSecret([]byte{uint8(i)}).PubKey()
		val, err := stakingtypes.NewValidator(
			sdk.ValAddress(pk.Address()).String(),
			pk,
			stakingtypes.Description{},
		)
		require.NoError(t, err)
		// set random commission rate
		val.Commission.CommissionRates.Rate = math.LegacyNewDecWithPrec(tmrand.Int63n(100), 2)
		stakingKeeper.SetValidator(ctx, val)
		valNum++
	}

	validators, err = stakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Equal(t, valNum, len(validators))

	// pre-test min commission rate is 0
	params, err := stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, params.MinCommissionRate, math.LegacyZeroDec(), "non-zero previous min commission rate")

	// run the test and confirm the values have been updated
	require.NoError(t, v15.UpgradeMinCommissionRate(ctx, *stakingKeeper))

	newStakingParams, err := stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.NotEqual(t, newStakingParams.MinCommissionRate, math.LegacyZeroDec(), "failed to update min commission rate")
	require.Equal(t, newStakingParams.MinCommissionRate, math.LegacyNewDecWithPrec(5, 2), "failed to update min commission rate")

	validators, err = stakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	for _, val := range validators {
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

	validators, err := stakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	validator := validators[0]
	params, err := stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	bondDenom := params.BondDenom

	// create continuous vesting account
	origCoins := sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 100))
	addr := sdk.AccAddress([]byte("cosmos145hytrc49m0hn6fphp8d5h4xspwkawcuzmx498"))

	vestingAccount, err := vesting.NewContinuousVestingAccount(
		authtypes.NewBaseAccountWithAddress(addr),
		origCoins,
		now.Unix(),
		endTime.Unix(),
	)
	require.NoError(t, err)

	require.True(t, vestingAccount.GetVestingCoins(now).Equal(origCoins))

	accountKeeper.SetAccount(ctx, vestingAccount)

	// check vesting account balance was set correctly
	require.NoError(t, bankKeeper.ValidateBalance(ctx, addr))
	require.Empty(t, bankKeeper.GetAllBalances(ctx, addr))

	// send original vesting coin amount
	require.NoError(t, banktestutil.FundAccount(ctx, bankKeeper, addr, origCoins))
	require.True(t, origCoins.Equal(bankKeeper.GetAllBalances(ctx, addr)))

	initBal := bankKeeper.GetAllBalances(ctx, vestingAccount.GetAddress())
	require.True(t, initBal.Equal(origCoins))

	// save validator tokens
	oldValTokens := validator.Tokens

	// delegate all vesting account tokens
	_, err = stakingKeeper.Delegate(
		ctx,
		vestingAccount.GetAddress(),
		origCoins.AmountOf(bondDenom),
		stakingtypes.Unbonded,
		validator,
		true)
	require.NoError(t, err)

	// check that the validator's tokens and shares increased
	validators, err = stakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	validator = validators[0]
	del, err := stakingKeeper.GetDelegation(ctx, addr, sdk.ValAddress(validator.GetOperator()))
	require.NoError(t, err)
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

	require.True(t, currVestingCoins.Equal(origCoins.QuoInt(math.NewInt(2))))
	require.True(t, currVestedCoins.Equal(origCoins.QuoInt(math.NewInt(2))))

	// execute migration script
	require.NoError(t, v15.ClawbackVestingFunds(ctx, addr, &gaiaApp.AppKeepers))

	// check that the validator's delegation is removed and that
	// their total tokens decreased
	validators, err = stakingKeeper.GetAllValidators(ctx)
	validator = validators[0]
	_, err = stakingKeeper.GetDelegation(ctx, addr, sdk.ValAddress(validator.GetOperator()))
	require.ErrorIs(t, err, stakingtypes.ErrNoDelegation)
	require.Equal(
		t,
		validator.TokensFromShares(validator.DelegatorShares),
		math.LegacyNewDec(oldValTokens.Int64()),
	)

	// verify that all modules can end/begin blocks
	gaiaApp.EndBlocker(ctx)
	gaiaApp.BeginBlocker(ctx)

	// check that the resulting account is of BaseAccount type now
	account, ok := accountKeeper.GetAccount(ctx, addr).(*authtypes.BaseAccount)
	require.True(t, ok)
	// check that the account values are still the same
	require.EqualValues(t, account, vestingAccount.BaseAccount)

	// check that the account's balance still has the vested tokens
	require.True(t, bankKeeper.GetAllBalances(ctx, addr).Equal(currVestedCoins))
	// check that the community pool balance received the vesting tokens
	communityPool, err := distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	require.True(
		t,
		communityPool.CommunityPool.
			Equal(sdk.NewDecCoinsFromCoins(currVestingCoins...)),
	)

	// verify that normal operations work in banking and staking
	_, err = stakingKeeper.Delegate(
		ctx, addr,
		math.NewInt(30),
		stakingtypes.Unbonded,
		validator,
		true)
	require.NoError(t, err)

	newAddr := sdk.AccAddress([]byte("cosmos1qqp9myctmh8mh2y7gynlsnw4y2wz3s3089dak6"))
	err = bankKeeper.SendCoins(
		ctx,
		addr,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewInt(10))),
	)
	require.NoError(t, err)
}

func TestSetMinInitialDepositRatio(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	err := v15.SetMinInitialDepositRatio(ctx, *gaiaApp.GovKeeper)
	require.NoError(t, err)

	params, err := gaiaApp.GovKeeper.Params.Get(ctx)
	require.NoError(t, err)
	minInitialDepositRatioStr := params.MinInitialDepositRatio
	minInitialDepositRatio, err := math.LegacyNewDecFromStr(minInitialDepositRatioStr)
	require.NoError(t, err)
	require.True(t, minInitialDepositRatio.Equal(math.LegacyNewDecWithPrec(1, 1)))
}

func TestUpgradeEscrowAccounts(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	bankKeeper := gaiaApp.BankKeeper
	transferKeeper := gaiaApp.TransferKeeper

	escrowUpdates := v15.GetEscrowUpdates(ctx)

	// check escrow accounts are empty
	for _, update := range escrowUpdates {
		require.Empty(t, bankKeeper.GetAllBalances(ctx, sdk.MustAccAddressFromBech32(update.Address)))
		for _, coin := range update.Coins {
			require.Equal(t, math.ZeroInt(), transferKeeper.GetTotalEscrowForDenom(ctx, coin.Denom).Amount)
		}
	}

	// execute the upgrade
	v15.UpgradeEscrowAccounts(ctx, bankKeeper, transferKeeper)

	// check that new assets are minted and transferred to the escrow accounts
	numUpdate := 0
	for _, update := range escrowUpdates {
		for _, coin := range update.Coins {
			require.Equal(t, coin, bankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(update.Address), coin.Denom))
			// check that the total escrow amount for the denom is updated
			require.Equal(t, coin, transferKeeper.GetTotalEscrowForDenom(ctx, coin.Denom))
			numUpdate++
		}
	}

	// verify that all tree discrepancies are covered in the update
	require.Equal(t, 3, numUpdate)
}
