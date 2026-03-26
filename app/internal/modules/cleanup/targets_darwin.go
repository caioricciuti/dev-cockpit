//go:build darwin

package cleanup

import (
	"os"
	"path/filepath"
	"time"
)

func defaultTargets() []CleanupTarget {
	homeDir, _ := os.UserHomeDir()

	return []CleanupTarget{
		{
			Name:        "User Caches",
			Path:        filepath.Join(homeDir, "Library/Caches"),
			Description: "Application cache files (safe to remove)",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Trash",
			Path:        filepath.Join(homeDir, ".Trash"),
			Description: "Items in Trash",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Homebrew Cache",
			Path:        filepath.Join(homeDir, "Library/Caches/Homebrew"),
			Description: "Downloaded Homebrew installers",
			Timeout:     15 * time.Second,
		},
		{
			Name:        "npm Cache",
			Path:        filepath.Join(homeDir, ".npm"),
			Description: "npm package cache",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Yarn Cache",
			Path:        filepath.Join(homeDir, "Library/Caches/Yarn"),
			Description: "Yarn package cache",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Go Build Cache",
			Path:        filepath.Join(homeDir, "Library/Caches/go-build"),
			Description: "Go compilation cache",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Xcode Derived Data",
			Path:        filepath.Join(homeDir, "Library/Developer/Xcode/DerivedData"),
			Description: "Xcode build artifacts (can be large)",
			Timeout:     30 * time.Second,
		},
	}
}
