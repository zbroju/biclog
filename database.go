// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

// Error messages
const (
	ERR_FILE_ALREADY_EXISTS    = "gBicLog: file already exists."
	ERR_FILE_CANNOT_BE_CREATED = "gBicLog: file cannot be created."
)

// DB Properties
var DB_PROPERTIES = map[string]string{
	"applicationName": "gBicLog",
	"databaseVersion": "0.1",
}

type Database struct {
	FilePath  string
	DBHandler *sql.DB
}

func NewDatabase(filePath string) *Database {
	tmpDB := new(Database)
	tmpDB.FilePath = filePath
	return tmpDB
}

func (d *Database) CreateNewFile() error {
	// Check if file exist and if so - return error
	if _, err := os.Stat(d.FilePath); !os.IsNotExist(err) {
		return errors.New(ERR_FILE_ALREADY_EXISTS)
	}

	// Open file
	var fileErr error
	d.DBHandler, fileErr = sql.Open("sqlite3", d.FilePath)
	if fileErr != nil {
		return errors.New(ERR_FILE_CANNOT_BE_CREATED)
	}
	defer d.DBHandler.Close()

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
	_, err := d.DBHandler.Exec(sqlStmt)
	if err != nil {
		os.Remove(d.FilePath)
		return errors.New(ERR_FILE_CANNOT_BE_CREATED)
	}

	// Insert properties values
	tx, err := d.DBHandler.Begin()
	if err != nil {
		os.Remove(d.FilePath)
		return errors.New(ERR_FILE_CANNOT_BE_CREATED)
	}
	stmt, err := tx.Prepare("INSERT INTO properties VALUES (?,?);")
	if err != nil {
		os.Remove(d.FilePath)
		return errors.New(ERR_FILE_CANNOT_BE_CREATED)
	}
	defer stmt.Close()
	for key, value := range DB_PROPERTIES {
		_, err := stmt.Exec(key, value)
		if err != nil {
			tx.Rollback()
			os.Remove(d.FilePath)
			return errors.New(ERR_FILE_CANNOT_BE_CREATED)
		}
	}
	tx.Commit()

	// Return nil for error
	return nil
}
