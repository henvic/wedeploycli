// +build nocompile

package integration

func init() {
	println(`Skipping compilation: using "we" command available on system.`)
	binary = "we"
}
