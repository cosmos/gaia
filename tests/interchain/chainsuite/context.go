package chainsuite

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type testReporterKey struct{}

type relayerExecReporterKey struct{}

type loggerKey struct{}

type dockerKey struct{}

type dockerContext struct {
	NetworkID string
	Client    *client.Client
}

func WithDockerContext(ctx context.Context, d *dockerContext) context.Context {
	return context.WithValue(ctx, dockerKey{}, d)
}

func GetDockerContext(ctx context.Context) (*client.Client, string) {
	d, _ := ctx.Value(dockerKey{}).(*dockerContext)
	return d.Client, d.NetworkID
}

func WithTestReporter(ctx context.Context, r *testreporter.Reporter) context.Context {
	return context.WithValue(ctx, testReporterKey{}, r)
}

func GetTestReporter(ctx context.Context) *testreporter.Reporter {
	r, _ := ctx.Value(testReporterKey{}).(*testreporter.Reporter)
	return r
}

func WithRelayerExecReporter(ctx context.Context, r *testreporter.RelayerExecReporter) context.Context {
	return context.WithValue(ctx, relayerExecReporterKey{}, r)
}

func GetRelayerExecReporter(ctx context.Context) *testreporter.RelayerExecReporter {
	r, _ := ctx.Value(relayerExecReporterKey{}).(*testreporter.RelayerExecReporter)
	return r
}

func WithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func GetLogger(ctx context.Context) *zap.Logger {
	l, _ := ctx.Value(loggerKey{}).(*zap.Logger)
	return l
}

func NewSuiteContext(s *suite.Suite) (context.Context, error) {
	ctx := context.Background()

	dockerClient, dockerNetwork := interchaintest.DockerSetup(s.T())
	dockerContext := &dockerContext{
		NetworkID: dockerNetwork,
		Client:    dockerClient,
	}
	ctx = WithDockerContext(ctx, dockerContext)
	logger := zaptest.NewLogger(s.T())
	ctx = WithLogger(ctx, logger)

	f, err := interchaintest.CreateLogFile(fmt.Sprintf("%d.json", time.Now().Unix()))
	if err != nil {
		return nil, err
	}
	testReporter := testreporter.NewReporter(f)
	ctx = WithTestReporter(ctx, testReporter)
	relayerExecReporter := testReporter.RelayerExecReporter(s.T())
	ctx = WithRelayerExecReporter(ctx, relayerExecReporter)
	return ctx, nil
}
