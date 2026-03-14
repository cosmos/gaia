package ics

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	proto "github.com/cosmos/gogoproto/proto"
)

// customRegistrar is satisfied by the concrete *interfaceRegistry in the SDK.
// It lets us register stubs under explicit typeURLs without relying on
// proto.MessageName, which would return "" for unregistered stub types.
type customRegistrar interface {
	RegisterCustomTypeURL(iface any, typeURL string, impl proto.Message)
}

// resolve registers the stub in the global proto registry under fqName (so
// that proto.MessageName works for gRPC reflection) and returns the stub.
// The check makes this safe to call multiple times (e.g. when NewGaiaApp is
// instantiated more than once in a process, as in e2e test setup).
func resolve(fqName string, stub proto.Message) proto.Message {
	if proto.MessageType(fqName) == nil {
		proto.RegisterType(stub, fqName)
	}
	return stub
}

// RegisterInterfaces registers legacy ICS provider message and proposal stubs
// with the interface registry so that historical on-chain data containing these
// type URLs can be decoded after the ICS provider module has been removed.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	cr, ok := registry.(customRegistrar)
	if !ok {
		panic("interface registry does not support RegisterCustomTypeURL; cannot register legacy ICS stubs")
	}

	const pkg = "interchain_security.ccv.provider.v1."

	msgs := []struct {
		name string
		stub proto.Message
	}{
		{"MsgAssignConsumerKey", &MsgAssignConsumerKey{}},
		{"MsgConsumerAddition", &MsgConsumerAddition{}},
		{"MsgConsumerRemoval", &MsgConsumerRemoval{}},
		{"MsgConsumerModification", &MsgConsumerModification{}},
		{"MsgCreateConsumer", &MsgCreateConsumer{}},
		{"MsgUpdateConsumer", &MsgUpdateConsumer{}},
		{"MsgRemoveConsumer", &MsgRemoveConsumer{}},
		{"MsgChangeRewardDenoms", &MsgChangeRewardDenoms{}},
		{"MsgUpdateParams", &MsgUpdateParams{}},
		{"MsgSubmitConsumerMisbehaviour", &MsgSubmitConsumerMisbehaviour{}},
		{"MsgSubmitConsumerDoubleVoting", &MsgSubmitConsumerDoubleVoting{}},
		{"MsgOptIn", &MsgOptIn{}},
		{"MsgOptOut", &MsgOptOut{}},
		{"MsgSetConsumerCommissionRate", &MsgSetConsumerCommissionRate{}},
	}
	for _, m := range msgs {
		cr.RegisterCustomTypeURL((*sdk.Msg)(nil), "/"+pkg+m.name, resolve(pkg+m.name, m.stub))
	}

	proposals := []struct {
		name string
		stub proto.Message
	}{
		{"ConsumerAdditionProposal", &ConsumerAdditionProposal{}},
		{"ConsumerRemovalProposal", &ConsumerRemovalProposal{}},
		{"ConsumerModificationProposal", &ConsumerModificationProposal{}},
		{"ChangeRewardDenomsProposal", &ChangeRewardDenomsProposal{}},
		{"EquivocationProposal", &EquivocationProposal{}},
	}
	for _, p := range proposals {
		cr.RegisterCustomTypeURL((*govv1beta1.Content)(nil), "/"+pkg+p.name, resolve(pkg+p.name, p.stub))
	}
}
