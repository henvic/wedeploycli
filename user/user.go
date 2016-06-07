package user

import "os"

// GetHomeDir returns the user's ~ (home)
func GetHomeDir() string {
	if os.Getenv("WEDEPLOY_CUSTOM_HOME") != "" {
		return os.Getenv("WEDEPLOY_CUSTOM_HOME")
	}

	return getHomeDir()
}
