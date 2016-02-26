// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

// Error messages
const (
	ERR_FILE_ALREADY_EXISTS    = "gBicLog: file already exists.\n"
	ERR_FILE_CANNOT_BE_CREATED = "gBicLog: file cannot be created.\n"
	ERR_FILE_CANNOT_BE_OPEN    = "gBicLog: file cannot be open.\n"
	ERR_FILE_NOT_APP_DB        = "gBicLog: given file is not an appropriate gBicLog file.\n"
	ERR_WRITING_TO_FILE        = "gBicLog: error writing to file.\n"
)

// DB Properties
var DB_PROPERTIES = map[string]string{
	"applicationName": "gBicLog",
	"databaseVersion": "0.1",
}

type database struct {
	filePath  string
	dbHandler *sql.DB
}

func NewDatabase(filePath string) *database {
	tmpDB := new(database)
	tmpDB.filePath = filePath
	return tmpDB
}

func (d *database) isTheFileBicLogDB() bool {
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
			if DB_PROPERTIES[key] != "" && DB_PROPERTIES[key] != value {
				return false
			}
		}
	}

	return true
}

func (d *database) CreateNewFile() error {
	// Check if file exist and if so - return error
	if _, err := os.Stat(d.filePath); !os.IsNotExist(err) {
		return errors.New(ERR_FILE_ALREADY_EXISTS)
	}

	// Open file
	var fileErr error
	d.dbHandler, fileErr = sql.Open("sqlite3", d.filePath)
	if fileErr != nil {
		return errors.New(ERR_FILE_CANNOT_BE_CREATED)
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
		return errors.New(ERR_FILE_CANNOT_BE_CREATED)
	}

	// Insert properties values
	tx, err := d.dbHandler.Begin()
	if err != nil {
		os.Remove(d.filePath)
		return errors.New(ERR_FILE_CANNOT_BE_CREATED)
	}
	stmt, err := tx.Prepare("INSERT INTO properties VALUES (?,?);")
	if err != nil {
		os.Remove(d.filePath)
		return errors.New(ERR_FILE_CANNOT_BE_CREATED)
	}
	defer stmt.Close()
	for key, value := range DB_PROPERTIES {
		_, err := stmt.Exec(key, value)
		if err != nil {
			tx.Rollback()
			os.Remove(d.filePath)
			return errors.New(ERR_FILE_CANNOT_BE_CREATED)
		}
	}
	tx.Commit()

	// Return nil for error
	return nil
}

func (d *database) Open() error {
	var fileErr error
	d.dbHandler, fileErr = sql.Open("sqlite3", d.filePath)
	if fileErr != nil {
		return errors.New(ERR_FILE_CANNOT_BE_OPEN)
	}
	if d.isTheFileBicLogDB() == false {
		return errors.New(ERR_FILE_NOT_APP_DB)
	} else {
		return nil
	}
}

func (d *database) Close() {
	d.dbHandler.Close()
}

func (d *database) TypeAdd(name string) error {
	sqlStmt := fmt.Sprintf("INSERT INTO bicycle_types VALUES (NULL, '%s');", name)
	_, err := d.dbHandler.Exec(sqlStmt)
	if err != nil {
		return errors.New(ERR_WRITING_TO_FILE)
	} else {
		return nil
	}
}
