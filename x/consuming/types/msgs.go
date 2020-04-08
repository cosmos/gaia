package types

import (
	bandoracle "github.com/bandprotocol/chain/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgRequestData creates a new MsgRequestData instance.
func NewMsgRequestData(
	oracleScriptID bandoracle.OracleScriptID,
	sourceChannel string,
	calldata []byte,
	askCount int64,
	minCount int64,
	sender sdk.AccAddress,
) *MsgRequestData {
	return &MsgRequestData{
		OracleScriptID: int64(oracleScriptID),
		SourceChannel:  sourceChannel,
		Calldata:       calldata,
		AskCount:       askCount,
		MinCount:       minCount,
		Sender:         sender.String(),
	}
}

// Route implements the sdk.Msg interface for MsgRequestData.
func (msg MsgRequestData) Route() string { return RouterKey }

// Type implements the sdk.Msg interface for MsgRequestData.
func (msg MsgRequestData) Type() string { return "consuming" }

// ValidateBasic implements the sdk.Msg interface for MsgRequestData.
func (msg MsgRequestData) ValidateBasic() error {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	if sender.Empty() {
		return sdkerrors.Wrapf(ErrInvalidBasicMsg, "MsgRequestData: Sender address must not be empty.")
	}
	if msg.OracleScriptID <= 0 {
		return sdkerrors.Wrapf(ErrInvalidBasicMsg, "MsgRequestData: Oracle script id (%d) must be positive.", msg.OracleScriptID)
	}
	if msg.AskCount <= 0 {
		return sdkerrors.Wrapf(ErrInvalidBasicMsg,
			"MsgRequestData: Sufficient validator count (%d) must be positive.",
			msg.AskCount,
		)
	}
	if msg.AskCount < msg.MinCount {
		return sdkerrors.Wrapf(ErrInvalidBasicMsg,
			"MsgRequestData: Request validator count (%d) must not be less than sufficient validator count (%d).",
			msg.AskCount,
			msg.MinCount,
		)
	}
	return nil
}

// GetSigners implements the sdk.Msg interface for MsgRequestData.
func (msg MsgRequestData) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// GetSignBytes implements the sdk.Msg interface for MsgRequestData.
func (msg MsgRequestData) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}
