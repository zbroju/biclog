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
//DONE: command - category list
//DONE: command - category edit
//DONE: command - category delete
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
	"github.com/zbroju/gsqlitehandler"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Error messages
const (
	errSyntaxErrorInConfig = "syntax error in config file"
	errMissingFileFlag     = "missing information about data file. Specify it with --file or -f flag"
	errMissingNameFlag     = "missing name. Specify it with --name or -n flag"
	errMissingIdFlag       = "missing id. Specify it with --id or -i flag"
	errWritingToFile       = "error writing to file"
	errReadingFromFile     = "error reading to file"
	errNoBicycleWithID     = "no bicycle with given id"
	errNoCategoriesWithID  = "no trip categories with given id"
)

// Config file settings
const (
	confDataFile = "DATA_FILE"
	confVerbose  = "VERBOSE"
)

// Objects
const (
	objectBicycleType       = "bicycle_type"
	objectBicycleTypeAlias  = "bt"
	objectTripCategory      = "trip_category"
	objectTripCategoryAlias = "tc"
)

// Application internal settings
const (
	appName     = "bicLog"
	fsSeparator = "  "
)

// Headers titles
const (
	btIdHeader   = "ID"
	btNameHeader = "B.TYPE"
	tcIdHeader   = "ID"
	tcNameHeader = "T.CATEGORY"
)

// DB Properties
var dataFileProperties = map[string]string{
	"applicationName": "gBicLog",
	"databaseVersion": "1.0",
}

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
	configFile, err := os.Open(path.Join(os.Getenv("HOME"), ".blrc"))
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
	app.Name = appName
	app.Usage = "keeps track of you bike rides"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{"Marcin 'Zbroju' Zbroinski", "marcin@zbroinski.net"},
	}

	// Flags definitions
	flagVerbose := cli.BoolFlag{Name: "verbose, b", Usage: "show more output", Destination: &verbose}
	flagFile := cli.StringFlag{Name: "file, f", Value: dataFile, Usage: "data file"}
	flagName := cli.StringFlag{Name: "name, n", Value: "", Usage: "name"}
	flagId := cli.IntFlag{Name: "id, i", Value: -1, Usage: "ID of an object"}

	// Commands
	app.Commands = []cli.Command{
		{Name: "init",
			Aliases: []string{"I"},
			Flags:   []cli.Flag{flagVerbose, flagFile},
			Usage:   "Init a new data file specified by the user",
			Action:  cmdInit},
		{Name: "add", Aliases: []string{"A"}, Usage: "Add an object (bicycle, bicycle type, trip, trip category).",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagName},
					Usage:   "Add new bicycle type.",
					Action:  cmdTypeAdd},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagName},
					Usage:   "Add new trip category.",
					Action:  cmdCategoryAdd}}},
		{Name: "list", Aliases: []string{"L"}, Usage: "List objects (bicycles, bicycle types, trips, trips categories)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile},
					Usage:   "List available bicycle types.",
					Action:  cmdTypeList},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile},
					Usage:   "List available trip categories.",
					Action:  cmdCategoryList}}},
		{Name: "edit", Aliases: []string{"E"}, Usage: "Edit an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId, flagName},
					Usage:   "Edit bicycle type with given id.",
					Action:  cmdTypeEdit},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId, flagName},
					Usage:   "Edit trip category with given id.",
					Action:  cmdCategoryEdit}}},
		{Name: "delete", Aliases: []string{"D"}, Usage: "Delete an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId},
					Usage:   "Delete bicycle type with given id.",
					Action:  cmdTypeDelete},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId},
					Usage:   "Delete trip category with given id.",
					Action:  cmdCategoryDelete}}}}
	app.Run(os.Args)
}

func cmdInit(c *cli.Context) {
	// Check the obligatory parameters and exit if missing
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingFileFlag)
		return
	}

	// Create new file
	sqlCreateTables := `
CREATE TABLE bicycles (
 id INTEGER PRIMARY KEY
 , name TEXT
 , producer TEXT
 , model TEXT
 , bicycle_type_id INTEGER
 , production_year INTEGER
 , buying_date INTEGER
 , description TEXT
 , status INTEGER
 , size TEXT
 , weight REAL
 , initial_distance REAL
 , series_no TEXT
 , photo BLOB
);
CREATE TABLE trips (
 id INTEGER PRIMARY KEY
 , bicycle_id INTEGER
 , date INTEGER
 , title TEXT
 , trip_category_id INTEGER
 , distance REAL
 , duration INTEGER
 , description TEXT
 , hr_max INTEGER
 , hr_avg INTEGER
 , speed_max REAL
 , driveways REAL
 , calories INTEGER
 , temperature REAL
);
CREATE TABLE bicycle_types (
 id INTEGER PRIMARY KEY
 , name text
);
CREATE TABLE trip_categories (
 id INTEGER PRIMARY KEY
 , name text
);
`
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.CreateNew(sqlCreateTables)
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
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}
	defer f.Close()

	// Add new type
	sqlAddType := fmt.Sprintf("INSERT INTO bicycle_types VALUES (NULL, '%s');", c.String("name"))
	_, err = f.Handler.Exec(sqlAddType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errWritingToFile)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "%s: added new bicycle type: %s\n", appName, c.String("name"))
	}

}

