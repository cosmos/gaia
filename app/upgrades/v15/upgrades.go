package v15

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	accountkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/gaia/v15/app/keepers"
)

// CreateUpgradeHandler returns a upgrade handler for Gaia v15
// which executes the following migrations:
//   - adhere to prop 826 which sets the minimum commission rate to 5% for all validators,
//     see https://www.mintscan.io/cosmos/proposals/826
//   - update the slashing module SigningInfos for which the consensus address is empty,
//     see https://github.com/cosmos/gaia/issues/1734.
//   - adhere to signal prop 860 which claws back vesting funds
//     see https://www.mintscan.io/cosmos/proposals/860
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

		if err := UpgradeMinCommissionRate(ctx, *keepers.StakingKeeper); err != nil {
			return nil, fmt.Errorf("failed migrating min commission rates: %s", err)
		}
		if err := UpgradeSigningInfos(ctx, keepers.SlashingKeeper); err != nil {
			return nil, fmt.Errorf("failed migrating signing infos: %s", err)
		}
		if err := ClawbackVestingFunds(
			ctx,
			sdk.MustAccAddressFromBech32("cosmos145hytrc49m0hn6fphp8d5h4xspwkawcuzmx498"),
			keepers); err != nil {
			return nil, fmt.Errorf("failed migrating vesting funds: %s", err)
		}
		if err := SetMinInitialDepositRatio(ctx, *keepers.GovKeeper); err != nil {
			return nil, fmt.Errorf("failed initializing the min initial deposit ratio: %s", err)
		}

		ctx.Logger().Info("Upgrade v15 complete")
		return vm, err
	}
}

// UpgradeMinCommissionRate sets the minimum commission rate staking parameter to 5%
// and updates the commission rate for all validators that have a commission rate less than 5%
// adhere to prop 826 which sets the minimum commission rate to 5% for all validators
// https://www.mintscan.io/cosmos/proposals/826
func UpgradeMinCommissionRate(ctx sdk.Context, sk stakingkeeper.Keeper) error {
	ctx.Logger().Info("Migrating min commission rate...")

	params := sk.GetParams(ctx)
	params.MinCommissionRate = sdk.NewDecWithPrec(5, 2)
	if err := sk.SetParams(ctx, params); err != nil {
		return err
	}

	for _, val := range sk.GetAllValidators(ctx) {
		if val.Commission.CommissionRates.Rate.LT(sdk.NewDecWithPrec(5, 2)) {
			// set the commission rate to 5%
			val.Commission.CommissionRates.Rate = sdk.NewDecWithPrec(5, 2)
			// set the max rate to 5% if it is less than 5%
			if val.Commission.CommissionRates.MaxRate.LT(sdk.NewDecWithPrec(5, 2)) {
				val.Commission.CommissionRates.MaxRate = sdk.NewDecWithPrec(5, 2)
			}
			val.Commission.UpdateTime = ctx.BlockHeader().Time
			sk.SetValidator(ctx, val)
		}
	}

	ctx.Logger().Info("Finished migrating min commission rate")
	return nil
}

// UpgradeSigningInfos updates the signing infos of validators for which
// the consensus address is missing
func UpgradeSigningInfos(ctx sdk.Context, sk slashingkeeper.Keeper) error {
	ctx.Logger().Info("Migrating signing infos...")

	signingInfos := []slashingtypes.ValidatorSigningInfo{}

	// update consensus address in signing info
	// using the store key of validators
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
			ctx.Logger().Error("incorrect consensus address in signing info %s: %s", si.Address, err)
			continue
		}
		sk.SetValidatorSigningInfo(ctx, addr, si)
	}

	ctx.Logger().Info("Finished migrating signing infos")
	return nil
}

// ClawbackVestingFunds transfers the vesting tokens from the given vesting account
// to the community pool
func ClawbackVestingFunds(ctx sdk.Context, address sdk.AccAddress, keepers *keepers.AppKeepers) error {
	ctx.Logger().Info("Migrating vesting funds...")

	ak := keepers.AccountKeeper
	bk := keepers.BankKeeper
	dk := keepers.DistrKeeper
	sk := *keepers.StakingKeeper

	// get target account
	account := ak.GetAccount(ctx, address)

	// verify that it's a vesting account type
	vestAccount, ok := account.(*vesting.ContinuousVestingAccount)
	if !ok {
		ctx.Logger().Error(
			"failed migrating vesting funds: %s: %s",
			"provided account address isn't a vesting account: ",
			address.String(),
		)

		return nil
	}

	// returns if the account has no vesting coins of the bond denom
	vestingCoinToClawback := sdk.Coin{}
	if vc := vestAccount.GetVestingCoins(ctx.BlockTime()); !vc.Empty() {
		_, vestingCoinToClawback = vc.Find(sk.BondDenom(ctx))
	}

	if vestingCoinToClawback.IsNil() {
		ctx.Logger().Info(
			"%s: %s",
			"no vesting coins to migrate",
			"Finished migrating vesting funds",
		)

		return nil
	}

	// unbond all delegations from vesting account
	if err := forceUnbondAllDelegations(sk, bk, ctx, address); err != nil {
		return err
	}

	// transfers still vesting tokens of BondDenom to community pool
	if err := forceFundCommunityPool(
		ak,
		dk,
		bk,
		ctx,
		vestingCoinToClawback,
		address,
		keepers.GetKey(banktypes.StoreKey),
	); err != nil {
		return err
	}

	// overwrite vesting account using its embedded base account
	ak.SetAccount(ctx, vestAccount.BaseAccount)

	// validate account balance
	if err := bk.ValidateBalance(ctx, address); err != nil {
		return err
	}

	ctx.Logger().Info("Finished migrating vesting funds")
	return nil
}

