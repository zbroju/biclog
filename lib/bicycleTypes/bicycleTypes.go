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
	errTypeNotFound = "gBicLog: bicycle type not found or the name is ambiguous"
)

// Headers titles
const (
	idHeaderTitle   = "ID"
	nameHeaderTitle = "B.TYPE"
)

type BicycleType struct {
	Id   int
	Name string
}

type BicycleTypes []BicycleType

func New() BicycleTypes {
	return make(BicycleTypes, 0)
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

	if counter == 1 {
		return foundType, nil
	} else {
		return BicycleType{0, ""}, errors.New(errTypeNotFound)
	}
}

func (bt *BicycleTypes) GetFormattingStrings() (idHeader, nameHeader, idFS, nameFS string) {
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
