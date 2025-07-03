package telemetry

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/pborman/uuid"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"cosmossdk.io/log"

	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/version"
)

const (
	meterName   = "gaia"
	serviceName = "cosmoshub"
)

type (
	ValidatorInfo struct {
		ChainID     string
		IsValidator bool
		Moniker     string
		Address     crypto.Address
	}

	OtelClient struct {
		cfg      OtelConfig
		vi       ValidatorInfo
		gatherer prometheus.Gatherer
	}
)

type OtelOption func(*OtelClient)

func WithGatherer(g prometheus.Gatherer) OtelOption {
	return func(o *OtelClient) {
		o.gatherer = g
	}
}

func NewOtelClient(otelConfig OtelConfig, vi ValidatorInfo, opts ...OtelOption) *OtelClient {
	if vi.Moniker == "" {
		vi.Moniker = "UNKNOWN-" + uuid.NewUUID().String()
	}
	client := &OtelClient{
		cfg:      otelConfig,
		vi:       vi,
		gatherer: prometheus.DefaultGatherer,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (o *OtelClient) StartExporter(logger log.Logger) error {
	cfg := o.cfg
	if cfg.Disable {
		logger.Debug("otlp exporter is disabled")
		return nil
	}
	logger.Debug("starting otlp exporter")
	ctx := context.Background()

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(cfg.CollectorEndpoint),
		otlpmetrichttp.WithURLPath(cfg.CollectorMetricsURLPath),
	}
	if cfg.User != "" && cfg.Token != "" {
		opts = append(opts, otlpmetrichttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + formatBasicAuth(cfg.User, cfg.Token),
		}))
	} else if strings.HasPrefix(cfg.CollectorEndpoint, "localhost") || strings.HasPrefix(cfg.CollectorEndpoint, "host.docker.internal") {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("OTLP exporter setup failed: %w", err)
	}

	res, _ := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName(fmt.Sprintf("%s.%s", serviceName, o.vi.ChainID)),
		semconv.ServiceVersion(version.Version),
	))

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(cfg.PushInterval))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	meter := otel.Meter(meterName)

	go func() {
		gauges := make(map[string]otmetric.Float64Gauge)
		histograms := make(map[string]otmetric.Float64Histogram)
		ticker := time.NewTicker(cfg.PushInterval)
		for range ticker.C {
			if err := o.scrapePrometheusMetrics(ctx, logger, meter, gauges, histograms); err != nil {
				logger.Debug("error scraping metrics", "error", err)
			}
		}
	}()
	return nil
}

func (o *OtelClient) SetValidatorStatus(isVal bool) {
	o.vi.IsValidator = isVal
}

func (o *OtelClient) IsValidator() bool {
	return o.vi.IsValidator
}

func (o *OtelClient) GetValAddr() crypto.Address {
	return o.vi.Address
}

func (o *OtelClient) Enabled() bool {
	return !o.cfg.Disable
}

func (o *OtelClient) scrapePrometheusMetrics(ctx context.Context, logger log.Logger, meter otmetric.Meter, gauges map[string]otmetric.Float64Gauge, histograms map[string]otmetric.Float64Histogram) error {
	if !o.vi.IsValidator {
		return nil
	}
	metricFamilies, err := o.gatherer.Gather()
	if err != nil {
		logger.Debug("failed to gather prometheus metrics", "error", err)
		return err
	}

	monikerAttr := []attribute.KeyValue{
		{Key: "moniker", Value: attribute.StringValue(o.vi.Moniker)},
	}

	for _, mf := range metricFamilies {
		name := mf.GetName()
		for _, m := range mf.Metric {
			attrs := make([]attribute.KeyValue, 0, len(m.Label)+len(monikerAttr)+1)
			attrs = append(attrs, monikerAttr...)
			attrs = append(attrs, attribute.String("chain_id", o.vi.ChainID))

			for _, label := range m.Label {
				attrs = append(attrs, attribute.String(label.GetName(), label.GetValue()))
			}
			switch mf.GetType() {
			case dto.MetricType_GAUGE:
				recordGauge(ctx, logger, meter, gauges, name, mf.GetHelp(), m.Gauge.GetValue(), attrs)

			case dto.MetricType_COUNTER:
				recordGauge(ctx, logger, meter, gauges, name, mf.GetHelp(), m.Counter.GetValue(), attrs)

			case dto.MetricType_HISTOGRAM:
				recordHistogram(ctx, logger, meter, histograms, name, mf.GetHelp(), m.Histogram, attrs)

			case dto.MetricType_SUMMARY:
				recordSummary(ctx, logger, meter, gauges, name, mf.GetHelp(), m.Summary, attrs)

			default:
				continue
			}
		}
	}

	return nil
}

