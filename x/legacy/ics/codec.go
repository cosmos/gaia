package ics

import (
	"reflect"

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

// resolve returns the proto.Message to register for the given fully-qualified
// proto type name. If the type was already registered in the global proto
// registry (e.g. because the ICS dependency is still in go.mod and its
// generated init() ran first), we use that registered type so that
// proto.MessageName continues to work for gRPC reflection. Otherwise we
// register the stub in the global registry and return the stub.
func resolve(fqName string, stub proto.Message) proto.Message {
	if t := proto.MessageType(fqName); t != nil {
		return reflect.New(t.Elem()).Interface().(proto.Message)
	}
	proto.RegisterType(stub, fqName)
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
