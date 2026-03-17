package ics

import (
	"context"

	"google.golang.org/grpc"

	errorsmod "cosmossdk.io/errors"

	gogogrpc "github.com/cosmos/gogoproto/grpc"

	gaiaerrors "github.com/cosmos/gaia/v28/types/errors"
)

// legacyMsgServer is a placeholder gRPC server that rejects all legacy ICS
// provider messages with ErrDeprecatedMessage.
type legacyMsgServer struct{}

// RegisterLegacyMsgHandlers registers no-op handlers for all 14 legacy ICS
// provider message types with the given gRPC server (i.e., the app's
// MsgServiceRouter).
//
// Background: the Cosmos SDK BaseApp checks whether a message handler exists
// for every message in a transaction BEFORE running the ante handler chain.
// Without registered handlers the node returns ErrUnknownRequest immediately,
// which prevents RejectLegacyICSDecorator (position 2 in the ante chain) from
// ever firing and returning the more informative ErrDeprecatedMessage.
//
// With these handlers in place the pre-ante check passes, the ante decorator
// rejects the tx with ErrDeprecatedMessage, and the message handler itself is
// never executed.
//
// Must be called after RegisterInterfaces so that sdk.MsgTypeURL resolves
// correctly for the stubs during handler registration.
func RegisterLegacyMsgHandlers(s gogogrpc.Server) {
	s.RegisterService(&_LegacyICS_Msg_serviceDesc, &legacyMsgServer{})
}

func deprecated(_ context.Context, msg any) (any, error) {
	return nil, errorsmod.Wrapf(gaiaerrors.ErrDeprecatedMessage,
		"legacy ICS message type %T is no longer accepted (msg handler)", msg)
}

// methodDesc builds a grpc.MethodDesc for a single legacy ICS stub method.
// newReq allocates the concrete stub type so that %T in the error message is
// informative if the handler is ever reached. In practice it is not — the ante
// decorator fires first.
func methodDesc(methodName, fullMethod string, newReq func() any) grpc.MethodDesc {
	return grpc.MethodDesc{
		MethodName: methodName,
		Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
			req := newReq()
			if err := dec(req); err != nil {
				return nil, err
			}
			if interceptor == nil {
				return deprecated(ctx, req)
			}
			info := &grpc.UnaryServerInfo{Server: srv, FullMethod: fullMethod}
			return interceptor(ctx, req, info, func(c context.Context, r any) (any, error) {
				return deprecated(c, r)
			})
		},
	}
}

const msgSvcPrefix = "/interchain_security.ccv.provider.v1.Msg/"

// _LegacyICS_Msg_serviceDesc is a hand-crafted gRPC ServiceDesc for the
// ICS provider Msg service. The Handler for each method calls dec() with the
// correct stub type so that registerMsgServiceHandler captures the right
// sdk.MsgTypeURL during registration. The actual method body returns
// ErrDeprecatedMessage, but in practice it is never reached because the ante
// decorator fires first.
var _LegacyICS_Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "interchain_security.ccv.provider.v1.Msg",
	HandlerType: (*legacyMsgServer)(nil),
	Methods: []grpc.MethodDesc{
		methodDesc("AssignConsumerKey", msgSvcPrefix+"AssignConsumerKey", func() any { return new(MsgAssignConsumerKey) }),
		methodDesc("ConsumerAddition", msgSvcPrefix+"ConsumerAddition", func() any { return new(MsgConsumerAddition) }),
		methodDesc("ConsumerRemoval", msgSvcPrefix+"ConsumerRemoval", func() any { return new(MsgConsumerRemoval) }),
		methodDesc("ConsumerModification", msgSvcPrefix+"ConsumerModification", func() any { return new(MsgConsumerModification) }),
		methodDesc("CreateConsumer", msgSvcPrefix+"CreateConsumer", func() any { return new(MsgCreateConsumer) }),
		methodDesc("UpdateConsumer", msgSvcPrefix+"UpdateConsumer", func() any { return new(MsgUpdateConsumer) }),
		methodDesc("RemoveConsumer", msgSvcPrefix+"RemoveConsumer", func() any { return new(MsgRemoveConsumer) }),
		methodDesc("ChangeRewardDenoms", msgSvcPrefix+"ChangeRewardDenoms", func() any { return new(MsgChangeRewardDenoms) }),
		methodDesc("UpdateParams", msgSvcPrefix+"UpdateParams", func() any { return new(MsgUpdateParams) }),
		methodDesc("SubmitConsumerMisbehaviour", msgSvcPrefix+"SubmitConsumerMisbehaviour", func() any { return new(MsgSubmitConsumerMisbehaviour) }),
		methodDesc("SubmitConsumerDoubleVoting", msgSvcPrefix+"SubmitConsumerDoubleVoting", func() any { return new(MsgSubmitConsumerDoubleVoting) }),
		methodDesc("OptIn", msgSvcPrefix+"OptIn", func() any { return new(MsgOptIn) }),
		methodDesc("OptOut", msgSvcPrefix+"OptOut", func() any { return new(MsgOptOut) }),
		methodDesc("SetConsumerCommissionRate", msgSvcPrefix+"SetConsumerCommissionRate", func() any { return new(MsgSetConsumerCommissionRate) }),
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "interchain_security/ccv/provider/v1/legacy_stubs.proto",
}
