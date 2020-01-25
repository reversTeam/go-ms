package goms

import (
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/promauto"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
)

type GoMsMetrics struct {
	ServiceName string
	// requests *prometheus.GaugeVec
}
	// requests: promauto.NewGaugeVec(
	// prometheus.GaugeOpts{
	// 	Name: "fizzbuzz_request_sec",
	// 	Help: "Number of requests",
	// }, []string{"code", "method", "path"}),

func NewGoMsMetrics(serviceName string) *GoMsMetrics {
	return &GoMsMetrics{
		ServiceName: serviceName,
	}
}

func (o *GoMsMetrics) GetServiceName() string {
	return o.ServiceName
}