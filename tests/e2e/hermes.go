package e2e

import (
	"context"
	"fmt"
	"time"
)

func (s *IntegrationTestSuite) hermesClearPacket(configPath, chainID, portID, channelID string) (success bool) { //nolint:unparam
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	hermesCmd := []string{
		hermesBinary,
		"--json",
		fmt.Sprintf("--config=%s", configPath),
		"clear",
		"packets",
		fmt.Sprintf("--chain=%s", chainID),
		fmt.Sprintf("--channel=%s", channelID),
		fmt.Sprintf("--port=%s", portID),
	}

	if _, err := s.executeHermesCommand(ctx, hermesCmd); err != nil {
		s.T().Logf("failed to clear packets: %s", err)
		return false
	}

	return true
}

type RelayerPacketsOutput struct {
	Result struct {
		Dst struct {
			UnreceivedPackets []uint64 `json:"unreceived_packets"`
		} `json:"dst"`
		Src struct {
			UnreceivedPackets []uint64 `json:"unreceived_packets"`
		} `json:"src"`
	} `json:"result"`
	Status string `json:"status"`
}

func (s *IntegrationTestSuite) createConnection() {
	s.T().Logf("connecting %s and %s chains via IBC", s.chainA.id, s.chainB.id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	hermesCmd := []string{
		hermesBinary,
		"--json",
		"create",
		"connection",
		"--a-chain",
		s.chainA.id,
		"--b-chain",
		s.chainB.id,
	}

	_, err := s.executeHermesCommand(ctx, hermesCmd)
	s.Require().NoError(err, "failed to connect chains: %s", err)

	s.T().Logf("connected %s and %s chains via IBC", s.chainA.id, s.chainB.id)
}

func (s *IntegrationTestSuite) createChannel() {
	s.T().Logf("creating IBC transfer channel created between chains %s and %s", s.chainA.id, s.chainB.id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	hermesCmd := []string{
		hermesBinary,
		"--json",
		"create",
		"channel",
		"--a-chain", s.chainA.id,
		"--a-connection", "connection-0",
		"--a-port", "transfer",
		"--b-port", "transfer",
		"--channel-version", "ics20-1",
		"--order", "unordered",
	}

	_, err := s.executeHermesCommand(ctx, hermesCmd)
	s.Require().NoError(err, "failed to create IBC transfer channel between chains: %s", err)

	s.T().Logf("IBC transfer channel created between chains %s and %s", s.chainA.id, s.chainB.id)
}

// This function will complete the channel handshake in cases when ChanOpenInit was initiated
// by some transaction that was previously executed on the chain. For example,
// ICA MsgRegisterInterchainAccount will perform ChanOpenInit during its execution.
func (s *IntegrationTestSuite) completeChannelHandshakeFromTry(
	srcChain, dstChain,
	srcConnection, dstConnection,
	srcPort, dstPort,
	srcChannel, dstChannel string,
) {
	s.T().Logf("completing IBC channel handshake between: (%s, %s, %s, %s) and (%s, %s, %s, %s)",
		srcChain, srcConnection, srcPort, srcChannel,
		dstChain, dstConnection, dstPort, dstChannel)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	hermesCmd := []string{
		hermesBinary,
		"--json",
		"tx",
		"chan-open-try",
		"--dst-chain", dstChain,
		"--src-chain", srcChain,
		"--dst-connection", dstConnection,
		"--dst-port", dstPort,
		"--src-port", srcPort,
		"--src-channel", srcChannel,
	}

	_, err := s.executeHermesCommand(ctx, hermesCmd)
	s.Require().NoError(err, "failed to execute chan-open-try: %s", err)

	hermesCmd = []string{
		hermesBinary,
		"--json",
		"tx",
		"chan-open-ack",
		"--dst-chain", srcChain,
		"--src-chain", dstChain,
		"--dst-connection", srcConnection,
		"--dst-port", srcPort,
		"--src-port", dstPort,
		"--dst-channel", srcChannel,
		"--src-channel", dstChannel,
	}

	_, err = s.executeHermesCommand(ctx, hermesCmd)
	s.Require().NoError(err, "failed to execute chan-open-ack: %s", err)

	hermesCmd = []string{
		hermesBinary,
		"--json",
		"tx",
		"chan-open-confirm",
		"--dst-chain", dstChain,
		"--src-chain", srcChain,
		"--dst-connection", dstConnection,
		"--dst-port", dstPort,
		"--src-port", srcPort,
		"--dst-channel", dstChannel,
		"--src-channel", srcChannel,
	}

	_, err = s.executeHermesCommand(ctx, hermesCmd)
	s.Require().NoError(err, "failed to execute chan-open-confirm: %s", err)

	s.T().Logf("IBC channel handshake completed between: (%s, %s, %s, %s) and (%s, %s, %s, %s)",
		srcChain, srcConnection, srcPort, srcChannel,
		dstChain, dstConnection, dstPort, dstChannel)
}
