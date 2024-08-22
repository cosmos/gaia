package chainsuite

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"golang.org/x/mod/semver"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ccvclient "github.com/cosmos/interchain-security/v5/x/ccv/provider/client"

	sdkmath "cosmossdk.io/math"

	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type ConsumerBootstrapCb func(ctx context.Context, consumer *cosmos.CosmosChain)

type ConsumerConfig struct {
	ChainName             string
	Version               string
	Denom                 string
	ShouldCopyProviderKey [ValidatorCount]bool
	TopN                  int
	ValidatorSetCap       int
	ValidatorPowerCap     int
	spec                  *interchaintest.ChainSpec

	DuringDepositPeriod ConsumerBootstrapCb
	DuringVotingPeriod  ConsumerBootstrapCb
	BeforeSpawnTime     ConsumerBootstrapCb
	AfterSpawnTime      ConsumerBootstrapCb
}

type proposalWaiter struct {
	canDeposit chan struct{}
	isInVoting chan struct{}
	canVote    chan struct{}
	isPassed   chan struct{}
}

func (pw *proposalWaiter) waitForDepositAllowed() {
	<-pw.canDeposit
}

func (pw *proposalWaiter) startVotingPeriod() {
	close(pw.isInVoting)
}

func (pw *proposalWaiter) waitForVoteAllowed() {
	<-pw.canVote
}

func (pw *proposalWaiter) pass() {
	close(pw.isPassed)
}

func (pw *proposalWaiter) AllowVote() {
	close(pw.canVote)
}

func (pw *proposalWaiter) WaitForPassed() {
	<-pw.isPassed
}

func (pw *proposalWaiter) AllowDeposit() {
	close(pw.canDeposit)
}

func (pw *proposalWaiter) WaitForVotingPeriod() {
	<-pw.isInVoting
}

func newProposalWaiter() *proposalWaiter {
	return &proposalWaiter{
		canDeposit: make(chan struct{}),
		isInVoting: make(chan struct{}),
		canVote:    make(chan struct{}),
		isPassed:   make(chan struct{}),
	}
}

