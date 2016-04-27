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
//DONE: command - bicycle show details
//DONE: command - trip add
//DONE: command - trip list
//DONE: command - trip edit
//DONE: command - trip delete
//DONE: command - trip show details
//TODO: command - report summary
//TODO: command - report history
//TODO: command - report pie chart (share of bicycles)
//TODO: command - report bar chart (history)
//DONE: fix issue so that searching by bicycle name, trip category, bicycle type is irrespective of capitals
//TODO: move all function except for main() to /lib folder
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
	"strings"
	"time"
	"unicode/utf8"
)

// Error messages
const (
	errMissingFileFlag        = "missing information about data file. Specify it with --file or -f flag"
	errMissingTypeFlag        = "missing bicycle type. Specify it with --type or -t flag"
	errMissingCategoryFlag    = "missing trip category. Specify it with --category or -c flag"
	errMissingIdFlag          = "missing id. Specify it with --id or -i flag"
	errMissingBicycleFlag     = "missing bicycle. Specify it with --bicycle or -b flag"
	errMissingBicycleOrIdFlag = "missing bicycle or id flag. Specify it with --bicycle (-b) or --id (-i) flag"
	errMissingTitleFlag       = "missing trip title. Specify it with --title or -s flag"
	errMissingDistanceFlag    = "missing trip distance. Specify it with --distance or -d flag"
	errBothIdAndBicycleFlag   = "both bicycle and id flag specified. Specify only one of them."

	errWritingToFile              = "error writing to file"
	errReadingFromFile            = "error reading from file"
	errNoBicycleWithID            = "no bicycle with given id"
	errNoBicycleForName           = "no bicycle for given name"
	errBicycleNameIsAmbiguous     = "bicycle name is ambiguous"
	errNoBicycleTypeWithID        = "no bicycle type with given id"
	errNoCategoryWithID           = "no trip categories with given id"
	errNoCategoryForName          = "no trip category for given name"
	errNoBicycleStatusForId       = "no bicycle status for given id"
	errNoBicycleTypesForName      = "no bicycle types for given name"
	errNoBicycleStatus            = "unknown bicycle status"
	errBicycleStatusIsAmbiguous   = "given bicycle status is ambiguous"
	errBicycleTypeNameIsAmbiguous = "given bicycle type name is ambiguous"
	errCategoryNameIsAmbiguous    = "given trip category name is ambiguous"
	errNoTripWithID               = "no trip with given id"

	errWrongDurationFormat = "wrong duration format (should be: 00h00m00s or 00m00s)"

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

	nullDataValue = "-"

	notSetIntValue    int     = -1
	notSetFloatValue  float64 = -1
	notSetStringValue         = ""
)

// Logger
var (
	printUserMsg *log.Logger
	printError   *log.Logger
)

