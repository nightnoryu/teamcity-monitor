// Package teamcity is a small client for the parts of the TeamCity REST API
// needed to read the latest build of a build configuration.
package teamcity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-faster/errors"
)

const (
	// maxResponseBytes bounds how much of a TeamCity response body is read,
	// including error bodies, to avoid unbounded memory use on a
	// misbehaving or malicious server.
	maxResponseBytes = 64 * 1024

	// dateLayout is TeamCity's build startDate/finishDate format.
	dateLayout = "20060102T150405-0700"

	// state:any includes queued and running builds, not just finished ones,
	// so the most recent build is reported even while it's still in progress.
	// revisions(revision(vcsBranchName)) is a fallback for branchName: a
	// build re-run on an existing revision (no new VCS changes) can have an
	// empty branchName while its checked-out revision still names a branch.
	// snapshot-dependencies(...) is a second fallback for deploy-only build
	// configurations with no VCS checkout of their own (triggered purely via
	// a snapshot dependency on a build that does have one).
	buildFields = "build(number,status,statusText,state,canceled,failedToStart,branchName,startDate,finishDate,webUrl," +
		"triggered(type,user(username,name)),revisions(revision(vcsBranchName))," +
		"snapshot-dependencies(build(branchName,revisions(revision(vcsBranchName)))))"

	// refsHeadsPrefix is stripped from a raw VCS ref for display, e.g.
	// "refs/heads/main" -> "main".
	refsHeadsPrefix = "refs/heads/"

	auditFields = "auditEvent(comment,user(username,name))"

	// auditScanLimit bounds how many recent audit events for a project are
	// scanned for the specific "Value of the parameter X changed" comment.
	// TeamCity's audit API has no server-side filter for comment text or
	// parameter name, so this trades off "how far back we look" against
	// request size; audit events are returned newest first.
	auditScanLimit = 100

	valueChangedCommentPrefix = "Value of the parameter "
	valueChangedCommentSuffix = " changed"
)

// Status is a TeamCity build status. Only meaningful once the build has
// finished; use State to detect a build still in progress.
type Status string

// Build statuses as reported by TeamCity.
const (
	StatusSuccess Status = "SUCCESS"
	StatusFailure Status = "FAILURE"
	StatusError   Status = "ERROR"
)

// State is a TeamCity build's lifecycle state.
type State string

// Build states as reported by TeamCity.
const (
	StateQueued   State = "queued"
	StateRunning  State = "running"
	StateFinished State = "finished"
)

// Sentinel errors returned by Client methods, checkable with errors.Is.
var (
	// ErrNoBuilds means the build configuration has never run — a valid
	// steady state, not a transport or server failure.
	ErrNoBuilds     = errors.New("no builds found for build configuration")
	ErrUnauthorized = errors.New("teamcity: unauthorized")
	ErrNotFound     = errors.New("teamcity: build configuration not found")
	// ErrNoAuditRecord means no matching parameter-change event was found
	// within the scanned audit history — a valid steady state (parameter
	// was never manually changed, or the change predates what was scanned).
	ErrNoAuditRecord = errors.New("no matching audit record found")
)

// Build is the subset of a TeamCity build's data the dashboard needs.
type Build struct {
	Number        string
	Status        Status
	StatusText    string
	State         State
	Canceled      bool
	FailedToStart bool
	// Branch is the actual VCS branch the build ran on, as reported by
	// TeamCity, not a build parameter value.
	Branch      string
	StartedAt   time.Time
	FinishedAt  time.Time
	TriggeredBy string
	WebURL      string
}

// Client is a TeamCity REST API client authenticating with a bearer access
// token.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewClient builds a Client. httpClient must not be nil; callers are
// expected to set a request timeout on it.
func NewClient(baseURL, token string, httpClient *http.Client) *Client {
	return &Client{baseURL: baseURL, token: token, http: httpClient}
}

