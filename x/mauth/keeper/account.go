package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInterchainAccount invokes the InitInterchainAccount entrypoint.
// InitInterchainAccount binds a new controller port and initiates a new ICS-27 channel handshake
func (k Keeper) RegisterInterchainAccount(ctx sdk.Context, owner sdk.AccAddress, connectionID, counterpartyConnectionID string) error {
	return k.icaControllerKeeper.InitInterchainAccount(ctx, connectionID, counterpartyConnectionID, owner.String())
}
