package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

type GitHubWrapper struct {
	client *github.Client
}

type PullRequestMetadata struct {
	pr       *github.PullRequest
	repoName string
}

func NewGitHubWrapper(token string) *GitHubWrapper {
	wrapper := &GitHubWrapper{}
	wrapper.client = createGitHubClient(token)
	return wrapper
}

func createGitHubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func (o *GitHubWrapper) ListPullRequests(repos RepoMap) chan PullRequestMetadata {
	channel := make(chan PullRequestMetadata)

	go func() {
		o.retrievePullRequestsForAllRepos(repos, channel)
	}()
	return channel
}

func (o *GitHubWrapper) retrieveRepository(userName string, repoName string) *github.Repository {
	repo, _, err := o.client.Repositories.Get(context.Background(), userName, repoName)
	if err != nil {
		if strings.Contains(err.Error(), "401 Bad credentials") {
			fmt.Fprintf(os.Stderr, "Bad Credentials")
		}
		panic(err)
	}
	return repo
}

func (o *GitHubWrapper) retrievePullRequests(userName string, repoName string, repo *github.Repository) []*github.PullRequest {
	prState := &github.PullRequestListOptions{State: "open"}

	prs, _, err := o.client.PullRequests.List(context.Background(), userName, repo.GetName(), prState)
	if err != nil {
		panic(err)
	}
	return prs
}

func (o *GitHubWrapper) retrievePullRequestsForAllRepos(repos map[string][]string, channel chan PullRequestMetadata) {
	for userName, repoNames := range repos {
		for _, repoName := range repoNames {
			repo := o.retrieveRepository(userName, repoName)
			prs := o.retrievePullRequests(userName, repoName, repo)
			for _, pr := range prs {
				channel <- PullRequestMetadata{pr, repo.GetName()}
			}
		}
	}
}

func verifyToken(client *github.Client) {
	_, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		log.Fatal(err)
	}
}
