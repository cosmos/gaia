package chainsuite

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/relayer"
	"github.com/docker/docker/api/types/container"
	"github.com/tidwall/gjson"
)

type Relayer struct {
	ibc.Relayer
}

func NewRelayer(ctx context.Context, testName interchaintest.TestName) (*Relayer, error) {
	dockerClient, dockerNetwork := GetDockerContext(ctx)
	rly := interchaintest.NewBuiltinRelayerFactory(
		ibc.Hermes,
		GetLogger(ctx),
		relayer.CustomDockerImage("ghcr.io/informalsystems/hermes", "1.13.1", "2000:2000"),
	).Build(testName, dockerClient, dockerNetwork)
	return &Relayer{Relayer: rly}, nil
}

func (r *Relayer) SetupChainKeys(ctx context.Context, chain *Chain) error {
	rep := GetRelayerExecReporter(ctx)
	rpcAddr, grpcAddr := chain.GetRPCAddress(), chain.GetGRPCAddress()
	if !r.UseDockerNetwork() {
		rpcAddr, grpcAddr = chain.GetHostRPCAddress(), chain.GetHostGRPCAddress()
	}

	chainName := chain.Config().ChainID
	if err := r.AddChainConfiguration(ctx, rep, chain.Config(), chainName, rpcAddr, grpcAddr); err != nil {
		return err
	}

	return r.RestoreKey(ctx, rep, chain.Config(), chainName, chain.RelayerWallet.Mnemonic())
}

// SetMaxGas modifies the hermes config to set max_gas for all chains and restarts the relayer
func (r *Relayer) SetMaxGas(ctx context.Context, maxGas int) error {
	// Modify the config file
	cmd := fmt.Sprintf("sed -i 's/max_gas = [0-9]*/max_gas = %d/g' /home/hermes/.hermes/config.toml", maxGas)
	rs := r.Exec(ctx, GetRelayerExecReporter(ctx), []string{"sh", "-c", cmd}, nil)
	if rs.Err != nil {
		return fmt.Errorf("failed to set max_gas: %w", rs.Err)
	}

	// Get the relayer container
	dockerClient, _ := GetDockerContext(ctx)
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	// Find and restart the hermes container
	for _, c := range containers {
		for _, name := range c.Names {
			if len(name) > 0 && name[0] == '/' {
				name = name[1:] // Remove leading /
			}
			if len(name) > 6 && name[:6] == "hermes" {
				timeout := 10
				if err := dockerClient.ContainerRestart(ctx, c.ID, container.StopOptions{Timeout: &timeout}); err != nil {
					return fmt.Errorf("failed to restart hermes container %s: %w", c.ID, err)
				}
				GetLogger(ctx).Sugar().Infof("Restarted hermes container with max_gas=%d", maxGas)
				return nil
			}
		}
	}
	return fmt.Errorf("hermes container not found")
}

func (r *Relayer) GetTransferChannel(ctx context.Context, chain, counterparty *Chain) (*ibc.ChannelOutput, error) {
	return r.GetChannelWithPort(ctx, chain, counterparty, TransferPortID)
}

func (r *Relayer) GetChannelWithPort(ctx context.Context, chain, counterparty *Chain, portID string) (*ibc.ChannelOutput, error) {
	clients, err := r.GetClients(ctx, GetRelayerExecReporter(ctx), chain.Config().ChainID)
	if err != nil {
		return nil, err
	}
	for _, c := range clients {
		if c.ClientState.ChainID == counterparty.Config().ChainID {
			stdout, _, err := chain.GetNode().ExecQuery(ctx, "ibc", "connection", "connections")
			if err != nil {
				return nil, fmt.Errorf("error querying connections: %w", err)
			}
			connections := gjson.GetBytes(stdout, fmt.Sprintf("connections.#(client_id==\"%s\")#.id", c.ClientID)).Array()
			if len(connections) == 0 {
				continue
			}
			for _, connID := range connections {
				stdout, _, err := chain.GetNode().ExecQuery(ctx, "ibc", "channel", "connections", connID.String())
				if err != nil {
					return nil, err
				}
				channelJSON := gjson.GetBytes(stdout, fmt.Sprintf("channels.#(port_id==\"%s\")", portID)).String()
				if channelJSON != "" {
					channelOutput := &ibc.ChannelOutput{}
					if err := json.Unmarshal([]byte(channelJSON), channelOutput); err != nil {
						return nil, fmt.Errorf("error unmarshalling channel output %s: %w", channelJSON, err)
					}
					return channelOutput, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no channel found for port %s", portID)
}

func (r *Relayer) ClearTransferChannel(ctx context.Context, chainA, chainB *Chain) error {
	channel, err := r.GetTransferChannel(ctx, chainA, chainB)
	if err != nil {
		return err
	}
	rs := r.Exec(ctx, GetRelayerExecReporter(ctx), []string{
		"hermes", "clear", "packets", "--port", channel.PortID, "--channel", channel.ChannelID,
		"--chain", chainA.Config().ChainID,
	}, nil)
	if rs.Err != nil {
		return fmt.Errorf("error clearing packets: %w", rs.Err)
	}
	return nil
}

func relayerTransferPathFor(chainA, chainB *Chain) string {
	return fmt.Sprintf("transfer-%s-%s", chainA.Config().ChainID, chainB.Config().ChainID)
}
