// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package tripCategories

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Error messages
const (
	errCategoryNotFound      = "trip category not found"
	errCategoryAmbiguousName = "trip category name is ambiguous"
)

// Headers titles
const (
	idHeaderTitle   = "ID"
	nameHeaderTitle = "T.CATEGORY"
)

// Basic types
type TripCategory struct {
	Id   int
	Name string
}

type TripCategories []TripCategory

// Types constant/variables
var (
	nullType TripCategory = TripCategory{0, ""}
)

func New() TripCategories {
	return make(TripCategories, 0)
}

func (tc *TripCategories) GetWithId(id int) (TripCategory, error) {
	for _, t := range *tc {
		if t.Id == id {
			return t, nil
		}
	}

	return nullType, errors.New(errCategoryNotFound)
}

func (tc *TripCategories) GetWithName(name string) (TripCategory, error) {
	var counter int
	var foundCategory TripCategory

	for _, tmpCategory := range *tc {
		if strings.Contains(tmpCategory.Name, name) == true {
			counter++
			foundCategory = tmpCategory
		}
	}

	switch counter {
	case 0:
		return nullType, errors.New(errCategoryNotFound)
	case 1:
		return foundCategory, nil
	default:
		return nullType, errors.New(errCategoryAmbiguousName)
	}
}

func (tc *TripCategories) GetDisplayStrings() (idHeader, nameHeader, idFS, nameFS string) {
	// Find longest strings
	maxLenId := utf8.RuneCountInString(idHeaderTitle)
	maxLenName := utf8.RuneCountInString(nameHeaderTitle)
	for _, t := range *tc {
		if lId := utf8.RuneCountInString(string(t.Id)); lId > maxLenId {
			maxLenId = lId
		}
		if lName := utf8.RuneCountInString(t.Name); lName > maxLenName {
			maxLenName = lName
		}
	}

	// Build formatting strings
	idHeader = fmt.Sprintf(fmt.Sprintf("%%%ds", maxLenId), idHeaderTitle)
	nameHeader = fmt.Sprintf(fmt.Sprintf("%%%ds", maxLenName), nameHeaderTitle)
	idFS = fmt.Sprintf("%%%dd", maxLenId)
	nameFS = fmt.Sprintf("%%%ds", maxLenName)

	return idHeader, nameHeader, idFS, nameFS
}
