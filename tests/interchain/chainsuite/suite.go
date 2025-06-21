package chainsuite

import (
	"context"
	"reflect"
	"runtime"
	"strings"

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
	suiteName := s.getSuiteName()
	s.T().Logf("=== STARTING INTERCHAIN TEST SUITE: %s ===", suiteName)

	if s.Config.Scope == ChainScopeSuite {
		s.createChain()
	}

	s.T().Logf("=== SUITE SETUP COMPLETED: %s ===", suiteName)
}

func (s *Suite) TearDownSuite() {
	suiteName := s.getSuiteName()
	s.T().Logf("=== TEARDOWN SUITE: %s ===", suiteName)
}

func (s *Suite) TearDownTest() {
	suiteName := s.getSuiteName()
	s.T().Logf("=== TEARDOWN TEST: %s ===", suiteName)
}

func (s *Suite) BeforeTest(suiteName, testName string) {
	s.T().Logf("=== STARTING INTERCHAIN TEST: %s.%s ===", suiteName, testName)
}

func (s *Suite) AfterTest(suiteName, testName string) {
	if s.T().Failed() {
		s.T().Logf("=== FAILED INTERCHAIN TEST: %s.%s ===", suiteName, testName)
	} else {
		s.T().Logf("=== PASSED INTERCHAIN TEST: %s.%s ===", suiteName, testName)
	}
}

func (s *Suite) getSuiteName() string {
	// Get the suite type name through reflection
	suiteType := reflect.TypeOf(s.Suite).Elem()
	if suiteType.Kind() == reflect.Struct {
		return suiteType.Name()
	}

	// Fallback: try to extract from the test name
	if s.T() != nil && s.T().Name() != "" {
		// Extract suite name from test name (format: TestSuiteName/TestMethodName)
		parts := strings.Split(s.T().Name(), "/")
		if len(parts) > 0 {
			return strings.TrimPrefix(parts[0], "Test")
		}
	}

	// Final fallback: get from runtime caller
	pc, _, _, ok := runtime.Caller(2)
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			name := fn.Name()
			parts := strings.Split(name, ".")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}

	return "Unknown"
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
