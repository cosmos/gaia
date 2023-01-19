package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// nolint: exhaustruct
var (
	_ sdk.Msg = &MsgName{}
)

// NewMsgName returns a new MsgName
func NewMsgName(name string) *MsgName {
	return &MsgName{
		Name: name,
	}
}

// Route should return the name of the module
func (msg *MsgName) Route() string { return RouterKey }

// ValidateBasic performs stateless checks
func (msg *MsgName) ValidateBasic() (err error) {
	if msg.Name == "bob" {
		return fmt.Errorf("Bob get outta here, I'm trying to program")
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgName) GetSigners() []sdk.AccAddress {
	// TODO: get all the users who must sign the message for the blockchain to be
	// convinced that they all gave consent for state to be updated

	// acc, err := sdk.AccAddressFromBech32(msg.Signer)
	// if err != nil {
	// 	panic(err)
	// }
	return []sdk.AccAddress{}
}
