// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.

package src

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
