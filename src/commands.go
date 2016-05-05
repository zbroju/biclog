// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package src

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/zbroju/gsqlitehandler"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

func CmdInit(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check the obligatory parameters and exit if missing
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}

	// Create new file
	sqlCreateTables := `
CREATE TABLE bicycles (
 id INTEGER PRIMARY KEY
 , name TEXT
 , producer TEXT
 , model TEXT
 , bicycle_type_id INTEGER
 , production_year INTEGER
 , buying_date TEXT
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
 , date TEXT
 , title TEXT
 , trip_category_id INTEGER
 , distance REAL
 , duration TEXT
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
`
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.CreateNew(sqlCreateTables)
	if err != nil {
		printError.Fatalln(err)
	}

	// Show summary
	printUserMsg.Printf("created file %s.\n", c.String("file"))
}

func CmdTypeAdd(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check obligatory flags (file, name)
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)

	}
	if c.String("type") == NotSetStringValue {
		printError.Fatalln(errMissingTypeFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Add new type
	sqlAddType := fmt.Sprintf("INSERT INTO bicycle_types VALUES (NULL, '%s');", c.String("type"))
	_, err = f.Handler.Exec(sqlAddType)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}

	// Show summary
	printUserMsg.Printf("added new bicycle type: %s\n", c.String("type"))
}

func CmdTypeList(c *cli.Context) {
	// Get loggers
	_, printError := GetLoggers()

	// Check obligatory flags (file)
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Create formatting strings
	var maxLId, maxLName int
	err = f.Handler.QueryRow("SELECT max(length(id)), max(length(name)) FROM bicycle_types;").Scan(&maxLId, &maxLName)
	if err != nil {
		printError.Fatalln("no bicycle types")
	}
	if hlId := utf8.RuneCountInString(btIdHeader); maxLId < hlId {
		maxLId = hlId
	}
	if hlName := utf8.RuneCountInString(btNameHeader); maxLName < hlName {
		maxLName = hlName
	}
	fsId := fmt.Sprintf("%%%dv", maxLId)
	fsName := fmt.Sprintf("%%-%dv", maxLName)

	// List bicycle types
	rows, err := f.Handler.Query("SELECT id, name FROM bicycle_types ORDER BY name;")
	if err != nil {
		printError.Fatalln(errReadingFromFile)
	}
	defer rows.Close()

	line := strings.Join([]string{fsId, fsName}, FSSeparator) + "\n"
	fmt.Fprintf(os.Stdout, line, btIdHeader, btNameHeader)
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		fmt.Fprintf(os.Stdout, line, id, name)
	}
}

func CmdTypeEdit(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check obligatory flags
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == NotSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}
	newName := c.String("type")
	if newName == NotSetStringValue {
		printError.Fatalln(errMissingTypeFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Edit bicycle type
	sqlUpdateType := fmt.Sprintf("UPDATE bicycle_types SET name='%s' WHERE id=%d;", newName, id)
	r, err := f.Handler.Exec(sqlUpdateType)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoBicycleWithID)
	}

	// Show summary
	printUserMsg.Printf("change bicycle type name to '%s'\n", newName)
}

func CmdTypeDelete(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check obligatory flags
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == NotSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	//Check if it is possible to safely remove the bicycle type
	if typePossibleToDelete(f.Handler, id) == false {
		printError.Fatalln(errCannotRemoveBicycleType)
	}

	// Delete bicycle type
	sqlDeleteType := fmt.Sprintf("DELETE FROM bicycle_types WHERE id=%d;", id)
	r, err := f.Handler.Exec(sqlDeleteType)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoBicycleTypeWithID)
	}

	// Show summary
	printUserMsg.Printf("deleted bicycle type with id = %d\n", id)
}

