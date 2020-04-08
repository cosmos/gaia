package consuming

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/band-consumer/x/consuming/types"
)

// NewHandler creates the msg handler of this module, as required by Cosmos-SDK standard.
func NewHandler(k types.MsgServer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.MsgRequestData:
			res, err := k.RequestData(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		// case channeltypes.MsgPacket:
		// 	var responseData oracle.OracleResponsePacketData
		// 	if err := types.ModuleCdc.UnmarshalJSON(msg.GetData(), &responseData); err == nil {
		// 		fmt.Println("I GOT DATA", responseData.Result)
		// 		return &sdk.Result{Events: ctx.EventManager().Events().ToABCIEvents()}, nil
		// 	}
		// 	return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal oracle packet data")
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
