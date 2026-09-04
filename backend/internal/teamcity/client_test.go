package teamcity_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"teamcity-monitor/internal/teamcity"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *teamcity.Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return teamcity.NewClient(server.URL, "test-token", server.Client())
}

func TestLatestBuild_Success(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"count": 1,
			"build": [{
				"number": "42",
				"status": "SUCCESS",
				"statusText": "Success",
				"state": "finished",
				"branchName": "refs/heads/feature/order-export",
				"startDate": "20260902T110000+0300",
				"finishDate": "20260902T112000+0300",
				"webUrl": "https://teamcity/build/1",
				"triggered": {"type": "user", "user": {"username": "a.kovalev", "name": "Alex Kovalev"}}
			}]
		}`))
	})

	build, err := client.LatestBuild(t.Context(), "Alpha_Testing_Dev_Ru")
	require.NoError(t, err)

	require.Equal(t, "42", build.Number)
	require.Equal(t, teamcity.StatusSuccess, build.Status)
	require.Equal(t, teamcity.StateFinished, build.State)
	require.Equal(t, "a.kovalev", build.TriggeredBy)
	require.Equal(t, "feature/order-export", build.Branch, "refs/heads/ prefix stripped")

	wantStart := time.Date(2026, 9, 2, 11, 0, 0, 0, time.FixedZone("", 3*60*60))
	require.True(t, build.StartedAt.Equal(wantStart))
}

func TestLatestBuild_RunningBuildHasNoFinishDate(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"count": 1,
			"build": [{
				"number": "43", "status": "SUCCESS", "state": "running",
				"branchName": "feature/order-export",
				"startDate": "20260902T110000+0300", "finishDate": "",
				"triggered": {"type": "vcs"}
			}]
		}`))
	})

	build, err := client.LatestBuild(t.Context(), "id")
	require.NoError(t, err)
	require.Equal(t, teamcity.StateRunning, build.State)
	require.True(t, build.FinishedAt.IsZero())
	require.False(t, build.StartedAt.IsZero())
}

func TestLatestBuild_QueuedBuildHasNoDates(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"count": 1,
			"build": [{"number": "44", "state": "queued", "startDate": "", "finishDate": ""}]
		}`))
	})

	build, err := client.LatestBuild(t.Context(), "id")
	require.NoError(t, err)
	require.Equal(t, teamcity.StateQueued, build.State)
	require.True(t, build.StartedAt.IsZero())
	require.True(t, build.FinishedAt.IsZero())
}

func TestLatestBuild_MissingBranchNameAndRevisions(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"count": 1,
			"build": [{
				"number": "1", "status": "FAILURE", "statusText": "Failure", "state": "finished",
				"startDate": "20260902T110000+0300", "finishDate": "20260902T112000+0300",
				"triggered": {"type": "vcs"}
			}]
		}`))
	})

	build, err := client.LatestBuild(t.Context(), "id")
	require.NoError(t, err)
	require.Empty(t, build.Branch)
	require.Equal(t, "vcs", build.TriggeredBy)
	require.Equal(t, teamcity.StatusFailure, build.Status)
}

// A build re-run on an existing revision (TeamCity's "no changes" case) can
// report an empty top-level branchName; the revision's own branch is used.
func TestLatestBuild_FallsBackToRevisionBranchWhenBranchNameEmpty(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"count": 1,
			"build": [{
				"number": "57", "status": "SUCCESS", "state": "finished",
				"branchName": "",
				"startDate": "20260902T110000+0300", "finishDate": "20260902T112000+0300",
				"revisions": {"revision": [{"vcsBranchName": "refs/heads/master"}]}
			}]
		}`))
	})

	build, err := client.LatestBuild(t.Context(), "id")
	require.NoError(t, err)
	require.Equal(t, "master", build.Branch)
}

// A deploy-only build configuration with no VCS checkout of its own
// (triggered purely via a snapshot dependency) has no revisions either;
// the branch of the build it depends on is used instead.
func TestLatestBuild_FallsBackToSnapshotDependencyBranch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"count": 1,
			"build": [{
				"number": "56", "status": "SUCCESS", "state": "finished",
				"branchName": "", "revisions": {"revision": []},
				"snapshot-dependencies": {"build": [
					{"branchName": "", "revisions": {"revision": [{"vcsBranchName": "refs/heads/master"}]}}
				]}
			}]
		}`))
	})

	build, err := client.LatestBuild(t.Context(), "id")
	require.NoError(t, err)
	require.Equal(t, "master", build.Branch)
}

func TestLatestBuild_NoBuilds(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"count": 0, "build": []}`))
	})

	_, err := client.LatestBuild(t.Context(), "id")
	require.ErrorIs(t, err, teamcity.ErrNoBuilds)
}

func TestLatestBuild_MalformedDate(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"count": 1, "build": [{"number": "1", "status": "SUCCESS", "startDate": "not-a-date"}]}`))
	})

	_, err := client.LatestBuild(t.Context(), "id")
	require.Error(t, err)
}

func TestLatestBuild_Unauthorized(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, err := client.LatestBuild(t.Context(), "id")
	require.ErrorIs(t, err, teamcity.ErrUnauthorized)
}

func TestLatestBuild_NotFound(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := client.LatestBuild(t.Context(), "id")
	require.ErrorIs(t, err, teamcity.ErrNotFound)
}

func TestLatestBuild_ServerError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.LatestBuild(t.Context(), "id")
	require.Error(t, err)
}

func TestLastParameterChangeAuthor_Found(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "affectedProject:(id:Alpha_Testing)")
		assert.Contains(t, r.URL.RawQuery, "action:project_edit_settings")
		_, _ = w.Write([]byte(`{
			"auditEvent": [
				{"comment": "Subprojects order changed", "user": {"username": "someone"}},
				{"comment": "Value of the parameter alpha_dev_branch changed", "user": {"username": "a.kovalev", "name": "Alex Kovalev"}},
				{"comment": "Value of the parameter alpha_orange_branch changed", "user": {"username": "d.rybakov"}}
			]
		}`))
	})

	author, err := client.LastParameterChangeAuthor(t.Context(), "Alpha_Testing", "alpha_dev_branch")
	require.NoError(t, err)
	require.Equal(t, "a.kovalev", author)
}

func TestLastParameterChangeAuthor_FallsBackToNameWhenUsernameEmpty(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"auditEvent": [
				{"comment": "Value of the parameter alpha_dev_branch changed", "user": {"name": "Alex Kovalev"}}
			]
		}`))
	})

	author, err := client.LastParameterChangeAuthor(t.Context(), "Alpha_Testing", "alpha_dev_branch")
	require.NoError(t, err)
	require.Equal(t, "Alex Kovalev", author)
}

func TestLastParameterChangeAuthor_NoMatch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"auditEvent": [
				{"comment": "project settings were updated", "user": {"username": "someone"}}
			]
		}`))
	})

	_, err := client.LastParameterChangeAuthor(t.Context(), "Alpha_Testing", "alpha_dev_branch")
	require.ErrorIs(t, err, teamcity.ErrNoAuditRecord)
}

func TestLastParameterChangeAuthor_Unauthorized(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, err := client.LastParameterChangeAuthor(t.Context(), "Alpha_Testing", "alpha_dev_branch")
	require.ErrorIs(t, err, teamcity.ErrUnauthorized)
}
