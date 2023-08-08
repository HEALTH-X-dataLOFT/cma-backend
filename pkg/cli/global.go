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

// Package CLI provides utilities for building the command line interface.
package cli

// GlobalOptions are options that are valid for all commands.
type GlobalOptions struct {
	Debug    bool
	LogLevel string `help:"Set log level." enum:"debug,info,warn,error" default:"info" env:"LOGLEVEL"`
}
