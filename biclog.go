// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
//
//TASKS:
//DONE: change communication to 'log' library
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
//DONE: command - bicycle add
//DONE: command - bicycle list
//DONE: command - bicycle edit (remember about changing status to scrapped, sold and stolen)
//DONE: command - bicycle delete
//TODO: command - bicycle show details
//DONE: command - trip add
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
	"database/sql"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/zbroju/gprops"
	"github.com/zbroju/gsqlitehandler"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Error messages
const (
	errSyntaxErrorInConfig = "syntax error in config file"
	errMissingFileFlag     = "missing information about data file. Specify it with --file or -f flag"
	errMissingTypeFlag     = "missing bicycle type. Specify it with --type or -t flag"
	errMissingCategoryFlag = "missing trip category. Specify it with --category or -c flag"
	errMissingIdFlag       = "missing id. Specify it with --id or -i flag"
	errMissingBicycleFlag  = "missing bicycle. Specify it with --bicycle or -b flag"
	errMissingTitleFlag    = "missing trip title. Specify it with --title or -s flag"
	errMissingDistanceFlag = "missing trip distance. Specify it with --distance or -d flag"

	errWritingToFile              = "error writing to file"
	errReadingFromFile            = "error reading to file"
	errNoBicycleWithID            = "no bicycle with given id"
	errNoBicycleForName           = "no bicycle for given name"
	errBicycleNameisAmbiguous     = "bicycle name is ambiguous"
	errNoBicycleTypeWithID        = "no bicycle type with given id"
	errNoCategoryWithID           = "no trip categories with given id"
	errNoCategoryForName          = "no trip category for given name"
	errNoBicycleTypesForName      = "no bicycle types for given name"
	errNoBicycleStatus            = "unknown bicycle status"
	errBicycleTypeNameIsAmbiguous = "given bicycle type name is ambiguous"
	errCategoryNameIsAmbiguous    = "given trip category name is ambiguous"

	errCannotRemoveBicycleType = "cannot remove bicycle type because there are bicycles of this type"
	errCannotRemoveCategory    = "cannot remove category because there are trips with this category"
	errCannotRemoveBicycle     = "cannot remove bicycle because there are trips done on it"
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
	objectBicycle           = "bicycle"
	objectBicycleAlias      = "bc"
	objectTrip              = "trip"
	objectTripAlias         = "tr"
)

// Bicycle statuses
var bicycleStatuses = map[string]int{
	"owned":    1,
	"sold":     2,
	"scrapped": 3,
	"stolen":   4,
}

// Application internal settings
const (
	appName     = "biclog"
	fsSeparator = "  "

	notSetIntValue    int     = -1
	notSetFloatValue  float64 = -1
	notSetStringValue         = ""
)

// Logger
var (
	printUserMsg *log.Logger
	printError   *log.Logger
)

// Headers titles
const (
	btIdHeader   = "ID"
	btNameHeader = "B.TYPE"

	tcIdHeader   = "ID"
	tcNameHeader = "T.CATEGORY"

	bcIdHeader          = "ID"
	bcNameHeader        = "B.NAME"
	bcProducerHeader    = "PRODUCER"
	bcModelHeader       = "MODEL"
	bcProdYearHeader    = "PROD.YEAR"
	bcBuyingDateHeader  = "BUY.DATE"
	bdDescriptionHeader = "DESCRIPTION"
	bcStatusHeader      = "STATUS"
	bcSizeHeader        = "SIZE"
	bcWeightHeader      = "WEIGHT"
	bcInitDistHeader    = "INIT DIST."
	bcSeriesHeader      = "SERIES"
)

// DB Properties
var dataFileProperties = map[string]string{
	"applicationName": "gBicLog",
	"databaseVersion": "1.0",
}

// *************
// MAIN FUNCTION
// *************

