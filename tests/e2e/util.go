package e2e

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec/unknownproto"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

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

type Summary struct {
	Src PendingPackets `json:"src"`
	Dst PendingPackets `json:"dst"`
}

type PendingPackets struct {
	UnreceivedPackets []Collated `json:"unreceived_packets"`
	UnreceivedAcks    []Collated `json:"unreceived_acks"`
}

type Collated struct {
	Start Sequence `json:"start"`
	End   Sequence `json:"end"`
}

type Sequence struct {
	Value int `json:"value"`
}

func parsePendingPacketResult(output string) ([]Collated, error) {
	var summary Summary
	var res string
	index := strings.Index(output, "SUCCESS")
	if index != -1 {
		res = output[index:]
	} else {
		return []Collated{}, fmt.Errorf("unexpected query pending packet result")
	}

	err := json.Unmarshal([]byte(res), &summary)
	if err != nil {
		return []Collated{}, fmt.Errorf("Error parsing  query pending packet result %v\n:", err)
	}

	return summary.Src.UnreceivedPackets, nil
}