func CmdCategoryAdd(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check obligatory flags (file, name)
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	if c.String("category") == NotSetStringValue {
		printError.Fatalln(errMissingCategoryFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Add new category
	sqlAddCategory := fmt.Sprintf("INSERT INTO trip_categories VALUES (NULL, '%s');", c.String("category"))
	_, err = f.Handler.Exec(sqlAddCategory)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}

	// Show summary
	printUserMsg.Printf("added new trip category: %s\n", c.String("category"))
}

func CmdCategoryList(c *cli.Context) {
	// Get loggers
	_, printError := GetLoggers()

	// Check obligatory flags (file)
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Create formatting strings
	var maxLId, maxLName int
	err = f.Handler.QueryRow("SELECT max(length(id)), max(length(name)) FROM trip_categories;").Scan(&maxLId, &maxLName)
	if err != nil {
		printError.Fatalln("no trip categories")
	}
	if hlId := utf8.RuneCountInString(tcIdHeader); maxLId < hlId {
		maxLId = hlId
	}
	if hlName := utf8.RuneCountInString(tcNameHeader); maxLName < hlName {
		maxLName = hlName
	}
	fsId := fmt.Sprintf("%%%dv", maxLId)
	fsName := fmt.Sprintf("%%-%dv", maxLName)

	// List trip categories
	rows, err := f.Handler.Query("SELECT id, name FROM trip_categories ORDER BY name;")
	if err != nil {
		printError.Fatalln(errReadingFromFile)
	}
	defer rows.Close()

	line := strings.Join([]string{fsId, fsName}, FSSeparator) + "\n"
	fmt.Fprintf(os.Stdout, line, tcIdHeader, tcNameHeader)
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		fmt.Fprintf(os.Stdout, line, id, name)
	}
}

func CmdCategoryEdit(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check obligatory flags
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == NotSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}
	newName := c.String("category")
	if newName == NotSetStringValue {
		printError.Fatalln(errMissingCategoryFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Edit trip category
	sqlUpdateCategory := fmt.Sprintf("UPDATE trip_categories SET name='%s' WHERE id=%d;", newName, id)
	r, err := f.Handler.Exec(sqlUpdateCategory)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoCategoryWithID)
	}

	// Show summary
	printUserMsg.Printf("change trip category name to '%s'\n", newName)
}

func CmdCategoryDelete(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check obligatory flags
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == NotSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Check if it is possible to safely remove the category
	if categoryPossibleToDelete(f.Handler, id) == false {
		printError.Fatalln(errCannotRemoveCategory)
	}

	// Delete trip category
	sqlDeleteCategory := fmt.Sprintf("DELETE FROM trip_categories WHERE id=%d;", id)
	r, err := f.Handler.Exec(sqlDeleteCategory)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoCategoryWithID)
	}

	// Show summary
	printUserMsg.Printf("deleted trip category with id = %d\n", id)
}

func CmdBicycleAdd(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check obligatory flags (file, bicycle, bicycle type)
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	bName := c.String("bicycle")
	if bName == NotSetStringValue {
		printError.Fatalln(errMissingBicycleFlag)
	}
	bType := c.String("type")
	if bType == NotSetStringValue {
		printError.Fatalln(errMissingTypeFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Add new bicycle
	bTypeId, err := bicycleTypeIDForName(f.Handler, bType)
	if err != nil {
		printError.Fatalln(err)
	}
	sqlAddBicycle := fmt.Sprintf("BEGIN TRANSACTION;INSERT INTO bicycles (id, name, bicycle_type_id) VALUES (NULL, '%s', %d);", bName, bTypeId)
	bManufacturer := c.String("manufacturer")
	if bManufacturer != NotSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET producer='%s' WHERE id=last_insert_rowid();", bManufacturer)
	}
	bModel := c.String("model")
	if bModel != NotSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET model='%s' WHERE id=last_insert_rowid();", bModel)
	}
	bYear := c.Int("year")
	if bYear != NotSetIntValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET production_year=%d WHERE id=last_insert_rowid();", bYear)
	}
	bBought := c.String("bought")
	if bBought != NotSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET buying_date='%s' WHERE id=last_insert_rowid();", bBought)
	}
	bDesc := c.String("description")
	if bDesc != NotSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET description='%s' WHERE id=last_insert_rowid();", bDesc)
	}
	bStatus := c.String("status")
	if bStatus == NotSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET status=%d WHERE id=last_insert_rowid();", bicycleStatuses["owned"])
	} else {
		bStatusID, err := bicycleStatusNoForName(bStatus)
		if err != nil {
			printError.Fatalln(err)
		}
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET status=%d WHERE id=last_insert_rowid();", bStatusID)
	}
	bSize := c.String("size")
	if bSize != NotSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET size='%s' WHERE id=last_insert_rowid();", bSize)
	}
	bWeight := c.Float64("weight")
	if bWeight != NotSetFloatValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET weight=%f WHERE id=last_insert_rowid();", bWeight)
	}
	bIDist := c.Float64("init_distance")
	if bIDist != NotSetFloatValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET initial_distance=%f WHERE id=last_insert_rowid();", bIDist)
	}
	bSeries := c.String("series")
	if bSeries != NotSetStringValue {
		sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("UPDATE bicycles SET series_no='%s' WHERE id=last_insert_rowid();", bSeries)
	}
	sqlAddBicycle = sqlAddBicycle + fmt.Sprintf("COMMIT;")
	_, err = f.Handler.Exec(sqlAddBicycle)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}

	// Show summary
	printUserMsg.Printf("added new bicycle: %s\n", bName)
}

