package graphql

import (
	"context"

	"github.com/cockroachdb/errors"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

// DefaultReferencesPageSize is the reference result page size when no limit is supplied.
const DefaultReferencesPageSize = 100

// DefaultDiagnosticsPageSize is the diagnostic result page size when no limit is supplied.
const DefaultDiagnosticsPageSize = 100

// ErrIllegalLimit occurs when the user requests less than one object per page.
var ErrIllegalLimit = errors.New("illegal limit")

// ErrIllegalBounds occurs when a negative or zero-width bound is supplied by the user.
var ErrIllegalBounds = errors.New("illegal bounds")

// QueryResolver is the main interface to bundle-related operations exposed to the GraphQL API. This
// resolver concerns itself with GraphQL/API-specific behaviors (auth, validation, marshaling, etc.).
// All code intel-specific behavior is delegated to the underlying resolver instance, which is defined
// in the parent package.
type QueryResolver struct {
	resolver         resolvers.QueryResolver
	locationResolver *CachedLocationResolver
}

// NewQueryResolver creates a new QueryResolver with the given resolver that defines all code intel-specific
// behavior. A cached location resolver instance is also given to the query resolver, which should be used
// to resolve all location-related values.
func NewQueryResolver(resolver resolvers.QueryResolver, locationResolver *CachedLocationResolver) gql.GitBlobLSIFDataResolver {
	return &QueryResolver{
		resolver:         resolver,
		locationResolver: locationResolver,
	}
}

func (r *QueryResolver) ToGitTreeLSIFData() (gql.GitTreeLSIFDataResolver, bool) { return r, true }
func (r *QueryResolver) ToGitBlobLSIFData() (gql.GitBlobLSIFDataResolver, bool) { return r, true }

func (r *QueryResolver) Ranges(ctx context.Context, args *gql.LSIFRangesArgs) (gql.CodeIntelligenceRangeConnectionResolver, error) {
	if args.StartLine < 0 || args.EndLine < args.StartLine {
		return nil, ErrIllegalBounds
	}

	ranges, err := r.resolver.Ranges(ctx, int(args.StartLine), int(args.EndLine))
	if err != nil {
		return nil, err
	}

	return &CodeIntelligenceRangeConnectionResolver{
		ranges:           ranges,
		locationResolver: r.locationResolver,
	}, nil
}

func (r *QueryResolver) Definitions(ctx context.Context, args *gql.LSIFQueryPositionArgs) (gql.LocationConnectionResolver, error) {
	locations, err := r.resolver.Definitions(ctx, int(args.Line), int(args.Character))
	if err != nil {
		return nil, err
	}

	return NewLocationConnectionResolver(locations, nil, r.locationResolver), nil
}

func (r *QueryResolver) references(ctx context.Context, args *gql.LSIFPagedQueryPositionArgs) (_ []resolvers.AdjustedLocation, cursor string, _ error) {
	limit := derefInt32(args.First, DefaultReferencesPageSize)
	if limit <= 0 {
		return nil, "", ErrIllegalLimit
	}
	cursor, err := decodeCursor(args.After)
	if err != nil {
		return nil, "", err
	}

	return r.resolver.References(ctx, int(args.Line), int(args.Character), limit, cursor)
}

func (r *QueryResolver) References(ctx context.Context, args *gql.LSIFPagedQueryPositionArgs) (gql.LocationConnectionResolver, error) {
	locations, cursor, err := r.references(ctx, args)
	if err != nil {
		return nil, err
	}
	return NewLocationConnectionResolver(locations, strPtr(cursor), r.locationResolver), nil
}

func (r *QueryResolver) Hover(ctx context.Context, args *gql.LSIFQueryPositionArgs) (gql.HoverResolver, error) {
	text, rx, exists, err := r.resolver.Hover(ctx, int(args.Line), int(args.Character))
	if err != nil || !exists {
		return nil, err
	}

	return NewHoverResolver(text, ConvertRange(rx)), nil
}

func (r *QueryResolver) Diagnostics(ctx context.Context, args *gql.LSIFDiagnosticsArgs) (gql.DiagnosticConnectionResolver, error) {
	limit := derefInt32(args.First, DefaultDiagnosticsPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	diagnostics, totalCount, err := r.resolver.Diagnostics(ctx, limit)
	if err != nil {
		return nil, err
	}

	return NewDiagnosticConnectionResolver(diagnostics, totalCount, r.locationResolver), nil
}

func (r *QueryResolver) Monikers(ctx context.Context, args *gql.LSIFQueryPositionArgs) ([]gql.MonikerResolver, error) {
	monikers, err := r.resolver.MonikersAtPosition(ctx, int(args.Line), int(args.Character))
	if err != nil {
		return nil, err
	}

	monikerResolvers := make([]gql.MonikerResolver, len(monikers))
	for i, m := range monikers {
		monikerResolvers[i] = NewMonikerResolver(m.MonikerData)
	}
	return monikerResolvers, nil
}

func (r *QueryResolver) ExploreUsageURL(ctx context.Context, args *gql.LSIFQueryPositionArgs) (*string, error) {
	monikers, err := r.resolver.MonikersAtPosition(ctx, int(args.Line), int(args.Character))
	if err != nil {
		return nil, err
	}

	if len(monikers) == 0 {
		return nil, nil
	}

	moniker := monikers[0]
	repo, err := r.locationResolver.Repository(ctx, api.RepoID(moniker.Dump.RepositoryID))
	if err != nil {
		return nil, err
	}
	// TODO(sqs): move this URL generation to the Guide package
	u := repo.URL() + "@" + moniker.Dump.Commit + "/-/usage/symbol/" + moniker.Scheme + "/" + moniker.Identifier
	return &u, nil
}
