// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package src

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/zbroju/gsqlitehandler"
	"os"
	"strings"
	"unicode/utf8"
)

func ReportSummary(c *cli.Context) {
	// Get loggers
	_, printError := GetLoggers()

	// Check obligatory flags (file)
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// SQL query
	sqlSubQuery, err := sqlTripsSubQuery(f.Handler, c)
	if err != nil {
		printError.Fatalln(err)
	}
	sqlQueryData := fmt.Sprintf("SELECT bicycle, type, sum(distance) as distance from (%s) GROUP BY bicycle, type", sqlSubQuery)

	// Create formatting strings
	var maxLBicycle, maxLType, maxLDistance int
	err = f.Handler.QueryRow(fmt.Sprintf("SELECT max(length(bicycle)), max(length(type)), max(length(distance)) FROM (%s);", sqlQueryData)).Scan(&maxLBicycle, &maxLType, &maxLDistance)
	if err != nil {
		printError.Fatalln("no trips")
	}
	if hlBicycle := utf8.RuneCountInString(bcNameHeader); maxLBicycle < hlBicycle {
		maxLBicycle = hlBicycle
	}
	if hlType := utf8.RuneCountInString(btNameHeader); maxLType < hlType {
		maxLType = hlType
	}
	if hlDistance := utf8.RuneCountInString(trpDistanceHeader); maxLDistance < hlDistance {
		maxLDistance = hlDistance
	}
	fsBicycle := fmt.Sprintf("%%-%ds", maxLBicycle)
	fsType := fmt.Sprintf("%%-%ds", maxLType)
	fsDistanceHeader := fmt.Sprintf("%%%ds", maxLDistance)
	fsDistanceData := fmt.Sprintf("%%%d.1f", maxLDistance)

	// Print summary
	rows, err := f.Handler.Query(fmt.Sprintf("%s;", sqlQueryData))
	if err != nil {
		printError.Fatalln(errReadingFromFile)
	}
	defer rows.Close()

	lineHeader := strings.Join([]string{fsBicycle, fsType, fsDistanceHeader}, FSSeparator) + "\n"
	lineData := strings.Join([]string{fsBicycle, fsType, fsDistanceData}, FSSeparator) + "\n"
	fmt.Fprintf(os.Stdout, lineHeader, bcNameHeader, btNameHeader, trpDistanceHeader)
	var distanceTotal float64
	for rows.Next() {
		var bicycle, bType string
		var distance float64
		rows.Scan(&bicycle, &bType, &distance)
		fmt.Fprintf(os.Stdout, lineData, bicycle, bType, distance)
		distanceTotal += distance
	}

	// Print total distance
	fmt.Fprintf(os.Stdout, lineHeader, strings.Repeat("-", maxLBicycle), strings.Repeat("-", maxLType), strings.Repeat("-", maxLDistance))
	fmt.Fprintf(os.Stdout, lineData, "TOTAL", NotSetStringValue, distanceTotal)
}

