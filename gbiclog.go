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
//DONE: command - type list
//DONE: command - type edit
//DONE: command - type delete
//DONE: command - category add
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
	"github.com/zbroju/gbiclog/lib/bicycleTypes"
	"github.com/zbroju/gbiclog/lib/dataFile"
	"github.com/zbroju/gbiclog/lib/tripCategories"
	"github.com/zbroju/gprops"
	"os"
	"path"
	"strconv"
	"strings"
)

// Error messages
const (
	errSyntaxErrorInConfig = "syntax error in config file"
	errMissingFileFlag     = "missing information about data file. Specify it with --file or -f flag"
	errMissingNameFlag     = "missing name. Specify it with --name or -n flag"
	errMissingIdFlag       = "missing id. Specify it with --id or -i flag"
)

// Config settings
const (
	confDataFile = "DATA_FILE"
	confVerbose  = "VERBOSE"
)

// Application internal settings
const (
	appName     = "gBicLog"
	fsSeparator = "  "
)

// Objects
const (
	objectBicycleType      = "bicycle_type"
	objectBicycleTypeAlias = "bt"

	objectTripCategory      = "trip_category"
	objectTripCategoryAlias = "tc"
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
			fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errSyntaxErrorInConfig)
			return
		}
	}
	configFile.Close()
	dataFile := configSettings.GetOrDefault(confDataFile, "")
	verbose, err := strconv.ParseBool(configSettings.GetOrDefault(confVerbose, "false"))
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
	flagId := cli.IntFlag{
		Name:  "id, i",
		Value: -1,
		Usage: "ID of an object",
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
					Name:    objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagName},
					Usage:   "Add new bicycle type.",
					Action:  cmdTypeAdd,
				},
				{
					Name:    objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagName},
					Usage:   "Add new trip category.",
					Action:  cmdCategoryAdd,
				},
			},
		},
		{
			Name:    "list",
			Aliases: []string{"L"},
			Usage:   "List objects (bicycles, bicycle types, trips, trips categories)",
			Subcommands: []cli.Command{
				{
					Name:    objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile},
					Usage:   "List available bicycle types.",
					Action:  cmdTypeList,
				},
			},
		},
		{
			Name:    "edit",
			Aliases: []string{"E"},
			Usage:   "Edit an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{
					Name:    objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId, flagName},
					Usage:   "Edit bicycle type with given id.",
					Action:  cmdTypeEdit,
				},
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"D"},
			Usage:   "Delete an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{
					Name:    objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId},
					Usage:   "Delete bicycle type with given id.",
					Action:  cmdTypeDelete,
				},
			},
		},
	}
	app.Run(os.Args)
}

func cmdInit(c *cli.Context) {
	// Check the obligatory parameters and exit if missing
	if c.String("file") == "" {
		fmt.Fprint(os.Stderr, "%s: %s\n", appName, errMissingFileFlag)
		return
	}

	// Create new file
	f := dataFile.New(c.String("file"))
	err := f.CreateNew()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "%s: created file %s.\n", appName, c.String("file"))
	}
}

func cmdTypeAdd(c *cli.Context) {
	// Check obligatory flags (file, name)
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingFileFlag)
		return
	}
	if c.String("name") == "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingNameFlag)
		return
	}

	// Open data file
	f := dataFile.New(c.String("file"))
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}
	defer f.Close()

	// Add new type
	nt := bicycleTypes.BicycleType{0, c.String("name")}
	err = f.TypeAdd(nt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "%s: added new bicycle type: %s\n", appName, nt.Name)
	}

}

func cmdTypeList(c *cli.Context) {
	// Check obligatory flags (file)
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingFileFlag)
		return
	}

	// Open data file
	f := dataFile.New(c.String("file"))
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
	}
	defer f.Close()

	// List bicycle types
	types, err := f.TypeList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
	}
	if len(types) == 0 {
		fmt.Fprintf(os.Stdout, "%s: no bicycle types\n", appName)
		return
	}
	idH, nameH, idFS, nameFS := types.GetDisplayStrings()
	fmt.Fprintf(os.Stdout, strings.Join([]string{idH, nameH}, fsSeparator)+"\n")
	l := strings.Join([]string{idFS, nameFS}, fsSeparator) + "\n"
	for _, t := range types {
		fmt.Fprintf(os.Stdout, l, t.Id, t.Name)
	}
}

func cmdTypeEdit(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingFileFlag)
		return
	}
	id := c.Int("id")
	if id < 0 {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingIdFlag)
		return
	}
	newName := c.String("name")
	if newName == "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingNameFlag)
		return
	}

	// Open data file
	f := dataFile.New(c.String("file"))
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
	}
	defer f.Close()

	// Edit bicycle type
	btl, err := f.TypeList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}
	bt, err := btl.GetWithId(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}
	oldName := bt.Name
	bt.Name = newName

	err = f.TypeUpdate(bt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "%s: change bicycle type name from %s to %s\n", appName, oldName, newName)
	}
}

func cmdTypeDelete(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingFileFlag)
		return
	}
	id := c.Int("id")
	if id < 0 {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingIdFlag)
		return
	}

	// Open data file
	f := dataFile.New(c.String("file"))
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
	}
	defer f.Close()

	// Delete bicycle type
	btl, err := f.TypeList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}
	bt, err := btl.GetWithId(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}

	err = f.TypeDelete(bt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "%s: deleted bicycle type %s\n", appName, bt.Name)
	}
}

func cmdCategoryAdd(c *cli.Context) {
	// Check obligatory flags (file, name)
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingFileFlag)
		return
	}
	if c.String("name") == "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingNameFlag)
		return
	}

	// Open data file
	f := dataFile.New(c.String("file"))
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}
	defer f.Close()

	// Add new category
	nc := tripCategories.TripCategory{0, c.String("name")}
	err = f.CategoryAdd(nc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "%s: added new trip category: %s\n", appName, nc.Name)
	}

}
