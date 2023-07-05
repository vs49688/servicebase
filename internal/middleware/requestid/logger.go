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

package requestid

import (
	log "github.com/sirupsen/logrus"
)

const (
	DefaultLoggerFieldName = "request_id"
)

type requestIDLogger struct {
	field string
}

func NewLoggerHook(field string) log.Hook {
	return &requestIDLogger{field: field}
}

func (l *requestIDLogger) Levels() []log.Level {
	return log.AllLevels
}

func (l *requestIDLogger) Fire(entry *log.Entry) error {
	if entry.Context == nil || l.field == "" {
		return nil
	}

	if id := FromContext(entry.Context); id != "" {
		entry.Data[l.field] = id
	}

	return nil
}