func CmdBicycleList(c *cli.Context) {
	// Get loggers
	_, printError := GetLoggers()

	// Check obligatory flags (file)
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// SQL query
	sqlSubQuery, err := sqlBicyclesSubQuery(f.Handler, c)
	if err != nil {
		printError.Fatalln(err)
	}
	sqlQueryData := fmt.Sprintf("SELECT id, bicycle, producer, model, type FROM (%s)", sqlSubQuery)

	// Create formatting strings
	var lId, lName, lProducer, lModel, lType int
	maxQuery := fmt.Sprintf("SELECT max(length(id)), max(length(bicycle)), ifnull(max(length(producer)),0), ifnull(max(length(model)),0), ifnull(max(length(type)),0) FROM (%s);", sqlQueryData)
	err = f.Handler.QueryRow(maxQuery).Scan(&lId, &lName, &lProducer, &lModel, &lType)
	if err != nil {
		printError.Fatalln("no bicycles")
	}
	if hl := utf8.RuneCountInString(bcIdHeader); lId < hl {
		lId = hl
	}
	if hl := utf8.RuneCountInString(bcNameHeader); lName < hl {
		lName = hl
	}
	if hl := utf8.RuneCountInString(bcProducerHeader); lProducer < hl {
		lProducer = hl
	}
	if hl := utf8.RuneCountInString(bcModelHeader); lModel < hl {
		lModel = hl
	}
	if hl := utf8.RuneCountInString(btNameHeader); lType < hl {
		lType = hl
	}
	fsId := fmt.Sprintf("%%%dv", lId)
	fsName := fmt.Sprintf("%%-%dv", lName)
	fsProducer := fmt.Sprintf("%%-%dv", lProducer)
	fsModel := fmt.Sprintf("%%-%dv", lModel)
	fsType := fmt.Sprintf("%%-%dv", lType)

	// List bicycles
	rows, err := f.Handler.Query(fmt.Sprintf("%s;", sqlQueryData))
	fmt.Println(fmt.Sprintf("%s;", sqlQueryData))
	if err != nil {
		printError.Fatalln(errReadingFromFile)
	}
	defer rows.Close()
	line := strings.Join([]string{fsId, fsName, fsProducer, fsModel, fsType}, FSSeparator) + "\n"
	fmt.Fprintf(os.Stdout, line, bcIdHeader, bcNameHeader, bcProducerHeader, bcModelHeader, btNameHeader)

	for rows.Next() {
		var id int
		var name, producer, model, bicType string
		rows.Scan(&id, &name, &producer, &model, &bicType)
		fmt.Fprintf(os.Stdout, line, id, name, producer, model, bicType)
	}
}

