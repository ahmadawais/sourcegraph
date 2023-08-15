package codehosts

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type codeHostsMigrator struct {
	logger log.Logger
	store  *basestore.Store
	key    encryption.Key
}

var _ oobmigration.Migrator = &codeHostsMigrator{}

func NewMigratorWithDB(store *basestore.Store, key encryption.Key) *codeHostsMigrator {
	return &codeHostsMigrator{
		logger: log.Scoped("codeHostsMigrator", ""),
		store:  store,
		key:    key,
	}
}

func (m *codeHostsMigrator) ID() int                 { return 24 }
func (m *codeHostsMigrator) Interval() time.Duration { return time.Second * 3 }

// Progress returns the percentage (ranged [0, 1]) of external services that were migrated to the code_hosts table.
func (m *codeHostsMigrator) Progress(ctx context.Context, _ bool) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(codeHostsMigratorProgressQuery)))
	return progress, err
}

// Note: We explicitly also migrate deleted external services here, so that we can be sure by 5.3
// that there will be no more external_services without an associated code host so we can make
// the code_host_id column non-nullable.
const codeHostsMigratorProgressQuery = `
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		CAST(c1.count AS float) / CAST(c2.count AS float)
	END
FROM
	(SELECT COUNT(*) AS count FROM external_services WHERE code_host_id IS NOT NULL) c1,
	(SELECT COUNT(*) AS count FROM external_services WHERE code_host_id IS NULL) c2
`

// Up loads a set of external services without a populated code_host_id column and
// upserts a code_hosts entry to fill the value for the host.
func (m *codeHostsMigrator) Up(ctx context.Context) (err error) {
	var parseErrs error

	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// Commit transaction with non-parse errors. If we include parse errors in
		// this set prior to the tx.Done call, then we will always rollback the tx
		// and lose progress on the batch
		err = tx.Done(err)

		// Add non-"fatal" errors for callers
		err = errors.CombineErrors(err, parseErrs)
	}()

	// First, read the currently configured value for gitMaxCodehostRequestsPerSecond from
	// the site config. This value needs to be transferred to any code host that we create.
	gitMaxCodehostRequestsPerSecond := 0
	{
		row := tx.QueryRow(ctx, sqlf.Sprintf(currentSiteConfigQuery))
		var siteConfigContents string
		if err := row.Scan(&siteConfigContents); err != nil {
			// No site config could exist, skip in this case.
			if err != sql.ErrNoRows {
				return errors.Wrap(err, "failed to read current site config")
			}
		}
		if siteConfigContents != "" {
			var cfg siteConfiguration
			if err := jsonc.Unmarshal(siteConfigContents, &cfg); err != nil {
				return errors.Wrap(err, "failed to parse current site config")
			}
			if cfg.GitMaxCodehostRequestsPerSecond != nil {
				gitMaxCodehostRequestsPerSecond = *cfg.GitMaxCodehostRequestsPerSecond
			}
		}
	}

	type svc struct {
		ID           int
		Kind, Config string
	}
	svcs, err := func() (svcs []svc, err error) {
		// First, we load ALL external_services. This should be << 50 for most instances
		// so this should not cause bigger issues.
		rows, err := tx.Query(ctx, sqlf.Sprintf(codeHostsMigratorSelectQuery))
		if err != nil {
			return nil, err
		}
		defer func() { err = basestore.CloseRows(rows, err) }()

		for rows.Next() {
			var id int
			var kind, config, keyID string
			if err := rows.Scan(&id, &kind, &config, &keyID); err != nil {
				return nil, err
			}
			config, err = encryption.MaybeDecrypt(ctx, m.key, config, keyID)
			if err != nil {
				return nil, err
			}

			svcs = append(svcs, svc{ID: id, Kind: kind, Config: config})
		}

		return svcs, nil
	}()
	if err != nil {
		return err
	}

	// Nothing more to migrate!
	if len(svcs) == 0 {
		return nil
	}

	// Look at the first unmigrated external service.
	current := svcs[0]
	currentHostURL, err := getExternalServiceURL(current.Kind, current.Config)
	if err != nil {
		return err
	}
	lowestRateLimitPerHour, err := getRateLimitPerHour(current.Kind, current.Config)
	if err != nil {
		return err
	}
	if lowestRateLimitPerHour < 0 {
		lowestRateLimitPerHour = 0
	}
	svcsWithSameHost := []int{current.ID}

	// Find all other external services for the same (kind, url).
	for _, o := range svcs[1:] {
		if o.Kind == current.Kind {
			have, err := getExternalServiceURL(o.Kind, o.Config)
			if err != nil {
				return err
			}
			// TODO: Can we do this equality check or do we need to somehow clean
			// the URLs first? (case-sensitivity, trailing `/`, and so forth..)
			if have == currentHostURL {
				svcsWithSameHost = append(svcsWithSameHost, o.ID)
				// Find the smallest configured rate limit for the given host.
				rateLimitPerHour, err := getRateLimitPerHour(current.Kind, current.Config)
				if err != nil {
					return err
				}
				if rateLimitPerHour > 0 && rateLimitPerHour < lowestRateLimitPerHour {
					lowestRateLimitPerHour = rateLimitPerHour
				}
			}
		}
	}

	var apiInterval int
	if lowestRateLimitPerHour > 0 {
		apiInterval = 60 * 60 // limits used to always be one hour.
	}

	var gitInterval int
	if gitMaxCodehostRequestsPerSecond > 0 {
		gitInterval = 1 // always per second in site-config.
	}

	row := tx.QueryRow(ctx, sqlf.Sprintf(
		upsertCodeHostQuery,
		current.Kind,
		currentHostURL,
		NewNullInt(lowestRateLimitPerHour),
		NewNullInt(apiInterval),
		NewNullInt(gitMaxCodehostRequestsPerSecond),
		NewNullInt(gitInterval),
		currentHostURL,
	))

	var codeHostID int
	if err := row.Scan(&codeHostID); err != nil {
		return errors.Wrap(err, "failed to upsert code host")
	}

	return tx.Exec(ctx, sqlf.Sprintf(setCodeHostIDOnExternalServiceQuery, codeHostID, pq.Array(svcsWithSameHost)))
}

