// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package src

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/zbroju/gprops"
	"log"
	"os"
	"path"
	"strings"
)

func GetConfigSettings() (dataFile string, err error) {
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

func GetLoggers() (messageLogger *log.Logger, errorLogger *log.Logger) {
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
