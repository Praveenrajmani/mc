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
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/minio/cli"
	"github.com/minio/mc/pkg/console"
	"github.com/minio/mc/pkg/probe"
)

const (
	treeEntry     = "├─ "
	treeLastEntry = "└─ "
	treeNext      = "│"
	treeLevel     = "  "
)

// Structured message depending on the type of console.
type treeMessage struct {
	Entry        string
	IsDir        bool
	BranchString string
}

// Colorized message for console printing.
func (t treeMessage) String() string {
	//fmt.Printf("%s\n", console.Colorize("Dir", url))
	entryType := "File"
	if t.IsDir {
		entryType = "Dir"
	}
	return fmt.Sprintf("%s%s", t.BranchString, console.Colorize(entryType, t.Entry))
}

// JSON'ified message for scripting.
// Does No-op. JSON requests are redirected to `ls -r --json`
func (r treeMessage) JSON() string {
	return ""
}

var treeFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "files, f",
		Usage: "includes files in tree",
	},
	cli.IntFlag{
		Name:  "depth, d",
		Usage: "sets the depth threshold",
		Value: -1,
	},
}

// trees files and folders.
var treeCmd = cli.Command{
	Name:   "tree",
	Usage:  "list buckets and objects in a tree format",
	Action: mainTree,
	Before: setGlobalsFromContext,
	Flags:  append(treeFlags, globalFlags...),
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} [FLAGS] TARGET [TARGET ...]

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
EXAMPLES:
   1. List buckets on Amazon S3 cloud storage in a tree format.
      $ {{.HelpName}} myS3

   2. List all buckets on Amazon S3 cloud storage in a tree format.
      $ {{.HelpName}} myS3/mybucket/

   3. List all buckets on Amazon S3 cloud storage on Microsoft Windows in a tree format.
      $ {{.HelpName}} myS3\mybucket\
   
   4. List all buckets including the objects on Amazon S3 cloud storage in a tree format.
      $ {{.HelpName}} -f myS3/mybucket/
   
   5. Set the depth of the tree for listing.
      $ {{.HelpName}} -d 2 myS3/mybucket/

   6. List all the directories irrespective to the depth. -1 is the default value for depth.
      $ {{.HelpName}} -d -1 myS3/mybucket/
`,
}

// checkTreeSyntax - validate all the passed arguments
func checkTreeSyntax(ctx *cli.Context) {
	args := ctx.Args()

	depth := ctx.Int("depth")
	if depth != -1 && depth <= 0 {
		fatalIf(errInvalidArgument().Trace(args...), "depth should have a value greater than 0 or equal to -1")
	}

	if (args.Present()) && len(args) == 0 {
		args = []string{"."}
		return
	}

	for _, url := range args {
		if !isURLPrefixExists(url, false) {
			fatalIf(probe.NewError(errors.New("See 'mc tree -h' for help")),
				"'"+url+"' is not a valid 'alias[/bucket-name ...]'")
		}
	}
}

// doTree - list all entities inside a folder in a tree format.
func doTree(url string, level int, leaf bool, dirClosed map[int]bool, depth int, includeFiles bool) error {

	targetAlias, targetURL, _ := mustExpandAlias(url)
	if !strings.HasSuffix(targetURL, "/") {
		targetURL += "/"
	}

	clnt, err := newClientFromAlias(targetAlias, targetURL)
	fatalIf(err.Trace(targetURL), "Unable to initialize target `"+targetURL+"`.")

	prefixPath := clnt.GetURL().Path
	separator := string(clnt.GetURL().Separator)
	prefixPath = strings.TrimSuffix(prefixPath, prefixPath[strings.LastIndex(prefixPath, separator)+1:])

	bucketNameShowed := false
	var prev *clientContent
	show := func(end bool) error {
		var branchString string
		if level == 1 && !bucketNameShowed {
			bucketNameShowed = true
			printMsg(treeMessage{
				Entry:        url,
				IsDir:        true,
				BranchString: branchString,
			})
		}

		if level != 1 {
			for i := 1; i < level; i++ {
				if dirClosed[i] {
					branchString += " " + treeLevel
				} else {
					branchString += treeNext + treeLevel
				}
			}
		}

		if end {
			dirClosed[level] = true
			branchString += treeLastEntry
		} else {
			dirClosed[level] = false
			branchString += treeEntry
		}

		// Convert any os specific delimiters to "/".
		contentURL := filepath.ToSlash(prev.URL.Path)
		prefixPath = filepath.ToSlash(prefixPath)

		// Trim prefix of current working dir
		prefixPath = strings.TrimPrefix(prefixPath, "."+separator)

		if prev.Type.IsDir() {
			printMsg(treeMessage{
				Entry:        strings.TrimSuffix(strings.TrimPrefix(contentURL, prefixPath), "/"),
				IsDir:        true,
				BranchString: branchString,
			})
		} else {
			printMsg(treeMessage{
				Entry:        strings.TrimPrefix(contentURL, prefixPath),
				IsDir:        false,
				BranchString: branchString,
			})
		}

		if prev.Type.IsDir() {
			url := ""
			if targetAlias != "" {
				url = targetAlias + "/" + contentURL
			} else {
				url = contentURL
			}

			if depth == -1 || level <= depth {
				if err := doTree(url, level+1, end, dirClosed, depth, includeFiles); err != nil {
					return err
				}
			}
		}

		return nil
	}

	for content := range clnt.List(false, false, DirNone) {

		if !includeFiles && !content.Type.IsDir() {
			continue
		}

		if content.Err != nil {
			errorIf(content.Err.Trace(clnt.GetURL().String()), "Unable to tree.")
			continue
		}

		if prev != nil {
			if err := show(false); err != nil {
				return err
			}
		}

		prev = content
	}

	if prev != nil {
		if err := show(true); err != nil {
			return err
		}
	}

	return nil
}

// mainTree - is a handler for mc tree command
func mainTree(ctx *cli.Context) error {

	// check 'tree' cli arguments.
	checkTreeSyntax(ctx)

	console.SetColor("File", color.New(color.Bold))
	console.SetColor("Dir", color.New(color.FgCyan, color.Bold))

	args := ctx.Args()
	// mimic operating system tool behavior.
	if !ctx.Args().Present() {
		args = []string{"."}
	}

	includeFiles := ctx.Bool("files")

	var cErr error
	for _, targetURL := range args {
		if !globalJSON {
			dirMap := make(map[int]bool)
			if e := doTree(targetURL, 1, false, dirMap, ctx.Int("depth"), includeFiles); e != nil {
				cErr = e
			}
		} else {
			targetAlias, targetURL, _ := mustExpandAlias(targetURL)
			if !strings.HasSuffix(targetURL, "/") {
				targetURL += "/"
			}
			clnt, err := newClientFromAlias(targetAlias, targetURL)
			fatalIf(err.Trace(targetURL), "Unable to initialize target `"+targetURL+"`.")
			if e := doList(clnt, true, false); e != nil {
				cErr = e
			}
		}
	}
	return cErr
}
