// Package updater keeps Reka Shade up to date from GitHub Releases.
//
// Flow: ask the GitHub API for the latest release, compare versions,
// download the portable exe asset (only from GitHub hosts, over HTTPS),
// verify it against the SHA256SUMS.txt asset of the same release, then swap
// it in place of the running exe (see swap_windows.go) and restart.
package updater

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Repo is the GitHub repository releases are published to.
const Repo = "1337133713377/rekashade"

// AssetName is the portable executable attached to every release.
const AssetName = "RekaShade.exe"

// SumsName is the checksum file attached to every release.
const SumsName = "SHA256SUMS.txt"

// Release describes an available version.
type Release struct {
	Version     string `json:"version"`
	Name        string `json:"name"`
	Notes       string `json:"notes"`
	PublishedAt string `json:"publishedAt"`
	PageURL     string `json:"pageUrl"`
	AssetSize   int64  `json:"assetSize"`
	assetURL    string
	sumsURL     string
}

// Check is the result of an update check.
type Check struct {
	Current   string   `json:"current"`
	Available bool     `json:"available"`
	Latest    *Release `json:"latest,omitempty"`
	CheckedAt string   `json:"checkedAt"`
}

var allowedHosts = map[string]bool{
	"api.github.com":                        true,
	"github.com":                            true,
	"objects.githubusercontent.com":         true,
	"release-assets.githubusercontent.com":  true,
	"github-releases.githubusercontent.com": true,
}

func checkURL(u *url.URL) error {
	if u.Scheme != "https" || !allowedHosts[strings.ToLower(u.Hostname())] {
		return fmt.Errorf("refusing to download from %s", u.Host)
	}
	return nil
}

func client(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 5 {
				return errors.New("too many redirects")
			}
			return checkURL(req.URL)
		},
	}
}

func get(ctx context.Context, c *http.Client, rawURL string) (*http.Response, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if err := checkURL(u); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RekaShade-Updater")
	req.Header.Set("Accept", "application/vnd.github+json, application/octet-stream")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, &statusError{code: resp.StatusCode, status: resp.Status}
	}
	return resp, nil
}

type statusError struct {
	code   int
	status string
}

func (e *statusError) Error() string { return "GitHub answered " + e.status }

type ghRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	PublishedAt string `json:"published_at"`
	HTMLURL     string `json:"html_url"`
	Assets      []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
		Size int64  `json:"size"`
	} `json:"assets"`
}

// CheckLatest asks GitHub for the newest release.
func CheckLatest(ctx context.Context, current string) (Check, error) {
	res := Check{Current: current, CheckedAt: time.Now().UTC().Format(time.RFC3339)}
	resp, err := get(ctx, client(20*time.Second), "https://api.github.com/repos/"+Repo+"/releases/latest")
	var se *statusError
	if errors.As(err, &se) && se.code == http.StatusNotFound {
		return res, nil // no release published yet: nothing to update to
	}
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	var gh ghRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&gh); err != nil {
		return res, err
	}
	if gh.Draft || gh.Prerelease {
		return res, nil
	}
	rel := &Release{
		Version: strings.TrimPrefix(gh.TagName, "v"), Name: gh.Name, Notes: gh.Body,
		PublishedAt: gh.PublishedAt, PageURL: gh.HTMLURL,
	}
	for _, a := range gh.Assets {
		switch a.Name {
		case AssetName:
			rel.assetURL, rel.AssetSize = a.URL, a.Size
		case SumsName:
			rel.sumsURL = a.URL
		}
	}
	res.Latest = rel
	res.Available = rel.assetURL != "" && rel.sumsURL != "" && Compare(rel.Version, current) > 0
	return res, nil
}

// Compare compares two dotted versions ("1.2.10" > "1.2.9"). Anything that
// is not a number (e.g. "dev") sorts lowest.
func Compare(a, b string) int {
	pa, pb := parse(a), parse(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			if pa[i] > pb[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}

func parse(v string) [3]int {
	var out [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+ "); i >= 0 {
		v = v[:i]
	}
	for i, part := range strings.SplitN(v, ".", 3) {
		n, err := strconv.Atoi(part)
		if err != nil {
			return [3]int{-1, -1, -1}
		}
		out[i] = n
	}
	return out
}

func expectedSum(ctx context.Context, c *http.Client, sumsURL string) (string, error) {
	resp, err := get(ctx, c, sumsURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(io.LimitReader(resp.Body, 1<<20))
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == AssetName && len(f[0]) == 64 {
			return strings.ToLower(f[0]), nil
		}
	}
	return "", errors.New("checksum for " + AssetName + " not found in release")
}

// Download fetches the release exe into dir and verifies its SHA-256.
func Download(ctx context.Context, rel *Release, dir string, progress func(done, total int64)) (string, error) {
	if rel == nil || rel.assetURL == "" || rel.sumsURL == "" {
		return "", errors.New("this release has no downloadable build")
	}
	c := client(15 * time.Minute)
	want, err := expectedSum(ctx, c, rel.sumsURL)
	if err != nil {
		return "", err
	}
	resp, err := get(ctx, c, rel.assetURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	total := resp.ContentLength
	if total <= 0 {
		total = rel.AssetSize
	}
	if total > 512<<20 {
		return "", errors.New("update is unexpectedly large")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "RekaShade-"+rel.Version+".exe.download")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	var done int64
	buf := make([]byte, 64<<10)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				f.Close()
				os.Remove(path)
				return "", werr
			}
			h.Write(buf[:n])
			done += int64(n)
			if progress != nil {
				progress(done, total)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			f.Close()
			os.Remove(path)
			return "", rerr
		}
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return "", err
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != want {
		os.Remove(path)
		return "", errors.New("downloaded file failed its integrity check (SHA-256 mismatch)")
	}
	return path, nil
}
