package v15_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	accountkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/gaia/v15/app/helpers"
	v15 "github.com/cosmos/gaia/v15/app/upgrades/v15"
)

func TestMigrateMinCommissionRate(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	// set min commission rate to 0
	stakingParams := gaiaApp.StakingKeeper.GetParams(ctx)
	stakingParams.MinCommissionRate = sdk.ZeroDec()
	err := gaiaApp.StakingKeeper.SetParams(ctx, stakingParams)
	require.NoError(t, err)

	// confirm all commissions are 0
	stakingKeeper := gaiaApp.StakingKeeper

	for _, val := range stakingKeeper.GetAllValidators(ctx) {
		require.Equal(t, val.Commission.CommissionRates.Rate, sdk.ZeroDec(), "non-zero previous commission rate for validator %s", val.GetOperator())
	}

	// pre-test min commission rate is 0
	require.Equal(t, stakingKeeper.GetParams(ctx).MinCommissionRate, sdk.ZeroDec(), "non-zero previous min commission rate")

	// run the test and confirm the values have been updated
	v15.MigrateMinCommissionRate(ctx, *gaiaApp.AppKeepers.StakingKeeper)

	newStakingParams := gaiaApp.StakingKeeper.GetParams(ctx)
	require.NotEqual(t, newStakingParams.MinCommissionRate, sdk.ZeroDec(), "failed to update min commission rate")
	require.Equal(t, newStakingParams.MinCommissionRate, sdk.NewDecWithPrec(5, 2), "failed to update min commission rate")

	for _, val := range stakingKeeper.GetAllValidators(ctx) {
		require.Equal(t, val.Commission.CommissionRates.Rate, newStakingParams.MinCommissionRate, "failed to update update commission rate for validator %s", val.GetOperator())
	}

	// set one of the validators commission rate to 10% and ensure it is not updated
	updateValCommission := sdk.NewDecWithPrec(10, 2)
	updateVal := stakingKeeper.GetAllValidators(ctx)[0]
	updateVal.Commission.CommissionRates.Rate = updateValCommission
	stakingKeeper.SetValidator(ctx, updateVal)

	v15.MigrateMinCommissionRate(ctx, *gaiaApp.AppKeepers.StakingKeeper)
	for _, val := range stakingKeeper.GetAllValidators(ctx) {
		if updateVal.OperatorAddress == val.OperatorAddress {
			require.Equal(t, val.Commission.CommissionRates.Rate, updateValCommission, "should not update commission rate for validator %s", val.GetOperator())
		} else {
			require.Equal(t, val.Commission.CommissionRates.Rate, newStakingParams.MinCommissionRate, "failed to update update commission rate for validator %s", val.GetOperator())
		}
	}
}

func TestMigrateValidatorsSigningInfos(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	slashingKeeper := gaiaApp.SlashingKeeper

	signingInfosNum := 8
	emptyAddrCtr := 0

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

		if i <= signingInfosNum/2 {
			info.Address = ""
			emptyAddrCtr++
		}

		slashingKeeper.SetValidatorSigningInfo(ctx, consAddr, info)
		require.NoError(t, err)
	}

	// check signing info were correctly created
	slashingKeeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		if info.Address == "" {
			emptyAddrCtr--
		}

		return false
	})
	require.Zero(t, emptyAddrCtr)

	// upgrade signing infos
	v15.MigrateSigningInfos(ctx, slashingKeeper)

	// check that all signing info have the address field correctly updated
	slashingKeeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		require.NotEmpty(t, info.Address)
		require.Equal(t, address.String(), info.Address)

		return false
	})
}

