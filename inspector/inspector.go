package inspector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/findresource"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/templates"
	"github.com/wedeploy/cli/verbose"
)

// GetSpec of the passed type
func GetSpec(t interface{}) []string {
	var fields []string
	val := reflect.ValueOf(t)
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)

		if c := field.Name[0]; c >= 'a' && c <= 'z' {
			continue
		}

		fields = append(fields, fmt.Sprintf("%v %v", field.Name, field.Type))
	}

	return fields
}

// ContextOverview for the context visualization
type ContextOverview struct {
	ProjectID string
	Services  []services.ServiceInfo

	directory string
}

func (overview *ContextOverview) loadService() error {
	var servicePath, _, cerr = getServicePackage(overview.directory)

	if cerr == nil || os.IsNotExist(cerr) {
		return nil
	}

	if errwrap.GetType(cerr, &json.SyntaxError{}) != nil {
		return errwrap.Wrapf(`{{err}}.
The wedeploy.json file syntax is described at https://wedeploy.com/docs/deploy/configuring-deployments/`, cerr)
	}

	if strings.Contains(cerr.Error(), servicePath) {
		return cerr
	}

	return errwrap.Wrapf("can't load service on "+servicePath+": {{err}}", cerr)
}

// Load the context overview for a given directory
func (overview *ContextOverview) Load(directory string) (err error) {
	if directory, err = filepath.Abs(directory); err != nil {
		return err
	}

	overview.directory = directory

	if err := overview.loadService(); err != nil {
		return err
	}

	if err := overview.loadServicesList(); err != nil {
		return err
	}

	return overview.setUniqueProjectID()
}

func (overview *ContextOverview) loadServicesList() error {
	var list, err = services.GetListFromDirectory(overview.directory)

	if err != nil {
		return err
	}

	overview.Services = list
	return nil
}

// setUniqueProjectID and return error if not unique
func (overview *ContextOverview) setUniqueProjectID() error {
	var prevService services.ServiceInfo

	for i, sInfo := range overview.Services {
		if i > 0 && sInfo.ProjectID != overview.ProjectID {
			relCurrent, _ := filepath.Rel(overview.directory, sInfo.Location)
			relPrev, _ := filepath.Rel(overview.directory, prevService.Location)
			return fmt.Errorf(
				`services "%s" and "%s" must have the same project ID defined on "%s" and "%s" (currently: "%s" and "%s")`,
				prevService.ServiceID,
				sInfo.ServiceID,
				filepath.Join(relPrev, "wedeploy.json"),
				filepath.Join(relCurrent, "wedeploy.json"),
				overview.ProjectID,
				sInfo.ProjectID,
			)
		}

		prevService = sInfo
		overview.ProjectID = sInfo.ProjectID
	}

	return nil
}

// InspectContext on a given directory, filtering by format
func InspectContext(format, directory string) (string, error) {
	var overview = ContextOverview{}

	if err := overview.Load(directory); err != nil {
		return "", err
	}

	return templates.ExecuteOrList(format, overview)
}

func getServicePackage(directory string) (path string, p *services.Package, err error) {
	var servicePath, cerr = findresource.GetRootDirectory(
		directory,
		findresource.GetSysRoot(),
		"wedeploy.json")

	if cerr != nil {
		return "", nil, cerr
	}

	p, err = services.Read(servicePath)

	if err != nil {
		return servicePath, nil, err
	}

	return servicePath, p, nil
}

// InspectService on a given directory, filtering by format
func InspectService(format, directory string) (string, error) {
	var servicePath, service, cerr = getServicePackage(directory)

	switch {
	case os.IsNotExist(cerr):
		return "", errwrap.Wrapf("Inspection failure: can not find service", cerr)
	case cerr != nil:
		return "", cerr
	}

	verbose.Debug("Reading service at " + servicePath)
	return templates.ExecuteOrList(format, service)
}
