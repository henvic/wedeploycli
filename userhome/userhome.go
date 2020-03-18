package userhome

import (
	"os"

	"github.com/henvic/wedeploycli/envs"
)

// GetHomeDir returns the user's ~ (home)
func GetHomeDir() string {
	if os.Getenv(envs.CustomHome) != "" {
		return os.Getenv(envs.CustomHome)
	}

	return getHomeDir()
}
