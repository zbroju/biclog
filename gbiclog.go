// Written 2015 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
//
// TASKS:
//DONE: create scheme of DB
//DONE: config - data file name
//TODO: command - init data file
//TODO: checking if given file is a appropriate biclog file
//TODO: command - type add
//TODO: command - type list
//TODO: command - type edit
//TODO: command - type delete
//TODO: command - category add
//TODO: command - category list
//TODO: command - category edit
//TODO: command - category delete
//TODO: command - bicycle add
//TODO: command - bicycle list
//TODO: command - bicycle edit
//TODO: command - bicycle delete
//TODO: command - bicycle show details
//TODO: command - bicycle show photo
//TODO: command - trip add
//TODO: command - trip list
//TODO: command - trip edit
//TODO: command - trip delete
//TODO: command - trip show details
//TODO: command - report summary
//TODO: command - report history
//TODO: command - report pie chart (share of bicycles)
//TODO: command - report bar chart (history)
package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/zbroju/gprops"
	"os"
	"path"
	"strconv"
)

// Config settings
const (
	CONF_DATAFILE = "DATA_FILE"
	CONF_VERBOSE  = "VERBOSE"
)

func main() {
	cli.CommandHelpTemplate = `NAME:
   {{.HelpName}} - {{.Usage}}
USAGE:
   {{.HelpName}}{{if .Subcommands}} [subcommand]{{end}}{{if .Flags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{if .Description}}
DESCRIPTION:
   {{.Description}}{{end}}{{if .Flags}}
OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{ end }}{{if .Subcommands}}
SUBCOMMANDS:
    {{range .Subcommands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
{{end}}{{ end }}
`

	// Loading properties from config file if exists
	configSettings := gprops.New()
	configFile, err := os.Open(path.Join(os.Getenv("HOME"), ".gblrc"))
	if err == nil {
		err = configSettings.Load(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gBicLog: syntax error in %s. Exit.\n", configFile.Name())
			return
		}
	}
	configFile.Close()
	dataFile := configSettings.GetOrDefault(CONF_DATAFILE, "")
	verbose, err := strconv.ParseBool(configSettings.GetOrDefault(CONF_VERBOSE, "false"))
	if err != nil {
		verbose = false
	}

}