// Headings titles
const (
	btIdHeader   = "ID"
	btNameHeader = "TYPE"

	tcIdHeader   = "ID"
	tcNameHeader = "CATEGORY"

	bcIdHeader               = "ID"
	bcNameHeader             = "BICYCLE"
	bcProducerHeader         = "PRODUCER"
	bcModelHeader            = "MODEL"
	bcProductionYearHeading  = "PRODUCTION YEAR"
	bcBuyingDateHeading      = "BUYING DATE"
	bcDescriptionHeading     = "DESCRIPTION"
	bcStatusHeading          = "STATUS"
	bcSizeHeading            = "SIZE"
	bcWeightHeading          = "WEIGHT"
	bcInitialDistanceHeading = "INITIAL DISTANCE"
	bcSeriesHeading          = "SERIES"
	bcHeadingSize            = 20

	trpIdHeader            = "ID"
	trpDateHeader          = "DATE"
	trpTitleHeader         = "TITLE"
	trpDistanceHeader      = "DISTANCE"
	trpDurationHeading     = "DURATION"
	trpDescriptionHeading  = "DESCRIPTION"
	trpHrMaxHeading        = "HR MAX"
	trpHrAvgHeading        = "HR AVG"
	trpSpeedMaxHeading     = "MAX SPEED"
	trpDrivewaysHeading    = "DRIVEWAYS"
	trpCaloriesHeading     = "CALORIES"
	trpTemperatureHeading  = "TEMPERATURE"
	trpSpeedAverageHeading = "AVERAGE SPEED"
	trpHeadingSize         = 15
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
	//TODO: move logger to separate function and call it from respective functions where necessary

	//TODO: move reading properties to separate function but herein (in this file)
	// Loading properties from config file if exists
	configSettings := gprops.New()
	configFile, err := os.Open(path.Join(os.Getenv("HOME"), ".blrc"))
	if err == nil {
		err = configSettings.Load(configFile)
		if err != nil {
			printError.Fatalln(err)
		}
	}
	configFile.Close()
	dataFile := configSettings.GetOrDefault(confDataFile, notSetStringValue)

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
	flagAll := cli.BoolFlag{Name: "all, a", Usage: "switch to all"}
	flagDate := cli.StringFlag{Name: "date", Value: time.Now().Format("2006-01-02"), Usage: "date of trip (default: today)"}
	flagTitle := cli.StringFlag{Name: "title, s", Value: notSetStringValue, Usage: "trip title"}
	flagDistance := cli.Float64Flag{Name: "distance, r", Value: notSetFloatValue, Usage: "trip distance"}
	flagDuration := cli.StringFlag{Name: "duration, l", Value: notSetStringValue, Usage: "trip duration"}
	flagHRMax := cli.IntFlag{Name: "hrmax", Value: notSetIntValue, Usage: "hr max"}
	flagHRAvg := cli.IntFlag{Name: "hravg", Value: notSetIntValue, Usage: "hr average"}
	flagSpeedMax := cli.Float64Flag{Name: "speed_max", Value: notSetFloatValue, Usage: "maximum speed"}
	flagDriveways := cli.Float64Flag{Name: "driveways", Value: notSetFloatValue, Usage: "sum of driveways"}
	flagCalories := cli.IntFlag{Name: "calories", Value: notSetIntValue, Usage: "sum of calories burnt"}
	flagTemperature := cli.Float64Flag{Name: "temperature", Value: notSetFloatValue, Usage: "average temperature"}

	app.Commands = []cli.Command{
		{Name: "init",
			Aliases: []string{"I"},
			Flags:   []cli.Flag{flagFile},
			Usage:   "Init a new data file specified by the user",
			Action:  cmdInit},
		{Name: "add", Aliases: []string{"A"}, Usage: "Add an object (bicycle, bicycle type, trip, trip category).",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagFile, flagType},
					Usage:   "Add new bicycle type.",
					Action:  cmdTypeAdd},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagFile, flagCategory},
					Usage:   "Add new trip category.",
					Action:  cmdCategoryAdd},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagBicycle, flagManufacturer, flagModel, flagType, flagProductionYear, flagBuyingDate, flagDescription, flagSize, flagWeight, flagInitialDistance, flagSeries},
					Usage:   "Add new bicycle.",
					Action:  cmdBicycleAdd},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagTitle, flagBicycle, flagDate, flagCategory, flagDistance, flagDuration, flagDescription, flagHRMax, flagHRAvg, flagSpeedMax, flagDriveways, flagCalories, flagTemperature},
					Usage:   "Add new trip.",
					Action:  cmdTripAdd}}},
		{Name: "list", Aliases: []string{"L"}, Usage: "List objects (bicycles, bicycle types, trips, trips categories)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagFile},
					Usage:   "List available bicycle types.",
					Action:  cmdTypeList},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagFile},
					Usage:   "List available trip categories.",
					Action:  cmdCategoryList},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagAll},
					Usage:   "List available bicycles.",
					Action:  cmdBicycleList},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile},
					Usage:   "List available trips.",
					Action:  cmdTripList}}},
		{Name: "edit", Aliases: []string{"E"}, Usage: "Edit an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagType},
					Usage:   "Edit bicycle type with given id.",
					Action:  cmdTypeEdit},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagCategory},
					Usage:   "Edit trip category with given id.",
					Action:  cmdCategoryEdit},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagBicycle, flagManufacturer, flagModel, flagType, flagProductionYear, flagBuyingDate, flagDescription, flagStatus, flagSize, flagWeight, flagInitialDistance, flagSeries},
					Usage:   "Edit bicycle details.",
					Action:  cmdBicycleEdit},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagBicycle, flagDate, flagTitle, flagCategory, flagDistance, flagDuration, flagDescription, flagHRMax, flagHRAvg, flagSpeedMax, flagDriveways, flagCalories, flagTemperature},
					Usage:   "Edit trip details.",
					Action:  cmdTripEdit}}},
		{Name: "delete", Aliases: []string{"D"}, Usage: "Delete an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Delete bicycle type with given id.",
					Action:  cmdTypeDelete},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Delete trip category with given id.",
					Action:  cmdCategoryDelete},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Delete bicycle with given id.",
					Action:  cmdBicycleDelete},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Delete trip with given id.",
					Action:  cmdTripDelete}}},
		{Name: "show", Aliases: []string{"S"}, Usage: "Show details of an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagBicycle},
					Usage:   "Shows details of bicycle with given id or bicycle.",
					Action:  cmdBicycleShow},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Shows details of trip with given id.",
					Action:  cmdTripShow}}}}
	app.Run(os.Args)
}

