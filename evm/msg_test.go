package evm

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	types2 "github.com/cosmos/gaia/v23/types"
	"github.com/stretchr/testify/require"
	proto2 "google.golang.org/protobuf/proto"
	"testing"
)

// MockTx implements sdk.Tx for testing
type MockTx struct {
	msgs []sdk.Msg
}

func (tx MockTx) GetMsgsV2() ([]proto2.Message, error) {
	//TODO implement me
	panic("implement me")
}

func (tx MockTx) GetMsgs() []sdk.Msg           { return tx.msgs }
func (tx MockTx) ValidateBasic() error         { return nil }
func (tx MockTx) GetSigners() []sdk.AccAddress { return nil }

func TestCosmosMsgsFromMsgEthereumTx(t *testing.T) {
	// Initialize interface registry with precise registration order
	interfaceRegistry := types.NewInterfaceRegistry()

	// 1. Register base interfaces first
	interfaceRegistry.RegisterInterface(
		sdk.MsgInterfaceProtoName, // "cosmos.base.v1beta1.Msg"
		(*sdk.Msg)(nil),
	)

	// 2. Register TxData interface before its implementations
	interfaceRegistry.RegisterInterface(
		"ethermint.evm.v1.TxData",
		(*TxData)(nil),
	)

	// 3. Register concrete implementations in deterministic order
	interfaceRegistry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgEthereumTx{},
	)

	interfaceRegistry.RegisterImplementations(
		(*TxData)(nil),
		&LegacyTx{},
	)

	// 4. Register module-specific interfaces last
	banktypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)

	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Test setup...
	sender := sdk.AccAddress("sender_address_____")
	recipient := sdk.AccAddress("recipient_address_")
	validator := sdk.ValAddress("validator_address_")

	t.Run("successful conversion with multiple Cosmos messages", func(t *testing.T) {
		// Create messages with validation
		msgSend := &banktypes.MsgSend{
			FromAddress: sender.String(),
			ToAddress:   recipient.String(),
			Amount: sdk.NewCoins(
				sdk.NewCoin("atom", math.NewInt(1000000)),
			),
		}
		//err := msgSend.ValidateBasic()
		//require.NoError(t, err)

		msgDelegate := &stakingtypes.MsgDelegate{
			DelegatorAddress: sender.String(),
			ValidatorAddress: validator.String(),
			Amount:           sdk.NewCoin("atom", math.NewInt(500000)),
		}
		//err = msgDelegate.ValidateBasic()
		//require.NoError(t, err)

		// Pack messages with explicit type verification
		msgSendAny, err := types.NewAnyWithValue(msgSend)
		require.NoError(t, err)
		require.Equal(t, "/cosmos.bank.v1beta1.MsgSend", msgSendAny.TypeUrl)

		msgDelegateAny, err := types.NewAnyWithValue(msgDelegate)
		require.NoError(t, err)
		require.Equal(t, "/cosmos.staking.v1beta1.MsgDelegate", msgDelegateAny.TypeUrl)

		// Create and validate InnerCosmosMsgs
		innerMsgs := types2.InnerCosmosMsgs{
			Msgs: []*types.Any{msgSendAny, msgDelegateAny},
		}

		// Marshal with type verification
		innerMsgsBytes, err := cdc.Marshal(&innerMsgs)
		require.NoError(t, err)

		// Verify round-trip marshaling
		var verifyInnerMsgs types2.InnerCosmosMsgs
		err = cdc.Unmarshal(innerMsgsBytes, &verifyInnerMsgs)
		require.NoError(t, err)
		require.Equal(t, len(innerMsgs.Msgs), len(verifyInnerMsgs.Msgs))

		// Create LegacyTx
		legacyTx := &LegacyTx{
			Nonce:    1,
			GasPrice: nil,
			GasLimit: 200000,
			To:       "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			Amount:   nil,
			Data:     innerMsgsBytes,
			V:        []byte{1},
			R:        []byte{2},
			S:        []byte{3},
		}

		// Pack LegacyTx with type verification
		legacyTxAny, err := types.NewAnyWithValue(legacyTx)
		require.NoError(t, err)
		require.Contains(t, legacyTxAny.TypeUrl, "LegacyTx")

		// Create and verify MsgEthereumTx
		ethTx := &MsgEthereumTx{
			Data: legacyTxAny,
			Hash: "0x123",
			From: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		}
		//err = ethTx.ValidateBasic()
		require.NoError(t, err)

		// Execute test
		tx := &MockTx{msgs: []sdk.Msg{ethTx}}
		cosmosMsgs, err := CosmosMsgsFromMsgEthereumTx(tx, *cdc)
		require.NoError(t, err)
		require.Len(t, cosmosMsgs, 2)

		// Verify results
		msgSendResult, ok := cosmosMsgs[0].(*banktypes.MsgSend)
		require.True(t, ok)
		require.Equal(t, sender.String(), msgSendResult.FromAddress)
		require.Equal(t, recipient.String(), msgSendResult.ToAddress)

		msgDelegateResult, ok := cosmosMsgs[1].(*stakingtypes.MsgDelegate)
		require.True(t, ok)
		require.Equal(t, sender.String(), msgDelegateResult.DelegatorAddress)
		require.Equal(t, validator.String(), msgDelegateResult.ValidatorAddress)
	})

	t.Run("error with invalid message type in inner messages", func(t *testing.T) {
		// Create an Any with an unregistered message type
		invalidAny := &types.Any{
			TypeUrl: "/cosmos.unknown.v1beta1.MsgUnknown",
			Value:   []byte("invalid message"),
		}

		innerMsgs := types2.InnerCosmosMsgs{
			Msgs: []*types.Any{invalidAny},
		}

		innerMsgsBytes, err := cdc.Marshal(&innerMsgs)
		require.NoError(t, err)

		legacyTx := &LegacyTx{
			Data: innerMsgsBytes,
		}
		legacyTxAny, err := types.NewAnyWithValue(legacyTx)
		require.NoError(t, err)

		ethTx := &MsgEthereumTx{
			Data: legacyTxAny,
		}
		tx := &MockTx{msgs: []sdk.Msg{ethTx}}

		_, err = CosmosMsgsFromMsgEthereumTx(tx, *cdc)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no concrete type registered")
	})

	t.Run("error with zero inner messages", func(t *testing.T) {
		innerMsgs := types2.InnerCosmosMsgs{
			Msgs: []*types.Any{},
		}

		innerMsgsBytes, err := cdc.Marshal(&innerMsgs)
		require.NoError(t, err)

		legacyTx := &LegacyTx{
			Data: innerMsgsBytes,
		}
		legacyTxAny, err := types.NewAnyWithValue(legacyTx)
		require.NoError(t, err)

		ethTx := &MsgEthereumTx{
			Data: legacyTxAny,
		}
		tx := &MockTx{msgs: []sdk.Msg{ethTx}}

		cosmosMsgs, err := CosmosMsgsFromMsgEthereumTx(tx, *cdc)
		require.NoError(t, err)
		require.Len(t, cosmosMsgs, 0)
	})
}

