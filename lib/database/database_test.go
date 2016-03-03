// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package database

import (
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strings"
	"testing"
)

const (
	testDBFile = "testdb.sqlite"
)

func TestCreateNewFile(t *testing.T) {
	testdb := New(testDBFile)
	err := testdb.CreateNew()
	if err != nil {
		t.Errorf("%q", err)
	}
	defer os.Remove(testDBFile)

	// Test if a file was created
	if _, err := os.Stat(testDBFile); os.IsNotExist(err) {
		t.Errorf("Test file not created at all.")
	}

	// Open file
	err = testdb.Open()
	if err != nil {
		t.Errorf("%q", err)
	}
	defer testdb.Close()
}

func TestTypeAdd(t *testing.T) {
	// Setup
	testdb := New(testDBFile)
	err := testdb.CreateNew()
	if err != nil {
		t.Errorf("%q", err)
	}
	defer os.Remove(testDBFile)
	err = testdb.Open()
	if err != nil {
		t.Errorf("%q", err)
	}
	defer testdb.Close()

	// Test TypeAdd
	testedType := "road bike"
	err = testdb.TypeAdd(testedType)
	if err != nil {
		t.Errorf("%q", err)
	}
	rows, err := testdb.dbHandler.Query("SELECT * FROM bicycle_types;")
	if err != nil {
		t.Errorf("%q", err)
	}
	defer rows.Close()
	for rows.Next() {
		var insertedType string
		rows.Scan(nil, &insertedType)
		if strings.Compare(insertedType, testedType) == 0 {
			t.Errorf("Inserted type and read type do not match")
		}
	}
}
