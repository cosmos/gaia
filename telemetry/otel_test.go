package telemetry

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	otmetric "go.opentelemetry.io/otel/metric"

	"cosmossdk.io/log"
)

func TestSetValidatorStatus(t *testing.T) {
	oc := OtelClient{vi: ValidatorInfo{IsValidator: false}}

	oc.SetValidatorStatus(true)
	require.True(t, oc.vi.IsValidator)
	oc.SetValidatorStatus(false)
	require.False(t, oc.vi.IsValidator)
}

func TestIsValidator(t *testing.T) {
	oc := OtelClient{vi: ValidatorInfo{IsValidator: true}}
	require.True(t, oc.IsValidator())
	oc.vi.IsValidator = false
	require.False(t, oc.IsValidator())
}

func TestGetValAddr(t *testing.T) {
	addr := bytes.HexBytes("hello")
	oc := OtelClient{vi: ValidatorInfo{Address: addr}}
	gotAddr := oc.GetValAddr()
	require.Equal(t, addr, gotAddr)
}

func TestEnabled(t *testing.T) {
	oc := OtelClient{cfg: OtelConfig{Disable: true}}
	require.False(t, oc.Enabled())
	oc.cfg.Disable = false
	require.True(t, oc.Enabled())
}

func TestMonikerDefault(t *testing.T) {
	oc := NewOtelClient(OtelConfig{}, ValidatorInfo{})
	require.Greater(t, len(oc.vi.Moniker), len("UNKNOWN-1"))
}

func TestScrapePrometheusMetrics_GaugeIsRecorded(t *testing.T) {
	reg := prometheus.NewRegistry()
	attrs := createTestAttributes()
	// create and register a test gauge metric
	name := "test_metric_gauge"
	help := "A test gauge metric"
	testGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})
	gaugeValue := 3.1415
	testGauge.Set(gaugeValue)
	require.NoError(t, reg.Register(testGauge))

	// setup OtelClient with registry and IsValidator=true
	client := &OtelClient{
		cfg:      OtelConfig{Disable: false},
		vi:       ValidatorInfo{IsValidator: true, Moniker: "test-validator"},
		gatherer: reg,
	}

	ctx := context.Background()
	logger := log.NewNopLogger()

	mockMeter := &MockMeter{}
	mockGauge := &MockFloat64Gauge{}
	gauges := make(map[string]otmetric.Float64Gauge)
	histograms := make(map[string]otmetric.Float64Histogram)

	// expect gauge to be created and recorded
	mockMeter.On("Float64Gauge", name, []otmetric.Float64GaugeOption{otmetric.WithDescription(help)}).Return(mockGauge, nil)
	mockGauge.On("Record", ctx, gaugeValue, []otmetric.RecordOption{otmetric.WithAttributes(attrs...)}).Return()

	err := client.scrapePrometheusMetrics(ctx, logger, mockMeter, gauges, histograms)
	require.NoError(t, err)

	mockMeter.AssertExpectations(t)
	mockGauge.AssertExpectations(t)

	// verify gauge was added to the map
	require.Contains(t, gauges, "test_metric_gauge")
}

