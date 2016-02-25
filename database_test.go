// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package main

import (
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
	err = testdb.Open()
	if err != nil {
		t.Errorf("%q", err)
	}
	defer testdb.Close()

	// Test if the file is a correct data file
	if testdb.isTheFileBicLogDB() == false {
		t.Errorf("File is not a correct BicLog file.")
	}

	// Remove temporary files
	os.Remove(TEST_DB_FILE)
}
