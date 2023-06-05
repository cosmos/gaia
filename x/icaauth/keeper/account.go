package keeper

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"

	"github.com/althea-net/ibc-test-chain/v9/x/icaauth/types"
)

// RegisterInterchainAccount invokes the InitInterchainAccount entrypoint.
// InitInterchainAccount binds a new controller port and initiates a new ICS-27 channel handshake
func (k Keeper) RegisterInterchainAccount(ctx sdk.Context, owner sdk.AccAddress, connectionID string, version string) error {
	ctx.EventManager().EmitTypedEvent(
		&types.EventRegisterInterchainAccount{
			Owner:        owner.String(),
			ConnectionId: connectionID,
			Version:      version,
		},
	)
	return k.icaControllerKeeper.RegisterInterchainAccount(ctx, connectionID, owner.String(), version)
}

// GetInterchainAccountAddress fetches the interchain account address for given `connectionId` and `owner`
func (k *Keeper) GetInterchainAccountAddress(ctx sdk.Context, connectionID, owner string) (string, error) {
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return "", status.Errorf(codes.InvalidArgument, "invalid owner address: %s", err)
	}

	icaAddress, found := k.icaControllerKeeper.GetInterchainAccountAddress(ctx, connectionID, portID)

	if !found {
		return "", status.Errorf(codes.NotFound, "could not find account")
	}

	return icaAddress, nil
}