// tests for recordGauge function
func TestRecordGauge(t *testing.T) {
	ctx := context.Background()
	mockLogger := log.NewNopLogger()
	mockMeter := &MockMeter{}
	mockGauge := &MockFloat64Gauge{}

	gauges := make(map[string]otmetric.Float64Gauge)
	attrs := createTestAttributes()

	t.Run("creates new gauge when not exists", func(t *testing.T) {
		gaugeName := "test_gauge"
		help := "Test gauge description"
		value := 42.5
		gaugeOpt := []otmetric.Float64GaugeOption{otmetric.WithDescription(help)}
		recOpts := []otmetric.RecordOption{otmetric.WithAttributes(attrs...)}
		mockMeter.On("Float64Gauge", gaugeName, gaugeOpt).Return(mockGauge, nil)
		mockGauge.On("Record", ctx, value, recOpts).Return()

		recordGauge(ctx, mockLogger, mockMeter, gauges, gaugeName, help, value, attrs)

		// verify the gauge was added to the map
		require.Contains(t, gauges, gaugeName)
		require.Equal(t, mockGauge, gauges[gaugeName])

		mockMeter.AssertExpectations(t)
		mockGauge.AssertExpectations(t)
	})

	t.Run("reuses existing gauge", func(t *testing.T) {
		gaugeName := "existing_gauge"
		help := "Existing gauge description"
		value := 123.45

		// pre-populate the gauge in the map
		existingGauge := &MockFloat64Gauge{}
		gauges[gaugeName] = existingGauge

		existingGauge.On("Record", ctx, value, []otmetric.RecordOption{otmetric.WithAttributes(attrs...)}).Return()

		recordGauge(ctx, mockLogger, mockMeter, gauges, gaugeName, help, value, attrs)

		// verify no new gauge was created (meter should not be called)
		existingGauge.AssertExpectations(t)
	})

	t.Run("handles gauge creation error", func(t *testing.T) {
		gaugeName := "error_gauge"
		help := "Error gauge description"
		value := 99.9
		gaugeError := fmt.Errorf("failed to create gauge")

		mockMeter.On("Float64Gauge", gaugeName, []otmetric.Float64GaugeOption{otmetric.WithDescription(help)}).Return((*MockFloat64Gauge)(nil), gaugeError)

		recordGauge(ctx, mockLogger, mockMeter, gauges, gaugeName, help, value, attrs)

		// verify the gauge was not added to the map
		require.NotContains(t, gauges, gaugeName)

		mockMeter.AssertExpectations(t)
	})
}

// tests for recordHistogram
func TestRecordHistogram(t *testing.T) {
	ctx := context.Background()
	mockLogger := log.NewNopLogger()
	mockMeter := &MockMeter{}
	mockHistogram := &MockFloat64Histogram{}

	histograms := make(map[string]otmetric.Float64Histogram)
	attrs := createTestAttributes()

	t.Run("creates and records histogram with multiple buckets", func(t *testing.T) {
		histName := "test_histogram"
		help := "Test histogram description"

		// create test histogram data with multiple buckets
		upperBounds := []float64{1.0, 5.0, 10.0}
		hist := &dto.Histogram{
			Bucket: []*dto.Bucket{
				{UpperBound: floatPtr(1.0), CumulativeCount: uintPtr(5)},
				{UpperBound: floatPtr(5.0), CumulativeCount: uintPtr(10)},
				{UpperBound: floatPtr(10.0), CumulativeCount: uintPtr(15)},
				{UpperBound: floatPtr(math.Inf(1)), CumulativeCount: uintPtr(20)}, // +Inf bucket
			},
		}

		mockMeter.On("Float64Histogram", histName, []otmetric.Float64HistogramOption{otmetric.WithDescription(help), otmetric.WithExplicitBucketBoundaries(upperBounds...)}).Return(mockHistogram, nil)

		// expect multiple Record calls based on bucket counts
		// bucket 1: 5 records at value 0.5 (mid-point of 0 to 1.0)
		mockHistogram.On("Record", ctx, 0.5, []otmetric.RecordOption{otmetric.WithAttributes(attrs...)}).Return().Times(5)
		// bucket 2: 5 records at value 3.0 (mid-point of 1.0 to 5.0)
		mockHistogram.On("Record", ctx, 3.0, []otmetric.RecordOption{otmetric.WithAttributes(attrs...)}).Return().Times(5)
		// bucket 3: 5 records at value 7.5 (mid-point of 5.0 to 10.0)
		mockHistogram.On("Record", ctx, 7.5, []otmetric.RecordOption{otmetric.WithAttributes(attrs...)}).Return().Times(5)

		recordHistogram(ctx, mockLogger, mockMeter, histograms, histName, help, hist, attrs)

		// verify the histogram was added to the map
		require.Contains(t, histograms, histName)

		mockMeter.AssertExpectations(t)
		mockHistogram.AssertExpectations(t)
	})

	t.Run("reuses existing histogram", func(t *testing.T) {
		histName := "existing_histogram"
		help := "Existing histogram description"

		// pre-populate the histogram in the map
		existingHist := &MockFloat64Histogram{}
		histograms[histName] = existingHist

		hist := &dto.Histogram{
			Bucket: []*dto.Bucket{
				{UpperBound: floatPtr(2.0), CumulativeCount: uintPtr(3)},
				{UpperBound: floatPtr(math.Inf(1)), CumulativeCount: uintPtr(3)},
			},
		}

		existingHist.On("Record", ctx, 1.0, []otmetric.RecordOption{otmetric.WithAttributes(attrs...)}).Return().Times(3)

		recordHistogram(ctx, mockLogger, mockMeter, histograms, histName, help, hist, attrs)

		existingHist.AssertExpectations(t)
	})

	t.Run("handles histogram creation error", func(t *testing.T) {
		histName := "error_histogram"
		help := "Error histogram description"
		histError := fmt.Errorf("failed to create histogram")

		hist := &dto.Histogram{
			Bucket: []*dto.Bucket{
				{UpperBound: floatPtr(1.0), CumulativeCount: uintPtr(1)},
			},
		}

		mockMeter.On("Float64Histogram", histName, []otmetric.Float64HistogramOption{otmetric.WithDescription(help), otmetric.WithExplicitBucketBoundaries(1.0)}).Return((*MockFloat64Histogram)(nil), histError)

		recordHistogram(ctx, mockLogger, mockMeter, histograms, histName, help, hist, attrs)

		// verify the histogram was not added to the map
		require.NotContains(t, histograms, histName)

		mockMeter.AssertExpectations(t)
	})

	t.Run("handles empty histogram", func(t *testing.T) {
		histName := "empty_histogram"
		help := "Empty histogram description"

		hist := &dto.Histogram{
			Bucket: []*dto.Bucket{},
		}

		mockMeter.On("Float64Histogram", histName, []otmetric.Float64HistogramOption{otmetric.WithDescription(help), otmetric.WithExplicitBucketBoundaries([]float64{}...)}).Return(mockHistogram, nil)

		recordHistogram(ctx, mockLogger, mockMeter, histograms, histName, help, hist, attrs)

		// verify the histogram was added but no records were made
		require.Contains(t, histograms, histName)

		mockMeter.AssertExpectations(t)
		// mockHistogram should have no Record calls
	})
}

