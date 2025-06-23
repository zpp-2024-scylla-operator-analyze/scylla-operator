package symptoms

import (
	"net/url"
	"strconv"
)

const scyllaOperatorGithubRepositoryUrl = "https://github.com/scylladb/scylla-operator"

type ExternalIssueReference interface {
	Name() string
	Reference() (string, error)
}

type githubIssueReference struct {
	issueId           int
	baseRepositoryUrl string
}

func NewScyllaOperatorIssueReference(issueId int) ExternalIssueReference {
	return &githubIssueReference{
		issueId:           issueId,
		baseRepositoryUrl: scyllaOperatorGithubRepositoryUrl,
	}
}

func (ref *githubIssueReference) Name() string {
	return "scylla-operator#" + strconv.Itoa(ref.issueId)
}

func (ref *githubIssueReference) Reference() (string, error) {
	return url.JoinPath(ref.baseRepositoryUrl, "issues", strconv.Itoa(ref.issueId))
}