// undelegations successful - validators have less voting power
// vesting account updated - still vesting tokens are gone
// community pool updated - still vesting tokens are added
// the vesting account still has the already vested tokens
func TestMigrateUnvestedFunds(t *testing.T) {
	gaiaApp := helpers.Setup(t)

	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})

	bankKeeper := gaiaApp.BankKeeper
	accountKeeper := gaiaApp.AccountKeeper
	distrKeeper := gaiaApp.DistrKeeper
	stakingKeeper := gaiaApp.StakingKeeper

	validators := stakingKeeper.GetAllValidators(ctx)
	validator := validators[0]

	// original vesting coin amount
	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))

	// create a continuous vesting account
	addr := sdk.AccAddress([]byte("addr1_______________"))
	require.Error(t, bankKeeper.ValidateBalance(ctx, addr))

	bacc := authtypes.NewBaseAccountWithAddress(addr)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())

	accountKeeper.SetAccount(ctx, vacc)
	vestingCoins := vacc.GetVestingCoins(ctx.BlockTime())

	require.True(t, vestingCoins.IsEqual(origCoins))
	require.Empty(t, bankKeeper.GetAllBalances(ctx, addr))

	// fund vesting account
	require.NoError(t, banktestutil.FundAccount(bankKeeper, ctx, addr, origCoins))
	require.True(t, origCoins.IsEqual(bankKeeper.GetAllBalances(ctx, addr)))
	// err := distrKeeper.FundCommunityPool(ctx, balances, vacc.GetAddress())
	// require.Error(t, err)

	initBal := bankKeeper.GetAllBalances(ctx, vacc.GetAddress())
	require.True(t, initBal.IsEqual(origCoins))

	oldVP := validator.GetConsensusPower(validator.BondedTokens())
	fmt.Println("oldVp:", oldVP)
	oldValTokens := validator.Tokens

	fmt.Println("last power", stakingKeeper.GetLastTotalPower(ctx))
	fmt.Println("0", validator.String())

	// delegate vested and still vesting tokens
	_, err := stakingKeeper.Delegate(ctx, vacc.GetAddress(), origCoins.AmountOf("stake"), stakingtypes.Unbonded, validator, true)
	require.NoError(t, err)

	validator = stakingKeeper.GetAllValidators(ctx)[0]

	require.True(t, validator.Tokens.Equal(oldValTokens.Add(origCoins.AmountOf("stake"))))

	consAddr, err := validator.GetConsAddr()
	require.NoError(t, err)

	validatorUpdated, found := stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	fmt.Println("1", validatorUpdated.String())

	validator = stakingKeeper.GetAllValidators(ctx)[0]

	fmt.Println("2", validator.String())

	fmt.Println("last power", stakingKeeper.GetLastTotalPower(ctx))

	del, found := stakingKeeper.GetDelegation(ctx, addr, validator.GetOperator())
	require.True(t, found)
	require.Equal(t, validator.TokensFromShares(del.Shares), math.LegacyNewDec(origCoins.AmountOf("stake").Int64()))

	acc := accountKeeper.GetAccount(ctx, addr)
	vaccupdated, ok := acc.(banktypes.VestingAccount)
	require.True(t, ok)
	require.Equal(t, vaccupdated.GetDelegatedVesting(), origCoins)
	require.Empty(t, vaccupdated.GetDelegatedFree())

	// vest half of the tokens
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	vestingCoinsUpdated := vacc.GetVestingCoins(ctx.BlockTime())
	vestedCoins := vacc.GetVestedCoins(ctx.BlockTime())
	require.True(t, vestingCoinsUpdated.IsEqual(origCoins.QuoInt(math.NewInt(2))))
	require.True(t, vestedCoins.IsEqual(origCoins.QuoInt(math.NewInt(2))))

	// remove delegations

	returnAmount, err := stakingKeeper.Unbond(ctx, addr, validator.GetOperator(), del.Shares)
	require.NoError(t, err)

	nb := bankKeeper.GetAllBalances(ctx, accountKeeper.GetModuleAddress(types.BondedPoolName))
	fmt.Println(" bonded balances", nb.String())

	// transfer the validator tokens to the not bonded pool
	// doing stakingKeeper.bondedTokensToNotBonded
	if validator.IsBonded() {
		fmt.Println("bonded")
		fmt.Println("returnAmount", returnAmount.String())
		coins := sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), returnAmount))
		err = bankKeeper.SendCoinsFromModuleToModule(ctx, types.BondedPoolName, types.NotBondedPoolName, coins)
		require.NoError(t, err)
	}

	nb = bankKeeper.GetAllBalances(ctx, accountKeeper.GetModuleAddress(types.BondedPoolName))
	fmt.Println(" bonded balances", nb.String())

	nbp := bankKeeper.GetAllBalances(ctx, accountKeeper.GetModuleAddress(types.NotBondedPoolName))
	fmt.Println("not bonded balances", nbp.String())

	bondDenom := "stake" // k.GetParams(ctx).BondDenom
	amt := sdk.NewCoin(bondDenom, returnAmount)
	err = bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, types.NotBondedPoolName, addr, sdk.NewCoins(amt))
	require.NoError(t, err)

	require.True(t, origCoins.IsEqual(bankKeeper.GetAllBalances(ctx, addr)))

	acc = accountKeeper.GetAccount(ctx, addr)
	vaccupdated, ok = acc.(banktypes.VestingAccount)
	require.True(t, ok)
	require.Empty(t, vaccupdated.GetDelegatedVesting())
	require.Empty(t, vaccupdated.GetDelegatedFree())

	require.Empty(t, distrKeeper.GetFeePoolCommunityCoins(ctx))

	kvs := gaiaApp.GetKVStoreKey()

	vestingAmt := sdk.NewCoin(amt.Denom, amt.Amount.Quo(math.NewInt(2)))

	FundCommunityPool(
		accountKeeper,
		distrKeeper,
		bankKeeper,
		ctx,
		sdk.NewCoins(vestingAmt),
		addr,
		kvs[banktypes.StoreKey],
	)

	// err = bankKeeper.UndelegateCoinsFromModuleToAccount(
	// 	ctx, types.NotBondedPoolName, addr, sdk.NewCoins(amt))
	// require.NoError(t, err)

	// remove

	//// send to community pool
	// create fee pool
	// distrKeeper.SetFeePool(ctx, types.InitialFeePool())

	require.NotEmpty(t, distrKeeper.GetFeePoolCommunityCoins(ctx))
	require.True(t, sdk.NewDecCoinsFromCoins(vestingAmt).IsEqual(distrKeeper.GetFeePoolCommunityCoins(ctx)))

	fmt.Println("community pool balances", distrKeeper.GetFeePoolCommunityCoins(ctx))
	fmt.Println("community pool balances 2", bankKeeper.GetAllBalances(ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName)))

	fmt.Println("vesting account", bankKeeper.GetAllBalances(ctx, addr))

	// require.NoError(t, bankKeeper.ValidateBalance(ctx, addr))
	// require.NoError(t, bankKeeper.ValidateBalance(ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName)))

	vacc3 := accountKeeper.GetAccount(ctx, addr)
	fmt.Println("vesting account 6", vacc3.(*vesting.ContinuousVestingAccount).GetVestingCoins(ctx.BlockTime()))
	fmt.Println("vesting account 7 ", vacc3.(*vesting.ContinuousVestingAccount).GetVestedCoins(ctx.BlockTime()))

	vacc4 := vesting.NewContinuousVestingAccount(bacc, sdk.NewCoins(vestingAmt), vacc.StartTime, ctx.BlockTime().Unix())

	accountKeeper.SetAccount(ctx, vacc4)

	fmt.Println("vesting account 8", vacc4.GetVestingCoins(ctx.BlockTime()))
	fmt.Println("vesting account 9 ", vacc4.GetVestedCoins(ctx.BlockTime()))

	require.Equal(t, vacc3.GetAddress().String(), vacc4.Address)

	require.NoError(t, bankKeeper.ValidateBalance(ctx, addr))
	require.NoError(t, bankKeeper.ValidateBalance(ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName)))

}

