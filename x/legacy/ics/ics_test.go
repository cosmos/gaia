package ics_test

// Package ics_test verifies that the legacy ICS stub types work correctly with
// the Cosmos SDK codec machinery: specifically that historical governance
// proposals and transactions containing ICS provider type URLs can be decoded
// after the ICS provider module has been removed.

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/cosmos/gaia/v28/app/params"
	"github.com/cosmos/gaia/v28/x/legacy/ics"
)

// ---------------------------------------------------------------------------
// Proto-wire helpers
// ---------------------------------------------------------------------------

// appendLenField appends a proto3 length-delimited field (wire type 2) to buf.
// Used for string, bytes, and embedded-message fields.
func appendLenField(buf []byte, fieldNum uint32, data []byte) []byte {
	buf = binary.AppendUvarint(buf, uint64((fieldNum<<3)|2)) // tag
	buf = binary.AppendUvarint(buf, uint64(len(data)))       // length
	return append(buf, data...)
}

// appendVarintField appends a proto3 varint field (wire type 0) to buf.
// Used for uint32/uint64/bool/int32/int64 fields.
func appendVarintField(buf []byte, fieldNum uint32, value uint64) []byte {
	buf = binary.AppendUvarint(buf, uint64(fieldNum<<3)) // tag (wire type 0)
	buf = binary.AppendUvarint(buf, value)
	return buf
}

// encodeMsgUpdateConsumerProposal returns the proto3 wire encoding for a
// MsgUpdateConsumer proposal to opt-in to PSS.
// Field layout follows the ICS provider v1 proto definition:
//
// field 1  owner                  (string)
// field 2  consumer_id            (string)
// field 3  new_owner_address      (string)
// field 5  power_shaping_params   (embedded message)
//
//	field 2  validators_power_cap  (uint32)
//	field 3  validator_set_cap      (uint32)
//	field 4  allowlist              (repeated string)
func encodeMsgUpdateConsumerProposal() []byte {
	// Build PowerShapingParameters sub-message bytes.
	var ps []byte
	ps = appendVarintField(ps, 2, 1) // validators_power_cap = 1
	ps = appendVarintField(ps, 3, 3) // validator_set_cap = 3
	for _, addr := range []string{
		"cosmosvalcons12m5td27rwwy95drgk53w9pfhlxqqguqmlfph2g",
		"cosmosvalcons15yprks04304h8wg0x2fef53g50x9w2qa3c0hcd",
		"cosmosvalcons146zd98kguwau7y3mfrrs9k4fsthv9qct9mdnx0",
	} {
		ps = appendLenField(ps, 4, []byte(addr)) // allowlist entry
	}

	var msg []byte
	msg = appendLenField(msg, 1, []byte("cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")) // owner
	msg = appendLenField(msg, 2, []byte("1"))                                             // consumer_id
	msg = appendLenField(msg, 3, []byte("cosmos1arjwkww79m65csulawqngr7ngs4uqu5hx9ak2a")) // new_owner_address
	msg = appendLenField(msg, 5, ps)                                                      // power_shaping_parameters
	return msg
}