func ReportMonthly(c *cli.Context) {
	// Get loggers
	_, printError := GetLoggers()

	// Check obligatory flags (file)
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// SQL query
	sqlSubQuery, err := sqlTripsSubQuery(f.Handler, c)
	if err != nil {
		printError.Fatalln(err)
	}
	sqlQueryData := fmt.Sprintf("SELECT strftime('%%Y-%%m', date) as month, sum(distance) as distance from (%s) GROUP BY month ORDER BY month", sqlSubQuery)

	// Create formatting strings
	var maxLMonth, maxLDistance int
	err = f.Handler.QueryRow(fmt.Sprintf("SELECT max(length(month)), max(length(distance)) FROM (%s);", sqlQueryData)).Scan(&maxLMonth, &maxLDistance)
	if err != nil {
		printError.Fatalln("no trips")
	}
	if hlMonth := utf8.RuneCountInString(trpDateHeader); maxLMonth < hlMonth {
		maxLMonth = hlMonth
	}
	if hlDistance := utf8.RuneCountInString(trpDistanceHeader); maxLDistance < hlDistance {
		maxLDistance = hlDistance
	}
	fsMonth := fmt.Sprintf("%%-%ds", maxLMonth)
	fsDistanceHeader := fmt.Sprintf("%%%ds", maxLDistance)
	fsDistanceData := fmt.Sprintf("%%%d.1f", maxLDistance)

	// Print summary
	rows, err := f.Handler.Query(fmt.Sprintf("%s;", sqlQueryData))
	if err != nil {
		printError.Fatalln(errReadingFromFile)
	}
	defer rows.Close()

	lineHeader := strings.Join([]string{fsMonth, fsDistanceHeader}, FSSeparator) + "\n"
	lineData := strings.Join([]string{fsMonth, fsDistanceData}, FSSeparator) + "\n"
	fmt.Fprintf(os.Stdout, lineHeader, trpDateHeader, trpDistanceHeader)
	var distanceTotal float64
	for rows.Next() {
		var month string
		var distance float64
		rows.Scan(&month, &distance)
		fmt.Fprintf(os.Stdout, lineData, month, distance)
		distanceTotal += distance
	}

	// Print total distance
	fmt.Fprintf(os.Stdout, lineHeader, strings.Repeat("-", maxLMonth), strings.Repeat("-", maxLDistance))
	fmt.Fprintf(os.Stdout, lineData, "SUM.", distanceTotal)
}

func ReportYearly(c *cli.Context) {
	// Get loggers
	_, printError := GetLoggers()

	// Check obligatory flags (file)
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// SQL query
	sqlSubQuery, err := sqlTripsSubQuery(f.Handler, c)
	if err != nil {
		printError.Fatalln(err)
	}
	sqlQueryData := fmt.Sprintf("SELECT strftime('%%Y', date) as year, sum(distance) as distance from (%s) GROUP BY year ORDER BY year", sqlSubQuery)

	// Create formatting strings
	var maxYear, maxLDistance int
	err = f.Handler.QueryRow(fmt.Sprintf("SELECT max(length(year)), max(length(distance)) FROM (%s);", sqlQueryData)).Scan(&maxYear, &maxLDistance)
	if err != nil {
		printError.Fatalln("no trips")
	}
	if hlYear := utf8.RuneCountInString(trpDateHeader); maxYear < hlYear {
		maxYear = hlYear
	}
	if hlDistance := utf8.RuneCountInString(trpDistanceHeader); maxLDistance < hlDistance {
		maxLDistance = hlDistance
	}
	fsYear := fmt.Sprintf("%%-%ds", maxYear)
	fsDistanceHeader := fmt.Sprintf("%%%ds", maxLDistance)
	fsDistanceData := fmt.Sprintf("%%%d.1f", maxLDistance)

	// Print summary
	rows, err := f.Handler.Query(fmt.Sprintf("%s;", sqlQueryData))
	if err != nil {
		printError.Fatalln(errReadingFromFile)
	}
	defer rows.Close()

	lineHeader := strings.Join([]string{fsYear, fsDistanceHeader}, FSSeparator) + "\n"
	lineData := strings.Join([]string{fsYear, fsDistanceData}, FSSeparator) + "\n"
	fmt.Fprintf(os.Stdout, lineHeader, trpDateHeader, trpDistanceHeader)
	var distanceTotal float64
	for rows.Next() {
		var year string
		var distance float64
		rows.Scan(&year, &distance)
		fmt.Fprintf(os.Stdout, lineData, year, distance)
		distanceTotal += distance
	}

	// Print total distance
	fmt.Fprintf(os.Stdout, lineHeader, strings.Repeat("-", maxYear), strings.Repeat("-", maxLDistance))
	fmt.Fprintf(os.Stdout, lineData, "SUM.", distanceTotal)
}
