package chainsuite

import (
	"context"
	"os"
	"path"

	"github.com/stretchr/testify/suite"
	"golang.org/x/mod/semver"
)

type Suite struct {
	suite.Suite
	Config SuiteConfig
	Env    Environment
	Chain  *Chain
	ctx    context.Context
}

func NewSuite(config SuiteConfig) *Suite {
	env := GetEnvironment()
	defaultConfig := DefaultSuiteConfig(env)
	newCfg := defaultConfig.Merge(config)
	return &Suite{Config: newCfg, Env: env}
}

func (s *Suite) SetupSuite() {
	cwd, err := os.Getwd()
	s.Require().NoError(err)
	err = os.Setenv("IBCTEST_CONFIGURED_CHAINS", path.Join(cwd, "configuredChains.yaml"))
	s.Require().NoError(err)

	ctx, err := NewSuiteContext(&s.Suite)
	s.Require().NoError(err)
	s.ctx = ctx
	s.Chain, err = CreateChain(s.GetContext(), s.T(), s.Config.ChainSpec)
	s.Require().NoError(err)
	if s.Config.UpgradeOnSetup {
		s.UpgradeChain()
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
}
