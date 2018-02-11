package inspector

import (
	"errors"
	"fmt"
	"os"
	"reflect"

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
	Services []services.ServiceInfo
}

func (overview *ContextOverview) loadService(directory string) error {
	var servicePath, _, cerr = getServicePackage(directory)

	switch {
	case os.IsNotExist(cerr):
	case cerr != nil:
		return errwrap.Wrapf("can't load service on "+servicePath+": {{err}}", cerr)
	}

	return nil
}

// Load the context overview for a given directory
func (overview *ContextOverview) Load(directory string) error {

	if err := overview.loadService(directory); err != nil {
		return err
	}

	return overview.loadServicesList(directory)
}

func (overview *ContextOverview) loadServicesList(directory string) error {
	var list, err = services.GetListFromDirectory(directory)

	if err != nil {
		return err
	}

	overview.Services = list
	return nil
}

func (overview *ContextOverview) UniqueProjectID() (error) {
	fmt.Println("overview.Services", overview.Services)
	for _, sInfo := range overview.Services {
		if overview.ProjectID != "" && sInfo.ProjectID != overview.ProjectID {
			return errors.New("You can only have one unique Project ID in your wedeploy.json's")
		}
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

func getServicePackage(directory string) (path string, cp *services.ServicePackage, err error) {
	var servicePath, cerr = getServiceRootDirectory(directory)

	if cerr != nil {
		return "", nil, cerr
	}

	cp, err = services.Read(servicePath)

	if err != nil {
		return servicePath, nil, err
	}

	return servicePath, cp, nil
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

func getServiceRootDirectory(dir string) (string, error) {
	return getRootDirectory(dir, "wedeploy.json")
}

func getRootDirectory(dir, file string) (string, error) {
	return findresource.GetRootDirectory(dir, findresource.GetSysRoot(), file)
}
