package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHasUpdate(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
		wantErr bool
	}{
		{name: "older current means update", current: "2.0.0", latest: "2.1.0", want: true},
		{name: "equal means no update", current: "2.1.0", latest: "2.1.0", want: false},
		{name: "newer current means no update", current: "2.2.0", latest: "2.1.0", want: false},
		{name: "v prefix on both is accepted", current: "v2.0.0", latest: "v2.1.0", want: true},
		{name: "mixed prefixes normalise", current: "2.0.0", latest: "v2.1.0", want: true},
		{name: "dev build always updates", current: "dev", latest: "2.1.0", want: true},
		{name: "dev build updates even against older latest", current: "dev", latest: "0.0.1", want: true},
		{name: "patch bump is detected", current: "2.1.0", latest: "2.1.1", want: true},
		{name: "major bump is detected", current: "2.1.0", latest: "3.0.0", want: true},
		{name: "prerelease sorts below release", current: "2.1.0-alpha.1", latest: "2.1.0", want: true},
		{name: "release does not downgrade to prerelease", current: "2.1.0", latest: "2.1.0-alpha.1", want: false},
		{name: "invalid current is an error", current: "not-a-version", latest: "2.1.0", wantErr: true},
		{name: "invalid latest is an error", current: "2.1.0", latest: "not-a-version", wantErr: true},
		{name: "empty current is an error", current: "", latest: "2.1.0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HasUpdate(tt.current, tt.latest)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("HasUpdate(%q, %q) = %v, want an error", tt.current, tt.latest, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("HasUpdate(%q, %q) returned unexpected error: %v", tt.current, tt.latest, err)
			}
			if got != tt.want {
				t.Errorf("HasUpdate(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestFindAsset(t *testing.T) {
	release := &Release{
		TagName: "v2.1.0",
		Assets: []Asset{
			{Name: "devcockpit-darwin-arm64", BrowserDownloadURL: "https://example.invalid/bin"},
			{Name: "devcockpit-darwin-arm64.sha256", BrowserDownloadURL: "https://example.invalid/sum"},
		},
	}

	t.Run("finds the binary", func(t *testing.T) {
		got := release.FindAsset("devcockpit-darwin-arm64")
		if got == nil {
			t.Fatal("FindAsset returned nil for an asset that exists")
		}
		if got.BrowserDownloadURL != "https://example.invalid/bin" {
			t.Errorf("got URL %q, want the binary URL", got.BrowserDownloadURL)
		}
	})

	t.Run("does not confuse the checksum for the binary", func(t *testing.T) {
		got := release.FindAsset("devcockpit-darwin-arm64.sha256")
		if got == nil {
			t.Fatal("FindAsset returned nil for the checksum asset")
		}
		if got.BrowserDownloadURL != "https://example.invalid/sum" {
			t.Errorf("got URL %q, want the checksum URL", got.BrowserDownloadURL)
		}
	})

	t.Run("returns nil when absent", func(t *testing.T) {
		if got := release.FindAsset("devcockpit-linux-amd64"); got != nil {
			t.Errorf("FindAsset returned %+v for an absent asset, want nil", got)
		}
	})

	t.Run("returns nil on an empty release", func(t *testing.T) {
		empty := &Release{}
		if got := empty.FindAsset("anything"); got != nil {
			t.Errorf("FindAsset returned %+v on an empty release, want nil", got)
		}
	})
}

// writeFile is a helper that creates a file with the given contents and
// returns its path.
func writeFile(t *testing.T, dir, name, contents string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	return path
}

func sha256Of(t *testing.T, contents string) string {
	t.Helper()
	sum := sha256.Sum256([]byte(contents))
	return hex.EncodeToString(sum[:])
}

func TestVerifyChecksum(t *testing.T) {
	const payload = "this is the binary content"
	digest := sha256Of(t, payload)

	t.Run("accepts a matching digest", func(t *testing.T) {
		dir := t.TempDir()
		bin := writeFile(t, dir, "bin", payload)
		sum := writeFile(t, dir, "bin.sha256", digest+"  bin\n")

		if err := verifyChecksum(bin, sum); err != nil {
			t.Errorf("verifyChecksum rejected a matching digest: %v", err)
		}
	})

	t.Run("accepts a bare digest with no filename", func(t *testing.T) {
		dir := t.TempDir()
		bin := writeFile(t, dir, "bin", payload)
		sum := writeFile(t, dir, "bin.sha256", digest)

		if err := verifyChecksum(bin, sum); err != nil {
			t.Errorf("verifyChecksum rejected a bare digest: %v", err)
		}
	})

	t.Run("accepts an uppercase digest", func(t *testing.T) {
		dir := t.TempDir()
		bin := writeFile(t, dir, "bin", payload)
		sum := writeFile(t, dir, "bin.sha256", strings.ToUpper(digest)+"  bin\n")

		if err := verifyChecksum(bin, sum); err != nil {
			t.Errorf("verifyChecksum rejected an uppercase digest: %v", err)
		}
	})

	t.Run("rejects a tampered binary", func(t *testing.T) {
		dir := t.TempDir()
		bin := writeFile(t, dir, "bin", payload+" plus a backdoor")
		sum := writeFile(t, dir, "bin.sha256", digest+"  bin\n")

		err := verifyChecksum(bin, sum)
		if err == nil {
			t.Fatal("verifyChecksum accepted a binary that does not match its digest")
		}
		if !strings.Contains(err.Error(), "SECURITY ALERT") {
			t.Errorf("mismatch error should be explicit about the security failure, got: %v", err)
		}
	})

	t.Run("rejects an empty checksum file", func(t *testing.T) {
		dir := t.TempDir()
		bin := writeFile(t, dir, "bin", payload)
		sum := writeFile(t, dir, "bin.sha256", "")

		if err := verifyChecksum(bin, sum); err == nil {
			t.Error("verifyChecksum accepted an empty checksum file")
		}
	})

	t.Run("rejects a whitespace-only checksum file", func(t *testing.T) {
		dir := t.TempDir()
		bin := writeFile(t, dir, "bin", payload)
		sum := writeFile(t, dir, "bin.sha256", "   \n\t\n")

		if err := verifyChecksum(bin, sum); err == nil {
			t.Error("verifyChecksum accepted a whitespace-only checksum file")
		}
	})

	t.Run("fails when the checksum file is missing", func(t *testing.T) {
		dir := t.TempDir()
		bin := writeFile(t, dir, "bin", payload)

		if err := verifyChecksum(bin, filepath.Join(dir, "absent.sha256")); err == nil {
			t.Error("verifyChecksum succeeded with no checksum file present")
		}
	})

	t.Run("fails when the binary is missing", func(t *testing.T) {
		dir := t.TempDir()
		sum := writeFile(t, dir, "bin.sha256", digest)

		if err := verifyChecksum(filepath.Join(dir, "absent"), sum); err == nil {
			t.Error("verifyChecksum succeeded with no binary present")
		}
	})
}

func TestCopyFile(t *testing.T) {
	t.Run("copies contents exactly", func(t *testing.T) {
		dir := t.TempDir()
		const contents = "binary\x00bytes\xff and text"
		src := writeFile(t, dir, "src", contents)
		dst := filepath.Join(dir, "dst")

		if err := copyFile(src, dst); err != nil {
			t.Fatalf("copyFile returned an error: %v", err)
		}

		got, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("could not read the copy: %v", err)
		}
		if string(got) != contents {
			t.Errorf("copy does not match the source; got %q, want %q", got, contents)
		}
	})

	t.Run("errors when the source is missing", func(t *testing.T) {
		dir := t.TempDir()
		if err := copyFile(filepath.Join(dir, "absent"), filepath.Join(dir, "dst")); err == nil {
			t.Error("copyFile succeeded with a missing source")
		}
	})

	t.Run("errors when the destination is not writable", func(t *testing.T) {
		dir := t.TempDir()
		src := writeFile(t, dir, "src", "content")

		if err := copyFile(src, filepath.Join(dir, "no-such-dir", "dst")); err == nil {
			t.Error("copyFile succeeded writing into a directory that does not exist")
		}
	})
}
