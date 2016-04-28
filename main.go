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
//DONE: move all function except for main() to /lib folder
package main

import (
	"github.com/codegangsta/cli"
	"github.com/zbroju/biclog/src"
	"os"
	"time"
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

func main() {
	// Get error logger
	_, printError := src.GetLoggers()

	// Get config settings
	dataFile, err := src.GetConfigSettings()
	if err != nil {
		printError.Fatalln(err)
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
	app.Name = src.AppName
	app.Usage = "keeps track of you bike rides"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{"Marcin 'Zbroju' Zbroinski", "marcin@zbroinski.net"},
	}

	flagFile := cli.StringFlag{Name: "file, f", Value: dataFile, Usage: "data file"}
	flagType := cli.StringFlag{Name: "type, t", Value: src.NotSetStringValue, Usage: "bicycle type"}
	flagCategory := cli.StringFlag{Name: "category, c", Value: src.NotSetStringValue, Usage: "trip category"}
	flagId := cli.IntFlag{Name: "id, i", Value: src.NotSetIntValue, Usage: "ID of an object"}
	flagBicycle := cli.StringFlag{Name: "bicycle, b", Value: src.NotSetStringValue, Usage: "bicycle name"}
	flagManufacturer := cli.StringFlag{Name: "manufacturer", Value: src.NotSetStringValue, Usage: "bicycle manufacturer"}
	flagModel := cli.StringFlag{Name: "model", Value: src.NotSetStringValue, Usage: "bicycle model"}
	flagProductionYear := cli.IntFlag{Name: "year", Value: src.NotSetIntValue, Usage: "year when the bike was made"}
	flagBuyingDate := cli.StringFlag{Name: "bought", Value: src.NotSetStringValue, Usage: "date when the bike was bought"}
	flagDescription := cli.StringFlag{Name: "description, d", Value: src.NotSetStringValue, Usage: "more verbose description"}
	flagStatus := cli.StringFlag{Name: "status", Value: src.NotSetStringValue, Usage: "bicycle status (owned, sold, scrapped, stolen)"}
	flagSize := cli.StringFlag{Name: "size", Value: src.NotSetStringValue, Usage: "size of the bike"}
	flagWeight := cli.Float64Flag{Name: "weight", Value: src.NotSetFloatValue, Usage: "bike's weight"}
	flagInitialDistance := cli.Float64Flag{Name: "init_distance", Value: src.NotSetFloatValue, Usage: "initial distance of the bike"}
	flagSeries := cli.StringFlag{Name: "series", Value: src.NotSetStringValue, Usage: "series number"}
	flagAll := cli.BoolFlag{Name: "all, a", Usage: "switch to all"}
	flagDate := cli.StringFlag{Name: "date", Value: time.Now().Format("2006-01-02"), Usage: "date of trip (default: today)"}
	flagTitle := cli.StringFlag{Name: "title, s", Value: src.NotSetStringValue, Usage: "trip title"}
	flagDistance := cli.Float64Flag{Name: "distance, r", Value: src.NotSetFloatValue, Usage: "trip distance"}
	flagDuration := cli.StringFlag{Name: "duration, l", Value: src.NotSetStringValue, Usage: "trip duration"}
	flagHRMax := cli.IntFlag{Name: "hrmax", Value: src.NotSetIntValue, Usage: "hr max"}
	flagHRAvg := cli.IntFlag{Name: "hravg", Value: src.NotSetIntValue, Usage: "hr average"}
	flagSpeedMax := cli.Float64Flag{Name: "speed_max", Value: src.NotSetFloatValue, Usage: "maximum speed"}
	flagDriveways := cli.Float64Flag{Name: "driveways", Value: src.NotSetFloatValue, Usage: "sum of driveways"}
	flagCalories := cli.IntFlag{Name: "calories", Value: src.NotSetIntValue, Usage: "sum of calories burnt"}
	flagTemperature := cli.Float64Flag{Name: "temperature", Value: src.NotSetFloatValue, Usage: "average temperature"}

	app.Commands = []cli.Command{
		{Name: "init",
			Aliases: []string{"I"},
			Flags:   []cli.Flag{flagFile},
			Usage:   "Init a new data file specified by the user",
			Action:  src.CmdInit},
		{Name: "add", Aliases: []string{"A"}, Usage: "Add an object (bicycle, bicycle type, trip, trip category).",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagFile, flagType},
					Usage:   "Add new bicycle type.",
					Action:  src.CmdTypeAdd},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagFile, flagCategory},
					Usage:   "Add new trip category.",
					Action:  src.CmdCategoryAdd},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagBicycle, flagManufacturer, flagModel, flagType, flagProductionYear, flagBuyingDate, flagDescription, flagSize, flagWeight, flagInitialDistance, flagSeries},
					Usage:   "Add new bicycle.",
					Action:  src.CmdBicycleAdd},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagTitle, flagBicycle, flagDate, flagCategory, flagDistance, flagDuration, flagDescription, flagHRMax, flagHRAvg, flagSpeedMax, flagDriveways, flagCalories, flagTemperature},
					Usage:   "Add new trip.",
					Action:  src.CmdTripAdd}}},
		{Name: "list", Aliases: []string{"L"}, Usage: "List objects (bicycles, bicycle types, trips, trips categories)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagFile},
					Usage:   "List available bicycle types.",
					Action:  src.CmdTypeList},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagFile},
					Usage:   "List available trip categories.",
					Action:  src.CmdCategoryList},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagAll},
					Usage:   "List available bicycles.",
					Action:  src.CmdBicycleList},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile},
					Usage:   "List available trips.",
					Action:  src.CmdTripList}}},
		{Name: "edit", Aliases: []string{"E"}, Usage: "Edit an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagType},
					Usage:   "Edit bicycle type with given id.",
					Action:  src.CmdTypeEdit},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagCategory},
					Usage:   "Edit trip category with given id.",
					Action:  src.CmdCategoryEdit},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagBicycle, flagManufacturer, flagModel, flagType, flagProductionYear, flagBuyingDate, flagDescription, flagStatus, flagSize, flagWeight, flagInitialDistance, flagSeries},
					Usage:   "Edit bicycle details.",
					Action:  src.CmdBicycleEdit},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagBicycle, flagDate, flagTitle, flagCategory, flagDistance, flagDuration, flagDescription, flagHRMax, flagHRAvg, flagSpeedMax, flagDriveways, flagCalories, flagTemperature},
					Usage:   "Edit trip details.",
					Action:  src.CmdTripEdit}}},
		{Name: "delete", Aliases: []string{"D"}, Usage: "Delete an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Delete bicycle type with given id.",
					Action:  src.CmdTypeDelete},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Delete trip category with given id.",
					Action:  src.CmdCategoryDelete},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Delete bicycle with given id.",
					Action:  src.CmdBicycleDelete},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Delete trip with given id.",
					Action:  src.CmdTripDelete}}},
		{Name: "show", Aliases: []string{"S"}, Usage: "Show details of an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagBicycle},
					Usage:   "Shows details of bicycle with given id or bicycle.",
					Action:  src.CmdBicycleShow},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Shows details of trip with given id.",
					Action:  src.CmdTripShow}}}}
	app.Run(os.Args)
}