func copyWith[T any](a []T, t T) []T {
	b := make([]T, len(a))
	copy(b, a)
	b = append(b, t)
	return b
}

// tests for recordSummary function
func TestRecordSummary(t *testing.T) {
	ctx := context.Background()
	mockLogger := log.NewNopLogger()
	mockMeter := &MockMeter{}

	gauges := make(map[string]otmetric.Float64Gauge)
	attrs := createTestAttributes()

	t.Run("records summary with quantiles", func(t *testing.T) {
		summaryName := "test_summary"
		help := "Test summary description"

		summary := &dto.Summary{
			SampleCount: uintPtr(100),
			SampleSum:   floatPtr(500.5),
			Quantile: []*dto.Quantile{
				{Quantile: floatPtr(0.5), Value: floatPtr(5.0)},   // median
				{Quantile: floatPtr(0.95), Value: floatPtr(9.5)},  // 95th percentile
				{Quantile: floatPtr(0.99), Value: floatPtr(10.0)}, // 99th percentile
			},
		}

		// mock gauges for sum, count, and quantiles
		mockSumGauge := &MockFloat64Gauge{}
		mockCountGauge := &MockFloat64Gauge{}
		mockQuantileGauge := &MockFloat64Gauge{}

		// expect gauge creation for sum
		mockMeter.On("Float64Gauge", summaryName+"_sum", []otmetric.Float64GaugeOption{otmetric.WithDescription(help + " (summary sum)")}).Return(mockSumGauge, nil)
		mockSumGauge.On("Record", ctx, 500.5, []otmetric.RecordOption{otmetric.WithAttributes(copyWith(attrs, attribute.String("quantile", "0.5"))...)}).Return()

		// expect gauge creation for count
		mockMeter.On("Float64Gauge", summaryName+"_count", []otmetric.Float64GaugeOption{otmetric.WithDescription(help + " (summary count)")}).Return(mockCountGauge, nil)
		mockCountGauge.On("Record", ctx, 100.0, []otmetric.RecordOption{otmetric.WithAttributes(copyWith(attrs, attribute.String("quantile", "0.5"))...)}).Return()

		// expect gauge creation for quantiles (reused for all quantiles)
		attrs = append(attrs,
			attribute.String("quantile", "0.5"),
		)
		mockMeter.On("Float64Gauge", summaryName, []otmetric.Float64GaugeOption{otmetric.WithDescription(help + " (summary quantile)")}).Return(mockQuantileGauge, nil)
		mockQuantileGauge.On("Record", ctx, 5.0, []otmetric.RecordOption{otmetric.WithAttributes(copyWith(attrs, attribute.String("quantile", "0.5"))...)}).Return()   // 0.5 quantile
		mockQuantileGauge.On("Record", ctx, 9.5, []otmetric.RecordOption{otmetric.WithAttributes(copyWith(attrs, attribute.String("quantile", "0.95"))...)}).Return()  // 0.95 quantile
		mockQuantileGauge.On("Record", ctx, 10.0, []otmetric.RecordOption{otmetric.WithAttributes(copyWith(attrs, attribute.String("quantile", "0.99"))...)}).Return() // 0.99 quantile

		recordSummary(ctx, mockLogger, mockMeter, gauges, summaryName, help, summary, attrs)

		// verify all gauges were added to the map
		require.Contains(t, gauges, summaryName+"_sum")
		require.Contains(t, gauges, summaryName+"_count")
		require.Contains(t, gauges, summaryName)

		mockMeter.AssertExpectations(t)
		mockSumGauge.AssertExpectations(t)
		mockCountGauge.AssertExpectations(t)
		mockQuantileGauge.AssertExpectations(t)
	})

	t.Run("records summary without quantiles", func(t *testing.T) {
		summaryName := "simple_summary"
		help := "Simple summary description"

		summary := &dto.Summary{
			SampleCount: uintPtr(50),
			SampleSum:   floatPtr(250.0),
			Quantile:    []*dto.Quantile{}, // no quantiles
		}

		mockSumGauge := &MockFloat64Gauge{}
		mockCountGauge := &MockFloat64Gauge{}

		mockMeter.On("Float64Gauge", summaryName+"_sum", []otmetric.Float64GaugeOption{otmetric.WithDescription(help + " (summary sum)")}).Return(mockSumGauge, nil)
		mockSumGauge.On("Record", ctx, 250.0, []otmetric.RecordOption{otmetric.WithAttributes(attrs...)}).Return()

		mockMeter.On("Float64Gauge", summaryName+"_count", []otmetric.Float64GaugeOption{otmetric.WithDescription(help + " (summary count)")}).Return(mockCountGauge, nil)
		mockCountGauge.On("Record", ctx, 50.0, []otmetric.RecordOption{otmetric.WithAttributes(attrs...)}).Return()

		recordSummary(ctx, mockLogger, mockMeter, gauges, summaryName, help, summary, attrs)

		// verify only sum and count gauges were added
		require.Contains(t, gauges, summaryName+"_sum")
		require.Contains(t, gauges, summaryName+"_count")

		mockMeter.AssertExpectations(t)
		mockSumGauge.AssertExpectations(t)
		mockCountGauge.AssertExpectations(t)
	})

	t.Run("handles zero values", func(t *testing.T) {
		summaryName := "zero_summary"
		help := "Zero summary description"

		summary := &dto.Summary{
			SampleCount: uintPtr(0),
			SampleSum:   floatPtr(0.0),
			Quantile: []*dto.Quantile{
				{Quantile: floatPtr(0.5), Value: floatPtr(0.0)},
			},
		}

		mockSumGauge := &MockFloat64Gauge{}
		mockCountGauge := &MockFloat64Gauge{}
		mockQuantileGauge := &MockFloat64Gauge{}

		mockMeter.On("Float64Gauge", summaryName+"_sum", []otmetric.Float64GaugeOption{otmetric.WithDescription(help + " (summary sum)")}).Return(mockSumGauge, nil)
		mockSumGauge.On("Record", ctx, 0.0, []otmetric.RecordOption{otmetric.WithAttributes(attrs...)}).Return()

		mockMeter.On("Float64Gauge", summaryName+"_count", []otmetric.Float64GaugeOption{otmetric.WithDescription(help + " (summary count)")}).Return(mockCountGauge, nil)
		mockCountGauge.On("Record", ctx, 0.0, []otmetric.RecordOption{otmetric.WithAttributes(attrs...)}).Return()

		mockMeter.On("Float64Gauge", summaryName, []otmetric.Float64GaugeOption{otmetric.WithDescription(help + " (summary quantile)")}).Return(mockQuantileGauge, nil)
		mockQuantileGauge.On("Record", ctx, 0.0, []otmetric.RecordOption{otmetric.WithAttributes(copyWith(attrs, attribute.String("quantile", "0.5"))...)}).Return()

		recordSummary(ctx, mockLogger, mockMeter, gauges, summaryName, help, summary, attrs)

		mockMeter.AssertExpectations(t)
		mockSumGauge.AssertExpectations(t)
		mockCountGauge.AssertExpectations(t)
		mockQuantileGauge.AssertExpectations(t)
	})
}

