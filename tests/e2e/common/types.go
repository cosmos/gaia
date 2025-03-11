package common

import (
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type Helper struct {
	Suite        *suite.Suite
	TestCounters TestCounters
	Resources    Resources
}

type TestCounters struct {
	ProposalCounter           int
	ContractsCounter          int
	ContractsCounterPerSender map[string]uint64
}

type Resources struct {
	TmpDirs        []string
	ChainA         *Chain
	ChainB         *Chain
	DkrPool        *dockertest.Pool
	DkrNet         *dockertest.Network
	HermesResource *dockertest.Resource

	ValResources map[string][]*dockertest.Resource
}
