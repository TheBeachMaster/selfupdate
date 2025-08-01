package service

import (
	"net/http"

	"com.thebeachmaster/selfupdate/internal/pkg/github"
	"com.thebeachmaster/selfupdate/internal/pkg/updater"
)

type serviceHandler struct {
	updateService updater.Updater
	githubClient  github.GithubClient
}

func NewServiceHandler(opts ...string) ServiceHandler {
	var _githubClient github.GithubClient
	if len(opts) > 0 {
		// We only care about the first option - for now
		_githubClient = github.NewGithubClient(github.WithAccessToken(opts[0]))
	} else {
		_githubClient = github.NewGithubClient()
	}
	return &serviceHandler{
		githubClient:  _githubClient,
		updateService: updater.NewUpdater(),
	}
}

// CheckAppVersionHandler implements ServiceHandler.
func (s *serviceHandler) CheckAppVersionHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// TODO:
	}
}

// UpdateAppHandler implements ServiceHandler.
func (s *serviceHandler) UpdateAppHandler() http.HandlerFunc {
	panic("unimplemented")
}
