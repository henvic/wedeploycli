package deployment

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeployment(t *testing.T) {
	var (
		defaultErrStream = errStream
		defaultOutStream = outStream
	)

	defer func() {
		errStream = defaultErrStream
		outStream = defaultOutStream
	}()

	var deployTest = &DeployTest{
		t: t,
	}
	errStream = &deployTest.bufErrStream
	outStream = &deployTest.bufOutStream
	deployTest.Run()
}

type DeployTest struct {
	bufErrStream bytes.Buffer
	bufOutStream bytes.Buffer
	t            *testing.T
	deploy       *Deploy
}

func (dt *DeployTest) Run() {
	if err := os.RemoveAll("./mocks/project/.git"); err != nil {
		dt.t.Errorf("Unexpected error %v", err)
	}

	if err := os.RemoveAll("./mocks/project-repo"); err != nil {
		dt.t.Errorf("Unexpected error %v", err)
	}

	defer func() {
		_ = os.RemoveAll("./mocks/project/.git")
	}()

	defer func() {
		_ = os.RemoveAll("./mocks/project-repo")
	}()

	dt.deploy = &Deploy{
		Context:          context.Background(),
		ProjectID:        "project",
		Path:             "./mocks/project",
		Remote:           "wedeploy",
		GitRemoteAddress: filepath.Join(abs("./mocks/project-repo")),
	}

	dt.TryInitializeRepo()
	dt.TryCommit()
	dt.TryPush()

}

func (dt *DeployTest) TryInitializeRepo() {
	if err := dt.deploy.InitializeRepository(); err != nil {
		dt.t.Errorf("Unexpected error %v", err)
	}

	if _, err := os.Stat(filepath.Join("./mocks/project", ".git")); err != nil {
		dt.t.Errorf("Expected git directory to be initialized, got %v instead", err)
	}

	// try to initialize again
	// it should not initialize, as it already exists, but no error should happen
	if err := dt.deploy.InitializeRepository(); err != nil {
		dt.t.Errorf("Unexpected error %v", err)
	}
}

func (dt *DeployTest) TryCommit() {
	if err := dt.deploy.CheckCurrentBranchIsMaster(); err != nil {
		dt.t.Errorf("Unexpected error %v", err)
	}

	stage, stageErr := dt.deploy.CheckUncommittedChanges()
	var wantStage = "?? project.json\n"

	if stage != wantStage {
		dt.t.Errorf("Stage should be %v, got %v instead", wantStage, stage)
	}

	if stageErr != nil {
		dt.t.Errorf("Expected no error checking for uncommitted changes, got %v instead", stageErr)
	}

	if !dt.deploy.uncommittedChanges {
		dt.t.Errorf("Expected to have uncommited changes")
	}

	var wantOut = "You have uncommitted changes"
	if !strings.Contains(dt.bufOutStream.String(), wantOut) {
		dt.t.Errorf("Expected outStream to have %v, got %v instead", wantOut, dt.bufOutStream.String())
	}

	if _, err := os.Stat(filepath.Join("./mocks/project", ".git")); err != nil {
		dt.t.Errorf("Expected no error, got %v instead", err)
	}

	commit, commitErr := dt.deploy.Commit()

	if commitErr != nil {
		dt.t.Errorf("Expected no commit error, got %v instead", commitErr)
	}

	if len(commit) != 40 {
		dt.t.Errorf("Expected git commit to have 40 characters, got commit %v instead", commit)
	}

	if !strings.Contains(dt.bufOutStream.String(), "commit ") ||
		!strings.Contains(dt.bufOutStream.String(), "Deployment at") {
		dt.t.Errorf("Expected commit message not found")
	}
}

func (dt *DeployTest) cloneToRemote() {
	var cmd = exec.CommandContext(context.Background(),
		"git",
		"clone",
		"mocks/project",
		"mocks/project-repo")

	var err = cmd.Run()

	if err != nil {
		dt.t.Errorf("Error trying to clone mock: %v", err)
	}
}

func (dt *DeployTest) TryPush() {
	if err := dt.deploy.AddRemote(); err != nil {
		dt.t.Errorf("Expected no error when adding remote, got %v instead", err)
	}

	dt.cloneToRemote()

	if err := dt.deploy.Push(); err != nil {
		dt.t.Errorf("Expected no error, got %v instead", err)
	}

	if !strings.Contains(dt.bufErrStream.String(), "Everything up-to-date") {
		dt.t.Errorf("Expected git push message")
	}
}

func abs(path string) string {
	var abs, err = filepath.Abs(path)

	if err != nil {
		panic(err)
	}

	return abs
}
