package reporter

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	metrics "github.com/launchdarkly/go-metrics"
	"github.com/launchdarkly/go-metrics-cloudwatch/config"
)

type MockPutMetricsClient struct {
	metricsPut int
	requests   int
}

func (m *MockPutMetricsClient) PutMetricData(in *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {
	m.metricsPut += len(in.MetricData)
	m.requests += 1
	return &cloudwatch.PutMetricDataOutput{}, nil
}

func TestCloudwatchReporter(t *testing.T) {
	mock := &MockPutMetricsClient{}
	cfg := &config.Config{
		Client: mock,
		Filter: &config.NoFilter{},
	}
	registry := metrics.NewRegistry()
	for i := 0; i < 30; i++ {
		count := metrics.GetOrRegisterCounter(fmt.Sprintf("count-%d", i), registry)
		count.Inc(1)
	}

	EmitMetrics(registry, cfg)

	if mock.metricsPut < 30 || mock.requests < 2 {
		t.Fatal("No Metrics Put")
	}
}

func TestHistograms(t *testing.T) {
	mock := &MockPutMetricsClient{}
	filter := &config.NoFilter{}
	cfg := &config.Config{
		Client: mock,
		Filter: filter,
	}
	registry := metrics.NewRegistry()
	hist := metrics.GetOrRegisterHistogram(fmt.Sprintf("histo"), registry, metrics.NewUniformSample(1024))
	hist.Update(1000)
	hist.Update(500)
	EmitMetrics(registry, cfg)

	if mock.metricsPut < len(filter.Percentiles("")) {
		t.Fatal("No Metrics Put")
	}
}

func TestTimers(t *testing.T) {
	mock := &MockPutMetricsClient{}
	cfg := &config.Config{
		Client: mock,
		Filter: &config.NoFilter{},
	}
	registry := metrics.NewRegistry()
	timer := metrics.GetOrRegisterTimer(fmt.Sprintf("timer"), registry)
	timer.Update(10 * time.Second)
	EmitMetrics(registry, cfg)

	if mock.metricsPut < 7 {
		t.Fatal("No Metrics Put")
	}
}

func TestFilters(t *testing.T) {
	mock := &MockPutMetricsClient{}
	cfg := &config.Config{
		Client: mock,
		Filter: &config.AllFilter{},
	}
	registry := metrics.NewRegistry()
	timer := metrics.GetOrRegisterTimer(fmt.Sprintf("timer"), registry)
	timer.Update(10 * time.Second)
	EmitMetrics(registry, cfg)

	if mock.metricsPut > 0 {
		t.Fatal("Metrics Put")
	}
}