func (c *Client) endpoint(resource string, query url.Values) string {
	u, _ := url.Parse(c.baseURL) // configuration validation ensures a valid base URL
	u.Path = path.Join(u.Path, resource)
	if strings.HasSuffix(resource, "/") {
		u.Path += "/"
	}
	u.RawQuery = query.Encode()
	return u.String()
}

// LatestBuild fetches the most recent build of the given build type,
// including one still queued or running.
func (c *Client) LatestBuild(ctx context.Context, buildTypeID string) (Build, error) {
	// Personal builds are private runs; canceled and failed-to-start builds
	// are attempts operators need to see across all branches.
	endpoint := c.endpoint("app/rest/buildTypes/id:"+buildTypeID+"/builds/", url.Values{
		"locator": {"count:1,state:any,defaultFilter:false,branch:default:any,personal:false,canceled:any,failedToStart:any"},
		"fields":  {buildFields},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return Build{}, errors.Wrap(err, "build request")
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Build{}, errors.Wrap(err, "do request")
	}
	defer func() { _ = resp.Body.Close() }()

	if err := statusErr(resp.StatusCode); err != nil {
		return Build{}, err
	}

	var payload buildsResponse
	if err := json.NewDecoder(http.MaxBytesReader(nil, resp.Body, maxResponseBytes)).Decode(&payload); err != nil {
		return Build{}, errors.Wrap(err, "decode response")
	}

	if len(payload.Build) == 0 {
		return Build{}, ErrNoBuilds
	}

	return payload.Build[0].toBuild()
}

// LastParameterChangeAuthor returns the username of whoever most recently
// changed paramName on the given TeamCity project, per TeamCity's audit
// log. TeamCity's audit API doesn't record which parameter changed or its
// old/new value — only a human-readable comment — so this scans recent
// "project settings edited" events for the exact comment TeamCity generates
// for a single-parameter change: "Value of the parameter X changed".
func (c *Client) LastParameterChangeAuthor(ctx context.Context, projectID, paramName string) (string, error) {
	authors, err := c.LastParameterChangeAuthors(ctx, projectID, []string{paramName})
	if err != nil {
		return "", err
	}
	if author, ok := authors[paramName]; ok {
		return author, nil
	}
	return "", ErrNoAuditRecord
}

// LastParameterChangeAuthors scans one bounded audit page for all requested
// project parameters. Missing keys have no exact matching event in that page.
func (c *Client) LastParameterChangeAuthors(ctx context.Context, projectID string, paramNames []string) (map[string]string, error) {
	endpoint := c.endpoint("app/rest/audit", url.Values{
		"locator": {"affectedProject:(id:" + projectID + "),action:project_edit_settings,count:" + strconv.Itoa(auditScanLimit)},
		"fields":  {auditFields},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "audit request")
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}
	defer func() { _ = resp.Body.Close() }()

	if err := statusErr(resp.StatusCode); err != nil {
		return nil, err
	}

	var payload auditResponse
	if err := json.NewDecoder(http.MaxBytesReader(nil, resp.Body, maxResponseBytes)).Decode(&payload); err != nil {
		return nil, errors.Wrap(err, "decode response")
	}

	wanted := make(map[string]string, len(paramNames))
	for _, name := range paramNames {
		wanted[valueChangedCommentPrefix+name+valueChangedCommentSuffix] = name
	}
	authors := make(map[string]string, len(paramNames))
	for _, event := range payload.AuditEvent {
		if name, ok := wanted[event.Comment]; ok {
			if _, found := authors[name]; !found {
				authors[name] = event.User.displayName()
			}
		}
	}
	return authors, nil
}

type auditResponse struct {
	AuditEvent []auditEventDTO `json:"auditEvent"`
}

type auditEventDTO struct {
	Comment string  `json:"comment"`
	User    userDTO `json:"user"`
}

func statusErr(code int) error {
	switch {
	case code == http.StatusUnauthorized || code == http.StatusForbidden:
		return ErrUnauthorized
	case code == http.StatusNotFound:
		return ErrNotFound
	case code >= http.StatusBadRequest:
		return errors.Errorf("teamcity: unexpected status %d", code)
	default:
		return nil
	}
}

type buildsResponse struct {
	Build []buildDTO `json:"build"`
}

type buildDTO struct {
	Number               string                  `json:"number"`
	Status               Status                  `json:"status"`
	StatusText           string                  `json:"statusText"`
	State                State                   `json:"state"`
	Canceled             bool                    `json:"canceled"`
	FailedToStart        bool                    `json:"failedToStart"`
	BranchName           string                  `json:"branchName"`
	StartDate            string                  `json:"startDate"`
	FinishDate           string                  `json:"finishDate"`
	WebURL               string                  `json:"webUrl"`
	Triggered            triggeredDTO            `json:"triggered"`
	Revisions            revisionsDTO            `json:"revisions"`
	SnapshotDependencies snapshotDependenciesDTO `json:"snapshot-dependencies"`
}

type snapshotDependenciesDTO struct {
	Build []buildDTO `json:"build"`
}

type revisionsDTO struct {
	Revision []revisionDTO `json:"revision"`
}

type revisionDTO struct {
	VcsBranchName string `json:"vcsBranchName"`
}

func (r revisionsDTO) branchName() string {
	for _, rev := range r.Revision {
		if rev.VcsBranchName != "" {
			return rev.VcsBranchName
		}
	}
	return ""
}

type triggeredDTO struct {
	Type string  `json:"type"`
	User userDTO `json:"user"`
}

type userDTO struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

func (u userDTO) displayName() string {
	if u.Username != "" {
		return u.Username
	}
	return u.Name
}

func (b buildDTO) toBuild() (Build, error) {
	startedAt, err := parseOptionalDate(b.StartDate)
	if err != nil {
		return Build{}, errors.Wrap(err, "parse startDate")
	}

	finishedAt, err := parseOptionalDate(b.FinishDate)
	if err != nil {
		return Build{}, errors.Wrap(err, "parse finishDate")
	}

	return Build{
		Number:        b.Number,
		Status:        b.Status,
		StatusText:    b.StatusText,
		State:         b.State,
		Canceled:      b.Canceled,
		FailedToStart: b.FailedToStart,
		Branch:        normalizeBranch(b.branch()),
		StartedAt:     startedAt,
		FinishedAt:    finishedAt,
		TriggeredBy:   b.Triggered.triggeredBy(),
		WebURL:        b.WebURL,
	}, nil
}

// branch prefers the build's own branchName, falling back to its checked
// out revision's branch, then to the branch of the build it snapshot-depends
// on. The first fallback covers a build re-run on an existing revision
// (TeamCity reports "no changes" for it, but the revision still names a
// branch); the second covers a deploy-only build configuration with no VCS
// checkout of its own, triggered purely via a snapshot dependency.
func (b buildDTO) branch() string {
	if b.BranchName != "" {
		return b.BranchName
	}
	if branch := b.Revisions.branchName(); branch != "" {
		return branch
	}
	for _, dep := range b.SnapshotDependencies.Build {
		if branch := dep.branch(); branch != "" {
			return branch
		}
	}
	return ""
}

// normalizeBranch strips TeamCity's "refs/heads/" prefix for display, e.g.
// "refs/heads/main" -> "main". Other ref namespaces (tags, pull requests)
// are left as-is.
func normalizeBranch(branch string) string {
	return strings.TrimPrefix(branch, refsHeadsPrefix)
}

// parseOptionalDate parses a TeamCity date field that may be empty, as is
// the case for startDate/finishDate on a queued or still-running build.
func parseOptionalDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(dateLayout, value)
}

func (t triggeredDTO) triggeredBy() string {
	if name := t.User.displayName(); name != "" {
		return name
	}
	return t.Type
}
