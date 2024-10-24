package chainsuite

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

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
		relayer.CustomDockerImage("ghcr.io/informalsystems/hermes", "1.10.3", "2000:2000"),
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

func (r *Relayer) ConnectProviderConsumer(ctx context.Context, provider *Chain, consumer *Chain) error {
	icsPath := relayerICSPathFor(provider, consumer)
	rep := GetRelayerExecReporter(ctx)
	if err := r.GeneratePath(ctx, rep, consumer.Config().ChainID, provider.Config().ChainID, icsPath); err != nil {
		return err
	}

	consumerClients, err := r.GetClients(ctx, rep, consumer.Config().ChainID)
	if err != nil {
		return err
	}
	sort.Slice(consumerClients, func(i, j int) bool {
		return consumerClients[i].ClientID > consumerClients[j].ClientID
	})
	var consumerClient *ibc.ClientOutput
	for _, client := range consumerClients {
		if client.ClientState.ChainID == provider.Config().ChainID {
			consumerClient = client
			break
		}
	}
	if consumerClient == nil {
		return fmt.Errorf("consumer chain %s does not have a client tracking the provider chain %s", consumer.Config().ChainID, provider.Config().ChainID)
	}
	consumerClientID := consumerClient.ClientID

	providerClients, err := r.GetClients(ctx, rep, provider.Config().ChainID)
	if err != nil {
		return err
	}
	sort.Slice(providerClients, func(i, j int) bool {
		return providerClients[i].ClientID > providerClients[j].ClientID
	})

	var providerClient *ibc.ClientOutput
	for _, client := range providerClients {
		if client.ClientState.ChainID == consumer.Config().ChainID {
			providerClient = client
			break
		}
	}
	if providerClient == nil {
		return fmt.Errorf("provider chain %s does not have a client tracking the consumer chain %s for path %s on relayer %s",
			provider.Config().ChainID, consumer.Config().ChainID, icsPath, r)
	}
	providerClientID := providerClient.ClientID

	if err := r.UpdatePath(ctx, rep, icsPath, ibc.PathUpdateOptions{
		SrcClientID: &consumerClientID,
		DstClientID: &providerClientID,
	}); err != nil {
		return err
	}

	if err := r.CreateConnections(ctx, rep, icsPath); err != nil {
		return err
	}

	if err := r.CreateChannel(ctx, rep, icsPath, ibc.CreateChannelOptions{
		SourcePortName: "consumer",
		DestPortName:   "provider",
		Order:          ibc.Ordered,
		Version:        "1",
	}); err != nil {
		return err
	}

	tCtx, tCancel := context.WithTimeout(ctx, 30*CommitTimeout)
	defer tCancel()
	for tCtx.Err() == nil {
		var ch *ibc.ChannelOutput
		ch, err = r.GetTransferChannel(ctx, provider, consumer)
		if err == nil && ch != nil {
			break
		} else if err == nil {
			err = fmt.Errorf("channel not found")
		}
		time.Sleep(CommitTimeout)
	}
	return err
}

func relayerICSPathFor(chainA, chainB *Chain) string {
	return fmt.Sprintf("ics-%s-%s", chainA.Config().ChainID, chainB.Config().ChainID)
}

func relayerTransferPathFor(chainA, chainB *Chain) string {
	return fmt.Sprintf("transfer-%s-%s", chainA.Config().ChainID, chainB.Config().ChainID)
}
