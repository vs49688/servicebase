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
	"context"
	"io"
	"io/fs"

	log "github.com/sirupsen/logrus"
)

type ListenConfig struct {
	BindAddress       string
	BindNetwork       string
	SocketPermissions fs.FileMode
}

type wrapper interface {
	io.Closer

	Serve(ctx context.Context) error

	Log() *log.Entry
}

type MultiListener struct {
	logger  *log.Logger
	servers []wrapper
}