func TestUnmarshalInnerMessages(t *testing.T) {
	// Setup interface registry with explicit unpacking hooks
	interfaceRegistry := types.NewInterfaceRegistry()

	// Register unpacking interface
	interfaceRegistry.RegisterInterface(
		sdk.MsgInterfaceProtoName,
		(*sdk.Msg)(nil),
		&banktypes.MsgSend{},
	)

	// Register concrete type mappings
	interfaceRegistry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&banktypes.MsgSend{},
	)

	banktypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)

	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create and marshal original message
	originalMsg := &banktypes.MsgSend{
		FromAddress: "cosmos1sender",
		ToAddress:   "cosmos1recipient",
		Amount:      sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(1000))),
	}

	// Create Any with explicit value
	msgAny, err := types.NewAnyWithValue(originalMsg)
	require.NoError(t, err)
	require.NotNil(t, msgAny.GetCachedValue(), "Initial Any should have cached value")

	// Create InnerCosmosMsgs
	innerMsgs := types2.InnerCosmosMsgs{
		Msgs: []*types.Any{msgAny},
	}

	// Marshal to bytes
	bz, err := cdc.Marshal(&innerMsgs)
	require.NoError(t, err)

	// Now unmarshal back - this is where the issue occurs
	var unpackedMsgs types2.InnerCosmosMsgs
	err = cdc.Unmarshal(bz, &unpackedMsgs)
	require.NoError(t, err)

	// Check the cached value immediately after unmarshal
	require.Len(t, unpackedMsgs.Msgs, 1)
	cachedValue := unpackedMsgs.Msgs[0].GetCachedValue()
	t.Logf("Cached value after unmarshal: %v", cachedValue)

	// Attempt explicit unpacking
	var unpacked sdk.Msg
	err = interfaceRegistry.UnpackAny(unpackedMsgs.Msgs[0], &unpacked)
	require.NoError(t, err)
	t.Logf("Unpacked message: %v", unpacked)
}
