// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"testing"
)

const (
	TEST_DB_FILE = "testdb.sqlite"
)

func TestInit(t *testing.T) {
	testdb := NewDatabase(TEST_DB_FILE)
	err := testdb.CreateNewFile()
	if err != nil {
		t.Errorf("%q", err)
	}

	// Test if a file was created
	if _, err := os.Stat(TEST_DB_FILE); os.IsNotExist(err) {
		t.Errorf("Test file not created at all.")
	}

	// Open file
	db, err := sql.Open("sqlite3", TEST_DB_FILE)
	if err != nil {
		t.Errorf("error touching the file")
	}
	defer db.Close()

	// Test if the file is a correct data file
	rows, err := db.Query("SELECT KEY, VALUE FROM PROPERTIES;")
	if err != nil {
		t.Errorf("error preparing query")
	}
	defer rows.Close()
	if rows.Next() == false {
		t.Errorf("Properties table is empty.")
	} else {
		for rows.Next() {
			var key, value string
			err = rows.Scan(&key, &value)
			if err != nil {
				t.Errorf("Cannot read from PROPERTIES TABLE")
			}
			if DB_PROPERTIES[key] != "" && DB_PROPERTIES[key] != value {
				t.Errorf("Unexpected properties in file %q.", TEST_DB_FILE)
			}
		}
	}

	// Remove temporary files
	os.Remove(TEST_DB_FILE)
}
