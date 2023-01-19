package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AllInvariants collects any defined invariants below
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return ExampleInvariant(k)(ctx)

		/*
			Example additional invariants:
			res, stop := FutureInvariant(k)(ctx)
			if stop {
				return res, stop
			}
			return AnotherFutureInvariant(k)(ctx)
		*/
	}
}

// ExampleInvariant checks for incorrect things
func ExampleInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// TODO: Check for bad things
		return "", false
	}
}
