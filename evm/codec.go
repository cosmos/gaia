package evm

import (
	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
)

// UnpackTxData unpacks an Any into a TxData. It returns an error if the
// client state can't be unpacked into a TxData.
func UnpackTxData(any *codectypes.Any) (TxData, error) {
	if any == nil {
		return nil, errorsmod.Wrap(errortypes.ErrUnpackAny, "protobuf Any message cannot be nil")
	}

	// First try to use cached value
	if cachedValue := any.GetCachedValue(); cachedValue != nil { //todo: figure out why cached value was nil (improperly registered interfaces?)
		txData, ok := cachedValue.(TxData)
		if !ok {
			return nil, errorsmod.Wrapf(errortypes.ErrUnpackAny, "cannot cast cached value to TxData: %T", cachedValue)
		}
		return txData, nil
	}

	// If no cached value, unpack based on TypeUrl
	switch any.TypeUrl {
	case "/ethermint.evm.v1.LegacyTx":
		var legacyTx LegacyTx
		err := proto.Unmarshal(any.Value, &legacyTx)
		if err != nil {
			return nil, errorsmod.Wrapf(errortypes.ErrUnpackAny, "failed to unmarshal LegacyTx: %v", err)
		}
		return &legacyTx, nil

	case "/ethermint.evm.v1.AccessListTx":
		var accessListTx AccessListTx
		err := proto.Unmarshal(any.Value, &accessListTx)
		if err != nil {
			return nil, errorsmod.Wrapf(errortypes.ErrUnpackAny, "failed to unmarshal AccessListTx: %v", err)
		}
		return &accessListTx, nil

	case "/ethermint.evm.v1.DynamicFeeTx":
		var dynamicFeeTx DynamicFeeTx
		err := proto.Unmarshal(any.Value, &dynamicFeeTx)
		if err != nil {
			return nil, errorsmod.Wrapf(errortypes.ErrUnpackAny, "failed to unmarshal DynamicFeeTx: %v", err)
		}
		return &dynamicFeeTx, nil

	default:
		return nil, errorsmod.Wrapf(errortypes.ErrUnpackAny, "unsupported TypeUrl: %s", any.TypeUrl)
	}
}

// PackTxData constructs a new Any packed with the given tx data value. It returns
// an error if the client state can't be cast to a protobuf message or if the concrete
// implementation is not registered to the protobuf codec.
func PackTxData(txData TxData) (*codectypes.Any, error) {
	msg, ok := txData.(proto.Message)
	if !ok {
		return nil, errorsmod.Wrapf(errortypes.ErrPackAny, "cannot proto marshal %T", txData)
	}

	anyTxData, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, errorsmod.Wrap(errortypes.ErrPackAny, err.Error())
	}

	return anyTxData, nil
}
