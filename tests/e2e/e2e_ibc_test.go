package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func (s *IntegrationTestSuite) runIBCRelayer() {
	s.T().Log("starting Hermes relayer container...")

	tmpDir, err := os.MkdirTemp("", "gaia-e2e-testnet-hermes-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	gaiaAVal := s.chainA.validators[0]
	gaiaBVal := s.chainB.validators[0]

	gaiaARly := s.chainA.genesisAccounts[relayerAccountIndex]
	gaiaBRly := s.chainB.genesisAccounts[relayerAccountIndex]

	hermesCfgPath := path.Join(tmpDir, "hermes")

	s.Require().NoError(os.MkdirAll(hermesCfgPath, 0o755))
	_, err = copyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.hermesResource, err = s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s-%s-relayer", s.chainA.id, s.chainB.id),
			Repository: "ghcr.io/cosmos/hermes-e2e",
			Tag:        "1.0.0",
			NetworkID:  s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/root/hermes", hermesCfgPath),
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"3031/tcp": {{HostIP: "", HostPort: "3031"}},
			},
			Env: []string{
				fmt.Sprintf("GAIA_A_E2E_CHAIN_ID=%s", s.chainA.id),
				fmt.Sprintf("GAIA_B_E2E_CHAIN_ID=%s", s.chainB.id),
				fmt.Sprintf("GAIA_A_E2E_VAL_MNEMONIC=%s", gaiaAVal.mnemonic),
				fmt.Sprintf("GAIA_B_E2E_VAL_MNEMONIC=%s", gaiaBVal.mnemonic),
				fmt.Sprintf("GAIA_A_E2E_RLY_MNEMONIC=%s", gaiaARly.mnemonic),
				fmt.Sprintf("GAIA_B_E2E_RLY_MNEMONIC=%s", gaiaBRly.mnemonic),
				fmt.Sprintf("GAIA_A_E2E_VAL_HOST=%s", s.valResources[s.chainA.id][0].Container.Name[1:]),
				fmt.Sprintf("GAIA_B_E2E_VAL_HOST=%s", s.valResources[s.chainB.id][0].Container.Name[1:]),
			},
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x /root/hermes/hermes_bootstrap.sh && /root/hermes/hermes_bootstrap.sh",
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	endpoint := fmt.Sprintf("http://%s/state", s.hermesResource.GetHostPort("3031/tcp"))
	s.Require().Eventually(
		func() bool {
			resp, err := http.Get(endpoint)
			if err != nil {
				return false
			}

			defer resp.Body.Close()

			bz, err := io.ReadAll(resp.Body)
			if err != nil {
				return false
			}

			var respBody map[string]interface{}
			if err := json.Unmarshal(bz, &respBody); err != nil {
				return false
			}

			status := respBody["status"].(string)
			result := respBody["result"].(map[string]interface{})

			return status == "success" && len(result["chains"].([]interface{})) == 2
		},
		5*time.Minute,
		time.Second,
		"hermes relayer not healthy",
	)

	s.T().Logf("started Hermes relayer container: %s", s.hermesResource.Container.ID)

	// XXX: Give time to both networks to start, otherwise we might see gRPC
	// transport errors.
	time.Sleep(10 * time.Second)

	// create the client, connection and channel between the two Gaia chains
	s.createConnection()
	time.Sleep(10 * time.Second)
	s.createChannel()
}

