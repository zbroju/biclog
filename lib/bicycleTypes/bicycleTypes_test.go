// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package bicycleTypes

import (
	"testing"
)

func TestWithName(t *testing.T) {
	bicList := New()

	bicList = append(bicList, BicycleType{1, "road bike"})
	bicList = append(bicList, BicycleType{2, "city bike"})

	if _, err := bicList.GetWithName("road bike"); err != nil {
		t.Errorf("Exact match does not work.")
	}
	if _, err := bicList.GetWithName("road"); err != nil {
		t.Errorf("Part of the name does not match.")
	}
	if _, err := bicList.GetWithName("bike"); err == nil {
		t.Errorf("Common part does not return error as it should.")
	}
}

func TestGetFormattingStringsFields(t *testing.T) {
	bicList := New()

	bicList = append(bicList, BicycleType{1, "road bike"})
	bicList = append(bicList, BicycleType{2, "city"})

	_, _, fsId, fsName := bicList.GetFormattingStrings()
	if fsId != "%2d" {
		t.Errorf("formatting strings do not match. Is %s and expected %s", fsId, "%2d")
	}
	if fsName != "%9s" {
		t.Errorf("formatting strings do not match. Is %s and expected %s", fsName, "%9s")
	}

}
