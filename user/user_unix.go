// +build !windows

package user

import "os"

func getHomeDir() string {
	return os.Getenv("HOME")
}
