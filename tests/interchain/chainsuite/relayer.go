package chainsuite

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer"
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
		relayer.CustomDockerImage("ghcr.io/informalsystems/hermes", "1.10.1", "2000:2000"),
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

func (r *Relayer) GetTransferChannel(ctx context.Context, chain, counterparty *Chain) (*ibc.ChannelOutput, error) {
	return r.GetChannelWithPort(ctx, chain, counterparty, "transfer")
}

func (r *Relayer) GetChannelWithPort(ctx context.Context, chain, counterparty *Chain, portID string) (*ibc.ChannelOutput, error) {
	clients, err := r.GetClients(ctx, GetRelayerExecReporter(ctx), chain.Config().ChainID)
	if err != nil {
		return nil, err
	}
	var client *ibc.ClientOutput
	for _, c := range clients {
		if c.ClientState.ChainID == counterparty.Config().ChainID {
			client = c
			break
		}
	}
	if client == nil {
		return nil, fmt.Errorf("no client found for chain %s", counterparty.Config().ChainID)
	}

	stdout, _, err := chain.GetNode().ExecQuery(ctx, "ibc", "connection", "connections")
	if err != nil {
		return nil, fmt.Errorf("error querying connections: %w", err)
	}
	connections := gjson.GetBytes(stdout, fmt.Sprintf("connections.#(client_id==\"%s\")#.id", client.ClientID)).Array()
	if len(connections) == 0 {
		return nil, fmt.Errorf("no connections found for client %s", client.ClientID)
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
	return nil, fmt.Errorf("no channel found for port %s", portID)
}

func (r *Relayer) ClearCCVChannel(ctx context.Context, provider, consumer *Chain) error {
	var channel *ibc.ChannelOutput
	channel, err := r.GetChannelWithPort(ctx, consumer, provider, "consumer")
	if err != nil {
		return err
	}
	rs := r.Exec(ctx, GetRelayerExecReporter(ctx), []string{
		"hermes", "clear", "packets", "--port", "consumer", "--channel", channel.ChannelID,
		"--chain", consumer.Config().ChainID,
	}, nil)
	if rs.Err != nil {
		return fmt.Errorf("error clearing packets: %w", rs.Err)
	}
	return nil
}