// encodeConsumerAdditionProposal hand-encodes a ConsumerAdditionProposal in
// proto3 wire format, covering every field type used by the descriptor:
//
//	1-3  title/description/chain_id          string  (wire type 2)
//	4    initial_height                       message (wire type 2)
//	5-6  genesis_hash/binary_hash             bytes   (wire type 2)
//	7    spawn_time                           message (wire type 2)
//	8-10 unbonding/ccv_timeout/transfer durations message (wire type 2)
//	11   consumer_redistribution_fraction     string  (wire type 2)
//	12   blocks_per_distribution_transmission int64   (wire type 0)
//	13   historical_entries                   int64   (wire type 0)
//	14   distribution_transmission_channel    string  (wire type 2)
//	15   top_N                                uint32  (wire type 0)
//	16   validators_power_cap                 uint32  (wire type 0)
//	17   validator_set_cap                    uint32  (wire type 0)
//	18-19 allowlist/denylist                 repeated string (wire type 2)
//	20   min_stake                            uint64  (wire type 0)
//	21   allow_inactive_vals                  bool    (wire type 0)
func encodeConsumerAdditionProposal() []byte {
	// ibc.core.client.v1.Height: revision_number (f1 uint64), revision_height (f2 uint64)
	var height []byte
	height = appendVarintField(height, 1, 1)
	height = appendVarintField(height, 2, 1)

	// google.protobuf.Timestamp: seconds (f1 int64)
	var spawnTime []byte
	spawnTime = appendVarintField(spawnTime, 1, 1_700_000_000)

	// google.protobuf.Duration: seconds (f1 int64) — reused for all three duration fields
	var dur []byte
	dur = appendVarintField(dur, 1, 1_728_000)

	var p []byte
	p = appendLenField(p, 1, []byte("Add consumer chain"))                                    // title
	p = appendLenField(p, 2, []byte("Spawn a new consumer chain"))                            // description
	p = appendLenField(p, 3, []byte("consumer-1"))                                            // chain_id
	p = appendLenField(p, 4, height)                                                          // initial_height
	p = appendLenField(p, 5, []byte("genesis_hash_bytes"))                                    // genesis_hash
	p = appendLenField(p, 6, []byte("binary_hash_bytes"))                                     // binary_hash
	p = appendLenField(p, 7, spawnTime)                                                       // spawn_time
	p = appendLenField(p, 8, dur)                                                             // unbonding_period
	p = appendLenField(p, 9, dur)                                                             // ccv_timeout_period
	p = appendLenField(p, 10, dur)                                                            // transfer_timeout_period
	p = appendLenField(p, 11, []byte("0.75"))                                                 // consumer_redistribution_fraction
	p = appendVarintField(p, 12, 1000)                                                        // blocks_per_distribution_transmission (int64)
	p = appendVarintField(p, 13, 10000)                                                       // historical_entries (int64)
	p = appendLenField(p, 14, []byte(""))                                                     // distribution_transmission_channel
	p = appendVarintField(p, 15, 67)                                                          // top_N (uint32)
	p = appendVarintField(p, 16, 0)                                                           // validators_power_cap (uint32)
	p = appendVarintField(p, 17, 0)                                                           // validator_set_cap (uint32)
	p = appendLenField(p, 18, []byte("cosmosvalcons12m5td27rwwy95drgk53w9pfhlxqqguqmlfph2g")) // allowlist
	p = appendLenField(p, 19, []byte("cosmosvalcons15yprks04304h8wg0x2fef53g50x9w2qa3c0hcd")) // denylist
	p = appendVarintField(p, 20, 0)                                                           // min_stake (uint64)
	p = appendVarintField(p, 21, 0)                                                           // allow_inactive_vals (bool)
	return p
}

// msgUpdateConsumerAny returns a codectypes.Any wrapping the
// proposal-1014 MsgUpdateConsumer bytes under the ICS provider type URL.
func msgUpdateConsumerAny() *codectypes.Any {
	return &codectypes.Any{
		TypeUrl: "/interchain_security.ccv.provider.v1.MsgUpdateConsumer",
		Value:   encodeMsgUpdateConsumerProposal(),
	}
}

// ---------------------------------------------------------------------------
// Test 1 -- round-trip through the SDK tx decoder
// ---------------------------------------------------------------------------