// FundCommunityPool allows an account to directly fund the community fund pool.
// The amount is first added to the distribution module account and then directly
// added to the pool. An error is returned if the amount cannot be sent to the
// module account.
func FundCommunityPool(
	ak accountkeeper.AccountKeeper,
	dk distributionkeeper.Keeper,
	bk bankkeeper.Keeper,
	ctx sdk.Context,
	amount sdk.Coins,
	sender sdk.AccAddress,
	bs storetypes.StoreKey) error {

	// SendCoinsFromAccountToModule{
	recipientAcc := ak.GetModuleAccount(ctx, distributiontypes.ModuleName)
	if recipientAcc == nil {
		panic(fmt.Errorf("module account %s does not exist", distributiontypes.ModuleName))
	}
	// SendCoins{
	//k.subUnlockedCoins{
	err := SubUnlockedCoins(bk, ctx, sender, amount, bs)
	if err != nil {
		panic(err)
	}

	//}k.subUnlockedCoins

	err = AddCoins(bk, ctx, recipientAcc.GetAddress(), amount, bs)
	if err != nil {
		panic(err)
	}

	// Create account if recipient does not exist.
	//
	// NOTE: This should ultimately be removed in favor a more flexible approach
	// such as delegated fee messages.
	accExists := ak.HasAccount(ctx, recipientAcc.GetAddress())
	if !accExists {
		fmt.Println("feepool doesn't exist")
		ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, recipientAcc.GetAddress()))
	}
	// }SendCoins

	//} SendCoinsFromAccountToModule

	feePool := dk.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...)
	dk.SetFeePool(ctx, feePool)

	return nil
}

