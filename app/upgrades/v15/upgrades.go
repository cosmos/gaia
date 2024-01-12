package v15

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/module"
	accountkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"

	"github.com/cosmos/gaia/v15/app/keepers"
)

// CreateUpgradeHandler returns a upgrade handler for Gaia v15
// which executes the following migrations:
// * set the MinCommissionRate param of the staking module to %5
// * update the slashing module SigningInfos records with empty consensus address
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		MigrateMinCommissionRate(ctx, *keepers.StakingKeeper)
		MigrateSigningInfos(ctx, keepers.SlashingKeeper)
		MigrateVestingAccount(ctx, sdk.AccAddress{}, keepers)

		ctx.Logger().Info("Upgrade v15 complete")
		return vm, err
	}
}

// MigrateMinCommissionRate adheres to prop 826 https://www.mintscan.io/cosmos/proposals/826
// by setting the minimum commission rate staking parameter to 5%
// and updating the commission rate for all validators that have a commission rate less than 5%
func MigrateMinCommissionRate(ctx sdk.Context, sk stakingkeeper.Keeper) {
	params := sk.GetParams(ctx)
	params.MinCommissionRate = sdk.NewDecWithPrec(5, 2)
	if err := sk.SetParams(ctx, params); err != nil {
		panic(err)
	}

	for _, val := range sk.GetAllValidators(ctx) {
		val := val
		if val.Commission.CommissionRates.Rate.LT(sdk.NewDecWithPrec(5, 2)) {
			// set the commmision rate to 5%
			val.Commission.CommissionRates.Rate = sdk.NewDecWithPrec(5, 2)
			// set the max rate to 5% if it is less than 5%
			if val.Commission.CommissionRates.MaxRate.LT(sdk.NewDecWithPrec(5, 2)) {
				val.Commission.CommissionRates.MaxRate = sdk.NewDecWithPrec(5, 2)
			}
			val.Commission.UpdateTime = ctx.BlockHeader().Time
			sk.SetValidator(ctx, val)
		}
	}
}

// MigrateSigningInfos updates validators signing infos for which the consensus address
// is missing using their store key, which contains the consensus address of the validator,
// see https://github.com/cosmos/gaia/issues/1734.
func MigrateSigningInfos(ctx sdk.Context, sk slashingkeeper.Keeper) {
	signingInfos := []slashingtypes.ValidatorSigningInfo{}

	sk.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		if info.Address == "" {
			info.Address = address.String()
			signingInfos = append(signingInfos, info)
		}

		return false
	})

	for _, si := range signingInfos {
		addr, err := sdk.ConsAddressFromBech32(si.Address)
		if err != nil {
			panic(err)
		}
		sk.SetValidatorSigningInfo(ctx, addr, si)
	}
}

func MigrateVestingAccount(ctx sdk.Context, address sdk.AccAddress, keepers *keepers.AppKeepers) {
	ak := keepers.AccountKeeper
	bk := keepers.BankKeeper
	dk := keepers.DistrKeeper
	sk := *keepers.StakingKeeper

	// Unbond all delegations from vesting account
	err := forceUnbondAllDelegations(sk, bk, ctx, address)
	if err != nil {
		panic(err)
	}

	vacc, ok := ak.GetAccount(ctx, address).(*vesting.ContinuousVestingAccount)
	if !ok {
		panic(fmt.Errorf("incorrect continuous vesting account"))
	}

	// transfers still vesting tokens of BondDenom to community pool
	_, vestingCoins := vacc.GetVestingCoins(ctx.BlockTime()).Find(sk.BondDenom(ctx))

	// Unbond all delegations from vesting account
	err = forceFundCommunityPool(
		ak,
		dk,
		bk,
		ctx,
		vestingCoins,
		address,
		keepers.GetKey(banktypes.StoreKey),
	)
	if err != nil {
		panic(err)
	}

	// update continuous vesting account in order
	// to have all tokens vested
	_, vestedCoins := vacc.GetVestedCoins(ctx.BlockTime()).Find(sk.BondDenom(ctx))
	err = updateContinuousVestingAccount(
		ak,
		ctx,
		address,
		sdk.NewCoins(vestedCoins),
	)
	if err != nil {
		panic(err)
	}

	// validate vesting account
	err = bk.ValidateBalance(ctx, address)
	if err != nil {
		panic(err)
	}
}

