package extra

import (
	"fmt"
	"io/ioutil"
)

// License structure
type License struct {
	Name        string
	Package     string
	Notes       string
	LicensePath string
}

// Get license
func (l *License) Get() ([]byte, error) {
	var license, err = ioutil.ReadFile(l.LicensePath)

	if err != nil {
		return nil, err
	}

	var content = []byte(fmt.Sprintf("%s %s", l.Name, l.Package))

	if len(l.Notes) != 0 {
		content = append(content, []byte(fmt.Sprintf(" (%s)", l.Notes))...)
	}

	content = append(content, "\n"...)
	content = append(content, license...)

	return content, nil
}
