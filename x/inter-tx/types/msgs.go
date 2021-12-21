package types

import (
	fmt "fmt"
	"strings"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	proto "github.com/gogo/protobuf/proto"
)

var (
	_ sdk.Msg = &MsgDelegate{}
	_ sdk.Msg = &MsgRegisterAccount{}
	_ sdk.Msg = &MsgSend{}
	_ sdk.Msg = &MsgSubmitTx{}

	_ codectypes.UnpackInterfacesMessage = MsgSubmitTx{}
)

// NewMsgDelegate creates a new MsgDelegate instance
func NewMsgDelegate(owner sdk.AccAddress, amt sdk.Coin, interchainAccAddr, validatorAddr, connectionID, counterpartyConnectionID string) *MsgDelegate {
	return &MsgDelegate{
		InterchainAccount:        interchainAccAddr,
		Owner:                    owner,
		ValidatorAddress:         validatorAddr,
		Amount:                   amt,
		ConnectionId:             connectionID,
		CounterpartyConnectionId: counterpartyConnectionID,
	}
}

// GetSigners implements sdk.Msg
func (msg MsgDelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

// ValidateBasic implements sdk.Msg
func (msg MsgDelegate) ValidateBasic() error {
	if strings.TrimSpace(msg.InterchainAccount) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}

	if strings.TrimSpace(msg.ValidatorAddress) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing validator address")
	}

	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	_, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid bech32 validator address: %s", msg.ValidatorAddress)
	}

	return nil
}

// NewMsgRegisterAccount creates a new MsgRegisterAccount instance
func NewMsgRegisterAccount(owner, connectionID, counterpartyConnectionID string) *MsgRegisterAccount {
	return &MsgRegisterAccount{
		Owner:                    owner,
		ConnectionId:             connectionID,
		CounterpartyConnectionId: counterpartyConnectionID,
	}
}

// ValidateBasic implements sdk.Msg
func (msg MsgRegisterAccount) ValidateBasic() error {
	if strings.TrimSpace(msg.Owner) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgRegisterAccount) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{accAddr}
}

// NewMsgSend creates a new MsgSend instance
func NewMsgSend(owner sdk.AccAddress, amt sdk.Coins, interchainAccAddr, toAddr, connectionID, counterpartyConnectionID string) *MsgSend {
	return &MsgSend{
		InterchainAccount:        interchainAccAddr,
		Owner:                    owner,
		ToAddress:                toAddr,
		Amount:                   amt,
		ConnectionId:             connectionID,
		CounterpartyConnectionId: counterpartyConnectionID,
	}
}

// GetSigners implements sdk.Msg
func (msg MsgSend) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

// ValidateBasic implements sdk.Msg
func (msg MsgSend) ValidateBasic() error {
	if strings.TrimSpace(msg.InterchainAccount) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}

	if strings.TrimSpace(msg.ToAddress) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing recipient address")
	}

	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	return nil
}

// NewMsgSend creates a new MsgSend instance
func NewMsgSubmitTx(owner sdk.AccAddress, sdkMsg sdk.Msg, connectionID, counterpartyConnectionID string) (*MsgSubmitTx, error) {
	any, err := PackTxMsgAny(sdkMsg)
	if err != nil {
		return nil, err
	}

	return &MsgSubmitTx{
		Owner:                    owner,
		ConnectionId:             connectionID,
		CounterpartyConnectionId: counterpartyConnectionID,
		Msg:                      any,
	}, nil
}

// PackTxMsgAny marshals the sdk.Msg payload to a protobuf Any type
func PackTxMsgAny(sdkMsg sdk.Msg) (*codectypes.Any, error) {
	msg, ok := sdkMsg.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't proto marshal %T", sdkMsg)
	}

	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return any, nil
}

// UnpackInterfaces implements codectypes.UnpackInterfacesMessage
func (msg MsgSubmitTx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var (
		sdkMsg sdk.Msg
	)

	return unpacker.UnpackAny(msg.Msg, &sdkMsg)
}

// GetTxMsg fetches the cached any message
func (msg *MsgSubmitTx) GetTxMsg() sdk.Msg {
	sdkMsg, ok := msg.Msg.GetCachedValue().(sdk.Msg)
	if !ok {
		return nil
	}

	return sdkMsg
}

// GetSigners implements sdk.Msg
func (msg MsgSubmitTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

// ValidateBasic implements sdk.Msg
func (msg MsgSubmitTx) ValidateBasic() error {

	if len(msg.Msg.GetValue()) == 0 {
		return fmt.Errorf("can't execute an empty msg")
	}

	if msg.ConnectionId == "" {
		return fmt.Errorf("can't execute an empty ConnectionId")
	}

	if msg.CounterpartyConnectionId == "" {
		return fmt.Errorf("can't execute an empty CounterpartyConnectionId")
	}

	return nil
}
