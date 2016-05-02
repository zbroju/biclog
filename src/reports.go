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

	// SQL queries
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

	// List trip categories
	rows, err := f.Handler.Query(fmt.Sprintf("%s;", sqlQueryData))
	if err != nil {
		printError.Fatalln(errReadingFromFile)
	}
	defer rows.Close()

	lineHeader := strings.Join([]string{fsBicycle, fsType, fsDistanceHeader}, FSSeparator) + "\n"
	lineData := strings.Join([]string{fsBicycle, fsType, fsDistanceData}, FSSeparator) + "\n"
	fmt.Fprintf(os.Stdout, lineHeader, bcNameHeader, btNameHeader, trpDistanceHeader)
	for rows.Next() {
		var bicycle, bType string
		var distance float64
		rows.Scan(&bicycle, &bType, &distance)
		fmt.Fprintf(os.Stdout, lineData, bicycle, bType, distance)
	}
}