// ********
// Commands
// ********

func cmdInit(c *cli.Context) {
	// Check the obligatory parameters and exit if missing
	if c.String("file") == notSetStringValue {
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
 , buying_date TEXT
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
 , date TEXT
 , title TEXT
 , trip_category_id INTEGER
 , distance REAL
 , duration TEXT
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

	// Show summary
	printUserMsg.Printf("created file %s.\n", c.String("file"))
}

func cmdTypeAdd(c *cli.Context) {
	// Check obligatory flags (file, name)
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)

	}
	if c.String("type") == notSetStringValue {
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

	// Show summary
	printUserMsg.Printf("added new bicycle type: %s\n", c.String("type"))
}

func cmdTypeList(c *cli.Context) {
	// Check obligatory flags (file)
	if c.String("file") == notSetStringValue {
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
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == notSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}
	newName := c.String("type")
	if newName == notSetStringValue {
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

	// Show summary
	printUserMsg.Printf("change bicycle type name to '%s'\n", newName)
}

func cmdTypeDelete(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == notSetIntValue {
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

	// Show summary
	printUserMsg.Printf("deleted bicycle type with id = %d\n", id)
}

func cmdCategoryAdd(c *cli.Context) {
	// Check obligatory flags (file, name)
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	if c.String("category") == notSetStringValue {
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

	// Show summary
	printUserMsg.Printf("added new trip category: %s\n", c.String("category"))
}

func cmdCategoryList(c *cli.Context) {
	// Check obligatory flags (file)
	if c.String("file") == notSetStringValue {
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
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == notSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}
	newName := c.String("category")
	if newName == notSetStringValue {
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

	// Show summary
	printUserMsg.Printf("change trip category name to '%s'\n", newName)
}

func cmdCategoryDelete(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == notSetIntValue {
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

	// Show summary
	printUserMsg.Printf("deleted trip category with id = %d\n", id)
}

func cmdBicycleAdd(c *cli.Context) {
	// Check obligatory flags (file, bicycle, bicycle type)
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	bName := c.String("bicycle")
	if bName == notSetStringValue {
		printError.Fatalln(errMissingBicycleFlag)
	}
	bType := c.String("type")
	if bType == notSetStringValue {
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
	if bManufacturer != notSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET producer='%s' WHERE id=last_insert_rowid();", bManufacturer)
	}
	bModel := c.String("model")
	if bModel != notSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET model='%s' WHERE id=last_insert_rowid();", bModel)
	}
	bYear := c.Int("year")
	if bYear != notSetIntValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET production_year=%d WHERE id=last_insert_rowid();", bYear)
	}
	bBought := c.String("bought")
	if bBought != notSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET buying_date='%s' WHERE id=last_insert_rowid();", bBought)
	}
	bDesc := c.String("description")
	if bDesc != notSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET description='%s' WHERE id=last_insert_rowid();", bDesc)
	}
	bStatus := c.String("status")
	if bStatus == notSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET status=%d WHERE id=last_insert_rowid();", bicycleStatuses["owned"])
	} else {
		bStatusID, err := bicycleStatusNoForName(bStatus)
		if err != nil {
			printError.Fatalln(err)
		}
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET status=%d WHERE id=last_insert_rowid();", bStatusID)
	}
	bSize := c.String("size")
	if bSize != notSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET size='%s' WHERE id=last_insert_rowid();", bSize)
	}
	bWeight := c.Float64("weight")
	if bWeight != notSetFloatValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET weight=%f WHERE id=last_insert_rowid();", bWeight)
	}
	bIDist := c.Float64("init_distance")
	if bIDist != notSetFloatValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET initial_distance=%f WHERE id=last_insert_rowid();", bIDist)
	}
	bSeries := c.String("series")
	if bSeries != notSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET series_no='%s' WHERE id=last_insert_rowid();", bSeries)
	}
	sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("COMMIT;")
	_, err = f.Handler.Exec(sqlAddBicycle)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}

	// Show summary
	printUserMsg.Printf("added new bicycle: %s\n", bName)
}