func main() {
	// Set up logger
	printUserMsg = log.New(os.Stdout, fmt.Sprintf("%s: ", appName), 0)
	printError = log.New(os.Stderr, fmt.Sprintf("%s: ", appName), 0)

	// Loading properties from config file if exists
	configSettings := gprops.New()
	configFile, err := os.Open(path.Join(os.Getenv("HOME"), ".blrc"))
	if err == nil {
		err = configSettings.Load(configFile)
		if err != nil {
			printError.Fatalln(errSyntaxErrorInConfig)
		}
	}
	configFile.Close()
	dataFile := configSettings.GetOrDefault(confDataFile, "")
	verbose, err := strconv.ParseBool(configSettings.GetOrDefault(confVerbose, "false"))
	if err != nil {
		verbose = false
	}

	// Parse user commands and flags
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
	//TODO: reformat help so that no default values are shown after -h option

	app := cli.NewApp()
	app.Name = appName
	app.Usage = "keeps track of you bike rides"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{"Marcin 'Zbroju' Zbroinski", "marcin@zbroinski.net"},
	}

	flagVerbose := cli.BoolFlag{Name: "verbose, v", Usage: "show more output", Destination: &verbose}
	flagFile := cli.StringFlag{Name: "file, f", Value: dataFile, Usage: "data file"}
	flagType := cli.StringFlag{Name: "type, t", Value: notSetStringValue, Usage: "bicycle type"}
	flagCategory := cli.StringFlag{Name: "category, c", Value: notSetStringValue, Usage: "trip category"}
	flagId := cli.IntFlag{Name: "id, i", Value: notSetIntValue, Usage: "ID of an object"}
	flagBicycle := cli.StringFlag{Name: "bicycle, b", Value: notSetStringValue, Usage: "bicycle name"}
	flagManufacturer := cli.StringFlag{Name: "manufacturer", Value: notSetStringValue, Usage: "bicycle manufacturer"}
	flagModel := cli.StringFlag{Name: "model", Value: notSetStringValue, Usage: "bicycle model"}
	flagProductionYear := cli.IntFlag{Name: "year", Value: notSetIntValue, Usage: "year when the bike was made"}
	flagBuyingDate := cli.StringFlag{Name: "bought", Value: notSetStringValue, Usage: "date when the bike was bought"}
	flagDescription := cli.StringFlag{Name: "description, d", Value: notSetStringValue, Usage: "more verbose description"}
	flagStatus := cli.StringFlag{Name: "status", Value: notSetStringValue, Usage: "bicycle status (owned, sold, scrapped, stolen)"}
	flagSize := cli.StringFlag{Name: "size", Value: notSetStringValue, Usage: "size of the bike"}
	flagWeight := cli.Float64Flag{Name: "weight", Value: notSetFloatValue, Usage: "bike's weight"}
	flagInitialDistance := cli.Float64Flag{Name: "init_distance", Value: notSetFloatValue, Usage: "initial distance of the bike"}
	flagSeries := cli.StringFlag{Name: "series", Value: notSetStringValue, Usage: "series number"}
	flagDate := cli.StringFlag{Name: "date", Value: time.Now().Format("2006-01-02"), Usage: "date of trip (default: today)"}
	flagTitle := cli.StringFlag{Name: "title, s", Value: notSetStringValue, Usage: "trip title"}
	flagDistance := cli.Float64Flag{Name: "distance, r", Value: notSetFloatValue, Usage: "trip distance"}
	flagDuration := cli.DurationFlag{Name: "duration, l", Usage: "trip duration"}
	flagHRMax := cli.IntFlag{Name: "hrmax", Value: notSetIntValue, Usage: "hr max"}
	flagHRAvg := cli.IntFlag{Name: "hravg", Value: notSetIntValue, Usage: "hr average"}
	flagSpeedMax := cli.Float64Flag{Name: "speed_max", Value: notSetFloatValue, Usage: "maximum speed"}
	flagDriveways := cli.Float64Flag{Name: "driveways", Value: notSetFloatValue, Usage: "sum of driveways"}
	flagCalories := cli.IntFlag{Name: "calories", Value: notSetIntValue, Usage: "sum of calories burnt"}
	flagTemperature := cli.Float64Flag{Name: "temperature", Value: notSetFloatValue, Usage: "average temperature"}

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
					Flags:   []cli.Flag{flagVerbose, flagFile, flagType},
					Usage:   "Add new bicycle type.",
					Action:  cmdTypeAdd},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagCategory},
					Usage:   "Add new trip category.",
					Action:  cmdCategoryAdd},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagBicycle, flagManufacturer, flagModel, flagType, flagProductionYear, flagBuyingDate, flagDescription, flagSize, flagWeight, flagInitialDistance, flagSeries},
					Usage:   "Add new bicycle.",
					Action:  cmdBicycleAdd},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagTitle, flagBicycle, flagDate, flagCategory, flagDistance, flagDuration, flagDescription, flagHRMax, flagHRAvg, flagSpeedMax, flagDriveways, flagCalories, flagTemperature},
					Usage:   "Add new trip.",
					Action:  cmdTripAdd}}},
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
					Action:  cmdCategoryList},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile},
					Usage:   "List available bicycles.",
					Action:  cmdBicycleList}}},
		{Name: "edit", Aliases: []string{"E"}, Usage: "Edit an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId, flagType},
					Usage:   "Edit bicycle type with given id.",
					Action:  cmdTypeEdit},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId, flagCategory},
					Usage:   "Edit trip category with given id.",
					Action:  cmdCategoryEdit},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagBicycle, flagManufacturer, flagModel, flagType, flagProductionYear, flagBuyingDate, flagDescription, flagStatus, flagSize, flagWeight, flagInitialDistance, flagSeries},
					Usage:   "Edit bicycle details.",
					Action:  cmdBicycleEdit}}},
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
					Action:  cmdCategoryDelete},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId},
					Usage:   "Delete bicycle with given id.",
					Action:  cmdBicycleDelete}}}}
	app.Run(os.Args)
}

