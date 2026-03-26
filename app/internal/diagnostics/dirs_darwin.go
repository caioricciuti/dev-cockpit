//go:build darwin

package diagnostics

var cacheDirs = []struct {
	name string
	path string
}{
	{"Homebrew Cache", "Library/Caches/Homebrew"},
	{"npm Cache", ".npm/_cacache"},
	{"Yarn Cache", "Library/Caches/Yarn"},
	{"Go Module Cache", "go/pkg/mod/cache"},
	{"pip Cache", "Library/Caches/pip"},
	{"CocoaPods Cache", "Library/Caches/CocoaPods"},
	{"Xcode DerivedData", "Library/Developer/Xcode/DerivedData"},
}
