// Package pr resolves a pull-request URL to the Grove repo and branch needed
// to dispatch an agent against it.
package pr

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Info is what dispatch needs from a pull request.
type Info struct {
	Provider string // "github"
	Owner    string
	Repo     string // repository name on the provider
	Number   int
	URL      string
	Branch   string // head ref name
	Title    string
	Body     string
}

// OwnerRepo returns "owner/repo".
func (i Info) OwnerRepo() string { return i.Owner + "/" + i.Repo }

// Parse extracts provider, owner, repo, and number from a GitHub PR URL.
func Parse(raw string) (Info, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" {
		return Info{}, fmt.Errorf("invalid pull request URL %q", raw)
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	host := strings.ToLower(parsed.Hostname())
	if host != "github.com" || len(parts) < 4 || parts[2] != "pull" {
		return Info{}, fmt.Errorf("unsupported pull request URL %q (expected https://github.com/owner/repo/pull/N)", raw)
	}
	number, err := strconv.Atoi(parts[3])
	if err != nil || number <= 0 {
		return Info{}, fmt.Errorf("invalid pull request number in %q", raw)
	}
	return Info{
		Provider: "github",
		Owner:    parts[0],
		Repo:     strings.TrimSuffix(parts[1], ".git"),
		Number:   number,
		URL:      fmt.Sprintf("https://github.com/%s/%s/pull/%d", parts[0], strings.TrimSuffix(parts[1], ".git"), number),
	}, nil
}

// Outputter runs a command and returns its stdout.
type Outputter interface {
	Output(name string, args []string) ([]byte, error)
}

// Fetch fills Branch, Title, and Body using the gh CLI.
func Fetch(info Info, run Outputter) (Info, error) {
	out, err := run.Output("gh", []string{"pr", "view", info.URL, "--json", "headRefName,title,body,isCrossRepository"})
	if err != nil {
		return Info{}, fmt.Errorf("gh pr view %s: %w", info.URL, err)
	}
	var payload struct {
		HeadRefName       string `json:"headRefName"`
		Title             string `json:"title"`
		Body              string `json:"body"`
		IsCrossRepository bool   `json:"isCrossRepository"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		return Info{}, fmt.Errorf("parse gh pr view output: %w", err)
	}
	if payload.HeadRefName == "" {
		return Info{}, fmt.Errorf("gh pr view %s returned no head branch", info.URL)
	}
	if payload.IsCrossRepository {
		return Info{}, fmt.Errorf("pull request %s comes from a fork; its head branch is not on origin, so gw create --track cannot check it out", info.URL)
	}
	info.Branch = payload.HeadRefName
	info.Title = payload.Title
	info.Body = payload.Body
	return info, nil
}

// MatchRepo finds the Grove repo name whose remote is owner/repo, using the
// JSON emitted by `gw repos --json`.
func MatchRepo(reposJSON []byte, ownerRepo string) (string, error) {
	var entries []struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Remote      string `json:"remote"`
	}
	if err := json.Unmarshal(reposJSON, &entries); err != nil {
		return "", fmt.Errorf("parse gw repos output: %w", err)
	}
	want := strings.ToLower(ownerRepo)
	var matches []string
	for _, entry := range entries {
		if strings.ToLower(entry.DisplayName) == want || remoteOwnerRepo(entry.Remote) == want {
			matches = append(matches, entry.Name)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no configured Grove repo has remote %s; clone it into a repo_dir or run gw add-dir", ownerRepo)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("multiple Grove repos match %s: %s", ownerRepo, strings.Join(matches, ", "))
	}
}

func remoteOwnerRepo(remote string) string {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return ""
	}
	path := remote
	if i := strings.Index(remote, "://"); i >= 0 {
		path = remote[i+3:]
		if j := strings.Index(path, "/"); j >= 0 {
			path = path[j+1:]
		}
	} else if i := strings.Index(remote, ":"); i >= 0 {
		path = remote[i+1:]
	}
	path = strings.TrimSuffix(strings.Trim(path, "/"), ".git")
	return strings.ToLower(path)
}

// DefaultPrompt is used when --pr is given without --prompt.
func DefaultPrompt(info Info) string {
	return fmt.Sprintf("Review pull request #%d (%s). Its head branch %s is checked out in this workspace; compare it against the base branch and report findings.", info.Number, info.Title, info.Branch)
}

// Context is appended to any prompt so the agent knows what it is looking at.
func Context(info Info) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n\nPull request: %s\nTitle: %s\nHead branch: %s", info.URL, info.Title, info.Branch)
	if body := strings.TrimSpace(info.Body); body != "" {
		fmt.Fprintf(&b, "\n\nDescription:\n%s", body)
	}
	return b.String()
}
