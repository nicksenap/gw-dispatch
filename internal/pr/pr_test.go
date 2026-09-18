package pr

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	info, err := Parse("https://github.com/Acme/api.git/pull/42/files")
	if err != nil {
		t.Fatal(err)
	}
	if info.Owner != "Acme" || info.Repo != "api" || info.Number != 42 || info.URL != "https://github.com/Acme/api/pull/42" {
		t.Fatalf("info = %+v", info)
	}
	for _, bad := range []string{"https://gitlab.com/a/b/-/merge_requests/1", "https://github.com/a/b/issues/3", "github.com/a/b/pull/x", ""} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("Parse(%q) accepted", bad)
		}
	}
}

func TestMatchRepoUsesDisplayNameOrRemote(t *testing.T) {
	repos := []byte(`[
		{"name":"api","display_name":"acme/api","remote":"git@github.com:acme/api.git"},
		{"name":"web","display_name":"web","remote":"https://github.com/Acme/Web.git"},
		{"name":"local","display_name":"local","remote":""}]`)
	if got, _ := MatchRepo(repos, "acme/api"); got != "api" {
		t.Fatalf("got %q", got)
	}
	if got, _ := MatchRepo(repos, "acme/web"); got != "web" {
		t.Fatalf("got %q", got)
	}
	if _, err := MatchRepo(repos, "acme/none"); err == nil || !strings.Contains(err.Error(), "no configured Grove repo") {
		t.Fatalf("err = %v", err)
	}
}
