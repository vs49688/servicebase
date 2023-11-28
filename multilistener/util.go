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

package multilistener

import (
	"log/slog"
	"net"
	"os"
)

func Listen(cfg *ListenConfig, logger *slog.Logger) (net.Listener, error) {
	logger = logger.With(
		slog.String("bind_address", cfg.BindAddress),
		slog.String("bind_network", cfg.BindNetwork),
		slog.String("socket_permissions", cfg.SocketPermissions.String()),
	)

	ln, err := net.Listen(cfg.BindNetwork, cfg.BindAddress)
	if err != nil {
		logger.Error("error listening", slog.Any("error", err))
		return nil, err
	}

	if _, ok := ln.(*net.UnixListener); ok {
		logger.Debug("fixing socket permissions")

		// Can't use the *os.File returned from the listener, it doesn't work.
		// We have to do it via name.
		if err := os.Chmod(cfg.BindAddress, cfg.SocketPermissions); err != nil {
			logger.Error("unable to chmod socket", slog.Any("error", err))
			return nil, err
		}
	}

	return ln, nil
}
