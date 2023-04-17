package database

import (
	"context"
	"database/sql"
	"hash/fnv"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

/////
// TODO: There is no support for directory root now. The fix is simple though.

type OwnSignalStore interface {
	AddCommit(ctx context.Context, commit Commit) error
	FindRecentAuthors(ctx context.Context, repoID api.RepoID, path string) ([]RecentAuthor, error)
}

type Commit struct {
	RepoID       api.RepoID
	AuthorName   string
	AuthorEmail  string
	Timestamp    time.Time
	CommitSHA    string
	FilesChanged []string
}

type ownSignalStore struct {
	logger log.Logger
	*basestore.Store
}

func (s *ownSignalStore) AddCommit(ctx context.Context, commit Commit) error {
	return s.Store.WithTransact(context.Background(), func(tx *basestore.Store) error {
		// Get or create commit author
		var authorID int
		err := tx.QueryRow(
			ctx,
			sqlf.Sprintf(`SELECT id FROM commit_authors WHERE email = %s AND name = %s`, commit.AuthorEmail, commit.AuthorName),
		).Scan(&authorID)
		if err == sql.ErrNoRows {
			err = tx.QueryRow(
				ctx,
				sqlf.Sprintf(`INSERT INTO commit_authors (email, name) VALUES (%s, %s) RETURNING id`, commit.AuthorEmail, commit.AuthorName),
			).Scan(&authorID)
		}
		if err != nil {
			return err
		}

		// Get or create repo paths
		pathIDs := make([]int, len(commit.FilesChanged))
		for i, path := range commit.FilesChanged {
			pathPrefixes := strings.Split(path, "/")
			var parentPathID *int
			var pathID int
			for j := range pathPrefixes {
				pathPrefix := strings.Join(pathPrefixes[:j+1], "/")
				isDir := j < len(pathPrefixes)-1
				// Get or create repo path
				err = tx.QueryRow(
					ctx,
					sqlf.Sprintf(`
            SELECT id FROM repo_paths
            WHERE repo_id = %s
            AND absolute_path = %s
            `, commit.RepoID, pathPrefix),
				).Scan(&pathID)
				if err == sql.ErrNoRows {
					err = tx.QueryRow(
						ctx,
						sqlf.Sprintf(`
                INSERT INTO repo_paths (repo_id, absolute_path, is_dir, parent_id)
                    VALUES (%s, %s, %s, %s) RETURNING id`, commit.RepoID, pathPrefix, isDir, parentPathID),
					).Scan(&pathID)
				}
				if err != nil {
					return err
				}
				parentPathID = &pathID
			}
			pathIDs[i] = pathID
		}

		// Insert into own_signal_recent_contribution
		for _, pathID := range pathIDs {
			commitID := fnv.New32a()
			commitID.Write([]byte(commit.CommitSHA))
			q := sqlf.Sprintf(`INSERT INTO own_signal_recent_contribution (commit_author_id, changed_file_path_id,
				commit_timestamp, commit_id_hash) VALUES (%s, %s, %s, %s)`,
				authorID,
				pathID,
				commit.Timestamp,
				commitID.Sum32(),
			)
			err = tx.Exec(ctx, q)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

type RecentAuthor struct {
	AuthorName        string
	AuthorEmail       string
	ContributionCount int
}

func (s *ownSignalStore) FindRecentAuthors(ctx context.Context, repoID api.RepoID, path string) ([]RecentAuthor, error) {
	var authors []RecentAuthor
	q := sqlf.Sprintf(`
		SELECT a.name, a.email, g.contributions_count
		FROM commit_authors AS a
		INNER JOIN own_aggregate_recent_contribution AS g
		ON a.id = g.commit_author_id
		INNER JOIN repo_paths AS p
		ON p.id = g.changed_file_path_id
		WHERE p.repo_id = %s
		AND p.absolute_path = %s
		ORDER BY 3 DESC
	`, repoID, path)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a RecentAuthor
		if err := rows.Scan(&a.AuthorName, &a.AuthorEmail, &a.ContributionCount); err != nil {
			return nil, err
		}
		authors = append(authors, a)
	}
	return authors, nil
}

func OwnSignalsStoreWith(other basestore.ShareableStore) OwnSignalStore {
	return &ownSignalStore{Store: basestore.NewWithHandle(other.Handle())}
}
