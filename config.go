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
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"strconv"
	"time"
)

type ListenConfig struct {
	Enabled           *bool    `json:"enabled,omitempty"`
	BindAddress       string   `json:"bind_address,omitempty"`
	BindNetwork       string   `json:"bind_network,omitempty"`
	SocketPermissions FileMode `json:"socket_permissions,omitempty"`
}

func (l *ListenConfig) IsEnabled() bool {
	if l.Enabled == nil {
		return false
	}

	return *l.Enabled
}

type HTTPConfig struct {
	ListenConfig
	PathPrefix        string        `json:"path_prefix,omitempty"`
	DisableXFF        *bool         `json:"disable_xff,omitempty"`
	DisableMetrics    *bool         `json:"disable_metrics"`
	DisableHealth     *bool         `json:"disable_health"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout"`
}

func (cfg *HTTPConfig) GetDisableXFF() bool {
	if cfg == nil || cfg.DisableXFF == nil {
		return false
	}

	return *cfg.DisableXFF
}

func (cfg *HTTPConfig) GetDisableMetrics() bool {
	if cfg == nil || cfg.DisableMetrics == nil {
		return false
	}

	return *cfg.DisableMetrics
}

func (cfg *HTTPConfig) GetDisableHealth() bool {
	if cfg == nil || cfg.DisableHealth == nil {
		return false
	}

	return *cfg.DisableHealth
}

type GRPCConfig struct {
	ListenConfig
	DisableMetrics   *bool               `json:"disable_metrics,omitempty"`
	EnableReflection *bool               `json:"enable_reflection,omitempty"`
	Options          []grpc.ServerOption `json:"-"` // TODO: Make this configurable from JSON/command line
}

func (cfg *GRPCConfig) GetDisableMetrics() bool {
	if cfg == nil || cfg.DisableMetrics == nil {
		return false
	}

	return *cfg.DisableMetrics
}

func (cfg *GRPCConfig) GetEnableReflection() bool {
	if cfg == nil || cfg.EnableReflection == nil {
		return false
	}

	return *cfg.EnableReflection
}

type ServiceConfig struct {
	LogLevel         string        `json:"log_level,omitempty"`
	LogFormat        string        `json:"log_format,omitempty"`
	ShutdownTimeout  time.Duration `json:"shutdown_timeout"`
	HTTP             HTTPConfig    `json:"http"`
	GRPC             GRPCConfig    `json:"grpc"`
	DisableRequestID *bool         `json:"disable_request_id,omitempty"`
}

func (cfg *ServiceConfig) GetDisableRequestID() bool {
	if cfg == nil || cfg.DisableRequestID == nil {
		return false
	}

	return *cfg.DisableRequestID
}

