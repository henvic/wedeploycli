// +build !functional

package functional

func init() {
	println("Skipping all functional tests.")
	println("Use: go test -tags=functional")
}
