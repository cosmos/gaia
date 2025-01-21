package chainsuite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/tidwall/gjson"
	"golang.org/x/sync/errgroup"

	sdkmath "cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// This moniker is hardcoded into interchaintest
const validatorMoniker = "validator"

type Chain struct {
	*cosmos.CosmosChain
	ValidatorWallets []ValidatorWallet
	RelayerWallet    ibc.Wallet
}

type ValidatorWallet struct {
	Moniker        string
	Address        string
	ValoperAddress string
	ValConsAddress string
}

func chainFromCosmosChain(cosmos *cosmos.CosmosChain, relayerWallet ibc.Wallet) (*Chain, error) {
	c := &Chain{CosmosChain: cosmos}
	wallets, err := getValidatorWallets(context.Background(), c)
	if err != nil {
		return nil, err
	}
	c.ValidatorWallets = wallets
	c.RelayerWallet = relayerWallet
	return c, nil
}

// CreateChain creates a single new chain with the given version and returns the chain object.
func CreateChain(ctx context.Context, testName interchaintest.TestName, spec *interchaintest.ChainSpec) (*Chain, error) {
	cf := interchaintest.NewBuiltinChainFactory(
		GetLogger(ctx),
		[]*interchaintest.ChainSpec{spec},
	)

	chains, err := cf.Chains(testName.Name())
	if err != nil {
		return nil, err
	}
	cosmosChain := chains[0].(*cosmos.CosmosChain)
	relayerWallet, err := cosmosChain.BuildRelayerWallet(ctx, "relayer-"+cosmosChain.Config().ChainID)
	if err != nil {
		return nil, err
	}

	ic := interchaintest.NewInterchain().AddChain(cosmosChain, ibc.WalletAmount{
		Address: relayerWallet.FormattedAddress(),
		Denom:   cosmosChain.Config().Denom,
		Amount:  sdkmath.NewInt(ValidatorFunds),
	})

	dockerClient, dockerNetwork := GetDockerContext(ctx)

	if err := ic.Build(ctx, GetRelayerExecReporter(ctx), interchaintest.InterchainBuildOptions{
		Client:    dockerClient,
		NetworkID: dockerNetwork,
		TestName:  testName.Name(),
	}); err != nil {
		return nil, err
	}

	chain, err := chainFromCosmosChain(cosmosChain, relayerWallet)
	if err != nil {
		return nil, err
	}
	return chain, nil
}

func (c *Chain) GenerateTx(ctx context.Context, valIdx int, command ...string) (string, error) {
	command = append([]string{"tx"}, command...)
	command = append(command, "--generate-only", "--keyring-backend", "test", "--chain-id", c.Config().ChainID)
	command = c.Validators[valIdx].NodeCommand(command...)
	stdout, _, err := c.Validators[valIdx].Exec(ctx, command, nil)
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}

func (c *Chain) WaitForProposalStatus(ctx context.Context, proposalID string, status govv1.ProposalStatus) error {
	propID, err := strconv.ParseInt(proposalID, 10, 64)
	if err != nil {
		return err
	}
	chainHeight, err := c.Height(ctx)
	if err != nil {
		return err
	}
	// At 4s per block, 75 blocks is about 5 minutes.
	maxHeight := chainHeight + 75
	_, err = cosmos.PollForProposalStatusV1(ctx, c.CosmosChain, chainHeight, maxHeight, uint64(propID), status)
	return err
}

func (c *Chain) PassProposal(ctx context.Context, proposalID string) error {
	propID, err := strconv.ParseInt(proposalID, 10, 64)
	if err != nil {
		return err
	}
	err = c.VoteOnProposalAllValidators(ctx, uint64(propID), cosmos.ProposalVoteYes)
	if err != nil {
		return err
	}
	return c.WaitForProposalStatus(ctx, proposalID, govv1.StatusPassed)
}

