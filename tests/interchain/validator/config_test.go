package validator_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/gorilla/websocket"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

type ConfigSuite struct {
	*chainsuite.Suite
}

func (s *ConfigSuite) TestNoIndexingTransactions() {
	txIndex := make(testutil.Toml)
	txIndex["indexer"] = "null"
	configToml := make(testutil.Toml)
	configToml["tx_index"] = txIndex
	s.Require().NoError(s.Chain.ModifyConfig(
		s.GetContext(), s.T(),
		map[string]testutil.Toml{"config/config.toml": configToml},
		0,
	))

	balanceBefore, err := s.Chain.GetBalance(s.GetContext(),
		s.Chain.ValidatorWallets[0].Address,
		s.Chain.Config().Denom)
	s.Require().NoError(err)
	amount := 1_000_000
	cmd := s.Chain.GetNode().TxCommand(
		interchaintest.FaucetAccountKeyName, "bank", "send",
		interchaintest.FaucetAccountKeyName, s.Chain.ValidatorWallets[0].Address,
		fmt.Sprintf("%d%s", amount, s.Chain.Config().Denom),
	)
	stdout, _, err := s.Chain.GetNode().Exec(s.GetContext(), cmd, nil)
	s.Require().NoError(err)
	tx := cosmos.CosmosTx{}
	s.Require().NoError(json.Unmarshal(stdout, &tx))
	s.Require().Equal(0, tx.Code, tx.RawLog)
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain))

	txResult, err := s.Chain.Validators[1].GetTransaction(
		s.Chain.Validators[1].CliContext(),
		tx.TxHash,
	)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), txResult.Code, txResult.RawLog)

	balanceAfter, err := s.Chain.GetBalance(s.GetContext(),
		s.Chain.ValidatorWallets[0].Address,
		s.Chain.Config().Denom)
	s.Require().NoError(err)
	s.Require().Equal(balanceBefore.AddRaw(int64(amount)), balanceAfter)

	_, err = s.Chain.Validators[0].GetTransaction(
		s.Chain.Validators[0].CliContext(),
		tx.TxHash,
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "transaction indexing is disabled")
}

func (s *ConfigSuite) TestPrometheus() {
	err := s.enablePrometheus()
	s.Require().NoError(err)
	metrics, err := s.getPrometheusMetrics(0)
	s.Require().NoError(err)
	s.Require().Contains(metrics, "go_gc_duration_seconds")
}

func (s *ConfigSuite) TestPruningEverything() {
	appToml := make(testutil.Toml)
	appToml["pruning"] = "everything"
	statesync := make(testutil.Toml)
	statesync["snapshot-interval"] = 0
	appToml["state-sync"] = statesync
	s.Require().NoError(s.Chain.ModifyConfig(
		s.GetContext(), s.T(),
		map[string]testutil.Toml{"config/app.toml": appToml},
	))

	s.smokeTestTx()
}

func (s *ConfigSuite) TestPeerLimit() {
	err := s.enablePrometheus()
	s.Require().NoError(err)

	metrics, err := s.getPrometheusMetrics(0)
	s.Require().NoError(err)
	s.Require().Equal(float64(3), metrics["cometbft_p2p_peers"].GetMetric()[0].GetGauge().GetValue())

	peers := s.Chain.Nodes().PeerString(s.GetContext())
	peerList := strings.Split(peers, ",")
	for i, node := range s.Chain.Nodes() {
		if i > 0 {
			s.Require().NoError(node.SetPeers(s.GetContext(), peerList[0]))
		}
	}
	s.Require().NoError(s.Chain.Validators[0].SetPeers(s.GetContext(), ""))

	s.Require().NoError(s.Chain.StopAllNodes(s.GetContext()))
	s.Require().NoError(s.Chain.StartAllNodes(s.GetContext()))
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain))

	p2p := make(testutil.Toml)
	p2p["max_num_inbound_peers"] = 2
	// disable pex so that we can control the number of peers
	p2p["pex"] = false
	configToml := make(testutil.Toml)
	configToml["p2p"] = p2p
	err = s.Chain.ModifyConfig(
		s.GetContext(), s.T(),
		map[string]testutil.Toml{"config/config.toml": configToml},
		0,
	)
	s.Require().NoError(err)

	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 4, s.Chain))

	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		metrics, err = s.getPrometheusMetrics(0)
		assert.NoError(c, err)
		assert.Equal(c, float64(2), metrics["cometbft_p2p_peers"].GetMetric()[0].GetGauge().GetValue())
	}, 3*time.Minute, 10*time.Second)

	foundZero := false
	for i := 1; i < len(s.Chain.Validators); i++ {
		metrics, err = s.getPrometheusMetrics(i)
		s.Require().NoError(err)
		metric := metrics["cometbft_p2p_peers"].GetMetric()
		if (len(metric) == 0 || metric[0].GetGauge().GetValue() == float64(0)) && !foundZero {
			// only one node should have 0 peers
			foundZero = true
			continue
		}
		s.Require().GreaterOrEqual(len(metric), 1)
		s.Require().GreaterOrEqual(metric[0].GetGauge().GetValue(), float64(1))
	}
}

