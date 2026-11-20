package updater

import (
	"errors"
	"net/url"
	"testing"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.1", "1.0.0", 1},
		{"1.2.10", "1.2.9", 1},
		{"v2.0.0", "1.9.9", 1},
		{"1.0.0", "1.0.0", 0},
		{"1.0", "1.0.0", 0},
		{"1.0.0", "dev", 1},
		{"0.9.0", "1.0.0", -1},
		{"1.1.0-beta", "1.0.9", 1},
	}
	for _, c := range cases {
		if got := Compare(c.a, c.b); got != c.want {
			t.Errorf("Compare(%q,%q)=%d want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestOnlyGitHubHostsAllowed(t *testing.T) {
	ok := []string{
		"https://github.com/1337133713377/rekashade/releases/download/v1.0.0/RekaShade.exe",
		"https://objects.githubusercontent.com/x",
		"https://release-assets.githubusercontent.com/x",
	}
	bad := []string{
		"http://github.com/x",
		"https://evil.example.com/RekaShade.exe",
		"https://github.com.evil.example/x",
	}
	for _, s := range ok {
		u, _ := url.Parse(s)
		if err := checkURL(u); err != nil {
			t.Errorf("%s should be allowed: %v", s, err)
		}
	}
	for _, s := range bad {
		u, _ := url.Parse(s)
		if err := checkURL(u); err == nil {
			t.Errorf("%s should be refused", s)
		}
	}
}

func TestNoReleaseYetIsNotAnError(t *testing.T) {
	err := error(&statusError{code: 404, status: "404 Not Found"})
	var se *statusError
	if !errors.As(err, &se) || se.code != 404 {
		t.Fatal("status error should unwrap with its code")
	}
}
