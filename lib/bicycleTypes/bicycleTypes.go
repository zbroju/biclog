// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package bicycleTypes

import (
	"errors"
	"fmt"
	"strings"
)

// Error messages
const (
	errTypeNotFound      = "gBicLog: bicycle type not found\n"
	errTypeAmbiguousName = "gBicLog: bicycle type name is ambiguous\n"
)

// Headers titles
const (
	idHeaderTitle   = "ID"
	nameHeaderTitle = "B.TYPE"
)

// Basic types
type BicycleType struct {
	Id   int
	Name string
}

type BicycleTypes []BicycleType

// Types constant/variables
var (
	nullType BicycleType = BicycleType{0, ""}
)

func New() BicycleTypes {
	return make(BicycleTypes, 0)
}

func (bt *BicycleTypes) GetWithId(id int) (BicycleType, error) {
	for _, t := range *bt {
		if t.Id == id {
			return t, nil
		}
	}

	return nullType, errors.New(errTypeNotFound)
}

func (bt *BicycleTypes) GetWithName(name string) (BicycleType, error) {
	var counter int
	var foundType BicycleType

	for _, tmpType := range *bt {
		if strings.Contains(tmpType.Name, name) == true {
			counter++
			foundType = tmpType
		}
	}

	switch counter {
	case 0:
		return nullType, errors.New(errTypeNotFound)
	case 1:
		return foundType, nil
	default:
		return nullType, errors.New(errTypeAmbiguousName)
	}
}

func (bt *BicycleTypes) GetDisplayStrings() (idHeader, nameHeader, idFS, nameFS string) {
	// Find longest strings
	maxLenId := len(idHeaderTitle)
	maxLenName := len(nameHeaderTitle)
	for _, t := range *bt {
		if lId := len(string(t.Id)); lId > maxLenId {
			maxLenId = lId
		}
		if lName := len(t.Name); lName > maxLenName {
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
