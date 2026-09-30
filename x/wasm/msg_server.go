package wasm

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	gaiaerrors "github.com/cosmos/gaia/v28/types/errors"
)

// msgServer wraps the wasmd MsgServer to temporarily block MsgStoreCode,
// MsgInstantiateContract, MsgInstantiateContract2,
// MsgStoreAndInstantiateContract, and MsgStoreAndMigrateContract regardless
// of how they were dispatched (direct tx, ICA, authz, wasm-emitted
// submessages, gov-executed proposals, ...) — baseapp always routes message
// execution through the module's registered MsgServer, so blocking here is
// airtight across every path, unlike ante/wasm_disable_ante.go's ante-layer
// decorator, which only sees messages that are top-level (or
// authz.MsgExec-wrapped) in a directly broadcast tx. That ante decorator is
// kept as a cheap fail-fast rejection for the common direct-tx case
// (rejects at CheckTx, before any gas is spent); this msg-server layer is
// the authoritative backstop covering every other dispatch path. See
// ante/wasm_disable_ante.go for the full rationale.
//
// MsgStoreAndInstantiateContract's and MsgStoreAndMigrateContract's proto
// doc comments call them "governance operations," but their handlers
// (wasmd's x/wasm/keeper/msg_server.go, StoreAndInstantiateContract and
// StoreAndMigrateContract) never check req.Authority against the module's
// gov authority the way e.g. AddCodeUploadParamsAddresses does — any
// bech32 address falls through to the same DefaultAuthorizationPolicy a
// normal MsgStoreCode gets.
type msgServer struct {
	wasmtypes.MsgServer
}

var _ wasmtypes.MsgServer = &msgServer{}

// NewMsgServerImpl returns an implementation of the wasm MsgServer interface
// that blocks the store/instantiate family of messages, delegating
// everything else to the wrapped implementation.
func NewMsgServerImpl(keeper wasmtypes.MsgServer) wasmtypes.MsgServer {
	return &msgServer{MsgServer: keeper}
}

func (m *msgServer) StoreCode(_ context.Context, _ *wasmtypes.MsgStoreCode) (*wasmtypes.MsgStoreCodeResponse, error) {
	return nil, errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "MsgStoreCode is temporarily disabled")
}

func (m *msgServer) InstantiateContract(_ context.Context, _ *wasmtypes.MsgInstantiateContract) (*wasmtypes.MsgInstantiateContractResponse, error) {
	return nil, errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "MsgInstantiateContract is temporarily disabled")
}

func (m *msgServer) InstantiateContract2(_ context.Context, _ *wasmtypes.MsgInstantiateContract2) (*wasmtypes.MsgInstantiateContract2Response, error) {
	return nil, errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "MsgInstantiateContract2 is temporarily disabled")
}

func (m *msgServer) StoreAndInstantiateContract(_ context.Context, _ *wasmtypes.MsgStoreAndInstantiateContract) (*wasmtypes.MsgStoreAndInstantiateContractResponse, error) {
	return nil, errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "MsgStoreAndInstantiateContract is temporarily disabled")
}

func (m *msgServer) StoreAndMigrateContract(_ context.Context, _ *wasmtypes.MsgStoreAndMigrateContract) (*wasmtypes.MsgStoreAndMigrateContractResponse, error) {
	return nil, errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "MsgStoreAndMigrateContract is temporarily disabled")
}
