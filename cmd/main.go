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

package main

import (
	"os"
	_ "time/tzdata"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/cli"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server"
	"github.com/alecthomas/kong"
)

var ui struct {
	cli.GlobalOptions
	Server server.Command `cmd:"" help:"Run server"`
}

func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "--help")
	}

	ctx := kong.Parse(&ui)
	params := cli.GenParams(ui.GlobalOptions)
	ctx.BindTo(params, (*cli.Params)(nil))
	ctx.FatalIfErrorf(ctx.Run())
}