func cmdBicycleList(c *cli.Context) {
	// Check obligatory flags (file)
	if c.String("file") == notSetStringValue {
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
	var maxQuery string
	if c.Bool("all") == true {
		maxQuery = fmt.Sprintf("SELECT max(length(b.id)), max(length(b.name)), ifnull(max(length(b.producer)),0), ifnull(max(length(b.model)),0), ifnull(max(length(t.name)),0) FROM bicycles b LEFT JOIN bicycle_types t ON b.bicycle_type_id=t.id;")
	} else {
		maxQuery = fmt.Sprintf("SELECT max(length(b.id)), max(length(b.name)), ifnull(max(length(b.producer)),0), ifnull(max(length(b.model)),0), ifnull(max(length(t.name)),0) FROM bicycles b LEFT JOIN bicycle_types t ON b.bicycle_type_id=t.id WHERE b.status=%d;", bicycleStatuses["owned"])
	}
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
	fsId := fmt.Sprintf("%%%dv", lId)
	fsName := fmt.Sprintf("%%-%dv", lName)
	fsProducer := fmt.Sprintf("%%-%dv", lProducer)
	fsModel := fmt.Sprintf("%%-%dv", lModel)
	fsType := fmt.Sprintf("%%-%dv", lType)

	// List bicycles
	var listQuery string
	if c.Bool("all") == true {
		listQuery = fmt.Sprintf("SELECT b.id, ifnull(b.name,''), ifnull(b.producer,''), ifnull(b.model,''), ifnull(t.name,'') FROM bicycles b LEFT JOIN bicycle_types t ON b.bicycle_type_id=t.id;")
	} else {
		listQuery = fmt.Sprintf("SELECT b.id, ifnull(b.name,''), ifnull(b.producer,''), ifnull(b.model,''), ifnull(t.name,'') FROM bicycles b LEFT JOIN bicycle_types t ON b.bicycle_type_id=t.id WHERE b.status=%d;", bicycleStatuses["owned"])
	}
	rows, err := f.Handler.Query(listQuery)
	if err != nil {
		printError.Fatalln(errReadingFromFile)
	}
	defer rows.Close()
	line := strings.Join([]string{fsId, fsName, fsProducer, fsModel, fsType}, fsSeparator) + "\n"
	fmt.Fprintf(os.Stdout, line, bcIdHeader, bcNameHeader, bcProducerHeader, bcModelHeader, btNameHeader)

	for rows.Next() {
		var id int
		var name, producer, model, bicType string
		rows.Scan(&id, &name, &producer, &model, &bicType)
		fmt.Fprintf(os.Stdout, line, id, name, producer, model, bicType)
	}
}

//TODO: add to all queries for list/show clause 'ifnull()'
//TODO: add to all queries filters for all non-numbers attributes attributes
func cmdBicycleEdit(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == notSetIntValue {
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
	if bType != notSetStringValue {
		bTypeId, err := bicycleTypeIDForName(f.Handler, bType)
		if err != nil {
			printError.Fatalln(err)
		}
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET bicycle_type_id=%d WHERE id=%d;", bTypeId, id)
	}
	bStatus := c.String("status")
	if bStatus != notSetStringValue {
		bStatusId, err := bicycleStatusNoForName(bStatus)
		if err != nil {
			printError.Fatalln(err)
		}
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET status=%d WHERE id=%d;", bStatusId, id)
	}
	bName := c.String("bicycle")
	if bName != notSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET name='%s' WHERE id=%d;", bName, id)
	}
	bManufacturer := c.String("manufacturer")
	if bManufacturer != notSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET producer='%s' WHERE id=%d;", bManufacturer, id)
	}
	bModel := c.String("model")
	if bModel != notSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET model='%s' WHERE id=%d;", bModel, id)
	}
	bYear := c.Int("year")
	if bYear != notSetIntValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET production_year=%d WHERE id=%d;", bYear, id)
	}
	bBought := c.String("bought")
	if bBought != notSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET buying_date='%s' WHERE id=%d;", bBought, id)
	}
	bDesc := c.String("description")
	if bDesc != notSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET description='%s' WHERE id=%d;", bDesc, id)
	}
	bSize := c.String("size")
	if bSize != notSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET size='%s' WHERE id=%d;", bSize, id)
	}
	bWeight := c.Float64("weight")
	if bWeight != notSetFloatValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET weight=%f WHERE id=%d;", bWeight, id)
	}
	bIDist := c.Float64("init_distance")
	if bIDist != notSetFloatValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET initial_distance=%f WHERE id=%d;", bIDist, id)
	}
	bSeries := c.String("series")
	if bSeries != notSetStringValue {
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

	// Show summary
	printUserMsg.Printf("changed bicycle details\n")
}