func recordGauge(ctx context.Context, logger log.Logger, meter otmetric.Meter, gauges map[string]otmetric.Float64Gauge, name, help string, val float64, attrs []attribute.KeyValue) {
	g, ok := gauges[name]
	if !ok {
		var err error
		g, err = meter.Float64Gauge(name, otmetric.WithDescription(help))
		if err != nil {
			logger.Debug("failed to create gauge", "name", name, "error", err)
			return
		}
		gauges[name] = g
	}
	g.Record(ctx, val, otmetric.WithAttributes(attrs...))
}

func recordHistogram(ctx context.Context, logger log.Logger, meter otmetric.Meter, histograms map[string]otmetric.Float64Histogram, name, help string, h *dto.Histogram, monikerAttrs []attribute.KeyValue) {
	boundaries := make([]float64, 0, len(h.Bucket)) // excluding +Inf
	bucketCounts := make([]uint64, 0, len(h.Bucket))

	for _, bucket := range h.Bucket {
		if math.IsInf(bucket.GetUpperBound(), +1) {
			continue // Skip +Inf bucket boundary explicitly
		}
		boundaries = append(boundaries, bucket.GetUpperBound())
		bucketCounts = append(bucketCounts, bucket.GetCumulativeCount())
	}

	hist, ok := histograms[name]
	if !ok {
		var err error
		hist, err = meter.Float64Histogram(
			name,
			otmetric.WithDescription(help),
			otmetric.WithExplicitBucketBoundaries(boundaries...),
		)
		if err != nil {
			logger.Debug("failed to create histogram", "name", name, "error", err)
			return
		}
		histograms[name] = hist
	}

	prevCount := uint64(0)
	for i, count := range bucketCounts {
		countInBucket := count - prevCount
		prevCount = count

		// explicitly record the mid-point of the bucket as approximation:
		var value float64
		if i == 0 {
			value = boundaries[0] / 2.0
		} else {
			value = (boundaries[i-1] + boundaries[i]) / 2.0
		}

		// record `countInBucket` number of observations with moniker attributes
		for j := uint64(0); j < countInBucket; j++ {
			hist.Record(ctx, value, otmetric.WithAttributes(monikerAttrs...))
		}
	}
}

func recordSummary(ctx context.Context, logger log.Logger, meter otmetric.Meter, gauges map[string]otmetric.Float64Gauge, name, help string, s *dto.Summary, monikerAttrs []attribute.KeyValue) {
	recordGauge(ctx, logger, meter, gauges, name+"_sum", help+" (summary sum)", s.GetSampleSum(), monikerAttrs)
	recordGauge(ctx, logger, meter, gauges, name+"_count", help+" (summary count)", float64(s.GetSampleCount()), monikerAttrs)

	for _, q := range s.Quantile {
		// combine moniker attrs with quantile attr
		attrs := make([]attribute.KeyValue, len(monikerAttrs)+1)
		copy(attrs, monikerAttrs)
		attrs[len(monikerAttrs)] = attribute.String("quantile", fmt.Sprintf("%v", q.GetQuantile()))

		recordGauge(ctx, logger, meter, gauges, name, help+" (summary quantile)", q.GetValue(), attrs)
	}
}

func formatBasicAuth(username, token string) string {
	auth := username + ":" + token
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