func (p *Chain) AddConsumerChain(ctx context.Context, relayer *Relayer, config ConsumerConfig) (*Chain, error) {
	dockerClient, dockerNetwork := GetDockerContext(ctx)

	if len(config.ShouldCopyProviderKey) != ValidatorCount {
		return nil, fmt.Errorf("shouldCopyProviderKey must be the same length as the number of validators")
	}

	spawnTime := time.Now().Add(ChainSpawnWait)
	chainID := fmt.Sprintf("%s-%d", config.ChainName, len(p.Consumers)+1)

	proposalWaiter, errCh, err := p.SubmitConsumerAdditionProposal(ctx, chainID, config, spawnTime)
	if err != nil {
		return nil, err
	}

	if config.spec == nil {
		config.spec = p.DefaultConsumerChainSpec(ctx, chainID, config, spawnTime, proposalWaiter)
	}
	if semver.Compare(p.GetNode().ICSVersion(ctx), "v4.1.0") > 0 && config.spec.InterchainSecurityConfig.ProviderVerOverride == "" {
		config.spec.InterchainSecurityConfig.ProviderVerOverride = "v4.1.0"
	}
	cf := interchaintest.NewBuiltinChainFactory(
		GetLogger(ctx),
		[]*interchaintest.ChainSpec{config.spec},
	)
	chains, err := cf.Chains(p.GetNode().TestName)
	if err != nil {
		return nil, err
	}
	cosmosConsumer := chains[0].(*cosmos.CosmosChain)

	// We can't use AddProviderConsumerLink here because the provider chain is already built; we'll have to do everything by hand.
	p.Consumers = append(p.Consumers, cosmosConsumer)
	cosmosConsumer.Provider = p.CosmosChain

	relayerWallet, err := cosmosConsumer.BuildRelayerWallet(ctx, "relayer-"+cosmosConsumer.Config().ChainID)
	if err != nil {
		return nil, err
	}
	wallets := make([]ibc.Wallet, len(p.Validators)+1)
	wallets[0] = relayerWallet
	// This is a hack, but we need to create wallets for the validators that have the right moniker.
	for i := 1; i <= len(p.Validators); i++ {
		wallets[i], err = cosmosConsumer.BuildRelayerWallet(ctx, validatorMoniker)
		if err != nil {
			return nil, err
		}
	}
	walletAmounts := make([]ibc.WalletAmount, len(wallets))
	for i, wallet := range wallets {
		walletAmounts[i] = ibc.WalletAmount{
			Address: wallet.FormattedAddress(),
			Denom:   cosmosConsumer.Config().Denom,
			Amount:  sdkmath.NewInt(ValidatorFunds),
		}
	}
	ic := interchaintest.NewInterchain().
		AddChain(cosmosConsumer, walletAmounts...).
		AddRelayer(relayer, "relayer")

	if err := ic.Build(ctx, GetRelayerExecReporter(ctx), interchaintest.InterchainBuildOptions{
		Client:    dockerClient,
		NetworkID: dockerNetwork,
		TestName:  p.GetNode().TestName,
	}); err != nil {
		return nil, err
	}

	// The chain should be built now, so we gotta check for errors in passing the proposal.
	if err := <-errCh; err != nil {
		return nil, err
	}

	for i, val := range cosmosConsumer.Validators {
		if err := val.RecoverKey(ctx, validatorMoniker, wallets[i+1].Mnemonic()); err != nil {
			return nil, err
		}
	}
	consumer, err := chainFromCosmosChain(cosmosConsumer, relayerWallet)
	if err != nil {
		return nil, err
	}

	err = relayer.SetupChainKeys(ctx, consumer)
	if err != nil {
		return nil, err
	}
	rep := GetRelayerExecReporter(ctx)
	if err := relayer.StopRelayer(ctx, rep); err != nil {
		return nil, err
	}
	if err := relayer.StartRelayer(ctx, rep); err != nil {
		return nil, err
	}
	err = connectProviderConsumer(ctx, p, consumer, relayer)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func (p *Chain) DefaultConsumerChainSpec(ctx context.Context, chainID string, config ConsumerConfig, spawnTime time.Time, proposalWaiter *proposalWaiter) *interchaintest.ChainSpec {
	const (
		strideChain  = "stride"
		icsConsumer  = "ics-consumer"
		neutronChain = "neutron"
	)
	fullNodes := len(p.FullNodes)
	validators := len(p.Validators)

	chainType := config.ChainName
	version := config.Version
	denom := config.Denom
	shouldCopyProviderKey := config.ShouldCopyProviderKey

	bechPrefix := ""
	if chainType == icsConsumer {
		majorVersion, err := strconv.Atoi(version[1:2])
		if err != nil {
			// this really shouldn't happen unless someone misconfigured something
			panic(fmt.Sprintf("failed to parse major version from %s: %v", version, err))
		}
		if majorVersion >= 4 {
			bechPrefix = "consumer"
		}
	} else if chainType == strideChain {
		bechPrefix = "stride"
	}
	genesisOverrides := []cosmos.GenesisKV{
		cosmos.NewGenesisKV("app_state.slashing.params.signed_blocks_window", strconv.Itoa(SlashingWindowConsumer)),
		cosmos.NewGenesisKV("consensus_params.block.max_gas", "50000000"),
	}
	if config.TopN >= 0 {
		genesisOverrides = append(genesisOverrides, cosmos.NewGenesisKV("app_state.ccvconsumer.params.soft_opt_out_threshold", "0.0"))
	}
	if chainType == neutronChain {
		genesisOverrides = append(genesisOverrides,
			cosmos.NewGenesisKV("app_state.globalfee.params.minimum_gas_prices", []interface{}{
				map[string]interface{}{
					"amount": "0.005",
					"denom":  denom,
				},
			}),
		)
	}

	modifyGenesis := cosmos.ModifyGenesis(genesisOverrides)
	if chainType == strideChain {
		genesisOverrides = append(genesisOverrides,
			cosmos.NewGenesisKV("app_state.gov.params.voting_period", GovVotingPeriod.String()),
		)
		modifyGenesis = func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
			b, err := cosmos.ModifyGenesis(genesisOverrides)(cc, b)
			if err != nil {
				return nil, err
			}
			b, err = sjson.SetBytes(b, "app_state.epochs.epochs.#(identifier==\"day\").duration", "120s")
			if err != nil {
				return nil, err
			}
			return sjson.SetBytes(b, "app_state.epochs.epochs.#(identifier==\"stride_epoch\").duration", "30s")
		}
	}

	return &interchaintest.ChainSpec{
		Name:          chainType,
		Version:       version,
		ChainName:     chainID,
		NumFullNodes:  &fullNodes,
		NumValidators: &validators,
		ChainConfig: ibc.ChainConfig{
			Denom:         denom,
			GasPrices:     "0.005" + denom,
			GasAdjustment: 2.0,
			ChainID:       chainID,
			ConfigFileOverrides: map[string]any{
				"config/config.toml": DefaultConfigToml(),
			},
			PreGenesis: func(consumer ibc.Chain) error {
				if config.DuringDepositPeriod != nil {
					config.DuringDepositPeriod(ctx, consumer.(*cosmos.CosmosChain))
				}
				proposalWaiter.AllowDeposit()
				proposalWaiter.WaitForVotingPeriod()
				if config.DuringVotingPeriod != nil {
					config.DuringVotingPeriod(ctx, consumer.(*cosmos.CosmosChain))
				}
				proposalWaiter.AllowVote()
				proposalWaiter.WaitForPassed()
				tCtx, tCancel := context.WithDeadline(ctx, spawnTime)
				defer tCancel()
				if config.BeforeSpawnTime != nil {
					config.BeforeSpawnTime(tCtx, consumer.(*cosmos.CosmosChain))
				}
				// interchaintest will set up the validator keys right before PreGenesis.
				// Now we just need to wait for the chain to spawn before interchaintest can get the ccv file.
				// This wait is here and not there because of changes we've made to interchaintest that need to be upstreamed in an orderly way.
				GetLogger(ctx).Sugar().Infof("waiting for chain %s to spawn at %s", chainID, spawnTime)
				<-tCtx.Done()
				if err := testutil.WaitForBlocks(ctx, 2, p); err != nil {
					return err
				}
				if config.AfterSpawnTime != nil {
					config.AfterSpawnTime(ctx, consumer.(*cosmos.CosmosChain))
				}
				return nil
			},
			Bech32Prefix:         bechPrefix,
			ModifyGenesisAmounts: DefaultGenesisAmounts(denom),
			ModifyGenesis:        modifyGenesis,
			InterchainSecurityConfig: ibc.ICSConfig{
				ConsumerCopyProviderKey: func(i int) bool {
					return shouldCopyProviderKey[i]
				},
			},
		},
	}
}