func (c *Chain) ReplaceImagesAndRestart(ctx context.Context, version string) error {
	// bring down nodes to prepare for upgrade
	err := c.StopAllNodes(ctx)
	if err != nil {
		return err
	}

	// upgrade version on all nodes
	c.UpgradeVersion(ctx, c.GetNode().DockerClient, c.GetNode().Image.Repository, version)

	// start all nodes back up.
	// validators reach consensus on first block after upgrade height
	// and block production resumes.
	err = c.StartAllNodes(ctx)
	if err != nil {
		return err
	}

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 60*time.Second)
	defer timeoutCancel()
	err = testutil.WaitForBlocks(timeoutCtx, 5, c)
	if err != nil {
		return fmt.Errorf("failed to wait for blocks after upgrade: %w", err)
	}

	// Flush "successfully migrated key info" messages
	for _, val := range c.Validators {
		_, _, err := val.ExecBin(ctx, "keys", "list", "--keyring-backend", "test")
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Chain) Upgrade(ctx context.Context, upgradeName, version string) error {
	height, err := c.Height(ctx)
	if err != nil {
		return err
	}

	haltHeight := height + UpgradeDelta

	proposal := cosmos.SoftwareUpgradeProposal{
		Deposit:     GovDepositAmount, // greater than min deposit
		Title:       "Upgrade to " + upgradeName,
		Name:        upgradeName,
		Description: "Upgrade to " + upgradeName,
		Height:      haltHeight,
	}
	upgradeTx, err := c.UpgradeProposal(ctx, interchaintest.FaucetAccountKeyName, proposal)
	if err != nil {
		return err
	}
	if err := c.PassProposal(ctx, upgradeTx.ProposalID); err != nil {
		return err
	}

	height, err = c.Height(ctx)
	if err != nil {
		return err
	}

	// wait for the chain to halt. We're asking for blocks after the halt height, so we should time out.
	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, (time.Duration(haltHeight-height)+10)*CommitTimeout)
	defer timeoutCtxCancel()
	err = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height)+3, c)
	if err == nil {
		return errors.New("chain should not produce blocks after halt height")
	} else if timeoutCtx.Err() == nil {
		return errors.New("chain should not produce blocks after halt height")
	}

	height, err = c.Height(ctx)
	if err != nil {
		return err
	}

	// make sure that chain is halted; some chains may produce one more block after halt height
	if height-haltHeight > 1 {
		return fmt.Errorf("height %d is not within one block of halt height %d; chain isn't halted", height, haltHeight)
	}

	return c.ReplaceImagesAndRestart(ctx, version)
}

