package chainconfig

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/ethereum"
)

const (
	// ethereumPackageId is the package ID used by Kurtosis to find the Ethereum package we use for the testnet
	ethereumPackageId = "github.com/ethpandaops/ethereum-package@4.4.0"

	faucetPrivateKey = "0x04b9f63ecf84210c5366c66d68fa1f5da1fa4f634fad6dfc86178e4d79ff9e59"
)

var (
	kurtosisConfig = kurtosisNetworkParams{
		Participants: []kurtosisParticipant{
			{
				CLType:     "lodestar",
				CLImage:    "chainsafe/lodestar:v1.24.0",
				ELType:     "geth",
				ELImage:    "ethereum/client-go:v1.14.6",
				ELLogLevel: "info",
			},
		},
		NetworkParams: kurtosisNetworkConfigParams{
			Preset: "minimal",
		},
		WaitForFinalization: true,
	}
	executionService = fmt.Sprintf("el-1-%s-%s", kurtosisConfig.Participants[0].ELType, kurtosisConfig.Participants[0].CLType)
	consensusService = fmt.Sprintf("cl-1-%s-%s", kurtosisConfig.Participants[0].CLType, kurtosisConfig.Participants[0].ELType)
)

type EthKurtosisChain struct {
	RPC             string
	BeaconApiClient ethereum.BeaconAPIClient
	Faucet          *ecdsa.PrivateKey

	kurtosisCtx *kurtosis_context.KurtosisContext
	enclaveCtx  *enclaves.EnclaveContext
}

// To see all the configuration options: github.com/ethpandaops/ethereum-package
type kurtosisNetworkParams struct {
	Participants        []kurtosisParticipant       `json:"participants"`
	NetworkParams       kurtosisNetworkConfigParams `json:"network_params"`
	WaitForFinalization bool                        `json:"wait_for_finalization"`
}

type kurtosisParticipant struct {
	CLType     string `json:"cl_type"`
	CLImage    string `json:"cl_image"`
	ELType     string `json:"el_type"`
	ELImage    string `json:"el_image"`
	ELLogLevel string `json:"el_log_level"`
}

type kurtosisNetworkConfigParams struct {
	Preset string `json:"preset"`
}

// SpinUpKurtosisPoS spins up a kurtosis enclave with Etheruem PoS testnet using github.com/ethpandaops/ethereum-package
func SpinUpKurtosisPoS(ctx context.Context) (EthKurtosisChain, error) {
	faucet, err := crypto.ToECDSA(ethcommon.FromHex(faucetPrivateKey))
	if err != nil {
		return EthKurtosisChain{}, err
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return EthKurtosisChain{}, err
	}

	enclaveName := "ethereum-pos-testnet"
	enclaves, err := kurtosisCtx.GetEnclaves(ctx)
	if err != nil {
		return EthKurtosisChain{}, err
	}

	if enclaveInfos, found := enclaves.GetEnclavesByName()[enclaveName]; found {
		for _, enclaveInfo := range enclaveInfos {
			err = kurtosisCtx.DestroyEnclave(ctx, enclaveInfo.EnclaveUuid)
			if err != nil {
				return EthKurtosisChain{}, err
			}
		}
	}
	enclaveCtx, err := kurtosisCtx.CreateEnclave(ctx, enclaveName)
	if err != nil {
		return EthKurtosisChain{}, err
	}

	networkParamsJson, err := json.Marshal(kurtosisConfig)
	if err != nil {
		return EthKurtosisChain{}, err
	}
	starlarkResp, err := enclaveCtx.RunStarlarkRemotePackageBlocking(ctx, ethereumPackageId, &starlark_run_config.StarlarkRunConfig{
		SerializedParams: string(networkParamsJson),
	})
	if err != nil {
		return EthKurtosisChain{}, err
	}
	fmt.Println(starlarkResp.RunOutput)

	// exeuctionCtx is the service context (kurtosis concept) for the execution node that allows us to get the public ports
	executionCtx, err := enclaveCtx.GetServiceContext(executionService)
	if err != nil {
		return EthKurtosisChain{}, err
	}
	rpcPortSpec := executionCtx.GetPublicPorts()["rpc"]
	rpc := fmt.Sprintf("http://localhost:%d", rpcPortSpec.GetNumber())

	// consensusCtx is the service context (kurtosis concept) for the consensus node that allows us to get the public ports
	consensusCtx, err := enclaveCtx.GetServiceContext(consensusService)
	if err != nil {
		return EthKurtosisChain{}, err
	}
	beaconPortSpec := consensusCtx.GetPublicPorts()["http"]
	beaconRPC := fmt.Sprintf("http://localhost:%d", beaconPortSpec.GetNumber())

	// Wait for the chain to finalize
	beaconAPIClient := ethereum.NewBeaconAPIClient(beaconRPC)
	err = testutil.WaitForCondition(10*time.Minute, 5*time.Second, func() (bool, error) {
		finalizedBlocksResp, err := beaconAPIClient.GetFinalizedBlocks()
		fmt.Printf("Waiting for chain to finalize, finalizedBlockResp: %+v, err: %s\n", finalizedBlocksResp, err)
		if err != nil {
			return false, nil
		}
		if !finalizedBlocksResp.Finalized {
			return false, nil
		}

		header, err := beaconAPIClient.GetHeader(finalizedBlocksResp.Data.Message.Slot)
		if err != nil {
			return false, nil
		}
		bootstrap, err := beaconAPIClient.GetBootstrap(header.Root)
		if err != nil {
			return false, nil
		}

		return bootstrap.Data.Header.Beacon.Slot != 0, nil
	})
	if err != nil {
		return EthKurtosisChain{}, err
	}

	return EthKurtosisChain{
		RPC:             rpc,
		BeaconApiClient: beaconAPIClient,
		Faucet:          faucet,
		kurtosisCtx:     kurtosisCtx,
		enclaveCtx:      enclaveCtx,
	}, nil
}

func (e EthKurtosisChain) Destroy(ctx context.Context) {
	if err := e.kurtosisCtx.DestroyEnclave(ctx, string(e.enclaveCtx.GetEnclaveUuid())); err != nil {
		panic(err)
	}
}

func (e EthKurtosisChain) DumpLogs(ctx context.Context) error {
	enclaveServices, err := e.enclaveCtx.GetServices()
	if err != nil {
		return err
	}

	userServices := make(map[services.ServiceUUID]bool)
	serviceIdToName := make(map[services.ServiceUUID]string)
	for serviceName, servicesUUID := range enclaveServices {
		userServices[servicesUUID] = true
		serviceIdToName[servicesUUID] = string(serviceName)

	}

	stream, cancelFunc, err := e.kurtosisCtx.GetServiceLogs(ctx, string(e.enclaveCtx.GetEnclaveUuid()), userServices, false, true, 0, nil)
	if err != nil {
		return err
	}

	// Dump the stream chan into stdout
	fmt.Println("Dumping kurtosis logs")
	for {
		select {
		case logs, ok := <-stream:
			if !ok {
				return nil
			}
			for serviceID, serviceLog := range logs.GetServiceLogsByServiceUuids() {
				if serviceIdToName[serviceID] != executionService {
					continue
				}
				for _, log := range serviceLog {
					fmt.Printf("Service %s logs: %s\n", serviceIdToName[serviceID], log)
				}
			}
		case <-ctx.Done():
			cancelFunc()
			return nil
		}
	}
}
