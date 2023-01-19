package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrErrir = sdkerrors.Register(ModuleName, 1, "errir")
)