const codeHostsMigratorSelectQuery = `
SELECT id, kind, config, encryption_key_id FROM external_services WHERE code_host_id IS NULL ORDER BY id FOR UPDATE
`

const currentSiteConfigQuery = `
SELECT contents FROM critical_and_site_config WHERE type = 'site' ORDER BY id DESC LIMIT 1
`

const upsertCodeHostQuery = `
WITH inserted AS (
	INSERT INTO
		code_hosts
			(kind, url, api_rate_limit_quota, api_rate_limit_interval_seconds, git_rate_limit_quota, git_rate_limit_interval_seconds)
	VALUES
			(%s, %s, %s, %s, %s, %s)
	ON CONFLICT (url) DO NOTHING
	RETURNING
		id
)
SELECT
	id
FROM inserted

UNION

SELECT
	id
FROM code_hosts
WHERE url = %s
`

const setCodeHostIDOnExternalServiceQuery = `
UPDATE external_services
SET
	code_host_id = %s
WHERE
	id = ANY(%s)
`

func (*codeHostsMigrator) Down(context.Context) error {
	// non-destructive
	return nil
}

type siteConfiguration struct {
	GitMaxCodehostRequestsPerSecond *int `json:"gitMaxCodehostRequestsPerSecond,omitempty"`
}

func getExternalServiceURL(kind string, config string) (string, error) {
	type genericConfig struct {
		URL string `json:"url"`
	}
	var cfg genericConfig
	if err := json.Unmarshal([]byte(config), &cfg); err != nil {
		return "", errors.Wrap(err, "failed to parse external service config")
	}
	switch kind {
	case "AWSCODECOMMIT":
		// TODO: No concept of a URL, what do we use instead? Region? Access key?
		return "awscodecommit", nil
	case "AZUREDEVOPS":
		return cfg.URL, nil
	case "BITBUCKETCLOUD":
		return cfg.URL, nil
	case "BITBUCKETSERVER":
		return cfg.URL, nil
	case "GERRIT":
		return cfg.URL, nil
	case "GITHUB":
		return cfg.URL, nil
	case "GITLAB":
		return cfg.URL, nil
	case "GITOLITE":
		u, err := url.Parse(cfg.URL)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse URL for gitolite code host type")
		}
		// Clean the URL from any secrets.
		u.Path = ""
		u.RawQuery = ""
		u.RawFragment = ""
		u.User = nil
		return u.String(), nil
	case "GOMODULES":
		// TODO: Multiple URLs exist here.
		return "gomodules", nil
	case "JVMPACKAGES":
		// TODO: Multiple URLs exist here.
		return "jvmpackages", nil
	case "NPMPACKAGES":
		type npmConfig struct {
			Registry string `json:"registry"`
		}
		var cfg npmConfig
		if err := json.Unmarshal([]byte(config), &cfg); err != nil {
			return "", errors.Wrap(err, "failed to parse external service npm")
		}
		return cfg.Registry, nil
	case "OTHER":
		u, err := url.Parse(cfg.URL)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse URL for other code host type")
		}
		// Clean the URL from any secrets.
		u.Path = ""
		u.RawQuery = ""
		u.RawFragment = ""
		u.User = nil
		return u.String(), nil
	case "PAGURE":
		return cfg.URL, nil
	case "PERFORCE":
		type perforceConfig struct {
			P4Port string `json:"p4.port"`
		}
		var cfg perforceConfig
		if err := json.Unmarshal([]byte(config), &cfg); err != nil {
			return "", errors.Wrap(err, "failed to parse external service config")
		}
		return cfg.P4Port, nil
	case "PHABRICATOR":
		return cfg.URL, nil
	case "PYTHONPACKAGES":
		// TODO: Multiple URLs exist here.
		return "pythonpackages", nil
	case "RUBYPACKAGES":
		// TODO: No URL exists here.
		return "rubypackages", nil
	case "RUSTPACKAGES":
		// TODO: No URL exists here.
		return "rustpackages", nil
	case "SCIM":
		// TODO: I think this type shouldn't exist. Asked David on Slack.
		return "scim", nil
	case "LOCALGIT":
		// TODO: Localgit doesn't have a codehost. What do we do here? Should
		// we create a fake codehost of type local that has a bogus URL and we
		// add some custom rendering to the UI for this host type?
		return "http://localhost", nil
	}
	return "", errors.Newf("invalid external service kind %q", kind)
}