func (s *IntegrationTestSuite) sendIBC(c *chain, valIdx int, sender, recipient, token, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	ibcCmd := []string{
		gaiadBinary,
		txCommand,
		"ibc-transfer",
		"transfer",
		"transfer",
		"channel-0",
		recipient,
		token,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	s.T().Logf("sending %s from %s to %s (%s)", token, s.chainA.id, s.chainB.id, recipient)
	s.executeGaiaTxCommand(ctx, c, ibcCmd, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Log("successfully sent IBC tokens")
}

func (s *IntegrationTestSuite) createConnection() {
	s.T().Logf("connecting %s and %s chains via IBC", s.chainA.id, s.chainB.id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.hermesResource.Container.ID,
		User:         "root",
		Cmd: []string{
			"hermes",
			"create",
			"connection",
			"--a-chain",
			s.chainA.id,
			"--b-chain",
			s.chainB.id,
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(
		err,
		"failed connect chains; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Logf("connected %s and %s chains via IBC", s.chainA.id, s.chainB.id)
}

func (s *IntegrationTestSuite) createChannel() {
	s.T().Logf("connecting %s and %s chains via IBC", s.chainA.id, s.chainB.id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.hermesResource.Container.ID,
		User:         "root",
		Cmd: []string{
			"hermes",
			txCommand,
			"chan-open-init",
			"--dst-chain",
			s.chainA.id,
			"--src-chain",
			s.chainB.id,
			"--dst-connection",
			"connection-0",
			"--src-port=transfer",
			"--dst-port=transfer",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(
		err,
		"failed connect chains; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Logf("connected %s and %s chains via IBC", s.chainA.id, s.chainB.id)
}

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	time.Sleep(30 * time.Second)

	var ibcStakeDenom string

	s.Run("send_uatom_to_chainB", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			balances sdk.Coins
			err      error
		)

		address, err := s.chainA.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := address.String()

		address, err = s.chainB.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		recipient := address.String()

		tokenAmt := 3300000000
		s.sendIBC(s.chainA, 0, sender, recipient, strconv.Itoa(tokenAmt)+uatomDenom, fees.String())

		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainB.id][0].GetHostPort("1317/tcp"))

		s.Require().Eventually(
			func() bool {
				balances, err = queryGaiaAllBalances(chainBAPIEndpoint, recipient)
				s.Require().NoError(err)
				return balances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				ibcStakeDenom = c.Denom
				s.Require().Equal(int64(tokenAmt), c.Amount.Int64())
				break
			}
		}

		s.Require().NotEmpty(ibcStakeDenom)
	})
}

func (s *IntegrationTestSuite) TestBankTokenTransfer() {
	s.Run("send_photon_between_accounts", func() {
		var err error

		senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddress.String()

		recipientAddress, err := s.chainA.validators[1].keyInfo.GetAddress()
		s.Require().NoError(err)
		recipient := recipientAddress.String()

		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

		var (
			beforeSenderUAtomBalance    sdk.Coin
			beforeRecipientUAtomBalance sdk.Coin
		)

		s.Require().Eventually(
			func() bool {
				beforeSenderUAtomBalance, err = getSpecificBalance(chainAAPIEndpoint, sender, "uatom")
				s.Require().NoError(err)

				beforeRecipientUAtomBalance, err = getSpecificBalance(chainAAPIEndpoint, recipient, "uatom")
				s.Require().NoError(err)

				return beforeSenderUAtomBalance.IsValid() && beforeRecipientUAtomBalance.IsValid()
			},
			10*time.Second,
			5*time.Second,
		)

		s.execBankSend(s.chainA, 0, sender, recipient, tokenAmount.String(), fees.String(), false)

		s.Require().Eventually(
			func() bool {
				afterSenderUAtomBalance, err := getSpecificBalance(chainAAPIEndpoint, sender, "uatom")
				s.Require().NoError(err)

				afterRecipientUAtomBalance, err := getSpecificBalance(chainAAPIEndpoint, recipient, "uatom")
				s.Require().NoError(err)

				decremented := beforeSenderUAtomBalance.Sub(tokenAmount).Sub(fees).IsEqual(afterSenderUAtomBalance)
				incremented := beforeRecipientUAtomBalance.Add(tokenAmount).IsEqual(afterRecipientUAtomBalance)

				return decremented && incremented
			},
			time.Minute,
			5*time.Second,
		)
	})
}