func (c *Chain) GetValidatorPower(ctx context.Context, hexaddr string) (int64, error) {
	var power int64
	err := CheckEndpoint(ctx, c.GetHostRPCAddress()+"/validators", func(b []byte) error {
		power = gjson.GetBytes(b, fmt.Sprintf("result.validators.#(address==\"%s\").voting_power", hexaddr)).Int()
		if power == 0 {
			return fmt.Errorf("validator %s power not found; validators are: %s", hexaddr, string(b))
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return power, nil
}

func (c *Chain) GetValidatorHex(ctx context.Context, val int) (string, error) {
	json, err := c.Validators[val].ReadFile(ctx, "config/priv_validator_key.json")
	if err != nil {
		return "", err
	}
	providerHex := gjson.GetBytes(json, "address").String()
	return providerHex, nil
}

func getValidatorWallets(ctx context.Context, chain *Chain) ([]ValidatorWallet, error) {
	wallets := make([]ValidatorWallet, len(chain.Validators))
	lock := new(sync.Mutex)
	eg := new(errgroup.Group)
	for i := range chain.Validators {
		i := i
		eg.Go(func() error {
			// This moniker is hardcoded into the chain's genesis process.
			moniker := validatorMoniker
			address, err := chain.Validators[i].KeyBech32(ctx, moniker, "acc")
			if err != nil {
				return err
			}
			valoperAddress, err := chain.Validators[i].KeyBech32(ctx, moniker, "val")
			if err != nil {
				return err
			}
			valCons, _, err := chain.Validators[i].ExecBin(ctx, "comet", "show-address")
			if err != nil {
				return err
			}
			lock.Lock()
			defer lock.Unlock()
			wallets[i] = ValidatorWallet{
				Moniker:        moniker,
				Address:        address,
				ValoperAddress: valoperAddress,
				ValConsAddress: strings.TrimSpace(string(valCons)),
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return wallets, nil
}

func (c *Chain) QueryJSON(ctx context.Context, jsonPath string, query ...string) (gjson.Result, error) {
	stdout, _, err := c.GetNode().ExecQuery(ctx, query...)
	if err != nil {
		return gjson.Result{}, err
	}
	retval := gjson.GetBytes(stdout, jsonPath)
	if !retval.Exists() {
		return gjson.Result{}, fmt.Errorf("json path %s not found in query result %s", jsonPath, stdout)
	}
	return retval, nil
}

// GetProposalID parses the proposal ID from the tx; necessary when the proposal type isn't accessible to interchaintest yet
func (c *Chain) GetProposalID(ctx context.Context, txhash string) (string, error) {
	stdout, _, err := c.GetNode().ExecQuery(ctx, "tx", txhash)
	if err != nil {
		return "", err
	}
	result := struct {
		Events []abcitypes.Event `json:"events"`
	}{}
	if err := json.Unmarshal(stdout, &result); err != nil {
		return "", err
	}
	for _, event := range result.Events {
		if event.Type == "submit_proposal" {
			for _, attr := range event.Attributes {
				if string(attr.Key) == "proposal_id" {
					return string(attr.Value), nil
				}
			}
		}
	}
	return "", fmt.Errorf("proposal ID not found in tx %s", txhash)
}

func (c *Chain) hasOrderingFlag(ctx context.Context) (bool, error) {
	cmd := c.GetNode().BinCommand("tx", "interchain-accounts", "controller", "register", "--help")
	stdout, _, err := c.GetNode().Exec(ctx, cmd, nil)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(stdout), "ordering"), nil
}

func (c *Chain) GetICAAddress(ctx context.Context, srcAddress string, srcConnection string) string {
	var icaAddress string

	// it takes a moment for it to be created
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 90*time.Second)
	defer timeoutCancel()
	for timeoutCtx.Err() == nil {
		time.Sleep(5 * time.Second)
		stdout, _, err := c.GetNode().ExecQuery(timeoutCtx,
			"interchain-accounts", "controller", "interchain-account",
			srcAddress, srcConnection,
		)
		if err != nil {
			GetLogger(ctx).Sugar().Warnf("error querying interchain account: %s", err)
			continue
		}
		result := map[string]interface{}{}
		err = json.Unmarshal(stdout, &result)
		if err != nil {
			GetLogger(ctx).Sugar().Warnf("error unmarshalling interchain account: %s", err)
			continue
		}
		icaAddress = result["address"].(string)
		if icaAddress != "" {
			break
		}
	}
	return icaAddress
}

func (c *Chain) SetupICAAccount(ctx context.Context, host *Chain, relayer *Relayer, srcAddress string, valIdx int, initialFunds int64) (string, error) {
	srcChannel, err := relayer.GetTransferChannel(ctx, c, host)
	if err != nil {
		return "", err
	}
	srcConnection := srcChannel.ConnectionHops[0]

	hasOrdering, err := c.hasOrderingFlag(ctx)
	if err != nil {
		return "", err
	}

	if hasOrdering {
		_, err = c.Validators[valIdx].ExecTx(ctx, srcAddress,
			"interchain-accounts", "controller", "register",
			"--ordering", "ORDER_ORDERED", "--version", "",
			srcConnection,
		)
	} else {
		_, err = c.Validators[valIdx].ExecTx(ctx, srcAddress,
			"interchain-accounts", "controller", "register",
			srcConnection,
		)
	}
	if err != nil {
		return "", err
	}

	icaAddress := c.GetICAAddress(ctx, srcAddress, srcConnection)
	if icaAddress == "" {
		return "", fmt.Errorf("ICA address not found")
	}

	err = host.SendFunds(ctx, interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Denom:   host.Config().Denom,
		Amount:  sdkmath.NewInt(initialFunds),
		Address: icaAddress,
	})
	if err != nil {
		return "", err
	}

	return icaAddress, nil
}

func (c *Chain) AddLinkedChain(ctx context.Context, testName interchaintest.TestName, relayer *Relayer, spec *interchaintest.ChainSpec) (*Chain, error) {
	dockerClient, dockerNetwork := GetDockerContext(ctx)

	cf := interchaintest.NewBuiltinChainFactory(
		GetLogger(ctx),
		[]*interchaintest.ChainSpec{spec},
	)

	chains, err := cf.Chains(testName.Name())
	if err != nil {
		return nil, err
	}
	cosmosChainB := chains[0].(*cosmos.CosmosChain)
	relayerWallet, err := cosmosChainB.BuildRelayerWallet(ctx, "relayer-"+cosmosChainB.Config().ChainID)
	if err != nil {
		return nil, err
	}

	ic := interchaintest.NewInterchain().AddChain(cosmosChainB, ibc.WalletAmount{
		Address: relayerWallet.FormattedAddress(),
		Denom:   cosmosChainB.Config().Denom,
		Amount:  sdkmath.NewInt(ValidatorFunds),
	})

	if err := ic.Build(ctx, GetRelayerExecReporter(ctx), interchaintest.InterchainBuildOptions{
		Client:    dockerClient,
		NetworkID: dockerNetwork,
		TestName:  testName.Name(),
	}); err != nil {
		return nil, err
	}

	chainB, err := chainFromCosmosChain(cosmosChainB, relayerWallet)
	if err != nil {
		return nil, err
	}
	rep := GetRelayerExecReporter(ctx)
	if err := relayer.SetupChainKeys(ctx, chainB); err != nil {
		return nil, err
	}
	if err := relayer.StopRelayer(ctx, rep); err != nil {
		return nil, err
	}
	if err := relayer.StartRelayer(ctx, rep); err != nil {
		return nil, err
	}

	if err := relayer.GeneratePath(ctx, rep, c.Config().ChainID, chainB.Config().ChainID, relayerTransferPathFor(c, chainB)); err != nil {
		return nil, err
	}

	if err := relayer.LinkPath(ctx, rep, relayerTransferPathFor(c, chainB), ibc.CreateChannelOptions{
		DestPortName:   TransferPortID,
		SourcePortName: TransferPortID,
		Order:          ibc.Unordered,
	}, ibc.DefaultClientOpts()); err != nil {
		return nil, err
	}

	return chainB, nil
}

func (c *Chain) ModifyConfig(ctx context.Context, testName interchaintest.TestName, configChanges map[string]testutil.Toml, validators ...int) error {
	eg := errgroup.Group{}
	if len(validators) == 0 {
		validators = make([]int, len(c.Validators))
		for valIdx := range validators {
			validators[valIdx] = valIdx
		}
	}
	for _, i := range validators {
		val := c.Validators[i]
		eg.Go(func() error {
			for file, changes := range configChanges {
				if err := testutil.ModifyTomlConfigFile(
					ctx, GetLogger(ctx),
					val.DockerClient, testName.Name(), val.VolumeName,
					file, changes,
				); err != nil {
					return err
				}
			}
			if err := val.StopContainer(ctx); err != nil {
				return err
			}
			return val.StartContainer(ctx)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	time.Sleep(30 * time.Second)
	return nil
}