// helper function to create test attributes
func createTestAttributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		{Key: "moniker", Value: attribute.StringValue("test-validator")},
	}
}

// helper functions for creating pointers
func floatPtr(f float64) *float64 {
	return &f
}

func uintPtr(u uint64) *uint64 {
	return &u
}

// ############################################################################
// ### 							Mocks 										###
// ############################################################################

type MockMeter struct {
	otmetric.Meter
	mock.Mock
}

func (m *MockMeter) Float64Counter(name string, options ...otmetric.Float64CounterOption) (otmetric.Float64Counter, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Float64Counter), args.Error(1)
}

func (m *MockMeter) Float64Gauge(name string, options ...otmetric.Float64GaugeOption) (otmetric.Float64Gauge, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Float64Gauge), args.Error(1)
}

func (m *MockMeter) Float64Histogram(name string, options ...otmetric.Float64HistogramOption) (otmetric.Float64Histogram, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Float64Histogram), args.Error(1)
}

func (m *MockMeter) Float64ObservableCounter(name string, options ...otmetric.Float64ObservableCounterOption) (otmetric.Float64ObservableCounter, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Float64ObservableCounter), args.Error(1)
}

func (m *MockMeter) Float64ObservableGauge(name string, options ...otmetric.Float64ObservableGaugeOption) (otmetric.Float64ObservableGauge, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Float64ObservableGauge), args.Error(1)
}

