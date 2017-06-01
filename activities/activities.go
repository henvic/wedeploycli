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
	// BuildFailed state
	BuildFailed = "BUILD_FAILED"

	// BuildPending state
	BuildPending = "BUILD_PENDING"

	BuildStarted = "BUILD_STARTED"
	// BuildStarted state

	BuildSucceeded = "BUILD_SUCCEEDED"
	// BuildSucceeded state

	CollaboratorDeleted = "COLLABORATOR_DELETED"
	// CollaboratorDeleted state

	CollaboratorInvitationAccepted = "COLLABORATOR_INVITATION_ACCEPTED"
	// CollaboratorInvitationAccepted state

	CollaboratorInvitationDeleted = "COLLABORATOR_INVITATION_DELETED"
	// CollaboratorInvitationDeleted state

	CollaboratorInvitationSent = "COLLABORATOR_INVITATION_SENT"
	// CollaboratorInvitationSent state

	CollaboratorLeft = "COLLABORATOR_LEFT"
	// CollaboratorLeft state

	CustomDomainUpdated = "CUSTOM_DOMAIN_UPDATED"
	// CustomDomainUpdated state

	DeployFailed = "DEPLOY_FAILED"
	// DeployFailed state

	DeployPending = "DEPLOY_PENDING"
	// DeployPending state

	DeployStarted = "DEPLOY_STARTED"
	// DeployStarted state

	DeploySucceeded = "DEPLOY_SUCCEEDED"
	// DeploySucceeded state

	GithubProviderConnected = "GITHUB_PROVIDER_CONNECTED"
	// GithubProviderConnected state

	GithubProviderDisconnected = "GITHUB_PROVIDER_DISCONNECTED"
	// GithubProviderDisconnected state

	GithubRepositoryConnected = "GITHUB_REPOSITORY_CONNECTED"
	// GithubRepositoryConnected state

	GithubRepositoryDisconnected = "GITHUB_REPOSITORY_DISCONNECTED"
	// GithubRepositoryDisconnected state

	HomeServiceUpdated = "HOME_SERVICE_UPDATED"
	// HomeServiceUpdated state

	ProjectCreated = "PROJECT_CREATED"
	// ProjectCreated state

	ProjectRestarted = "PROJECT_RESTARTED"
	// ProjectRestarted state

	ProjectTransferred = "PROJECT_TRANSFERRED"
	// ProjectTransferred state

	ServiceCreated = "SERVICE_CREATED"
	// ServiceCreated state

	ServiceDeleted = "SERVICE_DELETED"
	// ServiceDeleted state

	ServiceEnvironmentVariablesUpdated = "SERVICE_ENVIRONMENT_VARIABLES_UPDATED"
	// ServiceEnvironmentVariablesUpdated state

	ServiceRestarted = "SERVICE_RESTARTED"
	// ServiceRestarted state

)

var activityTemplates = map[string]string{
	BuildFailed:                        "{{.Metadata.serviceId}} build failed on project {{.ProjectID}}",
	BuildPending:                       "{{.Metadata.serviceId}} build pending on project {{.ProjectID}}",
	BuildStarted:                       "{{.Metadata.serviceId}} build started on project {{.ProjectID}}",
	BuildSucceeded:                     "{{.Metadata.serviceId}} build successful on project {{.ProjectID}}",
	CollaboratorDeleted:                "{{.ProjectID}} project collaborator deleted",
	CollaboratorInvitationAccepted:     "{{.ProjectID}} project collaborator invitation accepted",
	CollaboratorInvitationDeleted:      "{{.ProjectID}} project collaborator invitation deleted",
	CollaboratorInvitationSent:         "{{.ProjectID}} project collaborator invitation sent",
	CollaboratorLeft:                   "{{.ProjectID}} project collaborator left",
	CustomDomainUpdated:                "{{.Metadata.serviceId}} custom domain updated on project {{.ProjectID}}",
	DeployFailed:                       "{{.Metadata.serviceId}} deployment failed on project {{.ProjectID}}",
	DeployPending:                      "{{.Metadata.serviceId}} deployment pending on project {{.ProjectID}}",
	DeployStarted:                      "{{.Metadata.serviceId}} deployment started on project {{.ProjectID}}",
	DeploySucceeded:                    "{{.Metadata.serviceId}} deployment successful on project {{.ProjectID}}",
	GithubProviderConnected:            "GitHub provider connected",
	GithubProviderDisconnected:         "GitHub provider disconnected",
	GithubRepositoryConnected:          "GitHub repository connected",
	GithubRepositoryDisconnected:       "GitHub repository disconnected",
	ProjectCreated:                     "{{.ProjectID}} project created",
	ProjectRestarted:                   "{{.ProjectID}} project restarted",
	ProjectTransferred:                 "{{.ProjectID}} project transferred",
	ServiceCreated:                     "{{.Metadata.serviceId}} service created on project {{.ProjectID}}",
	ServiceDeleted:                     "{{.Metadata.serviceId}} service deleted on project {{.ProjectID}}",
	ServiceEnvironmentVariablesUpdated: "{{.Metadata.serviceId}} service environment variables updated on project {{.ProjectID}}",
	ServiceRestarted:                   "{{.Metadata.serviceId}} service restarted on project {{.ProjectID}}",
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
