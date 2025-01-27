package keeper

import (
	"fmt"
	"sort"

	"github.com/gogo/status"
	"github.com/strangelove-ventures/tokenfactory/x/tokenfactory/types"
	"google.golang.org/grpc/codes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) mintTo(ctx sdk.Context, amount sdk.Coin, mintTo string, isSudo bool) error {
	// verify that denom is an x/tokenfactory denom
	if !isSudo {
		_, _, err := types.DeconstructDenom(amount.Denom)
		if err != nil {
			return err
		}
	}

	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(mintTo)
	if err != nil {
		return err
	}

	if k.bankKeeper.BlockedAddr(addr) {
		return fmt.Errorf("failed to mint to blocked address: %s", addr)
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName,
		addr,
		sdk.NewCoins(amount))
}

func (k Keeper) burnFrom(ctx sdk.Context, amount sdk.Coin, burnFrom string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(burnFrom)
	if err != nil {
		return err
	}

	if k.bankKeeper.BlockedAddr(addr) {
		return fmt.Errorf("failed to burn from blocked address: %s", addr)
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		addr,
		types.ModuleName,
		sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	return k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
}

func (k Keeper) forceTransfer(ctx sdk.Context, amount sdk.Coin, fromAddr string, toAddr string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	fromSdkAddr, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return err
	}

	sortedPermAddrs := make([]string, 0, len(k.permAddrs))
	for moduleName := range k.permAddrs {
		sortedPermAddrs = append(sortedPermAddrs, moduleName)
	}
	sort.Strings(sortedPermAddrs)

	for _, moduleName := range sortedPermAddrs {
		account := k.accountKeeper.GetModuleAccount(ctx, moduleName)
		if account == nil {
			return status.Errorf(codes.NotFound, "account %s not found", moduleName)
		}

		if account.GetAddress().Equals(fromSdkAddr) {
			return status.Errorf(codes.Internal, "send from module acc not available")
		}
	}

	toSdkAddr, err := sdk.AccAddressFromBech32(toAddr)
	if err != nil {
		return err
	}

	if k.bankKeeper.BlockedAddr(toSdkAddr) {
		return fmt.Errorf("failed to force transfer to blocked address: %s", toSdkAddr)
	}

	return k.bankKeeper.SendCoins(ctx, fromSdkAddr, toSdkAddr, sdk.NewCoins(amount))
}
