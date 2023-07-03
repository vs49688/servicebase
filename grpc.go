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
	grpcprommetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func createGRPCServer(cfg *GRPCConfig, registry *prometheus.Registry) (*grpc.Server, error) {
	var metrics *grpcprommetrics.ServerMetrics
	opts := cfg.Options

	if !cfg.DisableMetrics {
		metrics = grpcprommetrics.NewServerMetrics()

		opts = append(opts,
			grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()),
			grpc.StreamInterceptor(metrics.StreamServerInterceptor()),
		)
	}

	srv := grpc.NewServer(opts...)

	if cfg.EnableReflection {
		reflection.Register(srv)
	}

	if !cfg.DisableMetrics {
		metrics.InitializeMetrics(srv)

		if err := registry.Register(metrics); err != nil {
			return nil, err
		}
	}

	return srv, nil
}