func DefaultHTTPConfig() HTTPConfig {
	return HTTPConfig{
		ListenConfig: ListenConfig{
			Enabled:           AsPtr(true),
			BindNetwork:       "tcp",
			BindAddress:       "127.0.0.1:8080",
			SocketPermissions: 0600,
		},
		PathPrefix:        "",
		DisableXFF:        AsPtr(false),
		DisableMetrics:    AsPtr(false),
		DisableHealth:     AsPtr(false),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func (cfg *HTTPConfig) Flags() []cli.Flag {
	def := DefaultHTTPConfig()
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "http-enabled",
			Usage:   "http server enabled",
			EnvVars: []string{"HTTP_ENABLED"},
			Value:   def.IsEnabled(),
			Action: func(context *cli.Context, b bool) error {
				cfg.Enabled = AsPtr(b)
				return nil
			},
		},
		&cli.StringFlag{
			Name:        "http-bind-network",
			Usage:       "http bind network (see net.Listen())",
			EnvVars:     []string{"HTTP_BIND_NETWORK"},
			Destination: &cfg.BindNetwork,
			Value:       def.BindNetwork,
		},
		&cli.StringFlag{
			Name:        "http-bind-address",
			Usage:       "http bind address (see net.Listen())",
			EnvVars:     []string{"HTTP_BIND_ADDRESS"},
			Destination: &cfg.BindAddress,
			Value:       def.BindAddress,
		},
		&cli.StringFlag{
			Name:        "http-path-prefix",
			Usage:       "http path prefix",
			EnvVars:     []string{"HTTP_PATH_PREFIX"},
			Destination: &cfg.PathPrefix,
			Required:    false,
			Value:       def.PathPrefix,
		},
		&cli.StringFlag{
			Name:    "http-unix-socket-permissions",
			Usage:   "http unix socket permissions (only if socket)",
			EnvVars: []string{"HTTP_UNIX_SOCKET_PERMISSIONS"},
			Value:   strconv.FormatInt(int64(def.SocketPermissions), 8),
			Action: func(context *cli.Context, s string) error {
				return cfg.SocketPermissions.UnmarshalText([]byte(s))
			},
		},
		&cli.BoolFlag{
			Name:    "http-disable-xff",
			Usage:   "disable X-Forwarded-For handling",
			EnvVars: []string{"HTTP_DISABLE_XFF"},
			Value:   def.GetDisableXFF(),
			Action: func(context *cli.Context, b bool) error {
				cfg.DisableXFF = AsPtr(b)
				return nil
			},
		},
		&cli.BoolFlag{
			Name:    "http-disable-metrics",
			Usage:   "disable /metrics endpoint",
			EnvVars: []string{"HTTP_DISABLE_METRICS"},
			Value:   def.GetDisableMetrics(),
			Action: func(context *cli.Context, b bool) error {
				cfg.DisableMetrics = AsPtr(b)
				return nil
			},
		},
		&cli.BoolFlag{
			Name:    "http-disable-health",
			Usage:   "disable /health endpoint",
			EnvVars: []string{"HTTP_DISABLE_HEALTH"},
			Value:   def.GetDisableHealth(),
			Action: func(context *cli.Context, b bool) error {
				cfg.DisableHealth = AsPtr(b)
				return nil
			},
		},
		&cli.DurationFlag{
			Name:        "http-read-header-timeout",
			Usage:       "http read header timeout",
			EnvVars:     []string{"HTTP_READ_HEADER_TIMEOUT"},
			Destination: &cfg.ReadHeaderTimeout,
			Value:       def.ReadHeaderTimeout,
		},
	}
}

func DefaultGRPCConfig() GRPCConfig {
	return GRPCConfig{
		ListenConfig: ListenConfig{
			Enabled:           AsPtr(true),
			BindNetwork:       "tcp",
			BindAddress:       "127.0.0.1:50051",
			SocketPermissions: 0600,
		},
		DisableMetrics:   AsPtr(false),
		EnableReflection: AsPtr(false),
	}
}

func (cfg *GRPCConfig) Flags() []cli.Flag {
	def := DefaultGRPCConfig()
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "grpc-enabled",
			Usage:   "grpc server enabled",
			EnvVars: []string{"GRPC_ENABLED"},
			Value:   def.IsEnabled(),
			Action: func(context *cli.Context, b bool) error {
				cfg.Enabled = AsPtr(b)
				return nil
			},
		},
		&cli.StringFlag{
			Name:        "grpc-bind-network",
			Usage:       "grpc bind network (see net.Listen())",
			EnvVars:     []string{"GRPC_BIND_NETWORK"},
			Destination: &cfg.BindNetwork,
			Value:       def.BindNetwork,
		},
		&cli.StringFlag{
			Name:        "grpc-bind-address",
			Usage:       "grpc bind address (see net.Listen())",
			EnvVars:     []string{"GRPC_BIND_ADDRESS"},
			Destination: &cfg.BindAddress,
			Value:       def.BindAddress,
		},
		&cli.StringFlag{
			Name:    "grpc-unix-socket-permissions",
			Usage:   "grpc unix socket permissions (only if socket)",
			EnvVars: []string{"GRPC_UNIX_SOCKET_PERMISSIONS"},
			Value:   strconv.FormatInt(int64(def.SocketPermissions), 8),
			Action: func(context *cli.Context, s string) error {
				return cfg.SocketPermissions.UnmarshalText([]byte(s))
			},
		},
		&cli.BoolFlag{
			Name:    "grpc-disable-metrics",
			Usage:   "disable /metrics endpoint",
			EnvVars: []string{"GRPC_DISABLE_METRICS"},
			Value:   def.GetDisableMetrics(),
			Action: func(context *cli.Context, b bool) error {
				cfg.DisableMetrics = AsPtr(b)
				return nil
			},
		},
	}

}

