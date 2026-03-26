//go:build linux

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
			Path:        filepath.Join(homeDir, ".cache"),
			Description: "XDG cache directory (safe to remove)",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Trash",
			Path:        filepath.Join(homeDir, ".local/share/Trash"),
			Description: "Items in XDG Trash",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "npm Cache",
			Path:        filepath.Join(homeDir, ".npm"),
			Description: "npm package cache",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Yarn Cache",
			Path:        filepath.Join(homeDir, ".cache/yarn"),
			Description: "Yarn package cache",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "pip Cache",
			Path:        filepath.Join(homeDir, ".cache/pip"),
			Description: "Python pip cache",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Go Build Cache",
			Path:        filepath.Join(homeDir, ".cache/go-build"),
			Description: "Go compilation cache",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Gradle Cache",
			Path:        filepath.Join(homeDir, ".gradle/caches"),
			Description: "Gradle build cache",
			Timeout:     15 * time.Second,
		},
	}
}
