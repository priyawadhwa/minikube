package monitor

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// Client provides the context and client with necessary auth
// for interacting with the Github API
type Client struct {
	ctx context.Context
	*github.Client
	owner string
	repo  string
}

// NewClient returns a github client with the necessary auth
func NewClient(ctx context.Context, owner, repo string) *Client {
	githubToken := os.Getenv(GithubAccessTokenEnvVar)
	// Setup the token for github authentication
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	// Return a client instance from github
	client := github.NewClient(tc)
	return &Client{
		ctx:    ctx,
		Client: client,
		owner:  owner,
		repo:   repo,
	}
}

// CommentOnPR comments message on the PR
func (g *Client) CommentOnPR(pr int, message string) error {
	comment := &github.IssueComment{
		Body: &message,
	}

	log.Printf("Creating comment on PR %d: %s", pr, message)
	_, _, err := g.Client.Issues.CreateComment(g.ctx, g.owner, g.repo, pr, comment)
	if err != nil {
		return errors.Wrap(err, "creating github comment")
	}
	log.Printf("Successfully commented on PR %d.", pr)
	return nil
}

// ListOpenPRsWithLabel returns all open PRs with the specified label
func (g *Client) ListOpenPRsWithLabel(label string) ([]int, error) {
	// TODO: priyawadhwa@
	return []int{5694}, nil
}

// NewCommitsExist checks if new commits exist since minikube-bot
// commented on the PR. If so, return true.
func (g *Client) NewCommitsExist(pr int, login string) (bool, error) {
	lastCommentTime, err := g.timeOfLastComment(pr, login)
	if err != nil {
		return false, errors.Wrapf(err, "getting time of last comment by %s on pr %d", login, pr)
	}
	lastCommitTime, err := g.timeOfLastCommit(pr)
	if err != nil {
		return false, errors.Wrapf(err, "getting time of last commit on pr %d", pr)
	}
	return lastCommentTime.Before(lastCommitTime), nil
}

func (g *Client) timeOfLastCommit(pr int) (time.Time, error) {
	commits, _, err := g.Client.PullRequests.ListCommits(g.ctx, g.owner, g.repo, pr, &github.ListOptions{})
	if err != nil {
		return time.Time{}, err
	}
	lastCommit := commits[len(commits)-1]
	return lastCommit.GetCommit().GetAuthor().GetDate(), nil
}

func (g *Client) timeOfLastComment(pr int, login string) (time.Time, error) {
	comments, _, err := g.Client.Issues.ListComments(g.ctx, g.owner, g.repo, pr, &github.IssueListCommentsOptions{})
	if err != nil {
		return time.Time{}, err
	}
	// go through comments backwards to find the most recent
	for i := len(comments) - 1; i >= 0; i-- {
		c := comments[i]
		if u := c.GetUser(); u != nil {
			if u.GetLogin() == login {
				return c.GetCreatedAt(), nil
			}
		}
	}

	return time.Time{}, nil
}