func connectProviderConsumer(ctx context.Context, provider *Chain, consumer *Chain, relayer *Relayer) error {
	icsPath := relayerICSPathFor(provider, consumer)
	rep := GetRelayerExecReporter(ctx)
	if err := relayer.GeneratePath(ctx, rep, consumer.Config().ChainID, provider.Config().ChainID, icsPath); err != nil {
		return err
	}

	consumerClients, err := relayer.GetClients(ctx, rep, consumer.Config().ChainID)
	if err != nil {
		return err
	}

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

	providerClients, err := relayer.GetClients(ctx, rep, provider.Config().ChainID)
	if err != nil {
		return err
	}

	var providerClient *ibc.ClientOutput
	for _, client := range providerClients {
		if client.ClientState.ChainID == consumer.Config().ChainID {
			providerClient = client
			break
		}
	}
	if providerClient == nil {
		return fmt.Errorf("provider chain %s does not have a client tracking the consumer chain %s for path %s on relayer %s",
			provider.Config().ChainID, consumer.Config().ChainID, icsPath, relayer)
	}
	providerClientID := providerClient.ClientID

	if err := relayer.UpdatePath(ctx, rep, icsPath, ibc.PathUpdateOptions{
		SrcClientID: &consumerClientID,
		DstClientID: &providerClientID,
	}); err != nil {
		return err
	}

	if err := relayer.CreateConnections(ctx, rep, icsPath); err != nil {
		return err
	}

	if err := relayer.CreateChannel(ctx, rep, icsPath, ibc.CreateChannelOptions{
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
		ch, err = relayer.GetTransferChannel(ctx, provider, consumer)
		if err == nil && ch != nil {
			break
		} else if err == nil {
			err = fmt.Errorf("channel not found")
		}
		time.Sleep(CommitTimeout)
	}
	return err
}

func (p *Chain) SubmitConsumerAdditionProposal(ctx context.Context, chainID string, config ConsumerConfig, spawnTime time.Time) (*proposalWaiter, chan error, error) {
	propWaiter := newProposalWaiter()
	prop := ccvclient.ConsumerAdditionProposalJSON{
		Title:         fmt.Sprintf("Addition of %s consumer chain", chainID),
		Summary:       "Proposal to add new consumer chain",
		ChainId:       chainID,
		InitialHeight: clienttypes.Height{RevisionNumber: clienttypes.ParseChainID(chainID), RevisionHeight: 1},
		GenesisHash:   []byte("gen_hash"),
		BinaryHash:    []byte("bin_hash"),
		SpawnTime:     spawnTime,

		BlocksPerDistributionTransmission: 1000,
		CcvTimeoutPeriod:                  2419200000000000,
		TransferTimeoutPeriod:             3600000000000,
		ConsumerRedistributionFraction:    "0.75",
		HistoricalEntries:                 10000,
		UnbondingPeriod:                   1728000000000000,
		Deposit:                           strconv.Itoa(GovMinDepositAmount/2) + p.Config().Denom,
	}
	if config.TopN >= 0 {
		prop.TopN = uint32(config.TopN)
	}
	if config.ValidatorSetCap > 0 {
		prop.ValidatorSetCap = uint32(config.ValidatorSetCap)
	}
	if config.ValidatorPowerCap > 0 {
		prop.ValidatorsPowerCap = uint32(config.ValidatorPowerCap)
	}
	propTx, err := p.ConsumerAdditionProposal(ctx, interchaintest.FaucetAccountKeyName, prop)
	if err != nil {
		return nil, nil, err
	}
	errCh := make(chan error)
	go func() {
		defer close(errCh)
		if err := p.WaitForProposalStatus(ctx, propTx.ProposalID, govv1.StatusDepositPeriod); err != nil {
			errCh <- err
			return
		}
		propWaiter.waitForDepositAllowed()

		if _, err := p.GetNode().ExecTx(ctx, interchaintest.FaucetAccountKeyName, "gov", "deposit", propTx.ProposalID, prop.Deposit); err != nil {
			errCh <- err
			return
		}

		if err := p.WaitForProposalStatus(ctx, propTx.ProposalID, govv1.StatusVotingPeriod); err != nil {
			errCh <- err
			return
		}
		propWaiter.startVotingPeriod()
		propWaiter.waitForVoteAllowed()

		if err := p.PassProposal(ctx, propTx.ProposalID); err != nil {
			errCh <- err
			return
		}
		propWaiter.pass()
	}()
	return propWaiter, errCh, nil
}

func (p *Chain) CheckCCV(ctx context.Context, consumer *Chain, relayer *Relayer, amount, valIdx, blocksPerEpoch int) error {
	providerAddress := p.ValidatorWallets[valIdx]

	json, err := p.Validators[valIdx].ReadFile(ctx, "config/priv_validator_key.json")
	if err != nil {
		return err
	}
	providerHex := gjson.GetBytes(json, "address").String()
	json, err = consumer.Validators[valIdx].ReadFile(ctx, "config/priv_validator_key.json")
	if err != nil {
		return err
	}
	consumerHex := gjson.GetBytes(json, "address").String()

	providerPowerBefore, err := p.GetValidatorPower(ctx, providerHex)
	if err != nil {
		return err
	}

	_, err = p.Validators[valIdx].ExecTx(ctx, providerAddress.Moniker,
		"staking", "delegate",
		providerAddress.ValoperAddress, fmt.Sprintf("%d%s", amount, p.Config().Denom),
	)
	if err != nil {
		return err
	}

	if blocksPerEpoch > 1 {
		providerPower, err := p.GetValidatorPower(ctx, providerHex)
		if err != nil {
			return err
		}
		if providerPowerBefore >= providerPower {
			return fmt.Errorf("provider power did not increase after delegation")
		}
		consumerPower, err := consumer.GetValidatorPower(ctx, consumerHex)
		if err != nil {
			return err
		}
		if providerPower == consumerPower {
			return fmt.Errorf("consumer power updated too soon")
		}
		if err := testutil.WaitForBlocks(ctx, blocksPerEpoch, p); err != nil {
			return err
		}
	}

	tCtx, tCancel := context.WithTimeout(ctx, 15*time.Minute)
	defer tCancel()
	var retErr error
	for tCtx.Err() == nil {
		retErr = nil
		providerPower, err := p.GetValidatorPower(ctx, providerHex)
		if err != nil {
			return err
		}
		consumerPower, err := consumer.GetValidatorPower(ctx, consumerHex)
		if err != nil {
			return err
		}
		if providerPowerBefore >= providerPower {
			retErr = fmt.Errorf("provider power did not increase after delegation")
		} else if providerPower != consumerPower {
			retErr = fmt.Errorf("consumer power did not update after provider delegation")
		}
		if retErr == nil {
			break
		}
		time.Sleep(CommitTimeout)
	}
	return retErr
}

func relayerICSPathFor(chainA, chainB *Chain) string {
	return fmt.Sprintf("ics-%s-%s", chainA.Config().ChainID, chainB.Config().ChainID)
}

func (p *Chain) IsValoperJailed(ctx context.Context, valoper string) (bool, error) {
	out, _, err := p.Validators[0].ExecQuery(ctx, "staking", "validator", valoper)
	if err != nil {
		return false, err
	}
	if gjson.GetBytes(out, "jailed").Exists() {
		return gjson.GetBytes(out, "jailed").Bool(), nil
	}
	return gjson.GetBytes(out, "validator.jailed").Bool(), nil
}

func (p *Chain) IsValidatorJailedForConsumerDowntime(ctx context.Context, relayer Relayer, consumer *Chain, validatorIdx int) (jailed bool, err error) {
	if err = consumer.Validators[validatorIdx].StopContainer(ctx); err != nil {
		return
	}
	defer func() {
		err = consumer.Validators[validatorIdx].StartContainer(ctx)
	}()
	channel, err := relayer.GetChannelWithPort(ctx, consumer, p, "consumer")
	if err != nil {
		return
	}
	if err = testutil.WaitForBlocks(ctx, SlashingWindowConsumer+1, consumer); err != nil {
		return
	}
	rs := relayer.Exec(ctx, GetRelayerExecReporter(ctx), []string{
		"hermes", "clear", "packets", "--port", "consumer", "--channel", channel.ChannelID,
		"--chain", consumer.Config().ChainID,
	}, nil)
	if rs.Err != nil {
		return false, rs.Err
	}
	tCtx, tCancel := context.WithTimeout(ctx, 30*CommitTimeout)
	defer tCancel()
	for tCtx.Err() == nil {
		jailed, err = p.IsValoperJailed(ctx, p.ValidatorWallets[validatorIdx].ValoperAddress)
		if err != nil || jailed {
			return
		}
		time.Sleep(CommitTimeout)
	}
	return false, nil
}
