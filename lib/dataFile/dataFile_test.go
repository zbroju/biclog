// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package dataFile

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/zbroju/gbiclog/lib/bicycleTypes"
	"github.com/zbroju/gbiclog/lib/tripCategories"
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

func TestTypeEdit(t *testing.T) {
	// Setup
	db := New(testDBFile)
	err := db.CreateNew()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer os.Remove(testDBFile)
	err = db.Open()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer db.Close()

	// Test edit type
	bikeType := bicycleTypes.BicycleType{0, "road bike"}
	err = db.TypeAdd(bikeType)
	if err != nil {
		t.Errorf("%s", err)
	}

	typesList, err := db.TypeList()
	if err != nil {
		t.Errorf("%s", err)
	}
	bikeType, err = typesList.GetWithName("road bike")
	if err != nil {
		t.Errorf("%s", err)
	}
	bikeType.Name = "City Bike"
	err = db.TypeUpdate(bikeType)
	if err != nil {
		t.Errorf("%s", err)
	}
	typesList, err = db.TypeList()
	if err != nil {
		t.Errorf("%s", err)
	}
	updatedType, err := typesList.GetWithName("City")
	if err != nil {
		t.Errorf("%s", err)
	}
	if bikeType.Name != updatedType.Name {
		t.Errorf("Update type name does not match the expected one.")
	}
}

func TestTypeDelete(t *testing.T) {
	// Setup
	db := New(testDBFile)
	err := db.CreateNew()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer os.Remove(testDBFile)
	err = db.Open()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer db.Close()

	// Test deleting types
	bikeType := bicycleTypes.BicycleType{0, "road bike"}
	err = db.TypeAdd(bikeType)
	if err != nil {
		t.Errorf("%s", err)
	}

	btypes, err := db.TypeList()
	if err != nil {
		t.Errorf("%s", err)
	}
	bikeType, err = btypes.GetWithName("road")
	if err != nil {
		t.Errorf("%s", err)
	}
	err = db.TypeDelete(bikeType)
	if err != nil {
		t.Errorf("%s", err)
	}
	btypes, err = db.TypeList()
	if err != nil {
		t.Errorf("%s", err)
	}
	bikeType, err = btypes.GetWithName("road")
	if err == nil {
		t.Errorf("%s", err)
	}
}

func TestCategoryAdd(t *testing.T) {
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
	tc := tripCategories.TripCategory{0, "commuting"}
	err = testdb.CategoryAdd(tc)
	if err != nil {
		t.Errorf("%s", err)
	}
	rows, err := testdb.dbHandler.Query("SELECT * FROM trip_categories;")
	if err != nil {
		t.Errorf("%s", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ic string
		rows.Scan(nil, &ic)
		if ic == tc.Name {
			t.Errorf("Inserted type and read type do not match")
		}
	}
}

func TestCategoryList(t *testing.T) {
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
	training := tripCategories.TripCategory{0, "training"}
	commuting := tripCategories.TripCategory{0, "commuting"}
	err = testdb.CategoryAdd(training)
	if err != nil {
		t.Errorf("%s", err)
	}
	err = testdb.CategoryAdd(commuting)
	if err != nil {
		t.Errorf("%s", err)
	}
	tmpList, err := testdb.CategoryList()
	if err != nil {
		t.Errorf("%s", err)
	}
	if _, err := tmpList.GetWithName(training.Name); err != nil {
		t.Errorf("%s", err)
	}
	if _, err := tmpList.GetWithName(commuting.Name); err != nil {
		t.Errorf("%s", err)
	}
}

func TestCategoryEdit(t *testing.T) {
	// Setup
	db := New(testDBFile)
	err := db.CreateNew()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer os.Remove(testDBFile)
	err = db.Open()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer db.Close()

	// Test edit type
	cat := tripCategories.TripCategory{0, "training"}
	err = db.CategoryAdd(cat)
	if err != nil {
		t.Errorf("%s", err)
	}

	catList, err := db.CategoryList()
	if err != nil {
		t.Errorf("%s", err)
	}
	cat, err = catList.GetWithName("train")
	if err != nil {
		t.Errorf("%s", err)
	}
	cat.Name = "commuting"
	err = db.CategoryUpdate(cat)
	if err != nil {
		t.Errorf("%s", err)
	}
	catList, err = db.CategoryList()
	if err != nil {
		t.Errorf("%s", err)
	}
	updatedCat, err := catList.GetWithName("commut")
	if err != nil {
		t.Errorf("%s", err)
	}
	if cat.Name != updatedCat.Name {
		t.Errorf("Update type name does not match the expected one.")
	}
}

func TestCategoryDelete(t *testing.T) {
	// Setup
	db := New(testDBFile)
	err := db.CreateNew()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer os.Remove(testDBFile)
	err = db.Open()
	if err != nil {
		t.Errorf("%s", err)
	}
	defer db.Close()

	// Test deleting types
	cat := tripCategories.TripCategory{0, "commuting"}
	err = db.CategoryAdd(cat)
	if err != nil {
		t.Errorf("%s", err)
	}

	cats, err := db.CategoryList()
	if err != nil {
		t.Errorf("%s", err)
	}
	cat, err = cats.GetWithName("commut")
	if err != nil {
		t.Errorf("%s", err)
	}
	err = db.CategoryDelete(cat)
	if err != nil {
		t.Errorf("%s", err)
	}
	cats, err = db.CategoryList()
	if err != nil {
		t.Errorf("%s", err)
	}
	cat, err = cats.GetWithName("commut")
	if err == nil {
		t.Errorf("%s", err)
	}
}