func (m *MockMeter) Float64ObservableUpDownCounter(name string, options ...otmetric.Float64ObservableUpDownCounterOption) (otmetric.Float64ObservableUpDownCounter, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Float64ObservableUpDownCounter), args.Error(1)
}

func (m *MockMeter) Float64UpDownCounter(name string, options ...otmetric.Float64UpDownCounterOption) (otmetric.Float64UpDownCounter, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Float64UpDownCounter), args.Error(1)
}

func (m *MockMeter) Int64Counter(name string, options ...otmetric.Int64CounterOption) (otmetric.Int64Counter, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Int64Counter), args.Error(1)
}

func (m *MockMeter) Int64Gauge(name string, options ...otmetric.Int64GaugeOption) (otmetric.Int64Gauge, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Int64Gauge), args.Error(1)
}

func (m *MockMeter) Int64Histogram(name string, options ...otmetric.Int64HistogramOption) (otmetric.Int64Histogram, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Int64Histogram), args.Error(1)
}

func (m *MockMeter) Int64ObservableCounter(name string, options ...otmetric.Int64ObservableCounterOption) (otmetric.Int64ObservableCounter, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Int64ObservableCounter), args.Error(1)
}

func (m *MockMeter) Int64ObservableGauge(name string, options ...otmetric.Int64ObservableGaugeOption) (otmetric.Int64ObservableGauge, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Int64ObservableGauge), args.Error(1)
}

func (m *MockMeter) Int64ObservableUpDownCounter(name string, options ...otmetric.Int64ObservableUpDownCounterOption) (otmetric.Int64ObservableUpDownCounter, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Int64ObservableUpDownCounter), args.Error(1)
}

func (m *MockMeter) Int64UpDownCounter(name string, options ...otmetric.Int64UpDownCounterOption) (otmetric.Int64UpDownCounter, error) {
	args := m.Called(name, options)
	return args.Get(0).(otmetric.Int64UpDownCounter), args.Error(1)
}

func (m *MockMeter) RegisterCallback(callback otmetric.Callback, instruments ...otmetric.Observable) (otmetric.Registration, error) {
	args := m.Called(callback, instruments)
	return args.Get(0).(otmetric.Registration), args.Error(1)
}

type MockFloat64Gauge struct {
	otmetric.Float64Gauge
	mock.Mock
}

func (m *MockFloat64Gauge) Record(ctx context.Context, value float64, options ...otmetric.RecordOption) {
	m.Called(ctx, value, options)
}

type MockFloat64Histogram struct {
	otmetric.Float64Histogram
	mock.Mock
}

func (m *MockFloat64Histogram) Record(ctx context.Context, value float64, options ...otmetric.RecordOption) {
	m.Called(ctx, value, options)
}