// forceUnbondAllDelegations unbonds all the delegations from the  given account address,
// without waiting for an unbonding period
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
			return stakingtypes.ErrNoValidatorFound
		}

		returnAmount, err := sk.Unbond(ctx, delegator, valAddr, del.GetShares())
		if err != nil {
			return err
		}

		coins := sdk.NewCoins(sdk.NewCoin(sk.BondDenom(ctx), returnAmount))

		// transfer the validator tokens to the not bonded pool
		if validator.IsBonded() {
			// doing stakingKeeper.bondedTokensToNotBonded
			err = bk.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, coins)
			if err != nil {
				return err
			}
		}

		err = bk.UndelegateCoinsFromModuleToAccount(ctx, stakingtypes.NotBondedPoolName, delegator, coins)
		if err != nil {
			return err
		}
	}

	return nil
}

// forceFundCommunityPool sends the given coin from the sender account to the community pool
// even if the coin is locked.
// Note that it partially follows the logic of the FundCommunityPool method in
// https://github.com/cosmos/cosmos-sdk/blob/release%2Fv0.47.x/x/distribution/keeper/keeper.go#L155
func forceFundCommunityPool(
	ak accountkeeper.AccountKeeper,
	dk distributionkeeper.Keeper,
	bk bankkeeper.Keeper,
	ctx sdk.Context,
	amount sdk.Coin,
	sender sdk.AccAddress,
	bs storetypes.StoreKey,
) error {
	recipientAcc := ak.GetModuleAccount(ctx, distributiontypes.ModuleName)
	if recipientAcc == nil {
		return fmt.Errorf("%s:%s", sdkerrors.ErrUnknownAddress, distributiontypes.ModuleName)
	}

	senderBal := bk.GetBalance(ctx, sender, amount.Denom)
	if _, hasNeg := sdk.NewCoins(senderBal).SafeSub(amount); hasNeg {
		return fmt.Errorf(
			"%s: spendable balance %s is smaller than %s",
			sdkerrors.ErrInsufficientFunds,
			senderBal,
			amount,
		)
	}
	if err := setBalance(ctx, sender, senderBal.Sub(amount), bs); err != nil {
		return err
	}
	recipientBal := bk.GetBalance(ctx, recipientAcc.GetAddress(), amount.Denom)
	if err := setBalance(ctx, recipientAcc.GetAddress(), recipientBal.Add(amount), bs); err != nil {
		return err
	}

	accExists := ak.HasAccount(ctx, recipientAcc.GetAddress())
	if !accExists {
		ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, recipientAcc.GetAddress()))
	}

	feePool := dk.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount)...)
	dk.SetFeePool(ctx, feePool)

	return nil
}

// setBalance sets the coin balance for an account by address.
// Note that it follows the same logic of the setBalance method in
// https://github.com/cosmos/cosmos-sdk/blob/v0.47.7/x/bank/keeper/send.go#L337
func setBalance(
	ctx sdk.Context,
	addr sdk.AccAddress,
	balance sdk.Coin,
	bs storetypes.StoreKey,
) error {
	if !balance.IsValid() {
		return fmt.Errorf("%s:%s", sdkerrors.ErrInvalidCoins, balance.String())
	}

	store := ctx.KVStore(bs)
	accountStore := prefix.NewStore(store, banktypes.CreateAccountBalancesPrefix(addr))
	denomPrefixStore := prefix.NewStore(store, banktypes.CreateDenomAddressPrefix(balance.Denom))

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

// SetMinInitialDepositRatio sets the MinInitialDepositRatio param of the gov
// module to 10% - this is the proportion of the deposit value that must be paid
// at proposal submission.
func SetMinInitialDepositRatio(ctx sdk.Context, gk govkeeper.Keeper) error {
	ctx.Logger().Info("Initializing MinInitialDepositRatio...")

	params := gk.GetParams(ctx)
	params.MinInitialDepositRatio = sdk.NewDecWithPrec(1, 1).String() // 0.1 (10%)
	err := gk.SetParams(ctx, params)
	if err != nil {
		return err
	}

	ctx.Logger().Info("Finished initializing MinInitialDepositRatio...")

	return nil
}
