// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package flags

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	evmconfig "github.com/cosmos/gaia/v23/server/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Tendermint/cosmos-sdk full-node start flags
const (
	WithCometBFT = "with-cometbft"
	Address      = "address"
	Transport    = "transport"
	TraceStore   = "trace-store"
	CPUProfile   = "cpu-profile"
	// The type of database for application and snapshots databases
	AppDBBackend = "app-db-backend"
)

// GRPC-related flags.
const (
	GRPCOnly       = "grpc-only"
	GRPCEnable     = "grpc.enable"
	GRPCAddress    = "grpc.address"
	GRPCWebEnable  = "grpc-web.enable"
	GRPCWebAddress = "grpc-web.address"
)

// Cosmos API flags
const (
	RPCEnable         = "api.enable"
	EnabledUnsafeCors = "api.enabled-unsafe-cors"
)

// JSON-RPC flags
const (
	JSONRPCEnable              = "json-rpc.enable"
	JSONRPCAPI                 = "json-rpc.api"
	JSONRPCAddress             = "json-rpc.address"
	JSONWsAddress              = "json-rpc.ws-address"
	JSONRPCGasCap              = "json-rpc.gas-cap"
	JSONRPCAllowInsecureUnlock = "json-rpc.allow-insecure-unlock"
	JSONRPCEVMTimeout          = "json-rpc.evm-timeout"
	JSONRPCTxFeeCap            = "json-rpc.txfee-cap"
	JSONRPCFilterCap           = "json-rpc.filter-cap"
	JSONRPCLogsCap             = "json-rpc.logs-cap"
	JSONRPCBlockRangeCap       = "json-rpc.block-range-cap"
	JSONRPCHTTPTimeout         = "json-rpc.http-timeout"
	JSONRPCHTTPIdleTimeout     = "json-rpc.http-idle-timeout"
	JSONRPCAllowUnprotectedTxs = "json-rpc.allow-unprotected-txs"
	JSONRPCMaxOpenConnections  = "json-rpc.max-open-connections"
	// JSONRPCEnableMetrics enables EVM RPC metrics server.
	// Set to `metrics` which is hardcoded flag from go-ethereum.
	// https://github.com/ethereum/go-ethereum/blob/master/metrics/metrics.go#L35-L55
	JSONRPCEnableMetrics            = "metrics"
	JSONRPCFixRevertGasRefundHeight = "json-rpc.fix-revert-gas-refund-height"
)

// EVM flags
const (
	EVMTracer         = "evm.tracer"
	EVMMaxTxGasWanted = "evm.max-tx-gas-wanted"
)

// TLS flags
const (
	TLSCertPath = "tls.certificate-path"
	TLSKeyPath  = "tls.key-path"
)

