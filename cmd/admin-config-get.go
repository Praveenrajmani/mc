/*
 * MinIO Client (C) 2017-2019 MinIO, Inc.
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
	"strings"

	"github.com/minio/cli"
	json "github.com/minio/mc/pkg/colorjson"
	"github.com/minio/mc/pkg/probe"
	"github.com/minio/minio/pkg/madmin"
)

var adminConfigGetCmd = cli.Command{
	Name:   "get",
	Usage:  "get config of a MinIO server/cluster",
	Before: setGlobalsFromContext,
	Action: mainAdminConfigGet,
	Flags:  globalFlags,
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} TARGET

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
EXAMPLES:
  1. Get the current region setting on MinIO server.
     {{.Prompt}} {{.HelpName}} play/ region
     # US east region setting
     name="us-east-1" state="on"

  2. Get the current notification settings for MQTT target on MinIO server
     {{.Prompt}} {{.HelpName}} myminio/ notify_mqtt
     # Notification settings for MQTT broker
     notify_mqtt broker="" password="" queue_dir="" queue_limit="0" reconnect_interval="0s" state="off" keep_alive_interval="0s" qos="0" topic="" username=""

  3. Get the current compression settings on MinIO server
     {{.Prompt}} {{.HelpName}} myminio/ compression
     # Compression settings for csv and text files only
     compression extensions=".txt,.csv" mime_types="text/*" state="on"
`,
}

// configGetMessage container to hold locks information.
type configGetMessage struct {
	Status string         `json:"status"`
	Value  madmin.Targets `json:"value"`
}

// String colorized service status message.
func (u configGetMessage) String() string {
	return u.Value.String()
}

// JSON jsonified service status Message message.
func (u configGetMessage) JSON() string {
	u.Status = "success"
	statusJSONBytes, e := json.MarshalIndent(u, "", " ")
	fatalIf(probe.NewError(e), "Unable to marshal into JSON.")

	return string(statusJSONBytes)
}

// checkAdminConfigGetSyntax - validate all the passed arguments
func checkAdminConfigGetSyntax(ctx *cli.Context) {
	if !ctx.Args().Present() || len(ctx.Args()) > 2 {
		cli.ShowCommandHelpAndExit(ctx, "get", 1) // last argument is exit code
	}
}

func mainAdminConfigGet(ctx *cli.Context) error {

	checkAdminConfigGetSyntax(ctx)

	// Get the alias parameter from cli
	args := ctx.Args()
	aliasedURL := args.Get(0)

	// Create a new MinIO Admin Client
	client, err := newAdminClient(aliasedURL)
	fatalIf(err, "Unable to initialize admin connection.")

	// Call get config API
	buf, e := client.GetConfigKV(strings.Join(args.Tail(), " "))
	fatalIf(probe.NewError(e), "Cannot get server '%s' config", args.Tail())

	// Print
	printMsg(configGetMessage{
		Value: buf,
	})

	return nil
}