// Unbond vesting account delegations
func forceUnbondAllDelegations(
	sk stakingkeeper.Keeper,
	bk bankkeeper.Keeper,
	ctx sdk.Context,
	delegator sdk.AccAddress,
) error {

	dels := sk.GetDelegatorDelegations(ctx, delegator, 100)

	for _, del := range dels {
		valAddr := del.GetValidatorAddr()

		validator, found := sk.GetValidator(ctx, valAddr)
		if !found {
			return fmt.Errorf("unknown validator")
		}

		returnAmount, err := sk.Unbond(ctx, delegator, valAddr, del.GetShares())
		if err != nil {
			return err
		}

		// transfer the validator tokens to the not bonded pool
		// doing stakingKeeper.bondedTokensToNotBonded
		if validator.IsBonded() {
			coins := sdk.NewCoins(sdk.NewCoin(sk.BondDenom(ctx), returnAmount))
			err = bk.SendCoinsFromModuleToModule(ctx, types.BondedPoolName, types.NotBondedPoolName, coins)
			if err != nil {
				return err
			}
		}

		bondDenom := sk.GetParams(ctx).BondDenom
		amt := sdk.NewCoin(bondDenom, returnAmount)

		// call TrackDelegation() to update vesting account delegations
		err = bk.UndelegateCoinsFromModuleToAccount(ctx, types.NotBondedPoolName, delegator, sdk.NewCoins(amt))
		if err != nil {
			return err
		}
	}

	return nil
}

// CONTRACT: coin are in the bond denom
// Community pool / distribution module account already exists
func forceFundCommunityPool(
	ak accountkeeper.AccountKeeper,
	dk distributionkeeper.Keeper,
	bk bankkeeper.Keeper,
	ctx sdk.Context,
	amount sdk.Coin,
	sender sdk.AccAddress,
	bs storetypes.StoreKey,
) error {
	// SendCoinsFromAccountToModule{
	recipientAcc := ak.GetModuleAccount(ctx, distributiontypes.ModuleName)
	if recipientAcc == nil {
		panic(fmt.Errorf("module account %s does not exist", distributiontypes.ModuleName))
	}
	// SendCoins{
	//k.subUnlockedCoins{
	senderBal := bk.GetBalance(ctx, sender, amount.Denom)
	if _, hasNeg := sdk.NewCoins(senderBal).SafeSub(amount); hasNeg {
		return fmt.Errorf("spendable balance %s is smaller than %s",
			senderBal, amount,
		)
	}
	if err := setBalance(bk, ctx, sender, senderBal.Sub(amount), bs); err != nil {
		return err
	}
	//}k.subUnlockedCoins
	recipientBal := bk.GetBalance(ctx, recipientAcc.GetAddress(), amount.Denom)
	if err := setBalance(bk, ctx, recipientAcc.GetAddress(), recipientBal.Add(amount), bs); err != nil {
		return err
	}

	// Create account if recipient does not exist.
	//
	// NOTE: This should ultimately be removed in favor a more flexible approach
	// such as delegated fee messages.
	accExists := ak.HasAccount(ctx, recipientAcc.GetAddress())
	if !accExists {
		ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, recipientAcc.GetAddress()))
	}
	// }SendCoins
	//} SendCoinsFromAccountToModule

	feePool := dk.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount)...)
	dk.SetFeePool(ctx, feePool)

	return nil
}

// Update vesting account
func updateContinuousVestingAccount(
	ak accountkeeper.AccountKeeper,
	ctx sdk.Context,
	addr sdk.AccAddress,
	OrigCoins sdk.Coins,
) error {
	acc := ak.GetAccount(ctx, addr)
	vacc, ok := acc.(*vesting.ContinuousVestingAccount)
	if !ok {
		return fmt.Errorf("inccorect a continuous vesting account")
	}
	bacc := &authtypes.BaseAccount{
		Address:       acc.GetAddress().String(),
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
	}

	err := bacc.SetPubKey(acc.GetPubKey())
	if err != nil {
		return err
	}

	newVacc := vesting.NewContinuousVestingAccount(bacc, OrigCoins, vacc.StartTime, ctx.BlockTime().Unix())
	ak.SetAccount(ctx, newVacc)

	return nil
}

// setBalance sets the coin balance for an account by address.
func setBalance(
	bk bankkeeper.Keeper,
	ctx sdk.Context,
	addr sdk.AccAddress,
	balance sdk.Coin,
	bs storetypes.StoreKey,
) error {
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

func getContinuousVestingAccount(ak accountkeeper.AccountKeeper, ctx sdk.Context, addr sdk.AccAddress) *vesting.ContinuousVestingAccount {
	acc := ak.GetAccount(ctx, addr)
	vacc, ok := acc.(*vesting.ContinuousVestingAccount)
	if !ok {
		panic(fmt.Errorf("incorrect continuous vesting account"))
	}
	return vacc
}
