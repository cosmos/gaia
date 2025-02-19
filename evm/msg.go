package evm

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	txsigning "cosmossdk.io/x/tx/signing"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	evmv1 "github.com/cosmos/gaia/v23/api/ethermint/evm/v1"
	"github.com/cosmos/gaia/v23/types"
	gaiaerrors "github.com/cosmos/gaia/v23/types/errors"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	protov2 "google.golang.org/protobuf/proto"
	"math/big"
)

var MsgEthereumTxCustomGetSigner = txsigning.CustomGetSigner{
	MsgType: protov2.MessageName(&evmv1.MsgEthereumTx{}),
	Fn:      evmv1.GetSigners,
}

// AsTransaction creates an Ethereum Transaction type from the msg fields
func (msg MsgEthereumTx) AsTransaction() *ethtypes.Transaction {
	txData, err := UnpackTxData(msg.Data)
	if err != nil {
		return nil
	}

	return ethtypes.NewTx(txData.AsEthereumData())
}

// FromEthereumTx populates the message fields from the given ethereum transaction
func (msg *MsgEthereumTx) FromEthereumTx(tx *ethtypes.Transaction) error {
	txData, err := NewTxDataFromTx(tx)
	if err != nil {
		return err
	}

	anyTxData, err := PackTxData(txData)
	if err != nil {
		return err
	}

	msg.Data = anyTxData
	msg.Hash = tx.Hash().Hex()
	return nil
}

// ValidateBasic implements the sdk.Msg interface. It performs basic validation
// checks of a Transaction. If returns an error if validation fails.
func (msg MsgEthereumTx) ValidateBasic() error {
	if msg.From != "" {
		if err := types.ValidateAddress(msg.From); err != nil {
			return errorsmod.Wrap(err, "invalid from address")
		}
	}

	// Validate Size_ field, should be kept empty
	if msg.Size_ != 0 {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "tx size is deprecated")
	}

	txData, err := UnpackTxData(msg.Data)
	if err != nil {
		return errorsmod.Wrap(err, "failed to unpack tx data")
	}

	gas := txData.GetGas()

	// prevent txs with 0 gas to fill up the mempool
	if gas == 0 {
		return errorsmod.Wrap(gaiaerrors.ErrInvalidGasLimit, "gas limit must not be zero")
	}

	// prevent gas limit from overflow
	if g := new(big.Int).SetUint64(gas); !g.IsInt64() {
		return errorsmod.Wrap(gaiaerrors.ErrGasOverflow, "gas limit must be less than math.MaxInt64")
	}

	if err := txData.Validate(); err != nil {
		return err
	}

	// Validate Hash field after validated txData to avoid panic
	txHash := msg.AsTransaction().Hash().Hex()
	if msg.Hash != txHash {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "invalid tx hash %s, expected: %s", msg.Hash, txHash)
	}

	return nil
}

// BuildTx builds the canonical cosmos tx from ethereum msg
func (msg *MsgEthereumTx) BuildTx(b client.TxBuilder, evmDenom string) (signing.Tx, error) {
	builder, ok := b.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, errors.New("unsupported builder")
	}

	option, err := codectypes.NewAnyWithValue(&ExtensionOptionsEthereumTx{})
	if err != nil {
		return nil, err
	}

	txData, err := UnpackTxData(msg.Data)
	if err != nil {
		return nil, err
	}
	fees := make(sdk.Coins, 0)
	feeAmt := sdkmath.NewIntFromBigInt(txData.Fee())
	if feeAmt.Sign() > 0 {
		fees = append(fees, sdk.NewCoin(evmDenom, feeAmt))
	}

	builder.SetExtensionOptions(option)

	// A valid msg should have empty `From`
	msg.From = ""

	err = builder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(msg.GetGas())
	tx := builder.GetTx()
	return tx, nil
}

// todo: test this
func CosmosMsgsFromMsgEthereumTx(tx sdk.Tx, cdc codec.Codec) ([]sdk.Msg, error) {
	txMsgs := tx.GetMsgs()
	ethTxMsgs := make([]*MsgEthereumTx, len(txMsgs))
	for i, msg := range txMsgs {
		ethTxMsgs[i] = msg.(*MsgEthereumTx)
	}

	//already validated that there is only a single message of type MsgEthereumTx, so now we unpack the TxData from the message
	var txData TxData
	txData, err := UnpackTxData(ethTxMsgs[0].Data)
	if err != nil {
		return nil, err
	}

	rawInnerMsgs := txData.GetData()

	// Pre-allocate our InnerCosmosMsgs struct
	var innerAnyMsgs types.InnerCosmosMsgs
	err = cdc.Unmarshal(rawInnerMsgs, &innerAnyMsgs)
	if err != nil {
		return nil, err
	}

	// Initialize our result slice
	innerCosmosMsgs := make([]sdk.Msg, len(innerAnyMsgs.Msgs))

	// Process each message individually
	for i, anyMsg := range innerAnyMsgs.Msgs {
		// Safety check for nil message
		if anyMsg == nil {
			return nil, fmt.Errorf("message at index %d is nil", i)
		}

		var unpacked sdk.Msg

		if err = cdc.InterfaceRegistry().UnpackAny(anyMsg, &unpacked); err != nil { //todo: figure out why cache is nil when using binary codec unmarshalling
			return nil, err
		}

		innerCosmosMsgs[i] = unpacked
	}
	return innerCosmosMsgs, nil
}
