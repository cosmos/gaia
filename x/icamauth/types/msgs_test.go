package types

import (
	"errors"
	"reflect"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
			name: "",
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
	type fields struct {
		Owner        string
		ConnectionId string
		Msg          *codectypes.Any
	}
	tests := []struct {
		name   string
		fields fields
		want   types.Msg
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &MsgSubmitTx{
				Owner:        tt.fields.Owner,
				ConnectionId: tt.fields.ConnectionId,
				Msg:          tt.fields.Msg,
			}
			if got := msg.GetTxMsg(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTxMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMsgSubmitTx_UnpackInterfaces(t *testing.T) {
	type fields struct {
		Owner        string
		ConnectionId string
		Msg          *codectypes.Any
	}
	type args struct {
		unpacker codectypes.AnyUnpacker
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := MsgSubmitTx{
				Owner:        tt.fields.Owner,
				ConnectionId: tt.fields.ConnectionId,
				Msg:          tt.fields.Msg,
			}
			if err := msg.UnpackInterfaces(tt.args.unpacker); (err != nil) != tt.wantErr {
				t.Errorf("UnpackInterfaces() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgSubmitTx_ValidateBasic(t *testing.T) {
	type fields struct {
		Owner        string
		ConnectionId string
		Msg          *codectypes.Any
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := MsgSubmitTx{
				Owner:        tt.fields.Owner,
				ConnectionId: tt.fields.ConnectionId,
				Msg:          tt.fields.Msg,
			}
			if err := msg.ValidateBasic(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
