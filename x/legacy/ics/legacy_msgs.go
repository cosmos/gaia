// Package ics provides legacy stub types for interchain-security (ICS) provider
// module messages. These stubs allow historical governance proposals and
// transactions stored in state to be decoded and returned by queries after the
// ICS provider module has been removed. They do not preserve field data —
// Unmarshal is a no-op that discards bytes — but they prevent "unknown type URL"
// errors from breaking queries.
package ics

import "github.com/cosmos/gogoproto/jsonpb"

// stubMsg implements proto.Message and codec.ProtoMarshaler with no-op
// marshal/unmarshal behaviour.
type stubMsg struct{}

func (s *stubMsg) ProtoMessage()                                 {}
func (s *stubMsg) Reset()                                        {}
func (s *stubMsg) String() string                                { return "{}" }
func (s *stubMsg) Marshal() ([]byte, error)                      { return []byte{}, nil }
func (s *stubMsg) MarshalTo(dAtA []byte) (int, error)            { return 0, nil }
func (s *stubMsg) MarshalToSizedBuffer(dAtA []byte) (int, error) { return 0, nil }
func (s *stubMsg) Size() int                                     { return 0 }
func (s *stubMsg) Unmarshal(_ []byte) error                      { return nil }
func (s *stubMsg) ValidateBasic() error                          { return nil }

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler. This is the hook that
// gogoproto's jsonpb Unmarshaler calls before its reflection-based field
// parser, so it prevents "unknown field" errors when the JSON payload contains
// real ICS field names (e.g. from a historical signed transaction) that are not
// declared in the stub's descriptor.
func (s *stubMsg) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, _ []byte) error { return nil }

// ICS provider tx message stubs.

type (
	MsgAssignConsumerKey          struct{ stubMsg }
	MsgConsumerAddition           struct{ stubMsg }
	MsgConsumerRemoval            struct{ stubMsg }
	MsgConsumerModification       struct{ stubMsg }
	MsgCreateConsumer             struct{ stubMsg }
	MsgUpdateConsumer             struct{ stubMsg }
	MsgRemoveConsumer             struct{ stubMsg }
	MsgChangeRewardDenoms         struct{ stubMsg }
	MsgUpdateParams               struct{ stubMsg }
	MsgSubmitConsumerMisbehaviour struct{ stubMsg }
	MsgSubmitConsumerDoubleVoting struct{ stubMsg }
	MsgOptIn                      struct{ stubMsg }
	MsgOptOut                     struct{ stubMsg }
	MsgSetConsumerCommissionRate  struct{ stubMsg }
)

// Descriptor satisfies the descriptorIface required by the Cosmos SDK's
// unknownproto package for tx field validation. Returns a minimal gzipped
// FileDescriptorProto with no fields defined (all bytes treated as unknown
// non-criticals, which is allowed by the tx decoder).
func (m *MsgAssignConsumerKey) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgAssignConsumerKey}
}

func (m *MsgConsumerAddition) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgConsumerAddition}
}

func (m *MsgConsumerRemoval) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgConsumerRemoval}
}

func (m *MsgConsumerModification) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgConsumerModification}
}

func (m *MsgCreateConsumer) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgCreateConsumer}
}

func (m *MsgUpdateConsumer) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgUpdateConsumer}
}

func (m *MsgRemoveConsumer) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgRemoveConsumer}
}

func (m *MsgChangeRewardDenoms) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgChangeRewardDenoms}
}

func (m *MsgUpdateParams) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgUpdateParams}
}

func (m *MsgSubmitConsumerMisbehaviour) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgSubmitConsumerMisbehaviour}
}

func (m *MsgSubmitConsumerDoubleVoting) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgSubmitConsumerDoubleVoting}
}

func (m *MsgOptIn) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgOptIn}
}

func (m *MsgOptOut) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgOptOut}
}

func (m *MsgSetConsumerCommissionRate) Descriptor() ([]byte, []int) {
	return fileDescBytes, []int{idxMsgSetConsumerCommissionRate}
}
