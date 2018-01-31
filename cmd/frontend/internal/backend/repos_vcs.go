package backend

import (
	"context"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"

	"github.com/pkg/errors"

	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
)

// OpenVCS returns a handle to the underlying Git repository (i.e., on gitserver).
func (repos) OpenVCS(ctx context.Context, repo *types.Repo) (vcs.Repository, error) {
	if Mocks.Repos.OpenVCS != nil {
		return Mocks.Repos.OpenVCS(ctx, repo)
	}
	gitserverRepo, err := (repos{}).GitserverRepoInfo(ctx, repo)
	if err != nil {
		return nil, err
	}
	return gitcmd.Open(gitserverRepo.Name, gitserverRepo.URL), nil
}

// VCSForGitserverRepo returns a handle to the Git repository specified by repo.
func (repos) VCSForGitserverRepo(repo gitserver.Repo) vcs.Repository {
	return gitcmd.Open(repo.Name, repo.URL)
}

func (repos) GitserverRepoInfo(ctx context.Context, repo *types.Repo) (gitserver.Repo, error) {
	if envvar.SourcegraphDotComMode() {
		// If it is possible to 100% correctly determine it statically, use a fast path. This is used
		// to avoid a RepoLookup call for public GitHub.com repositories on Sourcegraph.com, which reduces
		// rate limit pressure significantly.
		//
		// This fails for private repositories, which require authentication in the URL userinfo.
		if strings.HasPrefix(strings.ToLower(string(repo.URI)), "github.com/") {
			return gitserver.Repo{Name: repo.URI, URL: "https://" + string(repo.URI)}, nil
		}
	}

	result, err := repoupdater.DefaultClient.RepoLookup(ctx, protocol.RepoLookupArgs{
		Repo:         repo.URI,
		ExternalRepo: repo.ExternalRepo,
	})
	if err != nil {
		return gitserver.Repo{}, err
	}
	return gitserver.Repo{Name: result.Repo.URI, URL: result.Repo.VCS.URL}, nil
}

// ResolveRev will return the absolute commit for a commit-ish spec in a repo.
// If no rev is specified, HEAD is used.
// Error cases:
// * Repo does not exist: vcs.RepoNotExistError
// * Commit does not exist: vcs.ErrRevisionNotFound
// * Empty repository: vcs.ErrRevisionNotFound
// * The user does not have permission: errcode.IsNotFound
// * Other unexpected errors.
func (s *repos) ResolveRev(ctx context.Context, repo *types.Repo, rev string) (commitID api.CommitID, err error) {
	if Mocks.Repos.ResolveRev != nil {
		return Mocks.Repos.ResolveRev(ctx, repo, rev)
	}

	ctx, done := trace(ctx, "Repos", "ResolveRev", map[string]interface{}{"repo": repo.URI, "rev": rev}, &err)
	defer done()

	vcsrepo, err := Repos.OpenVCS(ctx, repo)
	if err != nil {
		return "", err
	}
	return vcsrepo.ResolveRevision(ctx, rev)
}

func (s *repos) GetCommit(ctx context.Context, repo *types.Repo, commitID api.CommitID) (res *vcs.Commit, err error) {
	if Mocks.Repos.GetCommit != nil {
		return Mocks.Repos.GetCommit(ctx, repo, commitID)
	}

	ctx, done := trace(ctx, "Repos", "GetCommit", map[string]interface{}{"repo": repo.URI, "commitID": commitID}, &err)
	defer done()

	log15.Debug("svc.local.repos.GetCommit", "repo", repo.URI, "commitID", commitID)

	if !isAbsCommitID(commitID) {
		return nil, errors.Errorf("non-absolute CommitID for Repos.GetCommit: %v", commitID)
	}

	vcsrepo, err := Repos.OpenVCS(ctx, repo)
	if err != nil {
		return nil, err
	}

	return vcsrepo.GetCommit(ctx, commitID)
}

func isAbsCommitID(commitID api.CommitID) bool { return len(commitID) == 40 }
