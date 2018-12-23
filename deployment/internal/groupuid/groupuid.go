package groupuid

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
)

var (
	gitRemoteMessagePrefix     = "remote: " // git client prepends this when printing messages
	gitRemoteDeployPrefix      = "wedeploy="
	gitRemoteDeployErrorPrefix = "wedeployError="
)

// Extract group uid from remote message.
func Extract(s string) (string, error) {
	lines := strings.Split(s, "\n")

	for _, line := range lines {
		line = strings.TrimPrefix(line, gitRemoteMessagePrefix)

		if strings.HasPrefix(line, gitRemoteDeployPrefix) {
			// \x1b[K is showing up at the end of "remote: wedeploy=" on at least git 1.9
			line = strings.TrimSuffix(line, "\x1b[K\n")
			return extractGroupUIDFromBuild([]byte(strings.TrimPrefix(line, gitRemoteDeployPrefix)))
		}

		if strings.HasPrefix(line, gitRemoteDeployErrorPrefix) {
			return "", extractErrorFromBuild([]byte(strings.TrimPrefix(line, gitRemoteDeployErrorPrefix)))
		}
	}

	return "", errors.New("can't find deployment group UID response")
}

func extractErrorFromBuild(e []byte) error {
	var af apihelper.APIFault
	if errJSON := json.Unmarshal(e, &af); errJSON != nil {
		return fmt.Errorf(`can't process error message: "%s"`, bytes.TrimSpace(e))
	}

	return af
}

type buildDeploymentOnGitServer struct {
	GroupUID string `json:"groupUid"`
}

func extractGroupUIDFromBuild(e []byte) (groupUID string, err error) {
	var bds []buildDeploymentOnGitServer

	if errJSON := json.Unmarshal(e, &bds); errJSON != nil {
		return "", errwrap.Wrapf("deployment response is invalid: {{err}}", errJSON)
	}

	if len(bds) == 0 {
		return "", errors.New("found no build during deployment")
	}

	return bds[0].GroupUID, nil
}
