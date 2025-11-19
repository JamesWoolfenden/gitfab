package gitfab

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func FindGitRepository(startPath string) (string, error) {
	// Clean the path
	currentPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	// Walk up the directory tree looking for .git
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return currentPath, nil
		}

		// Get parent directory
		parentPath := filepath.Dir(currentPath)

		// If we've reached the root, stop
		if parentPath == currentPath {
			return "", fmt.Errorf("not a git repository (or any parent up to mount point)")
		}

		currentPath = parentPath
	}
}

func TranslateGit2Url(url string) string {
	// Handle SSH URLs (git@github.com:user/repo.git)
	if strings.HasPrefix(url, "git@") {
		// Split on the first colon to separate host from path
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 2 {
			// Replace git@ with https://
			host := strings.Replace(parts[0], "git@", "https://", 1)
			path := parts[1]
			// Remove .git suffix if present
			path = strings.TrimSuffix(path, ".git")
			url = host + "/" + path
		}
	} else if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		// Already an HTTP(S) URL, just remove .git suffix if present
		url = strings.TrimSuffix(url, ".git")
	} else if strings.HasPrefix(url, "ssh://") {
		// Handle ssh:// URLs (ssh://git@github.com/user/repo.git)
		url = strings.Replace(url, "ssh://git@", "https://", 1)
		url = strings.Replace(url, "ssh://", "https://", 1)
		url = strings.TrimSuffix(url, ".git")
	}

	return url
}
