package deployment

import (
	"context"
	"strings"

	"github.com/henvic/ctxsignal"
	"github.com/wedeploy/cli/services"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/deployment/internal/feedback"
	"github.com/wedeploy/cli/projects"
)

// ParamsFromRepository is used when deploying git from a repo.
type ParamsFromRepository struct {
	Repository string

	ProjectID string

	OnlyBuild    bool
	SkipProgress bool
	Quiet        bool
}

// DeployFromGitRepository deploys from a repository on the web.
func DeployFromGitRepository(ctx context.Context,
	wectx config.Context, params ParamsFromRepository) (
	services.ServiceInfoList, error) {
	projectsClient := projects.New(wectx)
	params.Repository = addSchemaGitRepoPath(params.Repository)

	build := projects.BuildRequestBody{
		Repository: params.Repository,

		Deploy: !params.OnlyBuild,
	}

	groupUID, builds, err := projectsClient.Build(ctx, params.ProjectID, build)

	if err != nil {
		return nil, err
	}

	sil := services.ServiceInfoList{}

	for _, b := range builds {
		sil = append(sil, services.ServiceInfo{
			ProjectID: params.ProjectID,
			ServiceID: b.ServiceID,
		})
	}

	var watch = &feedback.Watch{
		ConfigContext: wectx,

		ProjectID: params.ProjectID,
		GroupUID:  groupUID,

		Services: sil,

		SkipProgress: params.SkipProgress,
		Quiet:        params.Quiet,
	}

	var cancel context.CancelFunc
	ctx, cancel = ctxsignal.WithTermination(ctx)
	defer cancel()

	watch.Start(ctx)

	if params.SkipProgress {
		watch.PrintSkipProgress()
		return sil, nil
	}

	return sil, watch.Wait()
}

func addSchemaGitRepoPath(repo string) string {
	providers := []string{"github.com", "bitbucket.com", "gitlab.com"}

	for _, p := range providers {
		if strings.HasPrefix(repo, p+"/") {
			return "https://" + repo
		}
	}

	return repo
}
