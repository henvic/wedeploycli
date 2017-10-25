package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/wedeploy/cli/legal/extra"
	"github.com/wedeploy/cli/legal/internal/legal"
)

const codeFormat = `// Code generated by legal generator; DO NOT EDIT.

package legal

// Licenses (legal notices)
var Licenses = %s
`

const thirdPartyFormat = "Third party licenses\n\n%s"

func readExtraLicenses(buf *bytes.Buffer) error {
	for n, l := range extra.Licenses {
		var content, err = l.Get()

		if err != nil {
			return err
		}

		buf.Write(content)

		if n < len(extra.Licenses)-1 {
			buf.WriteString("\n\n")
		}
	}

	return nil
}

func main() {
	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(context.Background(), "vendorlicenses")
	cmd.Dir = ".."
	cmd.Stderr = os.Stderr
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	buf.WriteString("\n\n")

	if err := readExtraLicenses(buf); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	var text = buf.String()

	if err := saveCode(text); err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}

	if err := saveThirdPartyFile(text); err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
}

func saveCode(text string) error {
	var code = fmt.Sprintf(codeFormat, legal.FormatLicense(text))
	return ioutil.WriteFile("licenses.go", []byte(code), 0644)
}

func saveThirdPartyFile(text string) error {
	var licenses = fmt.Sprintf(thirdPartyFormat, text)
	return ioutil.WriteFile("../LICENSE-THIRD-PARTY", []byte(licenses), 0644)
}