// ********
// Commands
// ********

func cmdInit(c *cli.Context) {
	// Check the obligatory parameters and exit if missing
	if c.String("file") == "" {
		printError.Fatalln(errMissingFileFlag)
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
		printError.Fatalln(err)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		printUserMsg.Printf("created file %s.\n", c.String("file"))
	}
}

func cmdTypeAdd(c *cli.Context) {
	// Check obligatory flags (file, name)
	if c.String("file") == "" {
		printError.Fatalln(errMissingFileFlag)

	}
	if c.String("type") == "" {
		printError.Fatalln(errMissingTypeFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Add new type
	sqlAddType := fmt.Sprintf("INSERT INTO bicycle_types VALUES (NULL, '%s');", c.String("type"))
	_, err = f.Handler.Exec(sqlAddType)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		printUserMsg.Printf("added new bicycle type: %s\n", c.String("type"))
	}
}

func cmdTypeList(c *cli.Context) {
	// Check obligatory flags (file)
	if c.String("file") == "" {
		printError.Fatalln(errMissingFileFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Create formatting strings
	var maxLId, maxLName int
	err = f.Handler.QueryRow("SELECT max(length(id)), max(length(name)) FROM bicycle_types;").Scan(&maxLId, &maxLName)
	if err != nil {
		printError.Fatalln("no bicycle types")
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
		printError.Fatalln(errReadingFromFile)
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
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id < 0 {
		printError.Fatalln(errMissingIdFlag)
	}
	newName := c.String("type")
	if newName == "" {
		printError.Fatalln(errMissingTypeFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Edit bicycle type
	sqlUpdateType := fmt.Sprintf("UPDATE bicycle_types SET name='%s' WHERE id=%d;", newName, id)
	r, err := f.Handler.Exec(sqlUpdateType)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoBicycleWithID)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		printUserMsg.Printf("change bicycle type name to '%s'\n", newName)
	}
}

func cmdTypeDelete(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == "" {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id < 0 {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	//Check if it is possible to safely remove the bicycle type
	if typePossibleToDelete(f.Handler, id) == false {
		printError.Fatalln(errCannotRemoveBicycleType)
	}

	// Delete bicycle type
	sqlDeleteType := fmt.Sprintf("DELETE FROM bicycle_types WHERE id=%d;", id)
	r, err := f.Handler.Exec(sqlDeleteType)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoBicycleTypeWithID)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		printUserMsg.Printf("deleted bicycle type with id = %d\n", id)
	}
}

func cmdCategoryAdd(c *cli.Context) {
	// Check obligatory flags (file, name)
	if c.String("file") == "" {
		printError.Fatalln(errMissingFileFlag)
	}
	if c.String("category") == "" {
		printError.Fatalln(errMissingCategoryFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Add new category
	sqlAddCategory := fmt.Sprintf("INSERT INTO trip_categories VALUES (NULL, '%s');", c.String("category"))
	_, err = f.Handler.Exec(sqlAddCategory)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		printUserMsg.Printf("added new trip category: %s\n", c.String("category"))
	}
}

func cmdCategoryList(c *cli.Context) {
	// Check obligatory flags (file)
	if c.String("file") == "" {
		printError.Fatalln(errMissingFileFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Create formatting strings
	var maxLId, maxLName int
	err = f.Handler.QueryRow("SELECT max(length(id)), max(length(name)) FROM trip_categories;").Scan(&maxLId, &maxLName)
	if err != nil {
		printError.Fatalln("no trip categories")
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
		printError.Fatalln(errReadingFromFile)
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
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id < 0 {
		printError.Fatalln(errMissingIdFlag)
	}
	newName := c.String("category")
	if newName == "" {
		printError.Fatalln(errMissingCategoryFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Edit trip category
	sqlUpdateCategory := fmt.Sprintf("UPDATE trip_categories SET name='%s' WHERE id=%d;", newName, id)
	r, err := f.Handler.Exec(sqlUpdateCategory)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoCategoryWithID)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		printUserMsg.Printf("change trip category name to '%s'\n", newName)
	}
}

func cmdCategoryDelete(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == "" {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id < 0 {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Check if it is possible to safely remove the category
	if categoryPossibleToDelete(f.Handler, id) == false {
		printError.Fatalln(errCannotRemoveCategory)
	}

	// Delete trip category
	sqlDeleteCategory := fmt.Sprintf("DELETE FROM trip_categories WHERE id=%d;", id)
	r, err := f.Handler.Exec(sqlDeleteCategory)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoCategoryWithID)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		printUserMsg.Printf("deleted trip category with id = %d\n", id)
	}
}

func cmdBicycleAdd(c *cli.Context) {
	// Check obligatory flags (file, bicycle, bicycle type)
	if c.String("file") == "" {
		printError.Fatalln(errMissingFileFlag)
	}
	bName := c.String("bicycle")
	if bName == "" {
		printError.Fatalln(errMissingBicycleFlag)
	}
	bType := c.String("type")
	if bType == "" {
		printError.Fatalln(errMissingTypeFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Add new bicycle
	bTypeId, err := bicycleTypeIDForName(f.Handler, bType)
	if err != nil {
		printError.Fatalln(err)
	}
	sqlAddBicycle := fmt.Sprintf("BEGIN TRANSACTION;INSERT INTO bicycles (id, name, bicycle_type_id) VALUES (NULL, '%s', %d);", bName, bTypeId)
	bManufacturer := c.String("manufacturer")
	if bManufacturer != "" {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET producer='%s' WHERE id=last_insert_rowid();", bManufacturer)
	}
	bModel := c.String("model")
	if bModel != "" {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET model='%s' WHERE id=last_insert_rowid();", bModel)
	}
	bYear := c.Int("year")
	if bYear != 0 {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET production_year=%d WHERE id=last_insert_rowid();", bYear)
	}
	bBought := c.String("bought")
	if bBought != "" {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET buying_date='%s' WHERE id=last_insert_rowid();", bBought)
	}
	bDesc := c.String("description")
	if bDesc != "" {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET description='%s' WHERE id=last_insert_rowid();", bDesc)
	}
	bSize := c.String("size")
	if bSize != "" {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET size='%s' WHERE id=last_insert_rowid();", bSize)
	}
	bWeight := c.Float64("weight")
	if bWeight != 0 {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET weight=%f WHERE id=last_insert_rowid();", bWeight)
	}
	bIDist := c.Float64("init_distance")
	if bIDist != 0 {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET initial_distance=%f WHERE id=last_insert_rowid();", bIDist)
	}
	bSeries := c.String("series")
	if bSeries != "" {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET series_no='%s' WHERE id=last_insert_rowid();", bSeries)
	}
	sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET status=%d WHERE id=last_insert_rowid();COMMIT;", bicycleStatuses["owned"])
	//TODO: change possibility to add status and if missing set by default to owned
	_, err = f.Handler.Exec(sqlAddBicycle)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		printUserMsg.Printf("added new bicycle: %s\n", bName)
	}
}

func cmdBicycleList(c *cli.Context) {
	// Check obligatory flags (file)
	if c.String("file") == "" {
		printError.Fatalln(errMissingFileFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Create formatting strings
	var lId, lName, lProducer, lModel, lType int
	maxQuery := fmt.Sprintf("SELECT max(length(b.id)), max(length(b.name)), ifnull(max(length(b.producer)),0), ifnull(max(length(b.model)),0), ifnull(max(length(t.name)),0) FROM bicycles b LEFT JOIN bicycle_types t ON b.bicycle_type_id=t.id WHERE b.status=%d;", bicycleStatuses["owned"])
	//TODO: add condition that if flag --all then all bicycles are listed
	err = f.Handler.QueryRow(maxQuery).Scan(&lId, &lName, &lProducer, &lModel, &lType)
	if err != nil {
		printError.Fatalln("no bicycles")
	}
	if hl := utf8.RuneCountInString(bcIdHeader); lId < hl {
		lId = hl
	}
	if hl := utf8.RuneCountInString(bcNameHeader); lName < hl {
		lName = hl
	}
	if hl := utf8.RuneCountInString(bcProducerHeader); lProducer < hl {
		lProducer = hl
	}
	if hl := utf8.RuneCountInString(bcModelHeader); lModel < hl {
		lModel = hl
	}
	if hl := utf8.RuneCountInString(btNameHeader); lType < hl {
		lType = hl
	}
	fmtStrings := make(map[string]string)
	fmtStrings["id"] = fmt.Sprintf("%%%dv", lId)
	fmtStrings["name"] = fmt.Sprintf("%%-%dv", lName)
	fmtStrings["producer"] = fmt.Sprintf("%%-%dv", lProducer)
	fmtStrings["model"] = fmt.Sprintf("%%-%dv", lModel)
	fmtStrings["type"] = fmt.Sprintf("%%-%dv", lType)

	// List bicycles
	rows, err := f.Handler.Query(fmt.Sprintf("SELECT b.id, b.name, b.producer, b.model, t.name FROM bicycles b LEFT JOIN bicycle_types t ON b.bicycle_type_id=t.id WHERE b.status=%d;", bicycleStatuses["owned"]))
	if err != nil {
		printError.Fatalln(errReadingFromFile)
	}
	defer rows.Close()
	line := strings.Join([]string{fmtStrings["id"], fmtStrings["name"], fmtStrings["producer"], fmtStrings["model"], fmtStrings["type"]}, fsSeparator) + "\n"
	fmt.Fprintf(os.Stdout, line, bcIdHeader, bcNameHeader, bcProducerHeader, bcModelHeader, btNameHeader)

	for rows.Next() {
		var id int
		var name, producer, model, bicType string
		rows.Scan(&id, &name, &producer, &model, &bicType)
		fmt.Fprintf(os.Stdout, line, id, name, producer, model, bicType)
	}
}

func cmdBicycleEdit(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == "" {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id < 0 {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Edit bicycle
	sqlUpdateBicycle := fmt.Sprintf("BEGIN TRANSACTION;")
	bType := c.String("type")
	if bType != "" {
		bTypeId, err := bicycleTypeIDForName(f.Handler, bType)
		if err != nil {
			printError.Fatalln(err)
		}
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET bicycle_type_id=%d WHERE id=%d;", bTypeId, id)
	}
	bStatus := c.String("status")
	if bStatus != "" {
		bStatusId, err := bicycleStatusNoForName(bStatus)
		if err != nil {
			printError.Fatalln(err)
		}
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET status=%d WHERE id=%d;", bStatusId, id)
	}
	bName := c.String("bicycle")
	if bName != "" {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET name=%s WHERE id=%d;", bName, id)
	}
	bManufacturer := c.String("manufacturer")
	if bManufacturer != "" {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET producer='%s' WHERE id=%d;", bManufacturer, id)
	}
	bModel := c.String("model")
	if bModel != "" {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET model='%s' WHERE id=%d;", bModel, id)
	}
	bYear := c.Int("year")
	if bYear != 0 {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET production_year=%d WHERE id=%d;", bYear, id)
	}
	bBought := c.String("bought")
	if bBought != "" {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET buying_date='%s' WHERE id=%d;", bBought, id)
	}
	bDesc := c.String("description")
	if bDesc != "" {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET description='%s' WHERE id=%d;", bDesc, id)
	}
	bSize := c.String("size")
	if bSize != "" {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET size='%s' WHERE id=%d;", bSize, id)
	}
	bWeight := c.Float64("weight")
	if bWeight != 0 {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET weight=%f WHERE id=%d;", bWeight, id)
	}
	bIDist := c.Float64("init_distance")
	if bIDist != 0 {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET initial_distance=%f WHERE id=%d;", bIDist, id)
	}
	bSeries := c.String("series")
	if bSeries != "" {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET series_no='%s' WHERE id=%d;", bSeries, id)
	}
	sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("COMMIT;")
	r, err := f.Handler.Exec(sqlUpdateBicycle)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoBicycleWithID)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		printUserMsg.Printf("changed bicycle details\n")
	}
}

func cmdBicycleDelete(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == "" {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id < 0 {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Check if it is possible to safely delete the bicycle
	if bicyclePossibleToDelete(f.Handler, id) == false {
		printError.Fatalln(errCannotRemoveBicycle)
	}

	// Delete bicycle type
	sqlDeleteBicycle := fmt.Sprintf("DELETE FROM bicycles WHERE id=%d;", id)
	r, err := f.Handler.Exec(sqlDeleteBicycle)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoBicycleWithID)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		printUserMsg.Printf("deleted bicycle with id = %d\n", id)
	}
}

func cmdTripAdd(c *cli.Context) {
	//TODO: refactor all not-set values to constants in all functions
	// Check obligatory flags (file, title, bicycle, trip category, distance)
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	tTitle := c.String("title")
	if tTitle == notSetStringValue {
		printError.Fatalln(errMissingTitleFlag)
	}
	tBicycle := c.String("bicycle")
	if tBicycle == notSetStringValue {
		printError.Fatalln(errMissingBicycleFlag)
	}
	tCategory := c.String("category")
	if tCategory == notSetStringValue {
		printError.Fatalln(errMissingCategoryFlag)
	}
	tDistance := c.Float64("distance")
	if tDistance == notSetFloatValue {
		printError.Fatalln(errMissingDistanceFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Add new trip
	tBicycleId, err := bicycleIDForName(f.Handler, tBicycle)
	if err != nil {
		printError.Fatalln(err)
	}
	tCategoryId, err := tripCategoryIDForName(f.Handler, tCategory)
	if err != nil {
		printError.Fatalln(err)
	}

	sqlAddTrip := fmt.Sprintf("BEGIN TRANSACTION;")
	sqlAddTrip = sqlAddTrip + fmt.Sprintf("INSERT INTO trips (id, bicycle_id, date,title, trip_category_id, distance) VALUES (NULL, %d, '%s', '%s', %d, %f);", tBicycleId, c.String("date"), tTitle, tCategoryId, tDistance)
	tDuration := c.Float64("duration")
	if tDuration != notSetFloatValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET duration=%f WHERE id=last_insert_rowid();", tDuration)
	}
	tDescription := c.String("description")
	if tDescription != notSetStringValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET description='%s' WHERE id=last_insert_rowid();", tDescription)
	}
	tHRMax := c.Int("hrmax")
	if tHRMax != notSetIntValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET hr_max=%d WHERE id=last_insert_rowid();", tHRMax)
	}
	tHRAvg := c.Int("hravg")
	if tHRAvg != notSetIntValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET hr_avg=%d WHERE id=last_insert_rowid();", tHRAvg)
	}
	tSpeedMax := c.Float64("speed_max")
	if tSpeedMax != notSetFloatValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET speed_max=%f WHERE id=last_insert_rowid();", tSpeedMax)
	}
	tDriveways := c.Float64("driveways")
	if tDriveways != notSetFloatValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET driveways=%f WHERE id=last_insert_rowid();", tDriveways)
	}
	tCalories := c.Int("calories")
	if tCalories != notSetIntValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET calories=%d WHERE id=last_insert_rowid();", tCalories)
	}
	tTemperature := c.Float64("temperature")
	if tTemperature != notSetFloatValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET temperature=%f WHERE id=last_insert_rowid();", tTemperature)
	}
	sqlAddTrip = sqlAddTrip + fmt.Sprintf("COMMIT;")

	_, err = f.Handler.Exec(sqlAddTrip)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		printUserMsg.Printf("added new trip: '%s'\n", tTitle)
	}
}

// ********************
// Supportive Functions
// ********************

// bicycleIDForName returns bicycle id for a given (part of) name.
// db - SQL database handler
// n - bicycle name, or part of its name
func bicycleIDForName(db *sql.DB, n string) (int, error) {
	var id int = notSetIntValue

	// Find all IDs of types that match '*n*'
	sqlGetIdQuery := fmt.Sprintf("SELECT id FROM bicycles WHERE name LIKE '%%%s%%';", n)
	rows, err := db.Query(sqlGetIdQuery)
	if err != nil {
		return id, errors.New(errReadingFromFile)
	}
	defer rows.Close()

	var i int = 0
	for rows.Next() {
		rows.Scan(&id)
		i++
	}

	switch i {
	case 0:
		return id, errors.New(errNoBicycleForName)
	case 1:
		return id, nil
	default:
		return id, errors.New(errBicycleNameisAmbiguous)
	}
}

// bicycleTypeIDForName returns bicycle type id for a given (part of) name.
// db - SQL database handler
// n - bicycle type name, or part of its name
func bicycleTypeIDForName(db *sql.DB, n string) (int, error) {
	var id int = notSetIntValue

	// Find all IDs of types that match '*n*'
	sqlGetIdQuery := fmt.Sprintf("SELECT id FROM bicycle_types WHERE name LIKE '%%%s%%';", n)
	rows, err := db.Query(sqlGetIdQuery)
	if err != nil {
		return id, errors.New(errReadingFromFile)
	}
	defer rows.Close()

	var i int = 0
	for rows.Next() {
		rows.Scan(&id)
		i++
	}

	switch i {
	case 0:
		return id, errors.New(errNoBicycleTypesForName)
	case 1:
		return id, nil
	default:
		return id, errors.New(errBicycleTypeNameIsAmbiguous)
	}
}

// tripCategoryIDForName returns trip category id for a given (part of) name.
// db - SQL database handler
// n - trip category name, or part of its name
func tripCategoryIDForName(db *sql.DB, n string) (int, error) {
	var id int = notSetIntValue

	// Find all IDs of types that match '*n*'
	sqlGetIdQuery := fmt.Sprintf("SELECT id FROM trip_categories WHERE name LIKE '%%%s%%';", n)
	rows, err := db.Query(sqlGetIdQuery)
	if err != nil {
		return id, errors.New(errReadingFromFile)
	}
	defer rows.Close()

	var i int = 0
	for rows.Next() {
		rows.Scan(&id)
		i++
	}

	switch i {
	case 0:
		return id, errors.New(errNoCategoryForName)
	case 1:
		return id, nil
	default:
		return id, errors.New(errCategoryNameIsAmbiguous)
	}
}

// typePossibleToDelete returns false if there is any bicycle of a type with given ID.
// db - SQL database handler
// id - bicycle type ID
func typePossibleToDelete(db *sql.DB, id int) bool {
	var n int

	// Check how many bicycle are of that type
	nQuery := fmt.Sprintf("SELECT count(id) FROM bicycles WHERE bicycle_type_id=%d;", id)
	err := db.QueryRow(nQuery).Scan(&n)
	if err != nil {
		return false
	}

	// If there is any bicycle of that type - return false
	if n != 0 {
		return false
	}

	return true

}

// categoryPossibleToDelete returns false if there is any trip done on a bicycle with given ID.
// db - SQL database handler
// id - category ID
func categoryPossibleToDelete(db *sql.DB, id int) bool {
	var n int

	// Check how many trips are classified with this category
	nQuery := fmt.Sprintf("SELECT count(id) FROM trips WHERE trip_category_id=%d;", id)
	err := db.QueryRow(nQuery).Scan(&n)
	if err != nil {
		return false
	}

	// If there is any trip classified with this category - return false
	if n != 0 {
		return false
	}

	return true

}

// bicyclePossibleToDelete returns false if there is any trip done on a bicycle with given ID.
// db - SQL database handler
// id - bicycle ID
func bicyclePossibleToDelete(db *sql.DB, id int) bool {
	var n int

	// Check how many trips are done on this bicycle
	nQuery := fmt.Sprintf("SELECT count(id) FROM trips WHERE bicycle_id=%d;", id)
	err := db.QueryRow(nQuery).Scan(&n)
	if err != nil {
		return false
	}

	// If there is any trip done on this bike - return false
	if n != 0 {
		return false
	}

	return true

}

// bicycleStatusNoForName returns status id for given (part of) status name
// n - (part of) status name
func bicycleStatusNoForName(n string) (int, error) {
	var val int = -1

	for key, val := range bicycleStatuses {
		if strings.Contains(key, n) {
			return val, nil
		}
	}

	return val, errors.New(errNoBicycleStatus)
}
