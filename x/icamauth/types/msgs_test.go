package types

import (
	"errors"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestMsgRegisterAccount_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgRegisterAccount
		err  error
	}{
		{
			name: "empty owner",
			msg: MsgRegisterAccount{
				Owner: "",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid bech32 owner address",
			msg: MsgRegisterAccount{
				Owner: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid mesage",
			msg: MsgRegisterAccount{
				Owner: "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, tt.err, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPackTxMsgAny(t *testing.T) {
	tests := []struct {
		name   string
		sdkMsg types.Msg
		want   *codectypes.Any
		err    error
	}{
		{
			name:   "empty bank send message",
			sdkMsg: &banktypes.MsgSend{},
			want: &codectypes.Any{
				TypeUrl: "/cosmos.bank.v1beta1.MsgSend",
				Value:   []byte{},
			},
			err: nil,
		}, {
			name: "bank send message",
			sdkMsg: &banktypes.MsgSend{
				FromAddress: "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
				ToAddress:   "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
				Amount:      sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(10))),
			},
			want: &codectypes.Any{
				TypeUrl: "/cosmos.bank.v1beta1.MsgSend",
				Value:   []byte{0xa, 0x2d, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x31, 0x61, 0x36, 0x7a, 0x6c, 0x79, 0x76, 0x70, 0x6e, 0x6b, 0x73, 0x78, 0x38, 0x77, 0x72, 0x36, 0x77, 0x7a, 0x38, 0x77, 0x65, 0x6d, 0x75, 0x72, 0x32, 0x78, 0x65, 0x38, 0x7a, 0x79, 0x68, 0x30, 0x79, 0x78, 0x65, 0x68, 0x32, 0x37, 0x61, 0x12, 0x2d, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x31, 0x61, 0x36, 0x7a, 0x6c, 0x79, 0x76, 0x70, 0x6e, 0x6b, 0x73, 0x78, 0x38, 0x77, 0x72, 0x36, 0x77, 0x7a, 0x38, 0x77, 0x65, 0x6d, 0x75, 0x72, 0x32, 0x78, 0x65, 0x38, 0x7a, 0x79, 0x68, 0x30, 0x79, 0x78, 0x65, 0x68, 0x32, 0x37, 0x61, 0x1a, 0xa, 0xa, 0x4, 0x61, 0x74, 0x6f, 0x6d, 0x12, 0x2, 0x31, 0x30},
			},
			err: nil,
		}, {
			name:   "nil message",
			sdkMsg: nil,
			err:    sdkerrors.ErrPackAny,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PackTxMsgAny(tt.sdkMsg)
			if tt.err != nil {
				require.ErrorIs(t, tt.err, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want.TypeUrl, got.TypeUrl)
			require.Equal(t, tt.want.Value, got.Value)
		})
	}
}

func TestMsgSubmitTx_GetSigners(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSubmitTx
		want []sdk.AccAddress
		err  error
	}{
		{
			name: "valid signers",
			msg: MsgSubmitTx{
				Owner: "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
			},
			want: []types.AccAddress{[]byte{238, 133, 242, 48, 51, 180, 12, 119, 15, 78, 17, 221, 157, 240, 106, 54, 78, 34, 93, 228}},
		}, {
			name: "empty address",
			msg: MsgSubmitTx{
				Owner: "",
			},
			err: errors.New("empty address string is not allowed"),
		}, {
			name: "invalid address",
			msg: MsgSubmitTx{
				Owner: "invalid_address",
			},
			err: errors.New("decoding bech32 failed: invalid separator index -1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				require.PanicsWithError(t, tt.err.Error(), func() {
					_ = tt.msg.GetSigners()
				})
				return
			}
			require.Equal(t, tt.want, tt.msg.GetSigners())
		})
	}
}

func TestMsgSubmitTx_GetTxMsg(t *testing.T) {
	t.Run("bank message", func(t *testing.T) {
		msg := &banktypes.MsgSend{
			FromAddress: "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
			ToAddress:   "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
			Amount:      sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(10))),
		}
		anyMsg, err := PackTxMsgAny(msg)
		require.NoError(t, err)
		submitTx := MsgSubmitTx{Msg: anyMsg}
		require.Equal(t, msg, submitTx.GetTxMsg())
	})

	t.Run("staking message", func(t *testing.T) {
		msg := &stakingtypes.MsgDelegate{
			DelegatorAddress: "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
			ValidatorAddress: "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
			Amount:           sdk.NewCoin("atom", sdk.NewInt(10)),
		}
		anyMsg, err := PackTxMsgAny(msg)
		require.NoError(t, err)
		submitTx := MsgSubmitTx{Msg: anyMsg}
		require.Equal(t, msg, submitTx.GetTxMsg())
	})
}

func TestMsgSubmitTx_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSubmitTx
		err  error
	}{
		{
			name: "empty owner",
			msg: MsgSubmitTx{
				Owner: "",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid bech32 owner address",
			msg: MsgSubmitTx{
				Owner: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid mesage",
			msg: MsgSubmitTx{
				Owner: "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, tt.err, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestNewMsgRegisterAccount(t *testing.T) {
	var (
		connectionID = "connection-0"
		owner        = "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a"
		version      = "1"
		got          = NewMsgRegisterAccount(owner, connectionID, version)
	)
	require.Equal(t, connectionID, got.ConnectionId)
	require.Equal(t, owner, got.Owner)
	require.Equal(t, version, got.Version)
}

func TestNewMsgSubmitTx(t *testing.T) {
	var (
		connectionID = "connection-0"
		owner        = "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a"
	)

	t.Run("valid message", func(t *testing.T) {
		msg := &banktypes.MsgSend{
			FromAddress: "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
			ToAddress:   "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
			Amount:      sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(10))),
		}
		got, err := NewMsgSubmitTx(msg, connectionID, owner)
		require.NoError(t, err)

		anyMsg, err := PackTxMsgAny(msg)
		require.NoError(t, err)

		want := &MsgSubmitTx{
			Owner:        owner,
			ConnectionId: connectionID,
			Msg:          anyMsg,
		}
		require.EqualValues(t, want, got)
	})

	t.Run("invalid message field", func(t *testing.T) {
		_, err := NewMsgSubmitTx(nil, connectionID, owner)
		require.ErrorIs(t, sdkerrors.ErrPackAny, err)
	})
}