func CmdBicycleEdit(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check obligatory flags
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == NotSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Edit bicycle
	sqlUpdateBicycle := fmt.Sprintf("BEGIN TRANSACTION;")
	bType := c.String("type")
	if bType != NotSetStringValue {
		bTypeId, err := bicycleTypeIDForName(f.Handler, bType)
		if err != nil {
			printError.Fatalln(err)
		}
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET bicycle_type_id=%d WHERE id=%d;", bTypeId, id)
	}
	bStatus := c.String("status")
	if bStatus != NotSetStringValue {
		bStatusId, err := bicycleStatusNoForName(bStatus)
		if err != nil {
			printError.Fatalln(err)
		}
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET status=%d WHERE id=%d;", bStatusId, id)
	}
	bName := c.String("bicycle")
	if bName != NotSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET name='%s' WHERE id=%d;", bName, id)
	}
	bManufacturer := c.String("manufacturer")
	if bManufacturer != NotSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET producer='%s' WHERE id=%d;", bManufacturer, id)
	}
	bModel := c.String("model")
	if bModel != NotSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET model='%s' WHERE id=%d;", bModel, id)
	}
	bYear := c.Int("year")
	if bYear != NotSetIntValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET production_year=%d WHERE id=%d;", bYear, id)
	}
	bBought := c.String("bought")
	if bBought != NotSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET buying_date='%s' WHERE id=%d;", bBought, id)
	}
	bDesc := c.String("description")
	if bDesc != NotSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET description='%s' WHERE id=%d;", bDesc, id)
	}
	bSize := c.String("size")
	if bSize != NotSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET size='%s' WHERE id=%d;", bSize, id)
	}
	bWeight := c.Float64("weight")
	if bWeight != NotSetFloatValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET weight=%f WHERE id=%d;", bWeight, id)
	}
	bIDist := c.Float64("init_distance")
	if bIDist != NotSetFloatValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET initial_distance=%f WHERE id=%d;", bIDist, id)
	}
	bSeries := c.String("series")
	if bSeries != NotSetStringValue {
		sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("UPDATE bicycles SET series_no='%s' WHERE id=%d;", bSeries, id)
	}
	sqlUpdateBicycle = sqlUpdateBicycle + fmt.Sprintf("COMMIT;")
	r, err := f.Handler.Exec(sqlUpdateBicycle)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoBicycleWithID)
	}

	// Show summary
	printUserMsg.Printf("changed bicycle details\n")
}

func CmdBicycleDelete(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check obligatory flags
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == NotSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Check if it is possible to safely delete the bicycle
	if bicyclePossibleToDelete(f.Handler, id) == false {
		printError.Fatalln(errCannotRemoveBicycle)
	}

	// Delete bicycle type
	sqlDeleteBicycle := fmt.Sprintf("DELETE FROM bicycles WHERE id=%d;", id)
	r, err := f.Handler.Exec(sqlDeleteBicycle)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoBicycleWithID)
	}

	// Show summary
	printUserMsg.Printf("deleted bicycle with id = %d\n", id)
}

