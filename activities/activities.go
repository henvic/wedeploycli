package activities

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/templates"
)

// Activity record
type Activity struct {
	ID         string            `json:"id"`
	CreatedAt  int64             `json:"createdAt"`
	Commit     string            `json:"commit"`
	ProjectID  string            `json:"projectId"`
	ProjectUID string            `json:"projectUid"`
	Type       string            `json:"type"`
	Metadata   map[string]string `json:"metadata"`
}

// Activities slice
type Activities []Activity

// Reverse activities slice
func (as Activities) Reverse() (ras []Activity) {
	for l := len(as) - 1; l >= 0; l-- {
		ras = append(ras, as[l])
	}

	return ras
}

// Filter for list
type Filter struct {
	Commit   string `json:"commit,omitempty"`
	GroupUID string `json:"groupUid,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Type     string `json:"type,omitempty"`
}

const (
	buildFailed                        = "BUILD_FAILED"
	buildPending                       = "BUILD_PENDING"
	buildStarted                       = "BUILD_STARTED"
	buildSucceeded                     = "BUILD_SUCCEEDED"
	collaboratorDeleted                = "COLLABORATOR_DELETED"
	collaboratorInvitationAccepted     = "COLLABORATOR_INVITATION_ACCEPTED"
	collaboratorInvitationDeleted      = "COLLABORATOR_INVITATION_DELETED"
	collaboratorInvitationSent         = "COLLABORATOR_INVITATION_SENT"
	collaboratorLeft                   = "COLLABORATOR_LEFT"
	customDomainUpdated                = "CUSTOM_DOMAIN_UPDATED"
	deployFailed                       = "DEPLOY_FAILED"
	deployPending                      = "DEPLOY_PENDING"
	deployStarted                      = "DEPLOY_STARTED"
	deploySucceeded                    = "DEPLOY_SUCCEEDED"
	githubProviderConnected            = "GITHUB_PROVIDER_CONNECTED"
	githubProviderDisconnected         = "GITHUB_PROVIDER_DISCONNECTED"
	githubRepositoryConnected          = "GITHUB_REPOSITORY_CONNECTED"
	githubRepositoryDisconnected       = "GITHUB_REPOSITORY_DISCONNECTED"
	homeServiceUpdated                 = "HOME_SERVICE_UPDATED"
	projectCreated                     = "PROJECT_CREATED"
	projectRestarted                   = "PROJECT_RESTARTED"
	projectTransferred                 = "PROJECT_TRANSFERRED"
	serviceCreated                     = "SERVICE_CREATED"
	serviceDeleted                     = "SERVICE_DELETED"
	serviceEnvironmentVariablesUpdated = "SERVICE_ENVIRONMENT_VARIABLES_UPDATED"
	serviceRestarted                   = "SERVICE_RESTARTED"
)

var activityTemplates = map[string]string{
	buildFailed:                        "{{.Metadata.serviceId}} build failed on project {{.ProjectID}}",
	buildPending:                       "{{.Metadata.serviceId}} build pending on project {{.ProjectID}}",
	buildStarted:                       "{{.Metadata.serviceId}} build started on project {{.ProjectID}}",
	buildSucceeded:                     "{{.Metadata.serviceId}} build successful on project {{.ProjectID}}",
	collaboratorDeleted:                "{{.ProjectID}} project collaborator deleted",
	collaboratorInvitationAccepted:     "{{.ProjectID}} project collaborator invitation accepted",
	collaboratorInvitationDeleted:      "{{.ProjectID}} project collaborator invitation deleted",
	collaboratorInvitationSent:         "{{.ProjectID}} project collaborator invitation sent",
	collaboratorLeft:                   "{{.ProjectID}} project collaborator left",
	customDomainUpdated:                "{{.Metadata.serviceId}} custom domain updated on project {{.ProjectID}}",
	deployFailed:                       "{{.Metadata.serviceId}} deployment failed on project {{.ProjectID}}",
	deployPending:                      "{{.Metadata.serviceId}} deployment pending on project {{.ProjectID}}",
	deployStarted:                      "{{.Metadata.serviceId}} deployment started on project {{.ProjectID}}",
	deploySucceeded:                    "{{.Metadata.serviceId}} deployment successful on project {{.ProjectID}}",
	githubProviderConnected:            "GitHub provider connected",
	githubProviderDisconnected:         "GitHub provider disconnected",
	githubRepositoryConnected:          "GitHub repository connected",
	githubRepositoryDisconnected:       "GitHub repository disconnected",
	projectCreated:                     "{{.ProjectID}} project created",
	projectRestarted:                   "{{.ProjectID}} project restarted",
	projectTransferred:                 "{{.ProjectID}} project transferred",
	serviceCreated:                     "{{.Metadata.serviceId}} service created on project {{.ProjectID}}",
	serviceDeleted:                     "{{.Metadata.serviceId}} service deleted on project {{.ProjectID}}",
	serviceEnvironmentVariablesUpdated: "{{.Metadata.serviceId}} service environment variables updated on project {{.ProjectID}}",
	serviceRestarted:                   "{{.Metadata.serviceId}} service restarted on project {{.ProjectID}}",
}

// List activities of a given project
func List(ctx context.Context, projectID string, f Filter) (activities Activities, err error) {
	if projectID == "" {
		return activities, projects.ErrEmptyProjectID
	}

	var request = apihelper.URL(ctx, "/projects/"+url.QueryEscape(projectID)+"/activities")
	apihelper.ParamsFromJSON(request, f)

	apihelper.Auth(request)

	if err = apihelper.Validate(request, request.Get()); err != nil {
		return nil, err
	}

	err = apihelper.DecodeJSON(request, &activities)
	return activities, err
}

// PrettyPrintList prints the activities in a formatted way
func PrettyPrintList(activities []Activity) {
	for _, a := range activities {
		var msg, err = getActivityMessage(a, activityTemplates)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
			continue
		}

		fmt.Fprintf(os.Stdout, "%v\n", msg)
	}
}

func getActivityMessage(a Activity, template map[string]string) (string, error) {
	at, ok := template[a.Type]

	if !ok {
		return a.Type, nil
	}

	return templates.Execute(at, a)
}
