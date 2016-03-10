// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package bicycleTypes

import (
	"testing"
)

func TestGetWithId(t *testing.T) {
	bicList := New()

	bt1 := BicycleType{1, "road bike"}
	bt2 := BicycleType{2, "city bike"}
	bicList = append(bicList, bt1)
	bicList = append(bicList, bt2)

	bt, err := bicList.GetWithId(1)
	if err != nil {
		t.Errorf("%s", err)
	}
	if bt.Name != bt1.Name {
		t.Errorf("the name does not match the id")
	}
}
func TestGetWithName(t *testing.T) {
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

func TestGetDisplayStrings(t *testing.T) {
	bicList := New()

	bicList = append(bicList, BicycleType{1, "road bike"})
	bicList = append(bicList, BicycleType{2, "city"})

	_, _, fsId, fsName := bicList.GetDisplayStrings()
	if fsId != "%2d" {
		t.Errorf("formatting strings do not match. Is %s and expected %s", fsId, "%2d")
	}
	if fsName != "%9s" {
		t.Errorf("formatting strings do not match. Is %s and expected %s", fsName, "%9s")
	}

}
