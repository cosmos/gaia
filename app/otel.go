package gaia

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"time"

	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const meterName = "cosmos-sdk-otlp-exporter"

type (
	OtelConfig struct {
		OtlpCollectorEndpoint       string        `mapstructure:"otlp-collector-endpoint"`
		OtlpCollectorMetricsURLPath string        `mapstructure:"otlp-collector-metrics-url-path"`
		OtlpUser                    string        `mapstructure:"otlp-user"`
		OtlpToken                   string        `mapstructure:"otlp-token"`
		OtlpServiceName             string        `mapstructure:"otlp-service-name"`
		OtlpPushInterval            time.Duration `mapstructure:"otlp-push-interval"`
	}

	ValidatorInfo struct {
		IsValidator bool
		Moniker     string
		Address     crypto.Address
	}

	OtelClient struct {
		vi ValidatorInfo
		sk *stakingkeeper.Keeper
	}
)

var DefaultOtelConfig = OtelConfig{
	OtlpCollectorEndpoint:       "localhost:4318",
	OtlpCollectorMetricsURLPath: "/v1/metrics",
	OtlpUser:                    "", // empty for local testing
	OtlpToken:                   "", // empty for local testing
	OtlpServiceName:             "gaia-test",
	OtlpPushInterval:            10 * time.Second,
}

func NewOtelClient(sk *stakingkeeper.Keeper, vi ValidatorInfo) *OtelClient {
	return &OtelClient{
		sk: sk,
		vi: vi,
	}
}

func (o *OtelClient) StartExporter(cfg OtelConfig) error {
	ctx := context.Background()

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(cfg.OtlpCollectorEndpoint),
		otlpmetrichttp.WithURLPath(cfg.OtlpCollectorMetricsURLPath),
	}
	if cfg.OtlpUser != "" && cfg.OtlpToken != "" {
		opts = append(opts, otlpmetrichttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + formatBasicAuth(cfg.OtlpUser, cfg.OtlpToken),
		}))
	} else {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	exporter, err := otlpmetrichttp.New(ctx,
		opts...,
	)
	if err != nil {
		log.Fatalf("OTLP exporter setup failed: %v", err)
	}

	res, _ := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName(cfg.OtlpServiceName),
	))

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter,
			metric.WithInterval(cfg.OtlpPushInterval))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	meter := otel.Meter(meterName)

	go func() {
		gauges := make(map[string]otmetric.Float64Gauge)
		histograms := make(map[string]otmetric.Float64Histogram)
		ticker := time.NewTicker(cfg.OtlpPushInterval)
		for {
			select {
			case <-ticker.C:
				if err := o.scrapePrometheusMetrics(ctx, meter, gauges, histograms); err != nil {
					log.Printf("error scraping metrics: %v", err)
				}
			}
		}
	}()
	return nil
}

func (o *OtelClient) SetValidatorStatus(isVal bool) {
	o.vi.IsValidator = isVal
}

func (o *OtelClient) GetValAddr() crypto.Address {
	return o.vi.Address
}

func (o *OtelClient) scrapePrometheusMetrics(ctx context.Context, meter otmetric.Meter, gauges map[string]otmetric.Float64Gauge, histograms map[string]otmetric.Float64Histogram) error {
	if !o.vi.IsValidator {
		return nil
	}
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		log.Printf("failed to gather prometheus metrics: %v", err)
		return err
	}

	for _, mf := range metricFamilies {
		name := mf.GetName()
		for _, m := range mf.Metric {
			switch mf.GetType() {
			case dto.MetricType_GAUGE:
				recordGauge(ctx, meter, gauges, name, mf.GetHelp(), m.Gauge.GetValue(), nil)

			case dto.MetricType_COUNTER:
				recordGauge(ctx, meter, gauges, name, mf.GetHelp(), m.Counter.GetValue(), nil)

			case dto.MetricType_HISTOGRAM:
				recordHistogram(ctx, meter, histograms, name, mf.GetHelp(), m.Histogram)

			case dto.MetricType_SUMMARY:
				recordSummary(ctx, meter, gauges, name, mf.GetHelp(), m.Summary)

			default:
				continue
			}
		}
	}

	return nil
}

func recordGauge(ctx context.Context, meter otmetric.Meter, gauges map[string]otmetric.Float64Gauge, name, help string, val float64, attrs []attribute.KeyValue) {
	g, ok := gauges[name]
	if !ok {
		var err error
		g, err = meter.Float64Gauge(name, otmetric.WithDescription(help))
		if err != nil {
			log.Printf("failed to create gauge %q: %v", name, err)
			return
		}
		gauges[name] = g
	}
	g.Record(ctx, val, otmetric.WithAttributes(attrs...))
}

func recordHistogram(ctx context.Context, meter otmetric.Meter, histograms map[string]otmetric.Float64Histogram, name, help string, h *dto.Histogram) {
	boundaries := make([]float64, 0, len(h.Bucket)-1) // excluding +Inf
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
			log.Printf("failed to create histogram %s: %v", name, err)
			return
		}
		histograms[name] = hist
	}

	prevCount := uint64(0)
	for i, count := range bucketCounts {
		countInBucket := count - prevCount
		prevCount = count

		// Explicitly record the mid-point of the bucket as approximation:
		var value float64
		if i == 0 {
			value = boundaries[0] / 2.0
		} else {
			value = (boundaries[i-1] + boundaries[i]) / 2.0
		}

		// Record `countInBucket` number of observations explicitly (approximation):
		for j := uint64(0); j < countInBucket; j++ {
			hist.Record(ctx, value)
		}
	}
}

func recordSummary(ctx context.Context, meter otmetric.Meter, gauges map[string]otmetric.Float64Gauge, name, help string, s *dto.Summary) {
	recordGauge(ctx, meter, gauges, name+"_sum", help+" (summary sum)", s.GetSampleSum(), nil)
	recordGauge(ctx, meter, gauges, name+"_count", help+" (summary count)", float64(s.GetSampleCount()), nil)

	for _, q := range s.Quantile {
		attrs := []attribute.KeyValue{
			attribute.String("quantile", fmt.Sprintf("%v", q.GetQuantile())),
		}
		recordGauge(ctx, meter, gauges, name, help+" (summary quantile)", q.GetValue(), attrs)
	}
}

func formatBasicAuth(username, token string) string {
	auth := username + ":" + token
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