func CmdBicycleShow(c *cli.Context) {
	// Get loggers
	_, printError := GetLoggers()

	// Check obligatory flags (file)
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	bcID := c.Int("id")
	bcBicycle := c.String("bicycle")
	if bcID == NotSetIntValue && bcBicycle == NotSetStringValue {
		printError.Fatalln(errMissingBicycleOrIdFlag)
	}
	if bcID != NotSetIntValue && bcBicycle != NotSetStringValue {
		printError.Fatalln(errBothIdAndBicycleFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Create formatting strings
	lineStr := fmt.Sprintf("%%-%ds%%-s\n", bcHeadingSize)
	lineInt := fmt.Sprintf("%%-%ds%%-d\n", bcHeadingSize)
	lineFloat := fmt.Sprintf("%%-%ds%%-.2f\n", bcHeadingSize)

	// Show bicycles
	if bcID == NotSetIntValue {
		bcID, err = bicycleIDForName(f.Handler, bcBicycle)
		if err != nil {
			printError.Fatalln(err)
		}
	}
	var (
		bId, bPYear, bStatId                                           int
		bName, bProducer, bModel, bType, bBDate, bDesc, bSize, bSeries string
		bWeight, bIDist                                                float64
	)
	showQuery := fmt.Sprintf("SELECT b.id, ifnull(b.name,''), ifnull(b.producer,''), ifnull(b.model,''), ifnull(t.name,''), ifnull(b.production_year,0), ifnull(b.buying_date,0), ifnull(b.description,''), ifnull(b.status,0), ifnull(b.size,''), ifnull(b.weight,0), ifnull(b.initial_distance,0), ifnull(b.series_no,'') FROM bicycles b LEFT JOIN bicycle_types t ON b.bicycle_type_id=t.id WHERE b.id=%d;", bcID)
	err = f.Handler.QueryRow(showQuery).Scan(&bId, &bName, &bProducer, &bModel, &bType, &bPYear, &bBDate, &bDesc, &bStatId, &bSize, &bWeight, &bIDist, &bSeries)
	if err != nil {
		printError.Fatalln(errNoBicycleWithID)
	}

	fmt.Printf(lineInt, bcIdHeader, bId)     // no need for if because it's obligatory
	fmt.Printf(lineStr, bcNameHeader, bName) // no need for if because it's obligatory
	if bProducer != NotSetStringValue {
		fmt.Printf(lineStr, bcProducerHeader, bProducer)
	} else {
		fmt.Printf(lineStr, bcProducerHeader, NullDataValue)
	}
	if bModel != NotSetStringValue {
		fmt.Printf(lineStr, bcModelHeader, bModel)
	} else {
		fmt.Printf(lineStr, bcModelHeader, NullDataValue)
	}
	fmt.Printf(lineStr, btNameHeader, bType) // no need for if because it's obligatory
	if bPYear != 0 {
		fmt.Printf(lineInt, bcProductionYearHeading, bPYear)
	} else {
		fmt.Printf(lineStr, bcProductionYearHeading, NullDataValue)
	}
	if bBDate != NotSetStringValue {
		fmt.Printf(lineStr, bcBuyingDateHeading, bBDate)
	} else {
		fmt.Printf(lineStr, bcBuyingDateHeading, NullDataValue)
	}
	fmt.Printf(lineStr, bcStatusHeading, bicycleStatusNameForID(bStatId)) // no need for if because it's obligatory
	if bSize != NotSetStringValue {
		fmt.Printf(lineStr, bcSizeHeading, bSize)
	} else {
		fmt.Printf(lineStr, bcSizeHeading, NullDataValue)
	}
	if bWeight != 0 {
		fmt.Printf(lineFloat, bcWeightHeading, bWeight)
	} else {
		fmt.Printf(lineStr, bcWeightHeading, NullDataValue)
	}
	if bIDist != 0 {
		fmt.Printf(lineFloat, bcInitialDistanceHeading, bIDist)
	} else {
		fmt.Printf(lineStr, bcInitialDistanceHeading, NullDataValue)
	}
	if bSeries != NotSetStringValue {
		fmt.Printf(lineStr, bcSeriesHeading, bSeries)
	} else {
		fmt.Printf(lineStr, bcSeriesHeading, NullDataValue)
	}
	if bDesc != NotSetStringValue {
		fmt.Printf(lineStr, bcDescriptionHeading, bDesc)
	} else {
		fmt.Printf(lineStr, bcDescriptionHeading, NullDataValue)
	}
}

func CmdTripAdd(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check obligatory flags (file, title, bicycle, trip category, distance)
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	tDate := c.String("date")
	if tDate == NotSetStringValue {
		tDate = time.Now().Format("2006-01-02")
	}
	tTitle := c.String("title")
	if tTitle == NotSetStringValue {
		printError.Fatalln(errMissingTitleFlag)
	}
	tBicycle := c.String("bicycle")
	if tBicycle == NotSetStringValue {
		printError.Fatalln(errMissingBicycleFlag)
	}
	tCategory := c.String("category")
	if tCategory == NotSetStringValue {
		printError.Fatalln(errMissingCategoryFlag)
	}
	tDistance := c.Float64("distance")
	if tDistance == NotSetFloatValue {
		printError.Fatalln(errMissingDistanceFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Add new trip
	tBicycleId, err := bicycleIDForName(f.Handler, tBicycle)
	if err != nil {
		printError.Fatalln(err)
	}
	tCategoryId, err := tripCategoryIDForName(f.Handler, tCategory)
	if err != nil {
		printError.Fatalln(err)
	}

	sqlAddTrip := fmt.Sprintf("BEGIN TRANSACTION;")
	sqlAddTrip = sqlAddTrip + fmt.Sprintf("INSERT INTO trips (id, bicycle_id, date,title, trip_category_id, distance) VALUES (NULL, %d, '%s', '%s', %d, %f);", tBicycleId, tDate, tTitle, tCategoryId, tDistance)
	tDuration := c.String("duration")
	if tDuration != NotSetStringValue {
		durationValue, err := time.ParseDuration(tDuration)
		if err != nil {
			printError.Fatalln(errWrongDurationFormat)
		}
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET duration='%s' WHERE id=last_insert_rowid();", durationValue.String())
	}
	tDescription := c.String("description")
	if tDescription != NotSetStringValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET description='%s' WHERE id=last_insert_rowid();", tDescription)
	}
	tHRMax := c.Int("hrmax")
	if tHRMax != NotSetIntValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET hr_max=%d WHERE id=last_insert_rowid();", tHRMax)
	}
	tHRAvg := c.Int("hravg")
	if tHRAvg != NotSetIntValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET hr_avg=%d WHERE id=last_insert_rowid();", tHRAvg)
	}
	tSpeedMax := c.Float64("speed_max")
	if tSpeedMax != NotSetFloatValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET speed_max=%f WHERE id=last_insert_rowid();", tSpeedMax)
	}
	tDriveways := c.Float64("driveways")
	if tDriveways != NotSetFloatValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET driveways=%f WHERE id=last_insert_rowid();", tDriveways)
	}
	tCalories := c.Int("calories")
	if tCalories != NotSetIntValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET calories=%d WHERE id=last_insert_rowid();", tCalories)
	}
	tTemperature := c.Float64("temperature")
	if tTemperature != NotSetFloatValue {
		sqlAddTrip = sqlAddTrip + fmt.Sprintf("UPDATE trips SET temperature=%f WHERE id=last_insert_rowid();", tTemperature)
	}
	sqlAddTrip = sqlAddTrip + fmt.Sprintf("COMMIT;")

	_, err = f.Handler.Exec(sqlAddTrip)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}

	// Show summary
	printUserMsg.Printf("added new trip: '%s'\n", tTitle)
}