func cmdBicycleDelete(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == notSetIntValue {
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

	// Show summary
	printUserMsg.Printf("deleted bicycle with id = %d\n", id)
}

func cmdBicycleShow(c *cli.Context) {
	// Check obligatory flags (file)
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	bcID := c.Int("id")
	bcBicycle := c.String("bicycle")
	if bcID == notSetIntValue && bcBicycle == notSetStringValue {
		printError.Fatalln(errMissingBicycleOrIdFlag)
	}
	if bcID != notSetIntValue && bcBicycle != notSetStringValue {
		printError.Fatalln(errBothIdAndBicycleFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Create formatting strings
	lineStr := fmt.Sprintf("%%-%ds%%-s\n", bcHeadingSize)
	lineInt := fmt.Sprintf("%%-%ds%%-d\n", bcHeadingSize)
	lineFloat := fmt.Sprintf("%%-%ds%%-.2f\n", bcHeadingSize)

	// Show bicycles
	if bcID == notSetIntValue {
		bcID, err = bicycleIDForName(f.Handler, bcBicycle)
		if err != nil {
			printError.Fatalln(err)
		}
	}
	var (
		bId, bPYear, bStatId                                           int
		bName, bProducer, bModel, bType, bBDate, bDesc, bSize, bSeries string
		bWeight, bIDist                                                float64
	)
	showQuery := fmt.Sprintf("SELECT b.id, ifnull(b.name,''), ifnull(b.producer,''), ifnull(b.model,''), ifnull(t.name,''), ifnull(b.production_year,0), ifnull(b.buying_date,0), ifnull(b.description,''), ifnull(b.status,0), ifnull(b.size,''), ifnull(b.weight,0), ifnull(b.initial_distance,0), ifnull(b.series_no,'') FROM bicycles b LEFT JOIN bicycle_types t ON b.bicycle_type_id=t.id WHERE b.id=%d;", bcID)
	err = f.Handler.QueryRow(showQuery).Scan(&bId, &bName, &bProducer, &bModel, &bType, &bPYear, &bBDate, &bDesc, &bStatId, &bSize, &bWeight, &bIDist, &bSeries)
	if err != nil {
		printError.Fatalln(errNoBicycleWithID)
	}

	fmt.Printf(lineInt, bcIdHeader, bId)     // no need for if because it's obligatory
	fmt.Printf(lineStr, bcNameHeader, bName) // no need for if because it's obligatory
	if bProducer != notSetStringValue {
		fmt.Printf(lineStr, bcProducerHeader, bProducer)
	} else {
		fmt.Printf(lineStr, bcProducerHeader, nullDataValue)
	}
	if bModel != notSetStringValue {
		fmt.Printf(lineStr, bcModelHeader, bModel)
	} else {
		fmt.Printf(lineStr, bcModelHeader, nullDataValue)
	}
	fmt.Printf(lineStr, btNameHeader, bType) // no need for if because it's obligatory
	if bPYear != 0 {
		fmt.Printf(lineInt, bcProductionYearHeading, bPYear)
	} else {
		fmt.Printf(lineStr, bcProductionYearHeading, nullDataValue)
	}
	if bBDate != notSetStringValue {
		fmt.Printf(lineStr, bcBuyingDateHeading, bBDate)
	} else {
		fmt.Printf(lineStr, bcBuyingDateHeading, nullDataValue)
	}
	fmt.Printf(lineStr, bcStatusHeading, bicycleStatusNameForID(bStatId)) // no need for if because it's obligatory
	if bSize != notSetStringValue {
		fmt.Printf(lineStr, bcSizeHeading, bSize)
	} else {
		fmt.Printf(lineStr, bcSizeHeading, nullDataValue)
	}
	if bWeight != 0 {
		fmt.Printf(lineFloat, bcWeightHeading, bWeight)
	} else {
		fmt.Printf(lineStr, bcWeightHeading, nullDataValue)
	}
	if bIDist != 0 {
		fmt.Printf(lineFloat, bcInitialDistanceHeading, bIDist)
	} else {
		fmt.Printf(lineStr, bcInitialDistanceHeading, nullDataValue)
	}
	if bSeries != notSetStringValue {
		fmt.Printf(lineStr, bcSeriesHeading, bSeries)
	} else {
		fmt.Printf(lineStr, bcSeriesHeading, nullDataValue)
	}
	if bDesc != notSetStringValue {
		fmt.Printf(lineStr, bcDescriptionHeading, bDesc)
	} else {
		fmt.Printf(lineStr, bcDescriptionHeading, nullDataValue)
	}
}

func cmdTripAdd(c *cli.Context) {
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
	tDuration := c.String("duration")
	if tDuration != notSetStringValue {
		durationValue, err := time.ParseDuration(tDuration)
		if err != nil {
			printError.Fatalln(errWrongDurationFormat)
		}
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET duration='%s' WHERE id=last_insert_rowid();", durationValue.String())
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

	// Show summary
	printUserMsg.Printf("added new trip: '%s'\n", tTitle)
}

func cmdTripList(c *cli.Context) {
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
	var lId, lDate, lTitle, lCategory, lBicycle, lDistance int
	maxQuery := fmt.Sprintf("SELECT max(length(t.id)), max(length(t.date)), max(length(t.title)), max(length(c.name)), max(length(b.name)), max(length(t.distance)) FROM trips t LEFT JOIN bicycles b ON t.bicycle_id=b.id LEFT JOIN trip_categories c ON t.trip_category_id=c.id;")
	err = f.Handler.QueryRow(maxQuery).Scan(&lId, &lDate, &lTitle, &lCategory, &lBicycle, &lDistance)
	if err != nil {
		printError.Fatalln("no trips")
	}
	if hl := utf8.RuneCountInString(bcIdHeader); lId < hl {
		lId = hl
	}
	if hl := utf8.RuneCountInString(trpDateHeader); lDate < hl {
		lDate = hl
	}
	if hl := utf8.RuneCountInString(trpTitleHeader); lTitle < hl {
		lTitle = hl
	}
	if hl := utf8.RuneCountInString(tcNameHeader); lCategory < hl {
		lCategory = hl
	}
	if hl := utf8.RuneCountInString(bcNameHeader); lBicycle < hl {
		lBicycle = hl
	}
	if hl := utf8.RuneCountInString(trpDistanceHeader); lDistance < hl {
		lDistance = hl
	}

	fsId := fmt.Sprintf("%%%dv", lId)
	fsDate := fmt.Sprintf("%%-%dv", lDate)
	fsTitle := fmt.Sprintf("%%-%dv", lTitle)
	fsCategory := fmt.Sprintf("%%-%dv", lCategory)
	fsBicycle := fmt.Sprintf("%%-%dv", lBicycle)
	fsDistance := fmt.Sprintf("%%%dv", lDistance)

	// List bicycles
	rows, err := f.Handler.Query(fmt.Sprintf("SELECT t.id, t.date, t.title, c.name, b.name, t.distance FROM trips t LEFT JOIN bicycles b ON t.bicycle_id=b.id LEFT JOIN trip_categories c ON t.trip_category_id=c.id;"))
	if err != nil {
		printError.Fatalln(errReadingFromFile)
	}
	defer rows.Close()
	line := strings.Join([]string{fsId, fsDate, fsCategory, fsBicycle, fsDistance, fsTitle}, fsSeparator) + "\n"
	fmt.Fprintf(os.Stdout, line, trpIdHeader, trpDateHeader, tcNameHeader, bcNameHeader, trpDistanceHeader, trpTitleHeader)

	for rows.Next() {
		var id int
		var date, title, category, bicycle string
		var distance float64
		rows.Scan(&id, &date, &title, &category, &bicycle, &distance)
		fmt.Fprintf(os.Stdout, line, id, date, category, bicycle, distance, title)
	}
}

func cmdTripEdit(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == notSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Edit trip
	sqlUpdateTrip := fmt.Sprintf("BEGIN TRANSACTION;")
	tCategory := c.String("category")
	if tCategory != notSetStringValue {
		tCategoryId, err := tripCategoryIDForName(f.Handler, tCategory)
		if err != nil {
			printError.Fatalln(err)
		}
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET trip_category_id=%d WHERE id=%d;", tCategoryId, id)
	}
	tBicycle := c.String("bicycle")
	if tBicycle != notSetStringValue {
		tBicycleId, err := bicycleIDForName(f.Handler, tBicycle)
		if err != nil {
			printError.Fatalln(err)
		}
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET bicycle_id=%d WHERE id=%d;", tBicycleId, id)
	}
	tDate := c.String("date")
	if tDate != notSetStringValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET date='%s' WHERE id=%d;", tDate, id)
	}
	tTitle := c.String("title")
	if tTitle != notSetStringValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET title='%s' WHERE id=%d;", tTitle, id)
	}
	tDistance := c.Float64("distance")
	if tDistance != notSetFloatValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET distance=%f WHERE id=%f;", tDistance, id)
	}
	tDuration := c.String("duration")
	if tDuration != notSetStringValue {
		durationValue, err := time.ParseDuration(tDuration)
		if err != nil {
			printError.Fatalln(errWrongDurationFormat)
		}
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET duration='%s' WHERE id=%d;", durationValue.String(), id)
	}
	tDescription := c.String("description")
	if tDescription != notSetStringValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET description='%s' WHERE id=%d;", tDescription, id)
	}
	tHrMax := c.Int("hrmax")
	if tHrMax != notSetIntValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET hr_max=%d WHERE id=%d;", tHrMax, id)
	}
	tHrAvg := c.Int("hravg")
	if tHrAvg != notSetIntValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET hr_avg=%d WHERE id=%d;", tHrAvg, id)
	}
	tSpeedMax := c.Float64("speed_max")
	if tSpeedMax != notSetFloatValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET speed_max=%f WHERE id=%d;", tSpeedMax, id)
	}
	tDriveways := c.Float64("driveways")
	if tDriveways != notSetFloatValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET driveways=%f WHERE id=%d;", tDriveways, id)
	}
	tCalories := c.Int("calories")
	if tCalories != notSetIntValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET calories=%d WHERE id=%d;", tCalories, id)
	}
	tTemperature := c.Float64("temperature")
	if tTemperature != notSetFloatValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET calories=%f WHERE id=%d;", tTemperature, id)
	}
	sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("COMMIT;")
	r, err := f.Handler.Exec(sqlUpdateTrip)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoBicycleWithID)
	}

	// Show summary
	printUserMsg.Printf("changed trip details\n")
}

