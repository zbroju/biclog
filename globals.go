// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.

package main

// Application internal settings
const (
	AppName       = "biclog"
	FSSeparator   = "  "
	NullDataValue = "-"

	NotSetIntValue    int     = -1
	NotSetFloatValue  float64 = -1
	NotSetStringValue         = ""
)

// Bicycle statuses
var bicycleStatuses = map[string]int{
	"owned":    1,
	"sold":     2,
	"scrapped": 3,
	"stolen":   4,
}

// Config file settings
const (
	confDataFile = "DATA_FILE"
)

// DB Properties
var dataFileProperties = map[string]string{
	"applicationName": "gBicLog",
	"databaseVersion": "1.0",
}

// Error messages
const (
	errMissingFileFlag        = "missing information about data file. Specify it with --file or -f flag"
	errMissingTypeFlag        = "missing bicycle type. Specify it with --type or -t flag"
	errMissingCategoryFlag    = "missing trip category. Specify it with --category or -c flag"
	errMissingIdFlag          = "missing id. Specify it with --id or -i flag"
	errMissingBicycleFlag     = "missing bicycle. Specify it with --bicycle or -b flag"
	errMissingBicycleOrIdFlag = "missing bicycle or id flag. Specify it with --bicycle (-b) or --id (-i) flag"
	errMissingTitleFlag       = "missing trip title. Specify it with --title or -s flag"
	errMissingDistanceFlag    = "missing trip distance. Specify it with --distance or -d flag"
	errBothIdAndBicycleFlag   = "both bicycle and id flag specified. Specify only one of them."

	errWritingToFile              = "error writing to file"
	errReadingFromFile            = "error reading from file"
	errNoBicycleWithID            = "no bicycle with given id"
	errNoBicycleForName           = "no bicycle for given name"
	errBicycleNameIsAmbiguous     = "bicycle name is ambiguous"
	errNoBicycleTypeWithID        = "no bicycle type with given id"
	errNoCategoryWithID           = "no trip categories with given id"
	errNoCategoryForName          = "no trip category for given name"
	errNoBicycleTypesForName      = "no bicycle types for given name"
	errNoBicycleStatus            = "unknown bicycle status"
	errBicycleStatusIsAmbiguous   = "given bicycle status is ambiguous"
	errBicycleTypeNameIsAmbiguous = "given bicycle type name is ambiguous"
	errCategoryNameIsAmbiguous    = "given trip category name is ambiguous"
	errNoTripWithID               = "no trip with given id"

	errWrongDurationFormat = "wrong duration format (should be: 00h00m00s or 00m00s)"

	errCannotRemoveBicycleType = "cannot remove bicycle type because there are bicycles of this type"
	errCannotRemoveCategory    = "cannot remove category because there are trips with this category"
	errCannotRemoveBicycle     = "cannot remove bicycle because there are trips done on it"
)

// Headings titles
const (
	btIdHeader   = "ID"
	btNameHeader = "TYPE"

	tcIdHeader   = "ID"
	tcNameHeader = "CATEGORY"

	bcIdHeader               = "ID"
	bcNameHeader             = "BICYCLE"
	bcProducerHeader         = "PRODUCER"
	bcModelHeader            = "MODEL"
	bcProductionYearHeading  = "PRODUCTION YEAR"
	bcBuyingDateHeading      = "BUYING DATE"
	bcDescriptionHeading     = "DESCRIPTION"
	bcStatusHeading          = "STATUS"
	bcSizeHeading            = "SIZE"
	bcWeightHeading          = "WEIGHT"
	bcInitialDistanceHeading = "INITIAL DISTANCE"
	bcSeriesHeading          = "SERIES"
	bcHeadingSize            = 20

	trpIdHeader            = "ID"
	trpDateHeader          = "DATE"
	trpTitleHeader         = "TITLE"
	trpDistanceHeader      = "DISTANCE"
	trpDurationHeading     = "DURATION"
	trpDescriptionHeading  = "DESCRIPTION"
	trpHrMaxHeading        = "HR MAX"
	trpHrAvgHeading        = "HR AVG"
	trpSpeedMaxHeading     = "MAX SPEED"
	trpDrivewaysHeading    = "DRIVEWAYS"
	trpCaloriesHeading     = "CALORIES"
	trpTemperatureHeading  = "TEMPERATURE"
	trpSpeedAverageHeading = "AVERAGE SPEED"
	trpHeadingSize         = 15
)

// Objects
const (
	objectBicycleType       = "bicycle_type"
	objectBicycleTypeAlias  = "bt"
	objectTripCategory      = "trip_category"
	objectTripCategoryAlias = "tc"
	objectBicycle           = "bicycle"
	objectBicycleAlias      = "bc"
	objectTrip              = "trip"
	objectTripAlias         = "tr"

	objectReportSummary      = "summary"
	objectReportSummaryAlias = "s"
	objectReportYearly       = "yearly"
	objectReportYearlyAlias  = "y"
	objectReportMonthly      = "monthly"
	objectReportMonthlyAlias = "m"
)