// TestRoundTripMsgUpdateConsumerTxDecode encodes a MsgUpdateConsumer
// proposal as the body of a minimal Cosmos SDK transaction, then
// feeds the raw bytes through the SDK TxDecoder -- with only the ICS legacy
// stubs registered on the interface registry -- and asserts that no error is
// returned.
//
// This verifies that historical transactions stored in state can be retrieved
// by query endpoints (which call TxDecoder internally) after the ICS provider
// module has been removed.
func TestRoundTripMsgUpdateConsumerTxDecode(t *testing.T) {
	encCfg := params.MakeEncodingConfig()
	ics.RegisterInterfaces(encCfg.InterfaceRegistry)

	// Build TxBody containing the MsgUpdateConsumer Any.
	body := &txtypes.TxBody{
		Messages: []*codectypes.Any{msgUpdateConsumerAny()},
		Memo:     "Update consumer ",
	}
	bodyBytes, err := body.Marshal()
	require.NoError(t, err)

	// Minimal AuthInfo (no signers, no fee) and one placeholder signature.
	authInfoBytes, err := (&txtypes.AuthInfo{}).Marshal()
	require.NoError(t, err)

	txRaw := &txtypes.TxRaw{
		BodyBytes:     bodyBytes,
		AuthInfoBytes: authInfoBytes,
		Signatures:    [][]byte{{}},
	}
	txBytes, err := txRaw.Marshal()
	require.NoError(t, err)

	_, err = encCfg.TxConfig.TxDecoder()(txBytes)
	require.NoError(t, err, "TxDecoder must not error on a tx containing a legacy ICS message")
}

// ---------------------------------------------------------------------------
// Test 1b -- JSON tx decode with real ICS field names (tx broadcast path)
// ---------------------------------------------------------------------------

// TestTxJSONDecoderWithRealICSFields reproduces the error seen when running
// `gaiad tx broadcast` against a build with ICS removed. The JSON transaction
// was originally built with the real ICS proto (which includes rich field
// names like "allowlisted_reward_denoms") and then signed. Without the
// UnmarshalJSONPB no-op on stubMsg, the SDK's jsonpb unmarshaler would fail
// with "unknown field %q in ics.MsgCreateConsumer" because AllowUnknownFields
// defaults to false and the stub registers no proto-tagged fields.
//
// After the fix, TxJSONDecoder must succeed and the ante handler is then able
// to reject the deprecated type URL with a clear ErrDeprecatedMessage error.
func TestTxJSONDecoderWithRealICSFields(t *testing.T) {
	encCfg := params.MakeEncodingConfig()
	ics.RegisterInterfaces(encCfg.InterfaceRegistry)

	// A minimal signed-tx JSON in the same shape as create-signed.json, i.e.
	// the MsgCreateConsumer body contains real ICS field names that do NOT
	// appear in the stub's proto descriptor.
	// The secp256k1 public key Any is omitted from auth_info because its
	// resolution is orthogonal to the ICS field-name issue under test; the
	// decoder resolves signer_infos only when a key is present.
	const signedTxJSON = `{
		"body": {
			"messages": [{
				"@type": "/interchain_security.ccv.provider.v1.MsgCreateConsumer",
				"submitter": "cosmos1r5v5srda7xfth3hn2s26txvrcrntldjumt8mhl",
				"chain_id": "test-consumer-1",
				"metadata": {
					"name": "Test consumer chain",
					"description": "The chain will join the ICS testnet as an opt-in chain.",
					"metadata": "ipfs://"
				},
				"initialization_parameters": {
					"initial_height": {"revision_number": "1", "revision_height": "1"},
					"spawn_time": "0001-01-01T00:00:00Z",
					"unbonding_period": "1728000s",
					"consumer_redistribution_fraction": "0.75"
				},
				"power_shaping_parameters": {
					"top_N": 0,
					"validators_power_cap": 0,
					"validator_set_cap": 0,
					"allowlist": [],
					"denylist": []
				},
				"allowlisted_reward_denoms": null,
				"infraction_parameters": null
			}],
			"memo": "",
			"timeout_height": "0"
		},
		"auth_info": {
			"signer_infos": [],
			"fee": {
				"amount": [{"denom": "uatom", "amount": "2005"}],
				"gas_limit": "400908"
			}
		},
		"signatures": []
	}`

	_, err := encCfg.TxConfig.TxJSONDecoder()([]byte(signedTxJSON))
	require.NoError(t, err,
		"TxJSONDecoder must not error on a JSON tx with real ICS field names; "+
			"the ante handler should be responsible for rejection, not the decoder")
}

// ---------------------------------------------------------------------------
// Test 2 -- proposals query path regression
// ---------------------------------------------------------------------------

