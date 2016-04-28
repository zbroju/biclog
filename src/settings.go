// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package src

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
