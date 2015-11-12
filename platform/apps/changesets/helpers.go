package changesets

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"google.golang.org/grpc"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	approuter "src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/platform"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/platform/putil"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

// GetRepoAndRevCommon retrieves common information about the repository, its
// revision and build status.
func GetRepoAndRevCommon(r *http.Request) (rc *handlerutil.RepoCommon, vc *handlerutil.RepoRevCommon, err error) {
	ctx := putil.Context(r)
	sg := sourcegraph.NewClientFromContext(ctx)

	rc = new(handlerutil.RepoCommon)
	rrs, ok := pctx.RepoRevSpec(ctx)
	if !ok {
		return nil, nil, errors.New("no repo found in context")
	}
	origSpec := rrs.RepoSpec
	rc.Repo, err = sg.Repos.Get(ctx, &origSpec)
	if err != nil {
		return nil, nil, err
	}
	spec := rc.Repo.RepoSpec()
	if origSpec.URI != "" && origSpec.URI != spec.URI {
		return nil, nil, &handlerutil.URLMovedError{spec.URI}
	}
	rc.RepoConfig, err = sg.Repos.GetConfig(ctx, &spec)
	if err != nil {
		return nil, nil, err
	}

	commit, err := sg.Repos.GetCommit(ctx, &rrs)
	if err != nil {
		return nil, nil, err
	}
	rrs.CommitID = string(commit.ID)
	if rrs.Rev == "" {
		rrs.Rev = rc.Repo.DefaultBranch
	}
	vc = &handlerutil.RepoRevCommon{RepoRevSpec: rrs}
	vc.RepoCommit, err = handlerutil.AugmentCommit(r, spec.URI, commit)
	if err != nil {
		return nil, nil, err
	}

	return
}

// writeJSON writes JSON to the given http.ResponseWriter.
func writeJSON(w http.ResponseWriter, v interface{}) error {
	w.Header().Set(platform.HTTPHeaderVerbatim, "true")
	w.Header().Set("Content-Type", "application/json")
	if err, ok := v.(error); ok {
		w.WriteHeader(errcode.HTTP(err))
		v = struct{ Error string }{Error: grpc.ErrorDesc(err)}
	}
	return json.NewEncoder(w).Encode(v)
}

// urlToRepoChangeset returns the relative URL of the changeset with given id.
func urlToRepoChangeset(repo string, changeset int64) (*url.URL, error) {
	subURL, err := router.Get(routeView).URL("ID", fmt.Sprint(changeset))
	if err != nil {
		return nil, err
	}
	return approuter.Rel.URLToOrError(approuter.RepoAppFrame, "Repo", repo, "App", appID, "AppPath", subURL.Path)
}