// TestProposalsQueryPathRegression simulates the code path executed when the
// governance gRPC query handler serialises a QueryProposalResponse to JSON and
// a client deserialises it.  It creates a govv1.Proposal whose Messages slice
// contains the MsgUpdateConsumer Any from proposal #1014, marshals the
// response to JSON using the SDK codec, and then unmarshals it back, asserting:
//
//  1. JSON marshaling succeeds without "unknown type URL" errors.
//  2. The proposal ID and message type URL survive the round-trip.
func TestProposalsQueryPathRegression(t *testing.T) {
	encCfg := params.MakeEncodingConfig()
	ics.RegisterInterfaces(encCfg.InterfaceRegistry)

	cdc := codec.NewProtoCodec(encCfg.InterfaceRegistry)

	proposal := &govv1.Proposal{
		Id:       10,
		Messages: []*codectypes.Any{msgUpdateConsumerAny()},
		Status:   govv1.StatusPassed,
		Title:    "Update consumer chain",
		Proposer: "cosmos1mrwtsv7p53k90ey2nej4glsv3gphujkh8fr0mx",
	}
	resp := &govv1.QueryProposalResponse{Proposal: proposal}

	// Simulate what the gRPC-gateway JSON transcoder does on the server side.
	bz, err := cdc.MarshalJSON(resp)
	require.NoError(t, err, "MarshalJSON must not error on a proposal containing a legacy ICS message")
	require.NotEmpty(t, bz)

	// Simulate the client-side decode.
	var decoded govv1.QueryProposalResponse
	err = cdc.UnmarshalJSON(bz, &decoded)
	require.NoError(t, err, "UnmarshalJSON must not error when decoding a legacy ICS proposal")
	require.NotNil(t, decoded.Proposal)
	require.Equal(t, uint64(10), decoded.Proposal.Id)
	require.Len(t, decoded.Proposal.Messages, 1)
	require.Equal(t,
		"/interchain_security.ccv.provider.v1.MsgUpdateConsumer",
		decoded.Proposal.Messages[0].TypeUrl,
	)
}

// ---------------------------------------------------------------------------
// Test 3 -- gRPC codec round-trip (binary proto + interface unpacking)
// ---------------------------------------------------------------------------

// TestGovGRPCQueryProposalCodecRoundTrip exercises the binary gRPC codec path:
//
//  1. Proto-marshal a QueryProposalResponse (server side serialisation).
//  2. Proto-unmarshal the bytes back (client side deserialisation).
//  3. Call UnpackInterfaces on the embedded Proposal to resolve each Any to
//     its concrete stub type via the interface registry.
//  4. Verify the resolved concrete type is *ics.MsgUpdateConsumer and that it
//     satisfies sdk.Msg -- confirming the stub is correctly registered.
//
// This matches the sequence performed by the Cosmos SDK gRPC client middleware
// when receiving a response containing a legacy ICS type URL.
func TestGovGRPCQueryProposalCodecRoundTrip(t *testing.T) {
	encCfg := params.MakeEncodingConfig()
	ics.RegisterInterfaces(encCfg.InterfaceRegistry)

	cdc := codec.NewProtoCodec(encCfg.InterfaceRegistry)

	proposal := &govv1.Proposal{
		Id:       1014,
		Messages: []*codectypes.Any{msgUpdateConsumerAny()},
		Status:   govv1.StatusPassed,
	}
	resp := &govv1.QueryProposalResponse{Proposal: proposal}

	// 1. Server side: binary marshal (what grpc.Server sends over the wire).
	bz, err := cdc.Marshal(resp)
	require.NoError(t, err, "proto-marshaling the gRPC response must not error")

	// 2. Client side: binary unmarshal.
	var decoded govv1.QueryProposalResponse
	err = cdc.Unmarshal(bz, &decoded)
	require.NoError(t, err, "proto-unmarshaling the gRPC response must not error")
	require.NotNil(t, decoded.Proposal)
	require.Equal(t, uint64(1014), decoded.Proposal.Id)

	// 3. Resolve each message Any to its concrete type (what SDK gRPC client
	//    middleware calls after Unmarshal).
	err = decoded.Proposal.UnpackInterfaces(encCfg.InterfaceRegistry)
	require.NoError(t, err, "UnpackInterfaces must not error for legacy ICS type URLs")

	msgAny := decoded.Proposal.Messages[0]
	require.Equal(t,
		"/interchain_security.ccv.provider.v1.MsgUpdateConsumer",
		msgAny.TypeUrl,
	)

	// 4. The cached value must be the registered stub and must satisfy sdk.Msg.
	cached := msgAny.GetCachedValue()
	require.NotNil(t, cached, "GetCachedValue must return the resolved stub")
	require.IsType(t, &ics.MsgUpdateConsumer{}, cached)

	sdkMsg, ok := cached.(sdk.Msg)
	require.True(t, ok, "the resolved stub must satisfy sdk.Msg")
	require.NotNil(t, sdkMsg)
}

