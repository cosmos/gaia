package gaia

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// ibc decorator becomes the outermost decorator, will be checked at last.
func AddExtraDecorator(handlers sdk.AnteHandler, decorator sdk.AnteDecorator) sdk.AnteHandler {
    //todo some checks ?
    return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
        return decorator.AnteHandle(ctx, tx, simulate, handlers)
    }
}
