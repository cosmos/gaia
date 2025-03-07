package e2e

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec/unknownproto"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

const (
	flagFrom            = "from"
	flagHome            = "home"
	flagFees            = "fees"
	flagGas             = "gas"
	flagOutput          = "output"
	flagChainID         = "chain-id"
	flagSpendLimit      = "spend-limit"
	flagGasAdjustment   = "gas-adjustment"
	flagFeeGranter      = "fee-granter"
	flagBroadcastMode   = "broadcast-mode"
	flagKeyringBackend  = "keyring-backend"
	flagAllowedMessages = "allowed-messages"
)

type flagOption func(map[string]interface{})

// withKeyValue add a new flag to command

func withKeyValue(key string, value interface{}) flagOption {
	return func(o map[string]interface{}) {
		o[key] = value
	}
}

func decodeTx(txBytes []byte) (*sdktx.Tx, error) {
	var raw sdktx.TxRaw

	// reject all unknown proto fields in the root TxRaw
	err := unknownproto.RejectUnknownFieldsStrict(txBytes, &raw, encodingConfig.InterfaceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := cdc.Unmarshal(txBytes, &raw); err != nil {
		return nil, err
	}

	var body sdktx.TxBody
	if err := cdc.Unmarshal(raw.BodyBytes, &body); err != nil {
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	var authInfo sdktx.AuthInfo

	// reject all unknown proto fields in AuthInfo
	err = unknownproto.RejectUnknownFieldsStrict(raw.AuthInfoBytes, &authInfo, encodingConfig.InterfaceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := cdc.Unmarshal(raw.AuthInfoBytes, &authInfo); err != nil {
		return nil, fmt.Errorf("failed to decode auth info: %w", err)
	}

	return &sdktx.Tx{
		Body:       &body,
		AuthInfo:   &authInfo,
		Signatures: raw.Signatures,
	}, nil
}

func concatFlags(originalCollection []string, commandFlags []string, generalFlags []string) []string {
	originalCollection = append(originalCollection, commandFlags...)
	originalCollection = append(originalCollection, generalFlags...)

	return originalCollection
}

func applyOptions(chainID string, options []flagOption) map[string]interface{} {
	opts := map[string]interface{}{
		flagKeyringBackend: "test",
		flagOutput:         "json",
		flagGas:            "auto",
		flagFrom:           "alice",
		flagBroadcastMode:  "sync",
		flagGasAdjustment:  "1.5",
		flagChainID:        chainID,
		flagHome:           gaiaHomePath,
		flagFees:           standardFees.String(),
	}
	for _, apply := range options {
		apply(opts)
	}
	return opts
}
