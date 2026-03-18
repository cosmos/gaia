package ics

import (
	"bytes"
	"compress/gzip"
	"fmt"

	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"

	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
)

// fileDescBytes holds a gzip-compressed FileDescriptorProto for all legacy ICS
// provider message stubs. It is used to satisfy the Descriptor() ([]byte,[]int)
// interface required by the Cosmos SDK's unknownproto package when decoding
// transactions and governance proposals that reference these type URLs.
//
// Each message declares its actual field numbers with wire-type-compatible types
// so that the unknownproto field checker does not reject stored historical data.
// All string/bytes/message fields are declared as TYPE_BYTES (wire type 2).
// Integer/bool fields are declared as TYPE_INT64 or TYPE_UINT32/UINT64/BOOL
// (wire type 0) to match the actual varint encoding on-chain.
var fileDescBytes []byte

// Message indices within fileDescBytes (order must match MessageType slice in init).
const (
	idxMsgAssignConsumerKey          = 0
	idxMsgConsumerAddition           = 1
	idxMsgConsumerRemoval            = 2
	idxMsgConsumerModification       = 3
	idxMsgCreateConsumer             = 4
	idxMsgUpdateConsumer             = 5
	idxMsgRemoveConsumer             = 6
	idxMsgChangeRewardDenoms         = 7
	idxMsgUpdateParams               = 8
	idxMsgSubmitConsumerMisbehaviour = 9
	idxMsgSubmitConsumerDoubleVoting = 10
	idxMsgOptIn                      = 11
	idxMsgOptOut                     = 12
	idxMsgSetConsumerCommissionRate  = 13
	idxConsumerAdditionProposal      = 14
	idxConsumerRemovalProposal       = 15
	idxConsumerModificationProposal  = 16
	idxChangeRewardDenomsProposal    = 17
	idxEquivocationProposal          = 18
)

// fB returns a BYTES-typed field descriptor (wire type 2 — string, bytes, message).
func fB(num int32) *descriptorpb.FieldDescriptorProto {
	name := fmt.Sprintf("f%d", num)
	typ := descriptorpb.FieldDescriptorProto_TYPE_BYTES
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	return &descriptorpb.FieldDescriptorProto{Name: &name, Number: &num, Type: &typ, Label: &label}
}

// fI returns an INT64-typed field descriptor (wire type 0 — varint).
func fI(num int32) *descriptorpb.FieldDescriptorProto {
	name := fmt.Sprintf("f%d", num)
	typ := descriptorpb.FieldDescriptorProto_TYPE_INT64
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	return &descriptorpb.FieldDescriptorProto{Name: &name, Number: &num, Type: &typ, Label: &label}
}

// fU32 returns a UINT32-typed field descriptor (wire type 0 — varint).
func fU32(num int32) *descriptorpb.FieldDescriptorProto {
	name := fmt.Sprintf("f%d", num)
	typ := descriptorpb.FieldDescriptorProto_TYPE_UINT32
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	return &descriptorpb.FieldDescriptorProto{Name: &name, Number: &num, Type: &typ, Label: &label}
}

// fU64 returns a UINT64-typed field descriptor (wire type 0 — varint).
func fU64(num int32) *descriptorpb.FieldDescriptorProto {
	name := fmt.Sprintf("f%d", num)
	typ := descriptorpb.FieldDescriptorProto_TYPE_UINT64
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	return &descriptorpb.FieldDescriptorProto{Name: &name, Number: &num, Type: &typ, Label: &label}
}

// fBool returns a BOOL-typed field descriptor (wire type 0 — varint).
func fBool(num int32) *descriptorpb.FieldDescriptorProto {
	name := fmt.Sprintf("f%d", num)
	typ := descriptorpb.FieldDescriptorProto_TYPE_BOOL
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	return &descriptorpb.FieldDescriptorProto{Name: &name, Number: &num, Type: &typ, Label: &label}
}

// msgServiceAnnotation returns a *descriptorpb.ServiceOptions with the
// cosmos.msg.v1.service = true extension set. This annotation is required by
// the SDK's proto annotation validator which runs at app init and warns/errors
// if any service named "Msg" in the merged proto registry (GlobalFiles + gogo
// registry) lacks this option.
func msgServiceAnnotation() *descriptorpb.ServiceOptions {
	opts := &descriptorpb.ServiceOptions{}
	protov2.SetExtension(opts, msgv1.E_Service, true)
	return opts
}

