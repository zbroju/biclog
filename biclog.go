// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
//
//TASKS:
//TODO: change communication to 'log' library
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
//TODO: command - bicycle list
//TODO: command - bicycle edit
//TODO: command - bicycle delete
//TODO: command - bicycle show details
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

	errWritingToFile              = "error writing to file"
	errReadingFromFile            = "error reading to file"
	errNoBicycleWithID            = "no bicycle with given id"
	errNoCategoriesWithID         = "no trip categories with given id"
	errNoBicycleTypesForName      = "no bicycle types for given name"
	errBicycleTypeNameIsAmbiguous = "given bicycle type name is ambiguous"
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
)

// Bicycle statuses
const (
	bicycleStatusOwned = 1
	//bicycleStatusSold     = 2
	//bicycleStatusStolen   = 3
	//bicycleStatusScrapped = 4
)

// Application internal settings
const (
	appName     = "biclog"
	fsSeparator = "  "
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
)

// DB Properties
var dataFileProperties = map[string]string{
	"applicationName": "gBicLog",
	"databaseVersion": "1.0",
}

func main() {
	// Set up logger
	printUserMsg = log.New(os.Stdout, fmt.Sprintf("%s: ", appName), log.LstdFlags)
	printError = log.New(os.Stderr, fmt.Sprintf("%s: ", appName), log.LstdFlags)

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

	app := cli.NewApp()
	app.Name = appName
	app.Usage = "keeps track of you bike rides"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{"Marcin 'Zbroju' Zbroinski", "marcin@zbroinski.net"},
	}

	flagVerbose := cli.BoolFlag{Name: "verbose, v", Usage: "show more output", Destination: &verbose}
	flagFile := cli.StringFlag{Name: "file, f", Value: dataFile, Usage: "data file"}
	flagType := cli.StringFlag{Name: "type, t", Value: "", Usage: "bicycle type"}
	flagCategory := cli.StringFlag{Name: "category, c", Value: "", Usage: "trip category"}
	flagId := cli.IntFlag{Name: "id, i", Value: -1, Usage: "ID of an object"}
	flagBicycle := cli.StringFlag{Name: "bicycle, b", Value: "", Usage: "bicycle name"}
	flagManufacturer := cli.StringFlag{Name: "manufacturer", Value: "", Usage: "bicycle manufacturer"}
	flagModel := cli.StringFlag{Name: "model", Value: "", Usage: "bicycle model"}
	flagProductionYear := cli.IntFlag{Name: "year", Value: 0, Usage: "year when the bike was made"}
	flagBuyingDate := cli.StringFlag{Name: "bought", Value: "", Usage: "date when the bike was bought"}
	flagDescription := cli.StringFlag{Name: "description, d", Value: "", Usage: "more verbose description"}
	flagSize := cli.StringFlag{Name: "size", Value: "", Usage: "size of the bike"}
	flagWeight := cli.Float64Flag{Name: "weight", Value: 0, Usage: "bike's weight"}
	flagInitialDistance := cli.Float64Flag{Name: "init_distance", Value: 0, Usage: "initial distance of the bike"}
	flagSeries := cli.StringFlag{Name: "series", Value: "", Usage: "series number"}

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
					Action:  cmdBicycleAdd}}},
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
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId, flagType},
					Usage:   "Edit bicycle type with given id.",
					Action:  cmdTypeEdit},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId, flagCategory},
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

	// Delete bicycle type
	sqlDeleteType := fmt.Sprintf("DELETE FROM bicycle_types WHERE id=%d;", id)
	r, err := f.Handler.Exec(sqlDeleteType)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoBicycleWithID)
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
		printError.Fatalln(errNoCategoriesWithID)
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

	// Delete trip category
	sqlDeleteCategory := fmt.Sprintf("DELETE FROM trip_categories WHERE id=%d;", id)
	r, err := f.Handler.Exec(sqlDeleteCategory)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoCategoriesWithID)
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
	sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET status=%d WHERE id=last_insert_rowid();COMMIT;", bicycleStatusOwned)
	_, err = f.Handler.Exec(sqlAddBicycle)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		printUserMsg.Printf("added new bicycle: %s\n", bName)
	}
}

func bicycleTypeIDForName(db *sql.DB, n string) (int, error) {
	var id int = -1

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
