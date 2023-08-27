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
	"net"
	"os"

	log "github.com/sirupsen/logrus"
)

func Listen(cfg *ListenConfig, logger *log.Logger) (net.Listener, error) {
	logEntry := logger.WithFields(log.Fields{
		"bind_address":       cfg.BindAddress,
		"bind_network":       cfg.BindNetwork,
		"socket_permissions": cfg.SocketPermissions.String(),
	})

	ln, err := net.Listen(cfg.BindNetwork, cfg.BindAddress)
	if err != nil {
		logEntry.WithError(err).Error("error listening")
		return nil, err
	}

	if _, ok := ln.(*net.UnixListener); ok {
		logEntry.Trace("fixing socket permissions")

		// Can't use the *os.File returned from the listener, it doesn't work.
		// We have to do it via name.
		if err := os.Chmod(cfg.BindAddress, cfg.SocketPermissions); err != nil {
			logEntry.WithError(err).Error("unable to chmod socket")
			return nil, err
		}
	}

	return ln, nil
}
