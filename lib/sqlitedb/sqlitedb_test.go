// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package sqlitedb

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/zbroju/gbiclog/lib/bicycleTypes"
	"os"
	"testing"
)

const (
	testDBFile = "testdb.sqlite"
)

func TestCreateNewFile(t *testing.T) {
	testdb := New(testDBFile)
	err := testdb.CreateNew()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer os.Remove(testDBFile)

	// Test if a file was created
	if _, err := os.Stat(testDBFile); os.IsNotExist(err) {
		t.Errorf("Test file not created at all.")
	}

	// Open file
	err = testdb.Open()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer testdb.Close()
}

func TestTypeAdd(t *testing.T) {
	// Setup
	testdb := New(testDBFile)
	err := testdb.CreateNew()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer os.Remove(testDBFile)
	err = testdb.Open()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer testdb.Close()

	// Test TypeAdd
	testedType := bicycleTypes.BicycleType{0, "road bike"}
	err = testdb.TypeAdd(testedType)
	if err != nil {
		t.Errorf("%s", err)
	}
	rows, err := testdb.dbHandler.Query("SELECT * FROM bicycle_types;")
	if err != nil {
		t.Errorf("%s", err)
	}
	defer rows.Close()
	for rows.Next() {
		var insertedType string
		rows.Scan(nil, &insertedType)
		if insertedType == testedType.Name {
			t.Errorf("Inserted type and read type do not match")
		}
	}
}

func TestTypeList(t *testing.T) {
	// Setup
	testdb := New(testDBFile)
	err := testdb.CreateNew()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer os.Remove(testDBFile)
	err = testdb.Open()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer testdb.Close()

	// Test TypeList
	roadBike := bicycleTypes.BicycleType{0, "Road Bike"}
	cityBike := bicycleTypes.BicycleType{0, "City Bike"}
	err = testdb.TypeAdd(roadBike)
	if err != nil {
		t.Errorf("%s", err)
	}
	err = testdb.TypeAdd(cityBike)
	if err != nil {
		t.Errorf("%s", err)
	}
	tmpList, err := testdb.TypeList()
	if err != nil {
		t.Errorf("%s", err)
	}
	if _, err := tmpList.GetWithName(roadBike.Name); err != nil {
		t.Errorf("%s", err)
	}
	if _, err := tmpList.GetWithName(cityBike.Name); err != nil {
		t.Errorf("%s", err)
	}
}
