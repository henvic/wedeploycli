package user

import (
	"os"
	"runtime"
)

// GetHomeDir returns the user's ~ (home)
// Extracted from Viper's util.go GetUserHomeDir method
func GetHomeDir() string {
	if os.Getenv("LAUNCHPAD_CUSTOM_HOME") != "" {
		return os.Getenv("LAUNCHPAD_CUSTOM_HOME")
	}

	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}

	return os.Getenv("HOME")
}
