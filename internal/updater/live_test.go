//go:build integration

package updater

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLiveRelease runs the real update path against the published GitHub
// release: find it, download RekaShade.exe and verify its SHA-256.
//
//	go test -tags integration -run TestLiveRelease -v ./internal/updater/
func TestLiveRelease(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	res, err := CheckLatest(ctx, "0.0.1")
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if res.Latest == nil {
		t.Skip("no release published yet")
	}
	t.Logf("latest release: %s (%s), asset %d bytes", res.Latest.Version, res.Latest.PublishedAt, res.Latest.AssetSize)
	if !res.Available {
		t.Fatal("a release exists but is not installable: RekaShade.exe or SHA256SUMS.txt asset missing")
	}
	dir := t.TempDir()
	path, err := Download(ctx, res.Latest, dir, nil)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Size() < 1<<20 {
		t.Fatalf("downloaded file looks wrong: %v %v", info, err)
	}
	t.Logf("downloaded and SHA-256 verified: %d bytes", info.Size())
}