// AddModuleInitFlags implements servertypes.ModuleInitFlags interface.
func AddModuleInitFlags(startCmd *cobra.Command) {
	startCmd.Flags().Bool(JSONRPCEnable, evmconfig.DefaultJSONRPCEnable, "Define if the JSON-RPC server should be enabled")
	startCmd.Flags().StringSlice(JSONRPCAPI, evmconfig.GetDefaultAPINamespaces(), "Defines a list of JSON-RPC namespaces that should be enabled")
	startCmd.Flags().String(JSONRPCAddress, evmconfig.DefaultJSONRPCAddress, "the JSON-RPC server address to listen on")
	startCmd.Flags().String(JSONWsAddress, evmconfig.DefaultJSONRPCWsAddress, "the JSON-RPC WS server address to listen on")
	startCmd.Flags().Uint64(JSONRPCGasCap, evmconfig.DefaultGasCap, "Sets a cap on gas that can be used in eth_call/estimateGas unit is aevmos (0=infinite)")                        //nolint:lll
	startCmd.Flags().Bool(JSONRPCAllowInsecureUnlock, evmconfig.DefaultJSONRPCAllowInsecureUnlock, "Allow insecure account unlocking when account-related RPCs are exposed by http") //nolint:lll
	startCmd.Flags().Float64(JSONRPCTxFeeCap, evmconfig.DefaultTxFeeCap, "Sets a cap on transaction fee that can be sent via the RPC APIs (1 = default 1 evmos)")                    //nolint:lll
	startCmd.Flags().Int32(JSONRPCFilterCap, evmconfig.DefaultFilterCap, "Sets the global cap for total number of filters that can be created")
	startCmd.Flags().Duration(JSONRPCEVMTimeout, evmconfig.DefaultEVMTimeout, "Sets a timeout used for eth_call (0=infinite)")
	startCmd.Flags().Duration(JSONRPCHTTPTimeout, evmconfig.DefaultHTTPTimeout, "Sets a read/write timeout for json-rpc http server (0=infinite)")
	startCmd.Flags().Duration(JSONRPCHTTPIdleTimeout, evmconfig.DefaultHTTPIdleTimeout, "Sets a idle timeout for json-rpc http server (0=infinite)")
	startCmd.Flags().Bool(JSONRPCAllowUnprotectedTxs, evmconfig.DefaultAllowUnprotectedTxs, "Allow for unprotected (non EIP155 signed) transactions to be submitted via the node's RPC when the global parameter is disabled") //nolint:lll
	startCmd.Flags().Int32(JSONRPCLogsCap, evmconfig.DefaultLogsCap, "Sets the max number of results can be returned from single `eth_getLogs` query")
	startCmd.Flags().Int32(JSONRPCBlockRangeCap, evmconfig.DefaultBlockRangeCap, "Sets the max block range allowed for `eth_getLogs` query")
	startCmd.Flags().Int(JSONRPCMaxOpenConnections, evmconfig.DefaultMaxOpenConnections, "Sets the maximum number of simultaneous connections for the server listener") //nolint:lll
	startCmd.Flags().Bool(JSONRPCEnableMetrics, false, "Define if EVM rpc metrics server should be enabled")
}

// AddTxFlags adds common flags for commands to post tx
func AddTxFlags(cmd *cobra.Command) (*cobra.Command, error) {
	cmd.PersistentFlags().String(flags.FlagChainID, "", "Specify Chain ID for sending Tx")
	cmd.PersistentFlags().String(flags.FlagFrom, "", "Name or address of private key with which to sign")
	cmd.PersistentFlags().String(flags.FlagFees, "", "Fees to pay along with transaction; eg: 10aevmos")
	cmd.PersistentFlags().String(flags.FlagGasPrices, "", "Gas prices to determine the transaction fee (e.g. 10aevmos)")
	cmd.PersistentFlags().String(flags.FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")                                                                                                   //nolint:lll
	cmd.PersistentFlags().Float64(flags.FlagGasAdjustment, flags.DefaultGasAdjustment, "adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored ") //nolint:lll
	cmd.PersistentFlags().StringP(flags.FlagBroadcastMode, "b", flags.BroadcastSync, "Transaction broadcasting mode (sync|async)")
	cmd.PersistentFlags().String(flags.FlagKeyringBackend, keyring.BackendOS, "Select keyring's backend")

	// --gas can accept integers and "simulate"
	// cmd.PersistentFlags().Var(&flags.GasFlagVar, "gas", fmt.Sprintf(
	//	"gas limit to set per-transaction; set to %q to calculate required gas automatically (default %d)",
	//	flags.GasFlagAuto, flags.DefaultGasLimit,
	// ))

	// viper.BindPFlag(flags.FlagTrustNode, cmd.Flags().Lookup(flags.FlagTrustNode))
	if err := viper.BindPFlag(flags.FlagNode, cmd.PersistentFlags().Lookup(flags.FlagNode)); err != nil {
		return nil, err
	}
	if err := viper.BindPFlag(flags.FlagKeyringBackend, cmd.PersistentFlags().Lookup(flags.FlagKeyringBackend)); err != nil {
		return nil, err
	}
	return cmd, nil
}