func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		LogLevel:        "info",
		LogFormat:       "text",
		ShutdownTimeout: 10 * time.Second,
		HTTP:            DefaultHTTPConfig(),
		GRPC:            DefaultGRPCConfig(),
	}
}

func (cfg *ServiceConfig) Flags() []cli.Flag {
	def := DefaultServiceConfig()

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Usage:       "logging level",
			EnvVars:     []string{"SERVICE_LOG_LEVEL"},
			Destination: &cfg.LogLevel,
			Value:       def.LogLevel,
		},
		&cli.StringFlag{
			Name:        "log-format",
			Usage:       "logging format (text/json)",
			EnvVars:     []string{"SERVICE_LOG_FORMAT"},
			Destination: &cfg.LogFormat,
			Value:       def.LogFormat,
		},
		&cli.DurationFlag{
			Name:        "shutdown-timeout",
			Usage:       "shutdown timeout",
			EnvVars:     []string{"SERVICE_SHUTDOWN_TIMEOUT"},
			Destination: &cfg.ShutdownTimeout,
			Value:       def.ShutdownTimeout,
		},
	}

	flags = append(flags, cfg.HTTP.Flags()...)
	flags = append(flags, cfg.GRPC.Flags()...)
	flags = append(flags, &cli.BoolFlag{
		Name:    "disable-request-id",
		Usage:   "disable request id handling (for both HTTP and GRPC)",
		EnvVars: []string{"SERVICE_DISABLE_REQUEST_ID"},
		Value:   def.GetDisableRequestID(),
		Action: func(context *cli.Context, b bool) error {
			cfg.DisableRequestID = AsPtr(b)
			return nil
		},
	})
	return flags
}

func MergeMap[T comparable, V any](left, right map[T]V) map[T]V {
	if left == nil && right != nil {
		left = map[T]V{}
	}

	for k, v := range right {
		left[k] = v
	}

	return left
}

func MergeString(left, right string) string {
	if right != "" {
		return right
	}

	return left
}

func MergeServiceConfig(left, right *ServiceConfig) *ServiceConfig {
	left.LogLevel = MergeString(left.LogLevel, right.LogLevel)
	left.LogFormat = MergeString(left.LogFormat, right.LogFormat)

	if right.ShutdownTimeout != 0 {
		left.ShutdownTimeout = right.ShutdownTimeout
	}

	MergeHTTPConfig(&left.HTTP, &right.HTTP)
	MergeGRPCConfig(&left.GRPC, &right.GRPC)

	if right.DisableRequestID != nil {
		left.DisableRequestID = AsPtr(*right.DisableRequestID)
	}

	return left
}

func MergeListenConfig(left, right *ListenConfig) *ListenConfig {
	left.BindNetwork = MergeString(left.BindNetwork, right.BindNetwork)
	left.BindAddress = MergeString(left.BindAddress, right.BindAddress)
	if right.SocketPermissions != 0 {
		left.SocketPermissions = right.SocketPermissions
	}

	if right.Enabled != nil {
		left.Enabled = AsPtr(*right.Enabled)
	}

	return left
}

func MergeHTTPConfig(left, right *HTTPConfig) *HTTPConfig {
	MergeListenConfig(&left.ListenConfig, &right.ListenConfig)
	left.PathPrefix = MergeString(left.PathPrefix, right.PathPrefix)

	if right.DisableXFF != nil {
		left.DisableXFF = AsPtr(*right.DisableXFF)
	}

	if right.DisableMetrics != nil {
		left.DisableMetrics = AsPtr(*right.DisableMetrics)
	}

	if right.DisableHealth != nil {
		left.DisableHealth = AsPtr(*right.DisableHealth)
	}

	if right.ReadHeaderTimeout != 0 {
		left.ReadHeaderTimeout = right.ReadHeaderTimeout
	}

	return left
}

func MergeGRPCConfig(left, right *GRPCConfig) *GRPCConfig {
	MergeListenConfig(&left.ListenConfig, &right.ListenConfig)

	if right.DisableMetrics != nil {
		left.DisableMetrics = AsPtr(*right.DisableMetrics)
	}

	if right.EnableReflection != nil {
		left.EnableReflection = AsPtr(*right.EnableReflection)
	}

	return left
}
