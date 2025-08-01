package github

import (
	"context"
	"fmt"
	"log"

	githubSDK "github.com/google/go-github/v74/github"
)

type GithubClientConfig struct {
	githubClient *githubSDK.Client
}

type GithubClientConfigOptions func(*GithubClientConfig)

func WithAccessToken(accessToken string) GithubClientConfigOptions {
	return func(gcc *GithubClientConfig) {
		gcc.githubClient = githubSDK.NewClient(nil).WithAuthToken(accessToken)
	}
}

func NewGithubClient(opts ...GithubClientConfigOptions) GithubClient {
	_client := &GithubClientConfig{
		githubClient: githubSDK.NewClient(nil),
	}

	for _, _opt := range opts {
		_opt(_client)
	}

	return _client
}

// GetLatestReleaseAsset implements GithubClient.
func (g *GithubClientConfig) GetLatestReleaseAsset(ctx context.Context, req *ReleaseRequest) (*githubSDK.ReleaseAsset, error) {
	_rel, _res, err := g.githubClient.Repositories.GetLatestRelease(ctx, req.RepoOwner, req.RepoName)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(_res.Response)

	if len(_rel.Assets) > 0 {
		return _rel.Assets[0], nil
	}

	return nil, fmt.Errorf("no assets")
}
