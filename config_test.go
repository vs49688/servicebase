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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigMerging(t *testing.T) {
	t.Parallel()

	t.Run("DisableXFF", func(t *testing.T) {
		t.Parallel()
		t.Run("true", func(t *testing.T) {
			left := &HTTPConfig{}

			right := &HTTPConfig{
				DisableXFF: AsPtr(true),
			}

			MergeHTTPConfig(left, right)

			assert.Equal(t, &HTTPConfig{DisableXFF: AsPtr(true)}, left)
		})

		t.Run("NotSpecified", func(t *testing.T) {
			left := &HTTPConfig{
				DisableXFF: AsPtr(true),
			}

			right := &HTTPConfig{}

			MergeHTTPConfig(left, right)

			assert.Equal(t, &HTTPConfig{DisableXFF: AsPtr(true)}, left)
		})
	})

	t.Run("MergeMap", func(t *testing.T) {
		left := map[string]string{"a": "b", "c": "d"}
		right := map[string]string{"a": "B", "e": "f"}
		MergeMap(left, right)
		assert.Equal(t, map[string]string{"a": "B", "c": "d", "e": "f"}, left)
	})

	t.Run("MergeString", func(t *testing.T) {
		assert.Equal(t, "a", MergeString("", "a"))
		assert.Equal(t, "a", MergeString("a", ""))
		assert.Equal(t, "b", MergeString("a", "b"))
	})
}
