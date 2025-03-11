package common

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec/unknownproto"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

const (
	FlagFrom            = "from"
	flagHome            = "home"
	FlagFees            = "fees"
	flagGas             = "gas"
	flagOutput          = "output"
	flagChainID         = "chain-id"
	FlagSpendLimit      = "spend-limit"
	flagGasAdjustment   = "gas-adjustment"
	FlagFeeGranter      = "fee-granter"
	flagBroadcastMode   = "broadcast-mode"
	flagKeyringBackend  = "keyring-backend"
	FlagAllowedMessages = "allowed-messages"
)

type FlagOption func(map[string]interface{})

// withKeyValue add a new flag to command

func WithKeyValue(key string, value interface{}) FlagOption {
	return func(o map[string]interface{}) {
		o[key] = value
	}
}

func DecodeTx(txBytes []byte) (*sdktx.Tx, error) {
	var raw sdktx.TxRaw

	// reject all unknown proto fields in the root TxRaw
	err := unknownproto.RejectUnknownFieldsStrict(txBytes, &raw, EncodingConfig.InterfaceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := Cdc.Unmarshal(txBytes, &raw); err != nil {
		return nil, err
	}

	var body sdktx.TxBody
	if err := Cdc.Unmarshal(raw.BodyBytes, &body); err != nil {
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	var authInfo sdktx.AuthInfo

	// reject all unknown proto fields in AuthInfo
	err = unknownproto.RejectUnknownFieldsStrict(raw.AuthInfoBytes, &authInfo, EncodingConfig.InterfaceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := Cdc.Unmarshal(raw.AuthInfoBytes, &authInfo); err != nil {
		return nil, fmt.Errorf("failed to decode auth info: %w", err)
	}

	return &sdktx.Tx{
		Body:       &body,
		AuthInfo:   &authInfo,
		Signatures: raw.Signatures,
	}, nil
}

func ConcatFlags(originalCollection []string, commandFlags []string, generalFlags []string) []string {
	originalCollection = append(originalCollection, commandFlags...)
	originalCollection = append(originalCollection, generalFlags...)

	return originalCollection
}

func ApplyOptions(chainID string, options []FlagOption) map[string]interface{} {
	opts := map[string]interface{}{
		flagKeyringBackend: "test",
		flagOutput:         "json",
		flagGas:            "auto",
		FlagFrom:           "alice",
		flagBroadcastMode:  "sync",
		flagGasAdjustment:  "1.5",
		flagChainID:        chainID,
		flagHome:           GaiaHomePath,
		FlagFees:           StandardFees.String(),
	}
	for _, apply := range options {
		apply(opts)
	}
	return opts
}
