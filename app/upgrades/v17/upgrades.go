package v17

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/gaia/v17/app/keepers"
)

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

		if err := MigrateRedelegations(ctx, *keepers.StakingKeeper); err != nil {
			//TODO: decide what to do with error
			ctx.Logger().Error("fail to migrate redelegations - tokenization of shares might me compromised after upgrade ")
		}

		ctx.Logger().Info("Upgrade v17 complete")
		return vm, nil
	}
}

func MigrateRedelegations(ctx sdk.Context, sk stakingkeeper.Keeper) error {
	delegatorAddrs := []string{}
	delegatorsToValReds := map[string]map[string][]types.Redelegation{}

	// require to ensure that redelegations are always iterated in the same order
	delegatorsToValAddr := map[string][]string{}

	// create special id for delegator validator pair
	// each new redelegations either find or create one new id
	// add new struct in list or update one

	// iterate over all redelegations, if the destination validator has liquid shares,
	// store the redelegation indexed by the delegator address
	sk.IterateRedelegations(ctx, func(index int64, red stakingtypes.Redelegation) (stop bool) {
		valDstAddr, err := sdk.ValAddressFromBech32(red.ValidatorDstAddress)
		if err != nil {
			ctx.Logger().Error("failed migration xxx")
		}
		val := sk.Validator(ctx, valDstAddr).(stakingtypes.Validator)

		// TODO: check if that's correct since it might just be the case now
		if val.LiquidShares.IsZero() {
			return false
		}

		// if new delegator address creates one
		delegatorReds, ok := delegatorsToValReds[red.DelegatorAddress]
		if !ok { // avoid duplicates
			// append first redelgation for the current delegator
			delegatorAddrs = append(delegatorAddrs, red.DelegatorAddress)
			delegatorsToValReds[red.DelegatorAddress] = map[string][]types.Redelegation{valDstAddr.String(): {red}}

			// add first validator destionation address for the current delegator
			delegatorsToValAddr[red.DelegatorAddress] = append(delegatorsToValAddr[red.DelegatorAddress], valDstAddr.String())
		} else {
			// check if some redelegation for the destination validator already exists
			if _, ok := delegatorReds[valDstAddr.String()]; !ok {
				delegatorsToValAddr[red.DelegatorAddress] = append(delegatorsToValAddr[red.DelegatorAddress], valDstAddr.String())
			}

			delegatorReds[valDstAddr.String()] = append(delegatorReds[valDstAddr.String()], red)
		}

		return false
	})

	// fix redelegations

	// iterate over the delegators address
	for _, delAddr := range delegatorAddrs {
		for _, valAddr := range delegatorsToValAddr[delAddr] {
			reds := delegatorsToValReds[delAddr][valAddr]
			// iterate over the delegators' redelegations
			// get the delegator's shares
			valAddr, err := sdk.ValAddressFromBech32(valAddr)
			if err != nil {
				ctx.Logger().Error("failed migration xxx")
			}
			del, delOk := sk.GetDelegation(ctx, sdk.MustAccAddressFromBech32(delAddr), valAddr)

			// get the delegator's unbonding shares
			ubd, ubdOk := sk.GetUnbondingDelegation(ctx, sdk.MustAccAddressFromBech32(delAddr), valAddr)

			redsSharesToDelete := sdk.ZeroDec()

			// Compute redelegations shares to delete
			switch {
			// case 1: del shares = 0, ubd shares == 0
			// 	=> delete all redelegations
			case !delOk && !ubdOk:
				for _, red := range reds {
					// note that the provider won't panic
					// when it'd receive a VSC mature packet for it
					sk.RemoveRedelegation(ctx, red)
					continue
				}
			// case 2: del shares > 0, ubd shares == 0
			case delOk && !ubdOk:
				redsSharesToDelete = SumRedelegationsShares(reds).Sub(del.Shares)
			// case 4: del shares == 0, ubd shares > 0
			case !delOk && ubdOk:
				redsSharesToDelete, err = ComputeRemainingRedelegatedSharesAfterUnbondings(
					sk,
					ctx,
					delAddr,
					valAddr,
					ubd,
					reds,
				)
				if err != nil {
					continue
				}
			// case 3:  del shares > 0, ubd shares > 0 and
			default:
				redsSharesRemaining, err := ComputeRemainingRedelegatedSharesAfterUnbondings(
					sk,
					ctx,
					delAddr,
					valAddr,
					ubd,
					reds,
				)
				if err != nil {
					continue
				}
				redsSharesToDelete = redsSharesRemaining.Sub(del.Shares)
			}

			// if redelegations shares is positive, it means that some redelegations
			// aren't linked to any delegation shares and must be deleted
			if redsSharesToDelete.IsPositive() {
				err = RemoveRemainingRedelegationsByAmount(
					sk,
					ctx,
					redsSharesToDelete,
					reds,
				)
				if err != nil {
					continue
				}
			}
		}
	}
	return nil
}