func CmdTripList(c *cli.Context) {
	// Get loggers
	_, printError := GetLoggers()

	// Check obligatory flags (file)
	if c.String("file") == "" {
		printError.Fatalln(errMissingFileFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// SQL query
	sqlSubQuery, err := sqlTripsSubQuery(f.Handler, c)
	if err != nil {
		printError.Fatalln(err)
	}
	sqlQueryData := fmt.Sprintf("SELECT id, date, title, category, bicycle, distance FROM (%s) ORDER BY date", sqlSubQuery)

	// Create formatting strings
	var lId, lDate, lTitle, lCategory, lBicycle, lDistance int
	maxQuery := fmt.Sprintf("SELECT max(length(id)), ifnull(max(length(date)),0), ifnull(max(length(title)),0), ifnull(max(length(category)),0), ifnull(max(length(bicycle)),0), ifnull(max(length(distance)),0) FROM (%s);", sqlQueryData)
	err = f.Handler.QueryRow(maxQuery).Scan(&lId, &lDate, &lTitle, &lCategory, &lBicycle, &lDistance)
	if err != nil {
		printError.Fatalln("no trips")
	}
	if hl := utf8.RuneCountInString(bcIdHeader); lId < hl {
		lId = hl
	}
	if hl := utf8.RuneCountInString(trpDateHeader); lDate < hl {
		lDate = hl
	}
	if hl := utf8.RuneCountInString(trpTitleHeader); lTitle < hl {
		lTitle = hl
	}
	if hl := utf8.RuneCountInString(tcNameHeader); lCategory < hl {
		lCategory = hl
	}
	if hl := utf8.RuneCountInString(bcNameHeader); lBicycle < hl {
		lBicycle = hl
	}
	if hl := utf8.RuneCountInString(trpDistanceHeader); lDistance < hl {
		lDistance = hl
	}

	fsId := fmt.Sprintf("%%%dv", lId)
	fsDate := fmt.Sprintf("%%-%dv", lDate)
	fsTitle := fmt.Sprintf("%%-%dv", lTitle)
	fsCategory := fmt.Sprintf("%%-%dv", lCategory)
	fsBicycle := fmt.Sprintf("%%-%dv", lBicycle)
	fsDistance := fmt.Sprintf("%%%dv", lDistance)

	// List bicycles
	rows, err := f.Handler.Query(fmt.Sprintf("%s;", sqlQueryData))
	if err != nil {
		printError.Fatalln(errReadingFromFile)
	}
	defer rows.Close()
	line := strings.Join([]string{fsId, fsDate, fsCategory, fsBicycle, fsDistance, fsTitle}, FSSeparator) + "\n"
	fmt.Fprintf(os.Stdout, line, trpIdHeader, trpDateHeader, tcNameHeader, bcNameHeader, trpDistanceHeader, trpTitleHeader)

	for rows.Next() {
		var id int
		var date, title, category, bicycle string
		var distance float64
		rows.Scan(&id, &date, &title, &category, &bicycle, &distance)
		fmt.Fprintf(os.Stdout, line, id, date, category, bicycle, distance, title)
	}
}

func CmdTripEdit(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check obligatory flags
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == NotSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Edit trip
	sqlUpdateTrip := fmt.Sprintf("BEGIN TRANSACTION;")
	tCategory := c.String("category")
	if tCategory != NotSetStringValue {
		tCategoryId, err := tripCategoryIDForName(f.Handler, tCategory)
		if err != nil {
			printError.Fatalln(err)
		}
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET trip_category_id=%d WHERE id=%d;", tCategoryId, id)
	}
	tBicycle := c.String("bicycle")
	if tBicycle != NotSetStringValue {
		tBicycleId, err := bicycleIDForName(f.Handler, tBicycle)
		if err != nil {
			printError.Fatalln(err)
		}
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET bicycle_id=%d WHERE id=%d;", tBicycleId, id)
	}
	tDate := c.String("date")
	if tDate != NotSetStringValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET date='%s' WHERE id=%d;", tDate, id)
	}
	tTitle := c.String("title")
	if tTitle != NotSetStringValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET title='%s' WHERE id=%d;", tTitle, id)
	}
	tDistance := c.Float64("distance")
	if tDistance != NotSetFloatValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET distance=%f WHERE id=%f;", tDistance, id)
	}
	tDuration := c.String("duration")
	if tDuration != NotSetStringValue {
		durationValue, err := time.ParseDuration(tDuration)
		if err != nil {
			printError.Fatalln(errWrongDurationFormat)
		}
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET duration='%s' WHERE id=%d;", durationValue.String(), id)
	}
	tDescription := c.String("description")
	if tDescription != NotSetStringValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET description='%s' WHERE id=%d;", tDescription, id)
	}
	tHrMax := c.Int("hrmax")
	if tHrMax != NotSetIntValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET hr_max=%d WHERE id=%d;", tHrMax, id)
	}
	tHrAvg := c.Int("hravg")
	if tHrAvg != NotSetIntValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET hr_avg=%d WHERE id=%d;", tHrAvg, id)
	}
	tSpeedMax := c.Float64("speed_max")
	if tSpeedMax != NotSetFloatValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET speed_max=%f WHERE id=%d;", tSpeedMax, id)
	}
	tDriveways := c.Float64("driveways")
	if tDriveways != NotSetFloatValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET driveways=%f WHERE id=%d;", tDriveways, id)
	}
	tCalories := c.Int("calories")
	if tCalories != NotSetIntValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET calories=%d WHERE id=%d;", tCalories, id)
	}
	tTemperature := c.Float64("temperature")
	if tTemperature != NotSetFloatValue {
		sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("UPDATE trips SET calories=%f WHERE id=%d;", tTemperature, id)
	}
	sqlUpdateTrip = sqlUpdateTrip + fmt.Sprintf("COMMIT;")
	r, err := f.Handler.Exec(sqlUpdateTrip)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoBicycleWithID)
	}

	// Show summary
	printUserMsg.Printf("changed trip details\n")
}