func getRateLimitPerHour(kind string, config string) (int, error) {
	type genericConfig struct {
		RateLimit *struct {
			Enabled         *bool `json:"enabled,omitempty"`
			RequestsPerHour *int  `json:"requestsPerHour,omitempty"`
		} `json:"rateLimit,omitempty"`
	}
	var cfg genericConfig
	if err := json.Unmarshal([]byte(config), &cfg); err != nil {
		return 0, errors.Wrap(err, "failed to parse external service config")
	}
	switch kind {
	case "AWSCODECOMMIT":
		// CodeCommit doesn't have a rate limit property.
		return 0, nil
	case "AZUREDEVOPS":
		// Azure DevOps doesn't have a rate limit property.
		return 0, nil
	case "BITBUCKETCLOUD":
		// The default for enabled is true, so either unset or true will be acceptable.
		if cfg.RateLimit != nil && (cfg.RateLimit.Enabled == nil || *cfg.RateLimit.Enabled) && cfg.RateLimit.RequestsPerHour != nil {
			return *cfg.RateLimit.RequestsPerHour, nil
		}
		return 0, nil
	case "BITBUCKETSERVER":
		// The default for enabled is true, so either unset or true will be acceptable.
		if cfg.RateLimit != nil && (cfg.RateLimit.Enabled == nil || *cfg.RateLimit.Enabled) && cfg.RateLimit.RequestsPerHour != nil {
			return *cfg.RateLimit.RequestsPerHour, nil
		}
		return 0, nil
	case "GERRIT":
		// Gerrit doesn't have a rate limit property.
		return 0, nil
	case "GITHUB":
		// The default for enabled is true, so either unset or true will be acceptable.
		if cfg.RateLimit != nil && (cfg.RateLimit.Enabled == nil || *cfg.RateLimit.Enabled) && cfg.RateLimit.RequestsPerHour != nil {
			return *cfg.RateLimit.RequestsPerHour, nil
		}
		return 0, nil
	case "GITLAB":
		// The default for enabled is true, so either unset or true will be acceptable.
		if cfg.RateLimit != nil && (cfg.RateLimit.Enabled == nil || *cfg.RateLimit.Enabled) && cfg.RateLimit.RequestsPerHour != nil {
			return *cfg.RateLimit.RequestsPerHour, nil
		}
		return 0, nil
	case "GITOLITE":
		// Gitolite doesn't have a rate limit property.
		return 0, nil
	case "GOMODULES":
		// The default for enabled is true, so either unset or true will be acceptable.
		if cfg.RateLimit != nil && (cfg.RateLimit.Enabled == nil || *cfg.RateLimit.Enabled) && cfg.RateLimit.RequestsPerHour != nil {
			return *cfg.RateLimit.RequestsPerHour, nil
		}
		return 0, nil
	case "JVMPACKAGES":
		type jvmConfig struct {
			Maven struct {
				RateLimit *struct {
					Enabled         *bool `json:"enabled,omitempty"`
					RequestsPerHour *int  `json:"requestsPerHour,omitempty"`
				} `json:"rateLimit,omitempty"`
			} `json:"maven"`
		}
		var cfg jvmConfig
		if err := json.Unmarshal([]byte(config), &cfg); err != nil {
			return 0, errors.Wrap(err, "failed to parse JVM external service config")
		}
		// The default for enabled is true, so either unset or true will be acceptable.
		if cfg.Maven.RateLimit != nil && (cfg.Maven.RateLimit.Enabled == nil || *cfg.Maven.RateLimit.Enabled) && cfg.Maven.RateLimit.RequestsPerHour != nil {
			return *cfg.Maven.RateLimit.RequestsPerHour, nil
		}
		return 0, nil
	case "NPMPACKAGES":
		// The default for enabled is true, so either unset or true will be acceptable.
		if cfg.RateLimit != nil && (cfg.RateLimit.Enabled == nil || *cfg.RateLimit.Enabled) && cfg.RateLimit.RequestsPerHour != nil {
			return *cfg.RateLimit.RequestsPerHour, nil
		}
		return 0, nil
	case "OTHER":
		// Other doesn't have a rate limit property.
		return 0, nil
	case "PAGURE":
		// The default for enabled is true, so either unset or true will be acceptable.
		if cfg.RateLimit != nil && (cfg.RateLimit.Enabled == nil || *cfg.RateLimit.Enabled) && cfg.RateLimit.RequestsPerHour != nil {
			return *cfg.RateLimit.RequestsPerHour, nil
		}
		return 0, nil
	case "PERFORCE":
		// The default for enabled is true, so either unset or true will be acceptable.
		if cfg.RateLimit != nil && (cfg.RateLimit.Enabled == nil || *cfg.RateLimit.Enabled) && cfg.RateLimit.RequestsPerHour != nil {
			return *cfg.RateLimit.RequestsPerHour, nil
		}
		return 0, nil
	case "PHABRICATOR":
		// Phabricator doesn't have a rate limit property.
		return 0, nil
	case "PYTHONPACKAGES":
		// The default for enabled is true, so either unset or true will be acceptable.
		if cfg.RateLimit != nil && (cfg.RateLimit.Enabled == nil || *cfg.RateLimit.Enabled) && cfg.RateLimit.RequestsPerHour != nil {
			return *cfg.RateLimit.RequestsPerHour, nil
		}
		return 0, nil
	case "RUBYPACKAGES":
		// The default for enabled is true, so either unset or true will be acceptable.
		if cfg.RateLimit != nil && (cfg.RateLimit.Enabled == nil || *cfg.RateLimit.Enabled) && cfg.RateLimit.RequestsPerHour != nil {
			return *cfg.RateLimit.RequestsPerHour, nil
		}
		return 0, nil
	case "RUSTPACKAGES":
		// The default for enabled is true, so either unset or true will be acceptable.
		if cfg.RateLimit != nil && (cfg.RateLimit.Enabled == nil || *cfg.RateLimit.Enabled) && cfg.RateLimit.RequestsPerHour != nil {
			return *cfg.RateLimit.RequestsPerHour, nil
		}
		return 0, nil
	case "SCIM":
	case "LOCALGIT":
		// TODO: What do we do here? Likely nothing?
		return 0, nil
	}
	return 0, errors.Newf("invalid external service kind %q", kind)
}
