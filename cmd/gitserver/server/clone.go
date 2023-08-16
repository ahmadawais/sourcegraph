package server

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func (s *Server) maybeStartClone(ctx context.Context, logger log.Logger, repoName api.RepoName) (notFound *protocol.NotFoundPayload, cloned bool) {
	dir := s.dir(repoName)
	if repoCloned(dir) {
		return nil, true
	}

	if conf.Get().DisableAutoGitUpdates {
		logger.Debug("not cloning on demand as DisableAutoGitUpdates is set")
		return &protocol.NotFoundPayload{}, false
	}

	cloneProgress, cloneInProgress := s.locker.Status(dir)
	if cloneInProgress {
		return &protocol.NotFoundPayload{
			CloneInProgress: true,
			CloneProgress:   cloneProgress,
		}, false
	}

	repo, err := s.DB.Repos().GetByName(ctx, repoName)
	if err != nil {
		logger.Debug("cannot find a repo in the database", log.String("repo", string(repoName)), log.Error(err))
		return &protocol.NotFoundPayload{}, false
	}
	_, _, err = s.DB.RepoUpdateJobs().Create(ctx, database.RepoUpdateJobOpts{RepoID: repo.ID})
	if err != nil {
		logger.Debug("error starting repo clone", log.String("repo", string(repoName)), log.Error(err))
		return &protocol.NotFoundPayload{CloneInProgress: false}, false
	}

	return &protocol.NotFoundPayload{
		CloneInProgress: true,
		CloneProgress:   cloneProgress,
	}, false
}
