// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package sqlitedb

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zbroju/gbiclog/lib/bicycleTypes"
	"os"
)

// Error messages
const (
	errFileAlreadyExists   = "gBicLog: file already exists.\n"
	errFileCannotBeCreated = "gBicLog: file cannot be created.\n"
	errFileCannotBeOpen    = "gBicLog: file cannot be open.\n"
	errFileNotAppDB        = "gBicLog: given file is not an appropriate gBicLog file.\n"
	errWritingToFile       = "gBicLog: error writing to file.\n"
	errReadingFromFile     = "gBicLog: error reading from file.\n"
)

// DB Properties
var dbProperties = map[string]string{
	"applicationName": "gBicLog",
	"databaseVersion": "0.1",
}

type Database struct {
	filePath  string
	dbHandler *sql.DB
}

func (d *Database) isTheFileBicLogDB() bool {
	rows, err := d.dbHandler.Query("SELECT KEY, VALUE FROM PROPERTIES;")
	if err != nil {
		return false
	}
	defer rows.Close()
	if rows.Next() == false {
		return false
	} else {
		for rows.Next() {
			var key, value string
			err = rows.Scan(&key, &value)
			if err != nil {
				return false
			}
			if dbProperties[key] != "" && dbProperties[key] != value {
				return false
			}
		}
	}

	return true
}

func New(filePath string) *Database {
	tmpDB := new(Database)
	tmpDB.filePath = filePath
	return tmpDB
}

func (d *Database) CreateNew() error {
	// Check if file exist and if so - return error
	if _, err := os.Stat(d.filePath); !os.IsNotExist(err) {
		return errors.New(errFileAlreadyExists)
	}

	// Open file
	var fileErr error
	d.dbHandler, fileErr = sql.Open("sqlite3", d.filePath)
	if fileErr != nil {
		return errors.New(errFileCannotBeCreated)
	}
	defer d.dbHandler.Close()

	// Create tables
	sqlStmt := `
	BEGIN TRANSACTION;
	CREATE TABLE properties (
		key TEXT
		, value TEXT);
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
	COMMIT;
	`
	_, err := d.dbHandler.Exec(sqlStmt)
	if err != nil {
		os.Remove(d.filePath)
		return errors.New(errFileCannotBeCreated)
	}

	// Insert properties values
	tx, err := d.dbHandler.Begin()
	if err != nil {
		os.Remove(d.filePath)
		return errors.New(errFileCannotBeCreated)
	}
	stmt, err := tx.Prepare("INSERT INTO properties VALUES (?,?);")
	if err != nil {
		os.Remove(d.filePath)
		return errors.New(errFileCannotBeCreated)
	}
	defer stmt.Close()
	for key, value := range dbProperties {
		_, err := stmt.Exec(key, value)
		if err != nil {
			tx.Rollback()
			os.Remove(d.filePath)
			return errors.New(errFileCannotBeCreated)
		}
	}
	tx.Commit()

	// Return nil for error
	return nil
}

func (d *Database) Open() error {
	var fileErr error
	d.dbHandler, fileErr = sql.Open("sqlite3", d.filePath)
	if fileErr != nil {
		return errors.New(errFileCannotBeOpen)
	}
	if d.isTheFileBicLogDB() == false {
		return errors.New(errFileNotAppDB)
	} else {
		return nil
	}
}

func (d *Database) Close() {
	d.dbHandler.Close()
}

func (d *Database) TypeAdd(bt bicycleTypes.BicycleType) error {
	sqlStmt := fmt.Sprintf("INSERT INTO bicycle_types VALUES (NULL, '%s');", bt.Name)
	_, err := d.dbHandler.Exec(sqlStmt)
	if err != nil {
		return errors.New(errWritingToFile)
	} else {
		return nil
	}
}

func (d *Database) TypeList() (bicycleTypes.BicycleTypes, error) {
	rows, err := d.dbHandler.Query("SELECT id, name FROM bicycle_types ORDER BY name;")
	if err != nil {
		return nil, errors.New(errReadingFromFile)
	}
	defer rows.Close()

	tmpList := bicycleTypes.New()
	for rows.Next() {
		var tmpType bicycleTypes.BicycleType
		rows.Scan(&tmpType.Id, &tmpType.Name)
		tmpList = append(tmpList, tmpType)
	}

	return tmpList, nil

}

func (d *Database) TypeUpdate(bt bicycleTypes.BicycleType) error {
	sqlStmt := fmt.Sprintf("UPDATE bicycle_types SET name='%s' WHERE id=%d;", bt.Name, bt.Id)
	_, err := d.dbHandler.Exec(sqlStmt)
	if err != nil {
		return errors.New(errWritingToFile)
	}

	return nil
}

func (d *Database) TypeDelete(bt bicycleTypes.BicycleType) error {
	sqlStmt := fmt.Sprintf("DELETE FROM bicycle_types WHERE id=%d;", bt.Id)
	_, err := d.dbHandler.Exec(sqlStmt)
	if err != nil {
		return errors.New(errWritingToFile)
	}

	return nil
}
