package chainsuite

import (
	"context"

	"github.com/stretchr/testify/suite"
	"golang.org/x/mod/semver"
)

type Suite struct {
	suite.Suite
	Config  SuiteConfig
	Env     Environment
	Chain   *Chain
	Relayer *Relayer
	ctx     context.Context
}

func NewSuite(config SuiteConfig) *Suite {
	env := GetEnvironment()
	defaultConfig := DefaultSuiteConfig(env)
	newCfg := defaultConfig.Merge(config)
	return &Suite{Config: newCfg, Env: env}
}

func (s *Suite) createChain() {
	ctx, err := NewSuiteContext(&s.Suite)
	s.Require().NoError(err)
	s.ctx = ctx
	s.Chain, err = CreateChain(s.GetContext(), s.T(), s.Config.ChainSpec)
	s.Require().NoError(err)
	if s.Config.CreateRelayer {
		s.Relayer, err = NewRelayer(s.GetContext(), s.T())
		s.Require().NoError(err)
		err = s.Relayer.SetupChainKeys(s.GetContext(), s.Chain)
		s.Require().NoError(err)
	}
	if s.Config.UpgradeOnSetup {
		s.UpgradeChain()
	}
}

func (s *Suite) SetupTest() {
	if s.Config.Scope == ChainScopeTest {
		s.createChain()
	}
}

func (s *Suite) SetupSuite() {
	if s.Config.Scope == ChainScopeSuite {
		s.createChain()
	}
}

func (s *Suite) GetContext() context.Context {
	s.Require().NotNil(s.ctx, "Tried to GetContext before it was set. SetupSuite must run first")
	return s.ctx
}

func (s *Suite) UpgradeChain() {
	GetLogger(s.GetContext()).Sugar().Infof("Upgrade %s from %s to %s", s.Env.UpgradeName, s.Env.OldGaiaImageVersion, s.Env.NewGaiaImageVersion)
	if s.Env.UpgradeName == semver.Major(s.Env.OldGaiaImageVersion) {
		// Not an on-chain upgrade, just replace the image.
		s.Require().NoError(s.Chain.ReplaceImagesAndRestart(s.GetContext(), s.Env.NewGaiaImageVersion))
	} else {
		s.Require().NoError(s.Chain.Upgrade(s.GetContext(), s.Env.UpgradeName, s.Env.NewGaiaImageVersion))
	}
	if s.Relayer != nil {
		s.Require().NoError(s.Relayer.StopRelayer(s.GetContext(), GetRelayerExecReporter(s.GetContext())))
		s.Require().NoError(s.Relayer.StartRelayer(s.GetContext(), GetRelayerExecReporter(s.GetContext())))
	}
}
