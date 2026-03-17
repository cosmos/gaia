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
		{
			MethodName: "AssignConsumerKey",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgAssignConsumerKey)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/AssignConsumerKey"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "ConsumerAddition",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgConsumerAddition)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/ConsumerAddition"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "ConsumerRemoval",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgConsumerRemoval)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/ConsumerRemoval"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "ConsumerModification",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgConsumerModification)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/ConsumerModification"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "CreateConsumer",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgCreateConsumer)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/CreateConsumer"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "UpdateConsumer",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgUpdateConsumer)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/UpdateConsumer"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "RemoveConsumer",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgRemoveConsumer)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/RemoveConsumer"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "ChangeRewardDenoms",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgChangeRewardDenoms)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/ChangeRewardDenoms"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "UpdateParams",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgUpdateParams)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/UpdateParams"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "SubmitConsumerMisbehaviour",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgSubmitConsumerMisbehaviour)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/SubmitConsumerMisbehaviour"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "SubmitConsumerDoubleVoting",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgSubmitConsumerDoubleVoting)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/SubmitConsumerDoubleVoting"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "OptIn",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgOptIn)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/OptIn"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "OptOut",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgOptOut)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/OptOut"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
		{
			MethodName: "SetConsumerCommissionRate",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := new(MsgSetConsumerCommissionRate)
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return deprecated(ctx, req)
				}
				return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/interchain_security.ccv.provider.v1.Msg/SetConsumerCommissionRate"}, func(c context.Context, r any) (any, error) { return deprecated(c, r) })
			},
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "interchain_security/ccv/provider/v1/legacy_stubs.proto",
}
