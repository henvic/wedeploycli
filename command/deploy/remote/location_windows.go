// +build windows

package deployremote

import "path/filepath"

func init() {
	// from https://wiki.carleton.edu/pages/viewpage.action?pageId=9961710
	forbiddenLocations = append(forbiddenLocations,
		filepath.Join(home, "My Documents"),
		filepath.Join(home, "My Documents", "My Pictures"),
		filepath.Join(home, "Application Data"),
		filepath.Join(home, "Local Settings", "Application Data"),
		filepath.Join(home, "Local Settings", "Temp"),
		filepath.Join(home, "AppData"),
		filepath.Join(home, "AppData", "Roaming"),
		filepath.Join(home, "AppData", "Local"),
		filepath.Join(home, "AppData", "Local", "Temp"),
		filepath.Join(home, "Start Menu", "Programs"),
		filepath.Join(home, "Favorites"),
		`C:\`,
		`C:\WINDOWS`,
		`C:\WINNT`,
		`C:\WINDOWS\Program Files`,
		`C:\WINDOWS\Documents and Settings`,
		`C:\WINDOWS\Users`,
		`C:\WINDOWS\System32`,
		`C:\Documents and Settings\All Users`,
		`C:\Documents and Settings\All Users\Application Data`,
		`C:\Documents and Settings\All Users\Start Menu\Programs`,
		`C:\Documents and Settings\All Users\Desktop`,
		`C:\Documents and Settings\All Users\Documents`,
		`C:\Documents and Settings\All Users\Documents\My Pictures`,
		`C:\Users\Public`)
}
