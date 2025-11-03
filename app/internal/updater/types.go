package updater

// Release represents a GitHub release
type Release struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	PublishedAt string  `json:"published_at"`
	Body        string  `json:"body"`
	Assets      []Asset `json:"assets"`
}

// Asset represents a release asset (binary, checksum, etc.)
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	ContentType        string `json:"content_type"`
}

// UpdateOptions configures the update behavior
type UpdateOptions struct {
	Force      bool   // Skip confirmation prompts
	CheckOnly  bool   // Check for updates without installing
	CurrentVer string // Current version
}

// FindAsset finds an asset by name in the release
func (r *Release) FindAsset(name string) *Asset {
	for i := range r.Assets {
		if r.Assets[i].Name == name {
			return &r.Assets[i]
		}
	}
	return nil
}
