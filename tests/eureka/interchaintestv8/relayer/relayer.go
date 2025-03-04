package relayer

import (
	"os"
	"os/exec"
	"time"

	grpc "google.golang.org/grpc"
	insecure "google.golang.org/grpc/credentials/insecure"

	relayertypes "github.com/srdtrk/solidity-ibc-eureka/e2e/v8/types/relayer"
)

// binaryPath returns the path to the relayer binary.
func binaryPath() string {
	return "relayer"
}

// StartRelayer starts the relayer with the given config file.
func StartRelayer(configPath string) (*os.Process, error) {
	cmd := exec.Command(binaryPath(), "start", "--config", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// run this command in the background
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	// wait for the relayer to start
	time.Sleep(5 * time.Second)

	return cmd.Process, nil
}

// GetGRPCClient returns a gRPC client for the relayer.
func GetGRPCClient(addr string) (relayertypes.RelayerServiceClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return relayertypes.NewRelayerServiceClient(conn), nil
}