func AddCoins(bk bankkeeper.Keeper, ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins, bs storetypes.StoreKey) error {
	if !amt.IsValid() {
		return fmt.Errorf("invalid coins") //sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	for _, coin := range amt {
		balance := bk.GetBalance(ctx, addr, coin.Denom)
		newBalance := balance.Add(coin)

		err := SetBalance(bk, ctx, addr, newBalance, bs)
		if err != nil {
			return err
		}
	}
	return nil
}

// setBalance sets the coin balance for an account by address.
func SetBalance(bk bankkeeper.Keeper, ctx sdk.Context, addr sdk.AccAddress, balance sdk.Coin, bs storetypes.StoreKey) error {
	if !balance.IsValid() {
		return fmt.Errorf(balance.String())
	}

	// 	accountStore := k.getAccountStore(ctx, addr)
	store := ctx.KVStore(bs)
	accountStore := prefix.NewStore(store, banktypes.CreateAccountBalancesPrefix(addr))

	// 	denomPrefixStore := k.getDenomAddressPrefixStore(ctx, balance.Denom)
	denomPrefixStore := prefix.NewStore(store, banktypes.CreateDenomAddressPrefix(balance.Denom))

	// x/bank invariants prohibit persistence of zero balances
	if balance.IsZero() {
		accountStore.Delete([]byte(balance.Denom))
		denomPrefixStore.Delete(address.MustLengthPrefix(addr))
	} else {
		amount, err := balance.Amount.Marshal()
		if err != nil {
			return err
		}

		accountStore.Set([]byte(balance.Denom), amount)

		// Store a reverse index from denomination to account address with a
		// sentinel value.
		denomAddrKey := address.MustLengthPrefix(addr)
		if !denomPrefixStore.Has(denomAddrKey) {
			denomPrefixStore.Set(denomAddrKey, []byte{0})
		}
	}

	return nil
}

// subUnlockedCoins removes the unlocked amt coins of the given account. An error is
// returned if the resulting balance is negative or the initial amount is invalid.
// A coin_spent event is emitted after.
func SubUnlockedCoins(bk bankkeeper.Keeper, ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins, bs storetypes.StoreKey) error {
	if !amt.IsValid() {
		return fmt.Errorf("amount isn't valid")
	}

	lockedCoins := sdk.Coins{}

	for _, coin := range amt {
		balance := bk.GetBalance(ctx, addr, coin.Denom)
		locked := sdk.NewCoin(coin.Denom, lockedCoins.AmountOf(coin.Denom))

		spendable, hasNeg := sdk.Coins{balance}.SafeSub(locked)
		if hasNeg {
			return fmt.Errorf(
				"locked amount exceeds account balance funds: %s > %s", locked, balance)
		}

		if _, hasNeg := spendable.SafeSub(coin); hasNeg {
			return fmt.Errorf("spendable balance %s is smaller than %s",
				spendable, coin,
			)
		}

		newBalance := balance.Sub(coin)

		if err := SetBalance(bk, ctx, addr, newBalance, bs); err != nil {
			return err
		}
	}

	return nil
}
