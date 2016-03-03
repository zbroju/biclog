// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
//
// TASKS:
//DONE: create scheme of DB
//DONE: config - data file name
//DONE: command - init data file
//DONE: checking if given file is a appropriate biclog file
//DONE: command - type add
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
	"github.com/zbroju/gBicLog/lib/database"
)

// Error messages
const (
	ERR_MISSING_FILE_FLAG = "gBicLog: missing information about data file. Specify it with --file or -f flag.\n"
	ERR_MISSING_NAME_FLAG = "gBicLog: missing name. Specify it with --name or -n flag.\n"
)

// Config settings
const (
	CONF_DATAFILE = "DATA_FILE"
	CONF_VERBOSE  = "VERBOSE"
)

func main() {
	cli.CommandHelpTemplate = `
	NAME:
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

	// Commandline arguments
	app := cli.NewApp()
	app.Name = "gBicLog"
	app.Usage = "keeps track of you bike rides"
	app.Version = "0.1 Alfa"
	app.Authors = []cli.Author{
		cli.Author{"Marcin 'Zbroju' Zbroinski", "marcin@zbroinski.net"},
	}

	// Flags definitions
	flagVerbose := cli.BoolFlag{
		Name:        "verbose, b",
		Usage:       "show more output",
		Destination: &verbose,
	}
	flagFile := cli.StringFlag{
		Name:  "file, f",
		Value: dataFile,
		Usage: "data file",
	}
	flagName := cli.StringFlag{
		Name:  "name, n",
		Value: "",
		Usage: "name",
	}

	// Commands
	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"I"},
			Flags:   []cli.Flag{flagVerbose, flagFile},
			Usage:   "Init a new data file specified by the user",
			Action:  cmdInit,
		},
		{
			Name:    "add",
			Aliases: []string{"A"},
			Usage:   "Add an object (bicycle, bicycle type, trip, trip category).",
			Subcommands: []cli.Command{
				{
					Name:    "bicycle_type",
					Aliases: []string{"bt"},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagName},
					Usage:   "Add new bicycle type.",
					Action:  cmdTypeAdd,
				},
			},
		},
	}
	app.Run(os.Args)
}

func cmdInit(c *cli.Context) {
	// Check the obligatory parameters and exit if missing
	if c.String("file") == "" {
		fmt.Fprint(os.Stderr, ERR_MISSING_FILE_FLAG)
		return
	}

	// Create new file
	dataFile := database.New(c.String("file"))
	err := dataFile.CreateNew()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%q", err)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "gBicLog: created file %s.\n", c.String("file"))
	}
}

func cmdTypeAdd(c *cli.Context) {
	// Check obligatory flags (file, name)
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, ERR_MISSING_FILE_FLAG)
		return
	}
	if c.String("name") == "" {
		fmt.Fprintf(os.Stderr, ERR_MISSING_NAME_FLAG)
		return
	}

	// Open data file
	dataFile:=database.New(c.String("file"))
	//dataFile := NewDatabase(c.String("file"))
	err := dataFile.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%q", err)
		return
	}
	defer dataFile.Close()

	// Add new type
	err = dataFile.TypeAdd(c.String("name"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%q", err)
		return
	}
}
