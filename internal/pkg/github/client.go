package github

import (
	"context"

	githubSDK "github.com/google/go-github/v74/github"
)

type ReleaseRequest struct {
	RepoOwner string
	RepoName  string
}

type GithubClient interface {
	GetLatestReleaseAsset(ctx context.Context, req *ReleaseRequest) (*githubSDK.ReleaseAsset, error)
}
