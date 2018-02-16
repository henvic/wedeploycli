package activities

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/templates"
)

// Client for the services
type Client struct {
	*apihelper.Client
}

// New Client
func New(wectx config.Context) *Client {
	return &Client{
		&apihelper.Client{
			Context: wectx,
		},
	}
}

// Activity record
type Activity struct {
	ID         string                 `json:"id"`
	CreatedAt  int64                  `json:"createdAt"`
	Commit     string                 `json:"commit"`
	ProjectID  string                 `json:"projectId"`
	ProjectUID string                 `json:"projectUid"`
	Type       string                 `json:"type"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Activities slice
type Activities []Activity

// Reverse activities slice
func (as Activities) Reverse() (ras []Activity) {
	for l := len(as) - 1; l >= 0; l-- {
		ras = append(ras, as[l])
	}

	sort.SliceStable(ras, func(i, j int) bool {
		return ras[i].CreatedAt < ras[j].CreatedAt
	})

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

	// BuildStarted state
	BuildStarted = "BUILD_STARTED"

	// BuildPushed state
	BuildPushed = "BUILD_PUSHED"

	// BuildSucceeded state
	BuildSucceeded = "BUILD_SUCCEEDED"

	// CollaboratorDeleted state
	CollaboratorDeleted = "COLLABORATOR_DELETED"

	// CollaboratorInvitationAccepted state
	CollaboratorInvitationAccepted = "COLLABORATOR_INVITATION_ACCEPTED"

	// CollaboratorInvitationDeleted state
	CollaboratorInvitationDeleted = "COLLABORATOR_INVITATION_DELETED"

	// CollaboratorInvitationSent state
	CollaboratorInvitationSent = "COLLABORATOR_INVITATION_SENT"

	// CollaboratorLeft state
	CollaboratorLeft = "COLLABORATOR_LEFT"

	// CustomDomainUpdated state
	CustomDomainUpdated = "CUSTOM_DOMAIN_UPDATED"

	// DeployFailed state
	DeployFailed = "DEPLOY_FAILED"

	// DeployCanceled state
	DeployCanceled = "DEPLOY_CANCELED"

	// DeployTimeout state
	DeployTimeout = "DEPLOY_TIMEOUT"

	// DeployRollback state
	DeployRollback = "DEPLOY_ROLLBACK"

	// DeployPending state
	DeployPending = "DEPLOY_PENDING"

	// DeployCreated state
	DeployCreated = "DEPLOY_CREATED"

	// DeployStarted state
	DeployStarted = "DEPLOY_STARTED"

	// DeploySucceeded state
	DeploySucceeded = "DEPLOY_SUCCEEDED"

	// GithubProviderConnected state
	GithubProviderConnected = "GITHUB_PROVIDER_CONNECTED"

	// GithubProviderDisconnected state
	GithubProviderDisconnected = "GITHUB_PROVIDER_DISCONNECTED"

	// GithubRepositoryConnected state
	GithubRepositoryConnected = "GITHUB_REPOSITORY_CONNECTED"

	// GithubRepositoryDisconnected state
	GithubRepositoryDisconnected = "GITHUB_REPOSITORY_DISCONNECTED"

	// HomeServiceUpdated state
	HomeServiceUpdated = "HOME_SERVICE_UPDATED"

	// ProjectCreated state
	ProjectCreated = "PROJECT_CREATED"

	// ProjectRestarted state
	ProjectRestarted = "PROJECT_RESTARTED"

	// ProjectTransferred state
	ProjectTransferred = "PROJECT_TRANSFERRED"

	// ServiceCreated state
	ServiceCreated = "SERVICE_CREATED"

	// ServiceDeleted state
	ServiceDeleted = "SERVICE_DELETED"

	// ServiceEnvironmentVariablesUpdated state
	ServiceEnvironmentVariablesUpdated = "SERVICE_ENVIRONMENT_VARIABLES_UPDATED"

	// ServiceRestarted state
	ServiceRestarted = "SERVICE_RESTARTED"
)

// FriendlyActivities related to project
var FriendlyActivities = map[string]string{
	BuildFailed:     "Build failed",
	BuildStarted:    "Build started",
	BuildPushed:     "Build pushed",
	BuildSucceeded:  "Build succeeded",
	DeployFailed:    "Deployment failed",
	DeployCanceled:  "Deployment canceled",
	DeployTimeout:   "Deployment timed out",
	DeployRollback:  "Deployment rollback",
	DeployCreated:   "Deployment created",
	DeployPending:   "Deployment pending",
	DeploySucceeded: "Deployment succeeded",
	DeployStarted:   "Deployment started",
}

var activityTemplates = map[string]string{
	BuildFailed:                        "{{.Metadata.serviceId}} build failed on project {{.ProjectID}}",
	BuildStarted:                       "{{.Metadata.serviceId}} build started on project {{.ProjectID}}",
	BuildPushed:                        "{{.Metadata.serviceId}} build pushed on project {{.ProjectID}}",
	BuildSucceeded:                     "{{.Metadata.serviceId}} build succeeded on project {{.ProjectID}}",
	CollaboratorDeleted:                "{{.ProjectID}} project collaborator deleted",
	CollaboratorInvitationAccepted:     "{{.ProjectID}} project collaborator invitation accepted",
	CollaboratorInvitationDeleted:      "{{.ProjectID}} project collaborator invitation deleted",
	CollaboratorInvitationSent:         "{{.ProjectID}} project collaborator invitation sent",
	CollaboratorLeft:                   "{{.ProjectID}} project collaborator left",
	CustomDomainUpdated:                "{{.Metadata.serviceId}} custom domain updated on project {{.ProjectID}}",
	DeployFailed:                       "{{.Metadata.serviceId}} deployment failed on project {{.ProjectID}}",
	DeployCanceled:                     "{{.Metadata.serviceId}} deployment canceled on project {{.ProjectID}}",
	DeployTimeout:                      "{{.Metadata.serviceId}} deployment timed out on project {{.ProjectID}}",
	DeployRollback:                     "{{.Metadata.serviceId}} deployment rollback on project {{.ProjectID}}",
	DeployPending:                      "{{.Metadata.serviceId}} deployment pending on project {{.ProjectID}}",
	DeployCreated:                      "{{.Metadata.serviceId}} deployment created on project {{.ProjectID}}",
	DeployStarted:                      "{{.Metadata.serviceId}} deployment started on project {{.ProjectID}}",
	DeploySucceeded:                    "{{.Metadata.serviceId}} deployment succeeded on project {{.ProjectID}}",
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
func (c *Client) List(ctx context.Context, projectID string, f Filter) (activities Activities, err error) {
	if projectID == "" {
		return activities, projects.ErrEmptyProjectID
	}

	var request = c.Client.URL(ctx, "/projects/"+url.QueryEscape(projectID)+"/activities")
	apihelper.ParamsFromJSON(request, f)

	c.Client.Auth(request)

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
