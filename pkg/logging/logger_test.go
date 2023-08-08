// Copyright 2025 HEALTH-X dataLOFT
//
// Licensed under the European Union Public Licence, Version 1.2 (the
// "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://eupl.eu/1.2/en/
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package logging provides logging utilities.
package logging_test

import (
	"testing"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/alecthomas/assert/v2"
)

func TestNewJSON(t *testing.T) {
	type args struct {
		requestedLevel string
	}
	tests := []struct {
		name     string
		args     args
		testFunc func(t testing.TB, fn func(), msgAndArgs ...interface{})
	}{
		{
			name: "Test if panics with wrong log level",
			args: args{
				requestedLevel: "wrong",
			},
			testFunc: assert.Panics,
		},
		{
			name: "Test if doesn't panic with correct log level",
			args: args{
				requestedLevel: "info",
			},
			testFunc: assert.NotPanics,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t, func() {
				logging.NewJSON(tt.args.requestedLevel, false)
			})
		})
	}
}
