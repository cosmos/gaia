package encoding_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gaia/v23/encoding"
	"testing"

	evmtypes "github.com/cosmos/gaia/v23/evm"
)

func TestTxEncoding(t *testing.T) {
	config := encoding.MakeConfig()

	legacyTx := &evmtypes.LegacyTx{
		Nonce:    0,
		GasPrice: nil,
		GasLimit: 21000,
	}

	anyData, err := types.NewAnyWithValue(legacyTx)
	if err != nil {
		t.Fatalf("Error packing: %v", err)
	}
	fmt.Printf("Type URL: %s\n", anyData.TypeUrl)

	msg := &evmtypes.MsgEthereumTx{
		Data: anyData,
	}

	// Try to unpack the Any field
	var txData evmtypes.TxData
	err = config.Codec.UnpackAny(msg.Data, &txData)
	if err != nil {
		fmt.Printf("Unpack error: %v\n", err)
	} else {
		fmt.Printf("Successfully unpacked to: %T\n", txData)
	}
}
