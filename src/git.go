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

// PageType represents the type of page to open
type PageType string

const (
	PageRepo     PageType = "repo"
	PagePipeline PageType = "pipeline"
)

// Platform represents the Git hosting platform
type Platform string

const (
	PlatformGitHub    Platform = "github"
	PlatformGitLab    Platform = "gitlab"
	PlatformBitbucket Platform = "bitbucket"
	PlatformUnknown   Platform = "unknown"
)

// DetectPlatform detects the Git hosting platform from a URL
func DetectPlatform(url string) Platform {
	if strings.Contains(url, "github.com") {
		return PlatformGitHub
	}
	if strings.Contains(url, "gitlab.com") || strings.Contains(url, "gitlab.") || strings.Contains(url, "code.pan.run") {
		return PlatformGitLab
	}
	if strings.Contains(url, "bitbucket.org") || strings.Contains(url, "bitbucket.") {
		return PlatformBitbucket
	}
	return PlatformUnknown
}

// GetPagePath returns the path suffix for a given platform and page type
func GetPagePath(platform Platform, pageType PageType) string {
	if pageType == PageRepo {
		return ""
	}

	switch platform {
	case PlatformGitHub:
		return "/actions"
	case PlatformGitLab:
		return "/-/pipelines"
	case PlatformBitbucket:
		return "/pipelines"
	default:
		return ""
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

// TranslateGit2UrlWithPage converts a Git URL to an HTTP(S) URL with optional page suffix
func TranslateGit2UrlWithPage(url string, pageType PageType) string {
	baseURL := TranslateGit2Url(url)
	platform := DetectPlatform(url)
	pagePath := GetPagePath(platform, pageType)
	return baseURL + pagePath
}
