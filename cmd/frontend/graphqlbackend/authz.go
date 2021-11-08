package graphqlbackend

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type AuthzResolver interface {
	// Mutations
	SetRepositoryPermissionsForUsers(ctx context.Context, args *RepoPermsArgs) (*EmptyResponse, error)
	ScheduleRepositoryPermissionsSync(ctx context.Context, args *RepositoryIDArgs) (*EmptyResponse, error)
	ScheduleUserPermissionsSync(ctx context.Context, args *UserPermissionsSyncArgs) (*EmptyResponse, error)

	// Queries
	AuthorizedUserRepositories(ctx context.Context, args *AuthorizedRepoArgs) (RepositoryConnectionResolver, error)
	UsersWithPendingPermissions(ctx context.Context) ([]string, error)
	AuthorizedUsers(ctx context.Context, args *RepoAuthorizedUserArgs) (UserConnectionResolver, error)

	// Helpers
	RepositoryPermissionsInfo(ctx context.Context, repoID graphql.ID) (PermissionsInfoResolver, error)
	UserPermissionsInfo(ctx context.Context, userID graphql.ID) (PermissionsInfoResolver, error)
}

type RepositoryIDArgs struct {
	Repository graphql.ID
}

type UserPermissionsSyncArgs struct {
	User    graphql.ID
	Options *struct {
		InvalidateCaches *bool
	}
}

type RepoPermsArgs struct {
	Repository      graphql.ID
	UserPermissions []struct {
		BindID     string
		Permission string
	}
}

type AuthorizedRepoArgs struct {
	Username *string
	Email    *string
	Perm     string
	First    int32
	After    *string
}

type PermissionsInfoResolver interface {
	Permissions() []string
	SyncedAt() *DateTime
	UpdatedAt() DateTime
}

// TODO: Remove usage of nolint

var (
	// subRepoPermsInstance should be initialized and used only via SubRepoPerms().
	//nolint:unused
	subRepoPermsInstance authz.SubRepoPermissionChecker
	//nolint:unused
	subRepoPermsOnce sync.Once
)

// subRepoPermsClient returns a global instance of the SubRepoPermissionsChecker for use in
// graphqlbackend only.
//nolint:unused
func subRepoPermsClient(db database.DB) authz.SubRepoPermissionChecker {
	subRepoPermsOnce.Do(func() {
		subRepoPermsInstance = authz.NewSubRepoPermsClient(database.SubRepoPerms(db))
	})
	return subRepoPermsInstance
}
