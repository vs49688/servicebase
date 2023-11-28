// Copyright 2023 Zane van Iperen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package servicebase

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net"
	"net/http"
)

type Metrics struct {
	Registry *prometheus.Registry
	requests *prometheus.CounterVec
}

type metricsLogger struct {
	logger *slog.Logger
}

func (m metricsLogger) Println(v ...interface{}) {
	m.logger.Error(fmt.Sprint(v...))
}

func configureMetrics(logger *slog.Logger) (Metrics, http.Handler, error) {
	metricsRegistry := prometheus.NewRegistry()

	// Add collector for Go stats
	if err := metricsRegistry.Register(collectors.NewGoCollector()); err != nil {
		return Metrics{}, nil, err
	}

	// Add collector for process stats
	if err := metricsRegistry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		return Metrics{}, nil, err
	}

	// Add build info
	if err := metricsRegistry.Register(collectors.NewBuildInfoCollector()); err != nil {
		return Metrics{}, nil, err
	}

	metricRequests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "root",
		Name:      "requests",
	}, []string{"path", "remote_ip", "host", "method", "user_agent"})

	if err := metricsRegistry.Register(metricRequests); err != nil {
		return Metrics{}, nil, err
	}

	return Metrics{Registry: metricsRegistry, requests: metricRequests}, promhttp.InstrumentMetricHandler(
		metricsRegistry,
		promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{ErrorLog: metricsLogger{logger: logger}}),
	), nil
}

func (m *Metrics) RecordHTTPRequest(req *http.Request) {
	remoteIP, _, _ := net.SplitHostPort(req.RemoteAddr)
	m.requests.With(prometheus.Labels{
		"path":       req.URL.Path,
		"remote_ip":  remoteIP,
		"host":       req.Host,
		"method":     req.Method,
		"user_agent": req.UserAgent(),
	}).Inc()
}
