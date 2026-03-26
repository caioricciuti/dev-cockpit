//go:build darwin

package cli

var cacheLocations = []struct {
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
	{"Gradle Cache", ".gradle/caches"},
	{"Maven Cache", ".m2/repository"},
	{"Docker Temp", "Library/Containers/com.docker.docker/Data"},
}
