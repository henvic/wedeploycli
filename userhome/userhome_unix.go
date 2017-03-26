// +build !windows

package userhome

import "os"

func getHomeDir() string {
	return os.Getenv("HOME")
}
