/*
 * MinIO Client (C) 2019 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/minio/cli"
	json "github.com/minio/mc/pkg/colorjson"
	"github.com/minio/mc/pkg/console"
	"github.com/minio/mc/pkg/probe"
)

var adminConfigHistoryRestoreCmd = cli.Command{
	Name:   "restore",
	Usage:  "restore a history key value on MinIO server",
	Before: setGlobalsFromContext,
	Action: mainAdminConfigHistoryRestore,
	Flags:  globalFlags,
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} TARGET RESTOREID

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
EXAMPLES:
  1. Restore 'restore-id' history key value on MinIO server.
     {{.Prompt}} {{.HelpName}} play/ <restore-id>
`,
}

// configHistoryRestoreMessage container to hold locks information.
type configHistoryRestoreMessage struct {
	Status      string `json:"status"`
	RestoreID   string `json:"restoreID"`
	targetAlias string
}

// String colorized service status message.
func (u configHistoryRestoreMessage) String() (msg string) {
	suggestion := fmt.Sprintf("mc admin service restart %s", u.targetAlias)
	msg += console.Colorize("ConfigHistoryRestoreMessage",
		fmt.Sprintf("Please restart your server with `%s`.\n", suggestion))
	msg += console.Colorize("ConfigHistoryRestoreMessage", "Restored "+u.RestoreID+" kv successfully.")
	return msg
}

// JSON jsonified service status Message message.
func (u configHistoryRestoreMessage) JSON() string {
	u.Status = "success"
	statusJSONBytes, e := json.MarshalIndent(u, "", " ")
	fatalIf(probe.NewError(e), "Unable to marshal into JSON.")

	return string(statusJSONBytes)
}

// checkAdminConfigHistoryRestoreSyntax - validate all the passed arguments
func checkAdminConfigHistoryRestoreSyntax(ctx *cli.Context) {
	if !ctx.Args().Present() || len(ctx.Args()) > 2 {
		cli.ShowCommandHelpAndExit(ctx, "restore", 1) // last argument is exit code
	}
}

func mainAdminConfigHistoryRestore(ctx *cli.Context) error {

	checkAdminConfigHistoryRestoreSyntax(ctx)

	console.SetColor("ConfigHistoryRestoreMessage", color.New(color.FgGreen))

	// Get the alias parameter from cli
	args := ctx.Args()
	aliasedURL := args.Get(0)

	// Create a new MinIO Admin Client
	client, err := newAdminClient(aliasedURL)
	fatalIf(err, "Unable to initialize admin connection.")

	// Call get config API
	fatalIf(probe.NewError(client.RestoreConfigHistoryKV(args.Get(1))), "Cannot restore server configuration.")

	// Print
	printMsg(configHistoryRestoreMessage{
		RestoreID:   args.Get(1),
		targetAlias: aliasedURL,
	})

	return nil
}
