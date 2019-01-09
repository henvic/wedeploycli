package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/ctxsignal"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/deployment/internal/copypkg"
	"github.com/wedeploy/cli/deployment/internal/feedback"
	"github.com/wedeploy/cli/deployment/transport"
	"github.com/wedeploy/cli/services"
)

// Transport for the deployment.
type Transport interface {
	Setup(context.Context, transport.Settings) error
	Init() error
	ProcessIgnored() (map[string]struct{}, error)
	Stage(services.ServiceInfoList) error
	Commit(message string) (hash string, err error)
	AddRemote() error
	Push() (groupUID string, err error)
	UploadDuration() time.Duration
	UserAgent() string
}

// Params for the deployment
type Params struct {
	ProjectID string
	ServiceID string

	Remote string

	Image       string
	CopyPackage string

	// Metadata type not used to avoid forcing Go's unordered map structure.
	// Metadata is only processed on the server-side (January 9th, 2019).
	Metadata json.RawMessage

	OnlyBuild    bool
	SkipProgress bool
	Quiet        bool
}

// Metadata for the deployment.
type Metadata map[string]json.RawMessage

// Deploy project
type Deploy struct {
	ctx           context.Context
	ConfigContext config.Context

	Transport
	Params

	Path     string
	Services services.ServiceInfoList

	groupUID string

	watch *feedback.Watch

	workDir string

	ignored map[string]struct{}
}

type changes struct {
	ServiceID string
	Image     string
}

func getPreparedServicePackage(c changes, path string) ([]byte, error) {
	// this smells a little bad because wedeploy.json is the responsibility of the services package
	// and I shouldn't be accessing it directly from here
	var sp = map[string]interface{}{}
	wedeployJSON, err := ioutil.ReadFile(filepath.Join(path, "wedeploy.json")) // #nosec
	switch {
	case err == nil:
		if err = json.Unmarshal(wedeployJSON, &sp); err != nil {
			return nil, errwrap.Wrapf("error parsing wedeploy.json on "+path+": {{err}}", err)
		}
	case os.IsNotExist(err):
		sp = map[string]interface{}{}
		err = nil
	default:
		return nil, err
	}

	sp["id"] = strings.ToLower(c.ServiceID)

	if c.Image != "" {
		sp["image"] = c.Image
	}

	delete(sp, "projectId")

	return json.MarshalIndent(&sp, "", "    ")
}

func (d *Deploy) copyGitPackage() error {
	_, _ = fmt.Fprintf(os.Stderr, "Debugging: copying (cloning) package file to %s\n", d.CopyPackage)

	var when = time.Now().Format("2006-01-02-15-04-05Z0700")
	var target = fmt.Sprintf("%s-%s", d.ProjectID, when)

	return copypkg.Copy(d.ctx, d.workDir, target)
}

// Do deployment
func (d *Deploy) Do(ctx context.Context, t Transport) (err error) {
	d.ctx = ctx
	d.Transport = t

	d.watch = &feedback.Watch{
		ConfigContext: d.ConfigContext,

		ProjectID: d.ProjectID,

		Services: d.Services,

		OnlyBuild:    d.OnlyBuild,
		SkipProgress: d.SkipProgress,
		Quiet:        d.Quiet,

		IsUpload: true,
	}

	d.workDir, err = ioutil.TempDir("", "wedeploy")

	if err != nil {
		return err
	}

	defer func() {
		if ec := d.CleanupPackage(); ec != nil {
			if err == nil {
				err = ec
				return
			}

			_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		}
	}()

	settings := transport.Settings{
		ConfigContext: d.ConfigContext,
		ProjectID:     d.ProjectID,
		Path:          d.Path,
		WorkDir:       d.workDir,
	}

	if err = d.Transport.Setup(d.ctx, settings); err != nil {
		return err
	}

	d.watch.Start(d.ctx)

	err = d.do()

	if d.SkipProgress && err != nil {
		return err
	}

	if d.SkipProgress {
		d.watch.PrintSkipProgress()
		return nil
	}

	if err != nil {
		d.watch.StopFailedUpload()
		return err
	}

	return d.watch.Wait()
}

func (d *Deploy) do() (err error) {
	if err = d.preparePackage(); err != nil {
		return err
	}

	d.watch.NotifyStart()

	err = d.uploadPackage()
	d.watch.GroupUID = d.GetGroupUID()

	return err
}

func (d *Deploy) preparePackage() (err error) {
	d.watch.NotifyPacking()

	if err = d.Transport.Init(); err != nil {
		return err
	}

	if d.ignored, err = d.Transport.ProcessIgnored(); err != nil {
		return err
	}

	if err = d.copyServices(); err != nil {
		return err
	}

	if err = d.Transport.Stage(d.Services); err != nil {
		return err
	}

	d.printPackageSize()

	if _, err = d.Transport.Commit(d.Info()); err != nil {
		return err
	}

	if d.CopyPackage != "" {
		if err = d.copyGitPackage(); err != nil {
			return errwrap.Wrapf("cannot copy git package: {{err}}", err)
		}
	}

	return nil
}

func (d *Deploy) copyServices() error {
	for _, s := range d.Services {
		err := d.copyServiceFiles(s.Location)

		if err != nil {
			return err
		}

		if err = d.prepareAndModifyServicePackage(s); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deploy) uploadPackage() (err error) {
	if err = d.Transport.AddRemote(); err != nil {
		return err
	}

	if d.groupUID, err = d.Transport.Push(); err != nil {
		if s, sctx := ctxsignal.Closed(d.ctx); sctx == nil {
			return errwrap.Wrapf("upload failed: signal: "+s.String(), err)
		}

		return err
	}

	d.watch.NotifyUploadComplete(d.Transport.UploadDuration())
	return nil
}

// CleanupPackage directory
func (d *Deploy) CleanupPackage() error {
	var tmpWorkDir = d.workDir

	if tmpWorkDir != "" {
		return os.RemoveAll(d.workDir)
	}

	return nil
}

// GetGroupUID gets the deployment group UID
func (d *Deploy) GetGroupUID() string {
	return d.groupUID
}
