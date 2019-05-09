package integration

import (
	"strings"
	"testing"
)

func TestAboutLegal(t *testing.T) {
	var cmd = &Command{
		Args: []string{"about", "legal"},
	}

	cmd.Run()

	if cmd.ExitCode != 0 {
		t.Errorf("Wanted exit code 0, got %v instead", cmd.ExitCode)
	}

	var header = `Legal Notices:

Copyright © 2016-present Liferay, Inc.
Liferay Cloud Platform CLI Software License Agreement

Liferay, the Liferay logo, WeDeploy, and WeDeploy logo
are trademarks of Liferay, Inc., registered in the U.S. and other countries.

Acknowledgements:
Portions of this Liferay software may utilize the following copyrighted material,
the use of which is hereby acknowledged.`

	var text = cmd.Stdout.String()

	if !strings.HasPrefix(text, header) {
		t.Errorf("Wanted header to be %v, got %v instead", header, cmd.Stdout)
	}

	var minTextLenght = 60000

	if len(text) < minTextLenght {
		t.Errorf("Licenses text lenght too small: expected at least %v, got %v", minTextLenght, len(text))
	}
}