func (s *ConfigSuite) TestWSConnectionLimit() {
	const connectionCount = 20
	u, err := url.Parse(s.Chain.GetHostRPCAddress())
	s.Require().NoError(err)
	u.Scheme = "ws"
	u.Path = "/websocket"
	canConnect := func() error {
		var eg errgroup.Group
		tCtx, tCancel := context.WithTimeout(s.GetContext(), 80*time.Second)
		defer tCancel()
		for i := 0; i < connectionCount; i++ {
			i := i
			eg.Go(func() error {
				c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
				if err != nil {
					return err
				}
				defer c.Close()
				err = c.WriteMessage(
					websocket.TextMessage,
					[]byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"subscribe","params":["tm.event='NewBlock'"],"id":%d}`, i)),
				)
				if err != nil {
					return err
				}
				for tCtx.Err() == nil {
					_, _, err = c.ReadMessage()
					if err != nil {
						return err
					}
				}
				return c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			})
		}
		return eg.Wait()
	}
	s.Require().NoError(canConnect())

	rpc := make(testutil.Toml)
	rpc["max_open_connections"] = 10
	configToml := make(testutil.Toml)
	configToml["rpc"] = rpc
	s.Require().NoError(s.Chain.ModifyConfig(
		s.GetContext(), s.T(),
		map[string]testutil.Toml{"config/config.toml": configToml},
	))

	s.Require().Error(canConnect())
}

func (s *ConfigSuite) TestDisableAPI() {
	_, _, err := s.Chain.Validators[0].Exec(s.GetContext(), []string{"wget", "-O", "-", s.Chain.GetAPIAddress() + "/cosmos/auth/v1beta1/accounts"}, nil)
	s.Require().NoError(err)

	apiToml := make(testutil.Toml)
	apiToml["enable"] = false
	appToml := make(testutil.Toml)
	appToml["api"] = apiToml
	s.Require().NoError(s.Chain.ModifyConfig(
		s.GetContext(), s.T(),
		map[string]testutil.Toml{"config/app.toml": appToml},
	))

	_, _, err = s.Chain.Validators[0].Exec(s.GetContext(), []string{"wget", "-O", "-", s.Chain.GetAPIAddress() + "/cosmos/auth/v1beta1/accounts"}, nil)
	s.Require().Error(err)
}

func TestConfig(t *testing.T) {
	nodes := 4
	s := &ConfigSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			UpgradeOnSetup: true,
			// Chains are gonna be test-scoped so that configuration changes
			// don't persist between tests.
			Scope: chainsuite.ChainScopeTest,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &nodes,
			},
		}),
	}
	suite.Run(t, s)
}

func (s *ConfigSuite) enablePrometheus() error {
	instrumentation := make(testutil.Toml)
	instrumentation["prometheus"] = true
	configToml := make(testutil.Toml)
	configToml["instrumentation"] = instrumentation
	return s.Chain.ModifyConfig(
		s.GetContext(), s.T(),
		map[string]testutil.Toml{"config/config.toml": configToml},
	)
}

func (s *ConfigSuite) getPrometheusMetrics(nodeIdx int) (map[string]*dto.MetricFamily, error) {
	prometheusHost := s.Chain.Validators[nodeIdx].HostName()
	prometheusHost = net.JoinHostPort(prometheusHost, "26660")
	stdout, _, err := s.Chain.Validators[nodeIdx].Exec(s.GetContext(),
		[]string{"wget", "-O", "-", "http://" + prometheusHost}, nil)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewBuffer(stdout)

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return nil, err
	}

	return mf, nil
}

// smokeTestTx does a basic bank send to verify that the validator is still working.
func (s *ConfigSuite) smokeTestTx() {
	txhash, err := s.Chain.GetNode().ExecTx(
		s.GetContext(), interchaintest.FaucetAccountKeyName,
		"bank", "send", interchaintest.FaucetAccountKeyName,
		s.Chain.ValidatorWallets[0].Address, "100"+s.Chain.Config().Denom,
	)
	s.Require().NoError(err)

	tx, err := s.Chain.GetTransaction(txhash)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), tx.Code)
	s.Require().Greater(tx.Height, int64(1))
}
