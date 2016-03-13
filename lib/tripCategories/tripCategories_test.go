// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package tripCategories

import (
	"testing"
)

func TestGetWithId(t *testing.T) {
	tripCatsList := New()

	bt1 := TripCategory{1, "commuting"}
	bt2 := TripCategory{2, "race training"}
	tripCatsList = append(tripCatsList, bt1)
	tripCatsList = append(tripCatsList, bt2)

	tc, err := tripCatsList.GetWithId(1)
	if err != nil {
		t.Errorf("%s", err)
	}
	if tc.Name != bt1.Name {
		t.Errorf("the name does not match the id")
	}
}

func TestGetWithName(t *testing.T) {
	tripCats := New()

	tripCats = append(tripCats, TripCategory{1, "commuting"})
	tripCats = append(tripCats, TripCategory{2, "race training"})

	if _, err := tripCats.GetWithName("race tra"); err != nil {
		t.Errorf("Exact match does not work.")
	}
	if _, err := tripCats.GetWithName("commut"); err != nil {
		t.Errorf("Part of the name does not match.")
	}
	if _, err := tripCats.GetWithName("ing"); err == nil {
		t.Errorf("Common part does not return error as it should.")
	}
}

func TestGetDisplayStrings(t *testing.T) {
	tripCats := New()

	tripCats = append(tripCats, TripCategory{1, "race training"})
	tripCats = append(tripCats, TripCategory{2, "commuting"})

	_, _, fsId, fsName := tripCats.GetDisplayStrings()
	if fsId != "%2d" {
		t.Errorf("formatting strings do not match. Is %s and expected %s", fsId, "%2d")
	}
	if fsName != "%13s" {
		t.Errorf("formatting strings do not match. Is %s and expected %s", fsName, "%13s")
	}

}