// ComputeRemainingRedelegatedSharesAfterUnbondings calculates the remaining redelegated shares for a given delegator address
// and validator address, considering a list of redelegations and an unbonding delegation.
// The redelegations are from the given delegator to the given validator, and the unbonding delegation
// is for the given validator and belongs to the given delegator.
// The function returns the total remaining delegation shares after computing over time
// the deposited redelegation shares and the shares withdrawn by the given unbonding delegation entries,
// based on their respective completion times.
func ComputeRemainingRedelegatedSharesAfterUnbondings(
	sk stakingkeeper.Keeper,
	ctx sdk.Context,
	delAddr string,
	valAddr sdk.ValAddress,
	ubd stakingtypes.UnbondingDelegation,
	reds []types.Redelegation,
) (sdk.Dec, error) {
	// delegationEntry defines an general entry representing either
	// the addition or withdrawal of delegation shares at completion time.
	type delegationEntry struct {
		creationTime time.Time
		shares       sdk.Dec
	}

	delegationEntries := []delegationEntry{}
	validator, found := sk.GetValidator(ctx, valAddr)
	if !found {
		return sdk.ZeroDec(), types.ErrNoValidatorFound
	}

	for _, red := range reds {
		// check that the redelegation has the given validator destination address
		if valAddr.String() != red.ValidatorDstAddress {
			return sdk.ZeroDec(), types.ErrBadRedelegationDst
		}
		// check that the redelegation has the given delegator address
		if delAddr != red.DelegatorAddress {
			return sdk.ZeroDec(), types.ErrBadDelegatorAddr
		}
		// store each redelegation entry as a delegation entry
		// adding shares at completion time
		for _, redEntry := range red.Entries {
			delegationEntries = append(delegationEntries, delegationEntry{
				redEntry.CompletionTime,
				redEntry.SharesDst,
			})
		}
	}

	for _, ubdEntry := range ubd.Entries {
		ubdEntryShares, err := validator.SharesFromTokens(ubdEntry.InitialBalance)
		if err != nil {
			return sdk.ZeroDec(), err
		}
		// store each unbonding delegation entry as a delegation entry
		// withdrawing shares at completion time, by using it's negative amount of shares
		delegationEntries = append(delegationEntries,
			delegationEntry{
				ubdEntry.CompletionTime,
				ubdEntryShares.Neg(),
			})
	}

	// sort delegation entries by completion time in ascending order
	sort.Slice(delegationEntries, func(i, j int) bool {
		return delegationEntries[i].creationTime.Before(delegationEntries[j].creationTime)
	})

	// Sum the shares of delegation entries, flooring negative values to zero.
	// This assumes that negative shares must have been taken from the initial delegation shares initially,
	// otherwise the withdrawing operation should have failed.
	remainingShares := sdk.ZeroDec()
	for _, entry := range delegationEntries {
		if remainingShares.Add(entry.shares).IsNegative() {
			remainingShares = sdk.ZeroDec()
			continue
		}

		remainingShares = remainingShares.Add(entry.shares)
	}

	return remainingShares, nil
}

func SumRedelegationsShares(reds []stakingtypes.Redelegation) sdk.Dec {
	redsShares := sdk.ZeroDec()
	for _, red := range reds {
		for _, entry := range red.Entries {
			redsShares = redsShares.Add(entry.SharesDst)
		}
	}
	return redsShares
}

// TODO: document and add UT
func RemoveRemainingRedelegationsByAmount(
	sk stakingkeeper.Keeper,
	ctx sdk.Context,
	amount sdk.Dec,
	reds []stakingtypes.Redelegation,
) error {
	type redEntry struct {
		redIdx   int
		entryIdx int
		entry    stakingtypes.RedelegationEntry
	}

	redEntries := []redEntry{}
	for redIdx, red := range reds {
		for entryIdx, entry := range red.Entries {
			redEntries = append(redEntries, redEntry{redIdx: redIdx, entryIdx: entryIdx, entry: entry})
		}
	}

	// sort delegation entries by completion time in descending order
	sort.Slice(redEntries, func(i, j int) bool {
		return redEntries[i].entry.CompletionTime.After(redEntries[j].entry.CompletionTime)
	})

	sharesDeleted := sdk.ZeroDec()
	lastRedIdx := len(reds) - 1
	for _, re := range redEntries {
		sharesDeleted = sharesDeleted.Add(re.entry.SharesDst)
		if sharesDeleted.GT(amount) {
			// update entry shares to shares deleted - amount
			reds[re.redIdx].Entries[re.entryIdx].SharesDst = sharesDeleted.Sub(amount)
			lastRedIdx = re.redIdx
			break
		}
		// remove entry shares to zero
		// note that since they are ordered its necessary the last entry
		// and therefore doesn't break the entries index
		reds[re.redIdx].RemoveEntry(int64(re.entryIdx))
	}

	// TODO: check if it can be optimized without too much complexity
	for idx, red := range reds {
		if len(red.Entries) == 0 {
			sk.RemoveRedelegation(ctx, red)
		} else {
			sk.SetRedelegation(ctx, red)
		}

		if idx == lastRedIdx {
			break
		}
	}

	return nil
}