func CmdTripDelete(c *cli.Context) {
	// Get loggers
	printUserMsg, printError := GetLoggers()

	// Check obligatory flags
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	id := c.Int("id")
	if id == NotSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Delete bicycle type
	sqlDeleteTrip := fmt.Sprintf("DELETE FROM trips WHERE id=%d;", id)
	r, err := f.Handler.Exec(sqlDeleteTrip)
	if err != nil {
		printError.Fatalln(errWritingToFile)
	}
	if i, _ := r.RowsAffected(); i == 0 {
		printError.Fatalln(errNoTripWithID)
	}

	// Show summary
	printUserMsg.Printf("deleted tirp with id = %d\n", id)
}

func CmdTripShow(c *cli.Context) {
	// Get loggers
	_, printError := GetLoggers()

	// Check obligatory flags (file)
	if c.String("file") == NotSetStringValue {
		printError.Fatalln(errMissingFileFlag)
	}
	tID := c.Int("id")
	if tID == NotSetIntValue {
		printError.Fatalln(errMissingIdFlag)
	}

	// Open data file
	f := gsqlitehandler.New(c.String("file"), dataFileProperties)
	err := f.Open()
	if err != nil {
		printError.Fatalln(err)
	}
	defer f.Close()

	// Create formatting strings
	lineStr := fmt.Sprintf("%%-%ds%%-s\n", trpHeadingSize)
	lineInt := fmt.Sprintf("%%-%ds%%-d\n", trpHeadingSize)
	lineFloat := fmt.Sprintf("%%-%ds%%-.1f\n", trpHeadingSize)

	// Show trip
	var (
		tId, tHrMax, tHrAvg, tCalories                    int
		bName, tDate, tTitle, tCategory, tDuration, tDesc string
		tDistance, tSpeedMax, tDriveways, tTemp           float64
	)
	showQuery := fmt.Sprintf("SELECT t.id, ifnull(b.name,''), ifnull(t.date,''), ifnull(t.title,''), ifnull(c.name,''), ifnull(t.distance,0), ifnull(t.duration,''), ifnull(t.description,''), ifnull(t.hr_max,0), ifnull(t.hr_avg,0), ifnull(t.speed_max,0), ifnull(t.driveways,0), ifnull(t.calories,0), ifnull(t.temperature,0) FROM trips t LEFT JOIN trip_categories c ON t.trip_category_id=c.id LEFT JOIN bicycles b ON t.bicycle_id=b.id WHERE t.id=%d;", tID)
	err = f.Handler.QueryRow(showQuery).Scan(&tId, &bName, &tDate, &tTitle, &tCategory, &tDistance, &tDuration, &tDesc, &tHrMax, &tHrAvg, &tSpeedMax, &tDriveways, &tCalories, &tTemp)
	if err != nil {
		printError.Fatalln(errNoTripWithID)
	}

	fmt.Printf(lineInt, trpIdHeader, tId)
	fmt.Printf(lineStr, bcNameHeader, bName)
	fmt.Printf(lineStr, trpDateHeader, tDate)
	fmt.Printf(lineStr, trpTitleHeader, tTitle)
	fmt.Printf(lineStr, tcNameHeader, tCategory)
	fmt.Printf(lineFloat, trpDistanceHeader, tDistance)
	if tDuration != NotSetStringValue {
		fmt.Printf(lineStr, trpDurationHeading, tDuration)
		durationValue, err := time.ParseDuration(tDuration)
		if err == nil {
			fmt.Printf(lineFloat, trpSpeedAverageHeading, tDistance/durationValue.Hours())
		}
	} else {
		fmt.Printf(lineStr, trpDurationHeading, NullDataValue)
		fmt.Printf(lineStr, trpSpeedAverageHeading, NullDataValue)
	}
	if tSpeedMax != 0 {
		fmt.Printf(lineFloat, trpSpeedMaxHeading, tSpeedMax)
	} else {
		fmt.Printf(lineStr, trpSpeedMaxHeading, NullDataValue)
	}
	if tDriveways != 0 {
		fmt.Printf(lineFloat, trpDrivewaysHeading, tDriveways)
	} else {
		fmt.Printf(lineStr, trpDrivewaysHeading, NullDataValue)
	}
	if tHrMax != 0 {
		fmt.Printf(lineInt, trpHrMaxHeading, tHrMax)
	} else {
		fmt.Printf(lineStr, trpHrMaxHeading, NullDataValue)
	}
	if tHrAvg != 0 {
		fmt.Printf(lineInt, trpHrAvgHeading, tHrAvg)
	} else {
		fmt.Printf(lineStr, trpHrAvgHeading, NullDataValue)
	}
	if tCalories != 0 {
		fmt.Printf(lineInt, trpCaloriesHeading, tCalories)
	} else {
		fmt.Printf(lineStr, trpCaloriesHeading, NullDataValue)
	}
	if tTemp != 0 {
		fmt.Printf(lineFloat, trpTemperatureHeading, tTemp)
	} else {
		fmt.Printf(lineStr, trpTemperatureHeading, NullDataValue)
	}
	if tDesc != NotSetStringValue {
		fmt.Printf(lineStr, trpDescriptionHeading, tDesc)
	} else {
		fmt.Printf(lineStr, trpDescriptionHeading, NullDataValue)
	}
}