func cmdTripDelete(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == notSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Delete bicycle type
	sqlDeleteTrip := fmt.Sprintf("DELETE FROM trips WHERE id=%d;", id)
	r, err := f.Handler.Exec(sqlDeleteTrip)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoTripWithID)
	}

	// Show summary
	printUserMsg.Printf("deleted tirp with id = %d\n", id)
}

func cmdTripShow(c *cli.Context) {
	// Check obligatory flags (file)
	if c.String("file") == notSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	tID := c.Int("id")
	if tID == notSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Create formatting strings
	lineStr := fmt.Sprintf("%%-%ds%%-s\n", trpHeadingSize)
	lineInt := fmt.Sprintf("%%-%ds%%-d\n", trpHeadingSize)
	lineFloat := fmt.Sprintf("%%-%ds%%-.1f\n", trpHeadingSize)

	// Show trip
	var (
		tId, tHrMax, tHrAvg, tCalories                    int
		bName, tDate, tTitle, tCategory, tDuration, tDesc string
		tDistance, tSpeedMax, tDriveways, tTemp           float64
	)
	showQuery := fmt.Sprintf("SELECT t.id, ifnull(b.name,''), ifnull(t.date,''), ifnull(t.title,''), ifnull(c.name,''), ifnull(t.distance,0), ifnull(t.duration,''), ifnull(t.description,''), ifnull(t.hr_max,0), ifnull(t.hr_avg,0), ifnull(t.speed_max,0), ifnull(t.driveways,0), ifnull(t.calories,0), ifnull(t.temperature,0) FROM trips t LEFT JOIN trip_categories c ON t.trip_category_id=c.id LEFT JOIN bicycles b ON t.bicycle_id=b.id WHERE t.id=%d;", tID)
	err = f.Handler.QueryRow(showQuery).Scan(&tId, &bName, &tDate, &tTitle, &tCategory, &tDistance, &tDuration, &tDesc, &tHrMax, &tHrAvg, &tSpeedMax, &tDriveways, &tCalories, &tTemp)
	if err != nil {
		printError.Fatalln(errNoTripWithID)
	}

	fmt.Printf(lineInt, trpIdHeader, tId)
	fmt.Printf(lineStr, bcNameHeader, bName)
	fmt.Printf(lineStr, trpDateHeader, tDate)
	fmt.Printf(lineStr, trpTitleHeader, tTitle)
	fmt.Printf(lineStr, tcNameHeader, tCategory)
	fmt.Printf(lineFloat, trpDistanceHeader, tDistance)
	if tDuration != notSetStringValue {
		fmt.Printf(lineStr, trpDurationHeading, tDuration)
		durationValue, err := time.ParseDuration(tDuration)
		if err == nil {
			fmt.Printf(lineFloat, trpSpeedAverageHeading, tDistance/durationValue.Hours())
		}
	} else {
		fmt.Printf(lineStr, trpDurationHeading, nullDataValue)
		fmt.Printf(lineStr, trpSpeedAverageHeading, nullDataValue)
	}
	if tSpeedMax != 0 {
		fmt.Printf(lineFloat, trpSpeedMaxHeading, tSpeedMax)
	} else {
		fmt.Printf(lineStr, trpSpeedMaxHeading, nullDataValue)
	}
	if tDriveways != 0 {
		fmt.Printf(lineFloat, trpDrivewaysHeading, tDriveways)
	} else {
		fmt.Printf(lineStr, trpDrivewaysHeading, nullDataValue)
	}
	if tHrMax != 0 {
		fmt.Printf(lineInt, trpHrMaxHeading, tHrMax)
	} else {
		fmt.Printf(lineStr, trpHrMaxHeading, nullDataValue)
	}
	if tHrAvg != 0 {
		fmt.Printf(lineInt, trpHrAvgHeading, tHrAvg)
	} else {
		fmt.Printf(lineStr, trpHrAvgHeading, nullDataValue)
	}
	if tCalories != 0 {
		fmt.Printf(lineInt, trpCaloriesHeading, tCalories)
	} else {
		fmt.Printf(lineStr, trpCaloriesHeading, nullDataValue)
	}
	if tTemp != 0 {
		fmt.Printf(lineFloat, trpTemperatureHeading, tTemp)
	} else {
		fmt.Printf(lineStr, trpTemperatureHeading, nullDataValue)
	}
	if tDesc != notSetStringValue {
		fmt.Printf(lineStr, trpDescriptionHeading, tDesc)
	} else {
		fmt.Printf(lineStr, trpDescriptionHeading, nullDataValue)
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
		return id, errors.New(errBicycleNameIsAmbiguous)
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
	var counter, val int

	for k, v := range bicycleStatuses {
		if strings.Contains(k, n) {
			val = v
			counter++
		}
	}

	switch counter {
	case 0:
		return notSetIntValue, errors.New(errNoBicycleStatus)
	case 1:
		return val, nil
	default:
		return notSetIntValue, errors.New(errBicycleStatusIsAmbiguous)
	}
}

// bicycleStatusNameForID returns status name for given number
// n - status no
func bicycleStatusNameForID(n int) string {

	for k, v := range bicycleStatuses {
		if n == v {
			return k
		}
	}
	return notSetStringValue
}

/*
QUERY: report summary
select b.name as bicycle, sum(t.distance) as distance from trips t left join bicycles b ON t.bicycle_id=b.id where 1=1 and t.date like '%2016%' group by bicycle;
*/
