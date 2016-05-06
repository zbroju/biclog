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
//DONE: command - report summary
//DONE: command - report history
//DONE: fix issue so that searching by bicycle name, trip category, bicycle type is irrespective of capitals
//DONE: move all function except for main() to /lib folder
package main

import (
	"github.com/codegangsta/cli"
	. "github.com/zbroju/biclog/src"
	"os"
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

	objectReportSummary      = "summary"
	objectReportSummaryAlias = "s"
	objectReportYearly       = "yearly"
	objectReportYearlyAlias  = "y"
	objectReportMonthly      = "monthly"
	objectReportMonthlyAlias = "m"
)

func main() {
	// Get error logger
	_, printError := GetLoggers()

	// Get config settings
	dataFile, err := GetConfigSettings()
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
	app.Name = AppName
	app.Usage = "keeps track of you bike rides"
	app.Version = "1.0.0"
	app.Authors = []cli.Author{
		cli.Author{"Marcin 'Zbroju' Zbroinski", "marcin@zbroinski.net"},
	}

	flagFile := cli.StringFlag{Name: "file, f", Value: dataFile, Usage: "data file"}
	flagType := cli.StringFlag{Name: "type, t", Value: NotSetStringValue, Usage: "bicycle type"}
	flagCategory := cli.StringFlag{Name: "category, c", Value: NotSetStringValue, Usage: "trip category"}
	flagId := cli.IntFlag{Name: "id, i", Value: NotSetIntValue, Usage: "ID of an object"}
	flagBicycle := cli.StringFlag{Name: "bicycle, b", Value: NotSetStringValue, Usage: "bicycle name"}
	flagManufacturer := cli.StringFlag{Name: "manufacturer", Value: NotSetStringValue, Usage: "bicycle manufacturer"}
	flagModel := cli.StringFlag{Name: "model", Value: NotSetStringValue, Usage: "bicycle model"}
	flagProductionYear := cli.IntFlag{Name: "year", Value: NotSetIntValue, Usage: "year when the bike was made"}
	flagBuyingDate := cli.StringFlag{Name: "bought", Value: NotSetStringValue, Usage: "date when the bike was bought"}
	flagDescription := cli.StringFlag{Name: "description, d", Value: NotSetStringValue, Usage: "more verbose description"}
	flagStatus := cli.StringFlag{Name: "status", Value: NotSetStringValue, Usage: "bicycle status (owned, sold, scrapped, stolen)"}
	flagSize := cli.StringFlag{Name: "size", Value: NotSetStringValue, Usage: "size of the bike"}
	flagWeight := cli.Float64Flag{Name: "weight", Value: NotSetFloatValue, Usage: "bike's weight"}
	flagInitialDistance := cli.Float64Flag{Name: "init_distance", Value: NotSetFloatValue, Usage: "initial distance of the bike"}
	flagSeries := cli.StringFlag{Name: "series", Value: NotSetStringValue, Usage: "series number"}
	flagAll := cli.BoolFlag{Name: "all, a", Usage: "switch to all"}
	flagDate := cli.StringFlag{Name: "date", Value: NotSetStringValue, Usage: "date of trip (default: today)"}
	flagTitle := cli.StringFlag{Name: "title, s", Value: NotSetStringValue, Usage: "trip title"}
	flagDistance := cli.Float64Flag{Name: "distance, r", Value: NotSetFloatValue, Usage: "trip distance"}
	flagDuration := cli.StringFlag{Name: "duration, l", Value: NotSetStringValue, Usage: "trip duration"}
	flagHRMax := cli.IntFlag{Name: "hrmax", Value: NotSetIntValue, Usage: "hr max"}
	flagHRAvg := cli.IntFlag{Name: "hravg", Value: NotSetIntValue, Usage: "hr average"}
	flagSpeedMax := cli.Float64Flag{Name: "speed_max", Value: NotSetFloatValue, Usage: "maximum speed"}
	flagDriveways := cli.Float64Flag{Name: "driveways", Value: NotSetFloatValue, Usage: "sum of driveways"}
	flagCalories := cli.IntFlag{Name: "calories", Value: NotSetIntValue, Usage: "sum of calories burnt"}
	flagTemperature := cli.Float64Flag{Name: "temperature", Value: NotSetFloatValue, Usage: "average temperature"}

	app.Commands = []cli.Command{
		{Name: "init",
			Aliases: []string{"I"},
			Flags:   []cli.Flag{flagFile},
			Usage:   "Init a new data file specified by the user",
			Action:  CmdInit},
		{Name: "add", Aliases: []string{"A"}, Usage: "Add an object (bicycle, bicycle type, trip, trip category).",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagFile, flagType},
					Usage:   "Add new bicycle type.",
					Action:  CmdTypeAdd},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagFile, flagCategory},
					Usage:   "Add new trip category.",
					Action:  CmdCategoryAdd},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagBicycle, flagManufacturer, flagModel, flagType, flagProductionYear, flagBuyingDate, flagDescription, flagSize, flagWeight, flagInitialDistance, flagSeries},
					Usage:   "Add new bicycle.",
					Action:  CmdBicycleAdd},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagTitle, flagBicycle, flagDate, flagCategory, flagDistance, flagDuration, flagDescription, flagHRMax, flagHRAvg, flagSpeedMax, flagDriveways, flagCalories, flagTemperature},
					Usage:   "Add new trip.",
					Action:  CmdTripAdd}}},
		{Name: "list", Aliases: []string{"L"}, Usage: "List objects (bicycles, bicycle types, trips, trips categories)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagFile},
					Usage:   "List available bicycle types.",
					Action:  CmdTypeList},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagFile},
					Usage:   "List available trip categories.",
					Action:  CmdCategoryList},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagBicycle, flagManufacturer, flagModel, flagType, flagAll},
					Usage:   "List available bicycles.",
					Action:  CmdBicycleList},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagType, flagCategory, flagBicycle, flagDate},
					Usage:   "List available trips.",
					Action:  CmdTripList}}},
		{Name: "edit", Aliases: []string{"E"}, Usage: "Edit an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagType},
					Usage:   "Edit bicycle type with given id.",
					Action:  CmdTypeEdit},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagCategory},
					Usage:   "Edit trip category with given id.",
					Action:  CmdCategoryEdit},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagBicycle, flagManufacturer, flagModel, flagType, flagProductionYear, flagBuyingDate, flagDescription, flagStatus, flagSize, flagWeight, flagInitialDistance, flagSeries},
					Usage:   "Edit bicycle details.",
					Action:  CmdBicycleEdit},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagBicycle, flagDate, flagTitle, flagCategory, flagDistance, flagDuration, flagDescription, flagHRMax, flagHRAvg, flagSpeedMax, flagDriveways, flagCalories, flagTemperature},
					Usage:   "Edit trip details.",
					Action:  CmdTripEdit}}},
		{Name: "delete", Aliases: []string{"D"}, Usage: "Delete an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{Name: objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Delete bicycle type with given id.",
					Action:  CmdTypeDelete},
				{Name: objectTripCategory,
					Aliases: []string{objectTripCategoryAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Delete trip category with given id.",
					Action:  CmdCategoryDelete},
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Delete bicycle with given id.",
					Action:  CmdBicycleDelete},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Delete trip with given id.",
					Action:  CmdTripDelete}}},
		{Name: "show", Aliases: []string{"S"}, Usage: "Show details of an object (bicycle, trip)",
			Subcommands: []cli.Command{
				{Name: objectBicycle,
					Aliases: []string{objectBicycleAlias},
					Flags:   []cli.Flag{flagFile, flagId, flagBicycle},
					Usage:   "Shows details of bicycle with given id or bicycle.",
					Action:  CmdBicycleShow},
				{Name: objectTrip,
					Aliases: []string{objectTripAlias},
					Flags:   []cli.Flag{flagFile, flagId},
					Usage:   "Shows details of trip with given id.",
					Action:  CmdTripShow}}},
		{Name: "report", Aliases: []string{"R"}, Usage: "Show report",
			Subcommands: []cli.Command{
				{Name: objectReportSummary,
					Aliases: []string{objectReportSummaryAlias},
					Flags:   []cli.Flag{flagFile, flagType, flagCategory, flagBicycle, flagDate},
					Usage:   "Shows summary of distance per bicycle.",
					Action:  ReportSummary},
				{Name: objectReportMonthly,
					Aliases: []string{objectReportMonthlyAlias},
					Flags:   []cli.Flag{flagFile, flagType, flagCategory, flagBicycle, flagDate},
					Usage:   "Shows summary of distance per month.",
					Action:  ReportMonthly},
				{Name: objectReportYearly,
					Aliases: []string{objectReportYearlyAlias},
					Flags:   []cli.Flag{flagFile, flagType, flagCategory, flagBicycle, flagDate},
					Usage:   "Shows summary of distance per year.",
					Action:  ReportYearly},
			}}}
	app.Run(os.Args)
}