func init() {
	name := func(s string) *string { return &s }

	fdp := &descriptorpb.FileDescriptorProto{
		Name:    name("interchain_security/ccv/provider/v1/legacy_stubs.proto"),
		Package: name("interchain_security.ccv.provider.v1"),
		Syntax:  name("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{
			// 0 — MsgAssignConsumerKey: all string fields (wire type 2)
			{Name: name("MsgAssignConsumerKey"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2), fB(3), fB(4), fB(5),
			}},
			// 1 — MsgConsumerAddition: string/bytes/message fields + int64/uint32/uint64/bool varints
			{Name: name("MsgConsumerAddition"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2), fB(3), fB(4), fB(5), fB(6), fB(7), fB(8), fB(9),
				fI(10), fI(11), fB(12), fU32(13), fU32(14), fU32(15),
				fB(16), fB(17), fB(18), fU64(19), fBool(20),
			}},
			// 2 — MsgConsumerRemoval: string + Timestamp + string (all wire type 2)
			{Name: name("MsgConsumerRemoval"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2), fB(3),
			}},
			// 3 — MsgConsumerModification: string fields + uint32/uint64/bool varints
			{Name: name("MsgConsumerModification"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2), fB(3), fU32(4), fU32(5), fU32(6),
				fB(7), fB(8), fB(9), fU64(10), fBool(11),
			}},
			// 4 — MsgCreateConsumer: all string/message fields (wire type 2)
			{Name: name("MsgCreateConsumer"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2), fB(3), fB(4), fB(5), fB(6), fB(7),
			}},
			// 5 — MsgUpdateConsumer: all string/message fields (wire type 2)
			{Name: name("MsgUpdateConsumer"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2), fB(3), fB(4), fB(5), fB(6), fB(7), fB(8), fB(9),
			}},
			// 6 — MsgRemoveConsumer: string fields (wire type 2)
			{Name: name("MsgRemoveConsumer"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2),
			}},
			// 7 — MsgChangeRewardDenoms: repeated string + string (all wire type 2)
			{Name: name("MsgChangeRewardDenoms"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2), fB(3),
			}},
			// 8 — MsgUpdateParams: string + message (all wire type 2)
			{Name: name("MsgUpdateParams"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2),
			}},
			// 9 — MsgSubmitConsumerMisbehaviour: string + message + string (wire type 2)
			{Name: name("MsgSubmitConsumerMisbehaviour"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2), fB(3),
			}},
			// 10 — MsgSubmitConsumerDoubleVoting: string + message + message + string (wire type 2)
			{Name: name("MsgSubmitConsumerDoubleVoting"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2), fB(3), fB(4),
			}},
			// 11 — MsgOptIn: all string fields (wire type 2)
			{Name: name("MsgOptIn"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2), fB(3), fB(4), fB(5),
			}},
			// 12 — MsgOptOut: all string fields (wire type 2)
			{Name: name("MsgOptOut"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2), fB(3), fB(4),
			}},
			// 13 — MsgSetConsumerCommissionRate: all string fields (wire type 2)
			{Name: name("MsgSetConsumerCommissionRate"), Field: []*descriptorpb.FieldDescriptorProto{
				fB(1), fB(2), fB(3), fB(4), fB(5),
			}},
			// 14-18 — governance proposal stubs (field data not required for proposals query)
			{Name: name("ConsumerAdditionProposal")},
			{Name: name("ConsumerRemovalProposal")},
			{Name: name("ConsumerModificationProposal")},
			{Name: name("ChangeRewardDenomsProposal")},
			{Name: name("EquivocationProposal")},
		},
		// Tx service descriptor: required so that baseapp's MsgServiceRouter can
		// register stub handlers via RegisterService (which calls
		// registerHybridHandler → HybridResolver.FindDescriptorByName).
		// Without the service definition in the file descriptor the hybrid handler
		// registration panics at startup.
		//
		// The cosmos.msg.v1.service = true annotation is required; on app init the
		// SDK calls msgservice.ValidateProtoAnnotations over the merged registry
		// (protoregistry.GlobalFiles + gogo registry) and panics/warns if any
		// service named "Msg" lacks this annotation.
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name:    name("Msg"),
				Options: msgServiceAnnotation(),

				Method: []*descriptorpb.MethodDescriptorProto{
					{Name: name("AssignConsumerKey"), InputType: name(".interchain_security.ccv.provider.v1.MsgAssignConsumerKey"), OutputType: name(".interchain_security.ccv.provider.v1.MsgAssignConsumerKey")},
					{Name: name("ConsumerAddition"), InputType: name(".interchain_security.ccv.provider.v1.MsgConsumerAddition"), OutputType: name(".interchain_security.ccv.provider.v1.MsgConsumerAddition")},
					{Name: name("ConsumerRemoval"), InputType: name(".interchain_security.ccv.provider.v1.MsgConsumerRemoval"), OutputType: name(".interchain_security.ccv.provider.v1.MsgConsumerRemoval")},
					{Name: name("ConsumerModification"), InputType: name(".interchain_security.ccv.provider.v1.MsgConsumerModification"), OutputType: name(".interchain_security.ccv.provider.v1.MsgConsumerModification")},
					{Name: name("CreateConsumer"), InputType: name(".interchain_security.ccv.provider.v1.MsgCreateConsumer"), OutputType: name(".interchain_security.ccv.provider.v1.MsgCreateConsumer")},
					{Name: name("UpdateConsumer"), InputType: name(".interchain_security.ccv.provider.v1.MsgUpdateConsumer"), OutputType: name(".interchain_security.ccv.provider.v1.MsgUpdateConsumer")},
					{Name: name("RemoveConsumer"), InputType: name(".interchain_security.ccv.provider.v1.MsgRemoveConsumer"), OutputType: name(".interchain_security.ccv.provider.v1.MsgRemoveConsumer")},
					{Name: name("ChangeRewardDenoms"), InputType: name(".interchain_security.ccv.provider.v1.MsgChangeRewardDenoms"), OutputType: name(".interchain_security.ccv.provider.v1.MsgChangeRewardDenoms")},
					{Name: name("UpdateParams"), InputType: name(".interchain_security.ccv.provider.v1.MsgUpdateParams"), OutputType: name(".interchain_security.ccv.provider.v1.MsgUpdateParams")},
					{Name: name("SubmitConsumerMisbehaviour"), InputType: name(".interchain_security.ccv.provider.v1.MsgSubmitConsumerMisbehaviour"), OutputType: name(".interchain_security.ccv.provider.v1.MsgSubmitConsumerMisbehaviour")},
					{Name: name("SubmitConsumerDoubleVoting"), InputType: name(".interchain_security.ccv.provider.v1.MsgSubmitConsumerDoubleVoting"), OutputType: name(".interchain_security.ccv.provider.v1.MsgSubmitConsumerDoubleVoting")},
					{Name: name("OptIn"), InputType: name(".interchain_security.ccv.provider.v1.MsgOptIn"), OutputType: name(".interchain_security.ccv.provider.v1.MsgOptIn")},
					{Name: name("OptOut"), InputType: name(".interchain_security.ccv.provider.v1.MsgOptOut"), OutputType: name(".interchain_security.ccv.provider.v1.MsgOptOut")},
					{Name: name("SetConsumerCommissionRate"), InputType: name(".interchain_security.ccv.provider.v1.MsgSetConsumerCommissionRate"), OutputType: name(".interchain_security.ccv.provider.v1.MsgSetConsumerCommissionRate")},
				},
			},
		},
	}

	b, err := protov2.Marshal(fdp)
	if err != nil {
		panic("legacyics: failed to marshal file descriptor: " + err.Error())
	}

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(b); err != nil {
		panic("legacyics: failed to gzip file descriptor: " + err.Error())
	}
	if err := w.Close(); err != nil {
		panic("legacyics: failed to close gzip writer: " + err.Error())
	}
	fileDescBytes = buf.Bytes()

	// Register with the protov2 registries so that the aminojson encoder
	// (used for the proposals query response marshaling) can resolve these
	// type URLs via protoregistry.GlobalTypes / GlobalFiles.
	fd, err := protodesc.NewFile(fdp, protoregistry.GlobalFiles)
	if err != nil {
		panic("legacyics: failed to build protov2 file descriptor: " + err.Error())
	}
	if err := protoregistry.GlobalFiles.RegisterFile(fd); err != nil {
		// "already registered" is harmless — happens if two init paths run.
		_ = err
	}
	msgs := fd.Messages()
	for i := 0; i < msgs.Len(); i++ {
		mt := dynamicpb.NewMessageType(msgs.Get(i))
		if err := protoregistry.GlobalTypes.RegisterMessage(mt); err != nil {
			_ = err
		}
	}
}
