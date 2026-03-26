//go:build linux

package diagnostics

var cacheDirs = []struct {
	name string
	path string
}{
	{"User Cache", ".cache"},
	{"npm Cache", ".npm/_cacache"},
	{"Yarn Cache", ".cache/yarn"},
	{"Go Module Cache", "go/pkg/mod/cache"},
	{"pip Cache", ".cache/pip"},
	{"Gradle Cache", ".gradle/caches"},
}
