// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"github.com/zbroju/gprops"
	"log"
	"os"
	"path"
	"strings"
)

//TODO: add report - chart of workload via gnuplot

// GetConfigSettings returns contents of settings file (~/.blrc)
func getConfigSettings() (dataFile string, err error) {
	// Read config file
	configSettings := gprops.New()
	configFile, err := os.Open(path.Join(os.Getenv("HOME"), ".blrc"))
	if err == nil {
		err = configSettings.Load(configFile)
		if err != nil {
			return NotSetStringValue, err
		}
	}
	configFile.Close()
	dataFile = configSettings.GetOrDefault(confDataFile, NotSetStringValue)

	return dataFile, nil
}

// GetLoggers returns two loggers for standard formatting of messages and errors
func getLoggers() (messageLogger *log.Logger, errorLogger *log.Logger) {
	messageLogger = log.New(os.Stdout, fmt.Sprintf("%s: ", AppName), 0)
	errorLogger = log.New(os.Stderr, fmt.Sprintf("%s: ", AppName), 0)

	return
}

// bicycleIDForName returns bicycle id for a given (part of) name.
// db - SQL database handler
// n - bicycle name, or part of its name
func bicycleIDForName(db *sql.DB, n string) (int, error) {
	var id int = NotSetIntValue

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
	var id int = NotSetIntValue

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
	var id int = NotSetIntValue

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
	if err := db.QueryRow(nQuery).Scan(&n); err != nil {
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
	if err := db.QueryRow(nQuery).Scan(&n); err != nil {
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
	if err := db.QueryRow(nQuery).Scan(&n); err != nil {
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
		return NotSetIntValue, errors.New(errNoBicycleStatus)
	case 1:
		return val, nil
	default:
		return NotSetIntValue, errors.New(errBicycleStatusIsAmbiguous)
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
	return NotSetStringValue
}

// sqlTripsSubQuery returns sql query string with all trips and
// associated data with filters for all relevant fields
func sqlTripsSubQuery(db *sql.DB, c *cli.Context) (sqlString string, err error) {
	sqlString = "SELECT" +
		" t.id as id" +
		",b.name as bicycle" +
		",bt.name as type" +
		",t.date as date" +
		",t.title as title" +
		",tc.name as category" +
		",t.distance as distance" +
		",t.duration as duration" +
		" FROM trips t LEFT JOIN bicycles b ON t.bicycle_id=b.id LEFT JOIN bicycle_types bt ON b.bicycle_type_id=bt.id LEFT JOIN trip_categories tc ON t.trip_category_id=tc.id"
	sqlString = fmt.Sprintf("%s WHERE 1=1", sqlString)

	bType := c.String("type")
	if bType != NotSetStringValue {
		bTypeID, err := bicycleTypeIDForName(db, bType)
		if err != nil {
			return NotSetStringValue, err
		}
		sqlString = fmt.Sprintf("%s AND bt.id=%d", sqlString, bTypeID)
	}

	tCategory := c.String("category")
	if tCategory != NotSetStringValue {
		tCategoryID, err := tripCategoryIDForName(db, tCategory)
		if err != nil {
			return NotSetStringValue, err
		}
		sqlString = fmt.Sprintf("%s AND tc.id=%d", sqlString, tCategoryID)
	}

	bName := c.String("bicycle")
	if bName != NotSetStringValue {
		sqlString = fmt.Sprintf("%s AND b.name LIKE '%%%s%%'", sqlString, bName)
	}

	tDate := c.String("date")
	if tDate != NotSetStringValue {
		sqlString = fmt.Sprintf("%s AND t.date LIKE '%%%s%%'", sqlString, tDate)
	}

	return sqlString, nil
}

// sqlReportSubQuery returns sql query string with all trips and
// associated data with filters for all relevant fields
func sqlBicyclesSubQuery(db *sql.DB, c *cli.Context) (sqlString string, err error) {
	sqlString = "SELECT" +
		" b.id as id" +
		",b.name as bicycle" +
		",b.producer as producer" +
		",b.model as model" +
		",t.name as type" +
		" FROM bicycles b LEFT JOIN bicycle_types t ON b.bicycle_type_id=t.id"
	sqlString = fmt.Sprintf("%s WHERE 1=1", sqlString)

	bName := c.String("bicycle")
	if bName != NotSetStringValue {
		sqlString = fmt.Sprintf("%s AND b.name LIKE '%%%s%%'", sqlString, bName)
	}

	bProducer := c.String("manufacturer")
	if bProducer != NotSetStringValue {
		sqlString = fmt.Sprintf("%s AND b.producer LIKE '%%%s%%'", sqlString, bProducer)
	}

	bModel := c.String("model")
	if bModel != NotSetStringValue {
		sqlString = fmt.Sprintf("%s AND b.model LIKE '%%%s%%'", sqlString, bModel)
	}

	bType := c.String("type")
	if bType != NotSetStringValue {
		bTypeID, err := bicycleTypeIDForName(db, bType)
		if err != nil {
			return NotSetStringValue, err
		}
		sqlString = fmt.Sprintf("%s AND t.id=%d", sqlString, bTypeID)
	}

	if c.Bool("all") == false {
		sqlString = fmt.Sprintf("%s AND b.status=%d", sqlString, bicycleStatuses["owned"])
	}

	return sqlString, nil
}