func cmdTypeList(c *cli.Context) {
	// Check obligatory flags (file)
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingFileFlag)
		return
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}
	defer f.Close()

	// Create formatting strings
	var maxLId, maxLName int
	err = f.Handler.QueryRow("SELECT max(length(id)), max(length(name)) FROM bicycle_types;").Scan(&maxLId, &maxLName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, "no bicycle types")
		return
	}
	if hlId := utf8.RuneCountInString(btIdHeader); maxLId < hlId {
		maxLId = hlId
	}
	if hlName := utf8.RuneCountInString(btNameHeader); maxLName < hlName {
		maxLName = hlName
	}
	fsId := fmt.Sprintf("%%%dv", maxLId)
	fsName := fmt.Sprintf("%%-%dv", maxLName)

	// List bicycle types
	rows, err := f.Handler.Query("SELECT id, name FROM bicycle_types ORDER BY name;")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s:  %s\n", appName, errReadingFromFile)
		return
	}
	defer rows.Close()

	line := strings.Join([]string{fsId, fsName}, fsSeparator) + "\n"
	fmt.Fprintf(os.Stdout, line, btIdHeader, btNameHeader)
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		fmt.Fprintf(os.Stdout, line, id, name)
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
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
	}
	defer f.Close()

	// Edit bicycle type
	sqlUpdateType := fmt.Sprintf("UPDATE bicycle_types SET name='%s' WHERE id=%d;", newName, id)
	r, err := f.Handler.Exec(sqlUpdateType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errWritingToFile)
		return
	}
	if i, _ := r.RowsAffected(); i == 0 {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errNoBicycleWithID)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "%s: change bicycle type name to '%s'\n", appName, newName)
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
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
	}
	defer f.Close()

	// Delete bicycle type
	sqlDeleteType := fmt.Sprintf("DELETE FROM bicycle_types WHERE id=%d;", id)
	r, err := f.Handler.Exec(sqlDeleteType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errWritingToFile)
		return
	}
	if i, _ := r.RowsAffected(); i == 0 {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errNoBicycleWithID)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "%s: deleted bicycle type with id = %d\n", appName, id)
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
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}
	defer f.Close()

	// Add new category
	sqlAddCategory := fmt.Sprintf("INSERT INTO trip_categories VALUES (NULL, '%s');", c.String("name"))
	_, err = f.Handler.Exec(sqlAddCategory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "%s: added new trip category: %s\n", appName, c.String("name"))
	}

}

func cmdCategoryList(c *cli.Context) {
	// Check obligatory flags (file)
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errMissingFileFlag)
		return
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
		return
	}
	defer f.Close()

	// Create formatting strings
	var maxLId, maxLName int
	err = f.Handler.QueryRow("SELECT max(length(id)), max(length(name)) FROM trip_categories;").Scan(&maxLId, &maxLName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, "no trip categories")
		return
	}
	if hlId := utf8.RuneCountInString(tcIdHeader); maxLId < hlId {
		maxLId = hlId
	}
	if hlName := utf8.RuneCountInString(tcNameHeader); maxLName < hlName {
		maxLName = hlName
	}
	fsId := fmt.Sprintf("%%%dv", maxLId)
	fsName := fmt.Sprintf("%%-%dv", maxLName)

	// List trip categories
	rows, err := f.Handler.Query("SELECT id, name FROM trip_categories ORDER BY name;")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s:  %s\n", appName, errReadingFromFile)
		return
	}
	defer rows.Close()

	line := strings.Join([]string{fsId, fsName}, fsSeparator) + "\n"
	fmt.Fprintf(os.Stdout, line, tcIdHeader, tcNameHeader)
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		fmt.Fprintf(os.Stdout, line, id, name)
	}
}

func cmdCategoryEdit(c *cli.Context) {
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
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
	}
	defer f.Close()

	// Edit trip category
	sqlUpdateCategory := fmt.Sprintf("UPDATE trip_categories SET name='%s' WHERE id=%d;", newName, id)
	r, err := f.Handler.Exec(sqlUpdateCategory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errWritingToFile)
		return
	}
	if i, _ := r.RowsAffected(); i == 0 {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errNoCategoriesWithID)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "%s: change trip category name to '%s'\n", appName, newName)
	}
}

func cmdCategoryDelete(c *cli.Context) {
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
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
	}
	defer f.Close()

	// Delete trip category
	sqlDeleteCategory := fmt.Sprintf("DELETE FROM trip_categories WHERE id=%d;", id)
	r, err := f.Handler.Exec(sqlDeleteCategory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errWritingToFile)
		return
	}
	if i, _ := r.RowsAffected(); i == 0 {
		fmt.Fprintf(os.Stderr, "%s: %s\n", appName, errNoCategoriesWithID)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "%s: deleted trip category with id = %d\n", appName, id)
	}
}