// ---------------------------------------------------------------------------
// Test 4 -- MsgSubmitProposal tx decode (historical proposal-submission tx)
// ---------------------------------------------------------------------------

// TestSubmitProposalTxWithICSContentTxDecode verifies that TxDecoder can decode
// a historical MsgSubmitProposal (govv1beta1) transaction whose content Any
// holds a ConsumerAdditionProposal.
//
// This is the exact path that fails when a client queries a legacy ICS
// proposal-submission tx by hash: the unknownproto field checker walks from
// TxBody.Messages → MsgSubmitProposal → content Any → proposal value bytes
// and validates each field against the descriptor of the resolved type.  If
// ConsumerAdditionProposal has no Descriptor() (or the wrong one), the check
// returns an error like "unknown field X for message ConsumerAdditionProposal".
func TestSubmitProposalTxWithICSContentTxDecode(t *testing.T) {
	encCfg := params.MakeEncodingConfig()
	// Register MsgSubmitProposal as sdk.Msg and the Content interface.
	govv1beta1.RegisterInterfaces(encCfg.InterfaceRegistry)
	// Register ICS proposal stubs as govv1beta1.Content implementations.
	ics.RegisterInterfaces(encCfg.InterfaceRegistry)

	// Encode a ConsumerAdditionProposal with representative field data for
	// every wire-type category used by the descriptor.
	contentAny := &codectypes.Any{
		TypeUrl: "/interchain_security.ccv.provider.v1.ConsumerAdditionProposal",
		Value:   encodeConsumerAdditionProposal(),
	}

	// Build a govv1beta1.MsgSubmitProposal and marshal it to bytes.
	submitMsg := &govv1beta1.MsgSubmitProposal{
		Content:  contentAny,
		Proposer: "cosmos1mrwtsv7p53k90ey2nej4glsv3gphujkh8fr0mx",
	}
	submitBytes, err := submitMsg.Marshal()
	require.NoError(t, err)

	// Wrap as an Any so it can appear in TxBody.Messages.
	submitAny := &codectypes.Any{
		TypeUrl: "/cosmos.gov.v1beta1.MsgSubmitProposal",
		Value:   submitBytes,
	}

	// Assemble TxRaw.
	body := &txtypes.TxBody{
		Messages: []*codectypes.Any{submitAny},
		Memo:     "submit ConsumerAdditionProposal",
	}
	bodyBytes, err := body.Marshal()
	require.NoError(t, err)

	authInfoBytes, err := (&txtypes.AuthInfo{}).Marshal()
	require.NoError(t, err)

	txRaw := &txtypes.TxRaw{
		BodyBytes:     bodyBytes,
		AuthInfoBytes: authInfoBytes,
		Signatures:    [][]byte{{}},
	}
	txBytes, err := txRaw.Marshal()
	require.NoError(t, err)

	_, err = encCfg.TxConfig.TxDecoder()(txBytes)
	require.NoError(t, err,
		"TxDecoder must decode a historical MsgSubmitProposal tx containing a "+
			"ConsumerAdditionProposal content Any without error")
}
