package gitfab

import (
	"fmt"
	"net/url"
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

func TranslateGit2Url(gitURL string) string {
	// Handle SSH URLs (git@github.com:user/repo.git)
	if strings.HasPrefix(gitURL, "git@") {
		// Split on the first colon to separate host from path
		parts := strings.SplitN(gitURL, ":", 2)
		if len(parts) == 2 {
			// Replace git@ with https://
			host := strings.Replace(parts[0], "git@", "", 1)
			path := strings.TrimSuffix(parts[1], ".git")

			// SECURITY: Properly construct URL using net/url
			u := &url.URL{
				Scheme: "https",
				Host:   host,
				Path:   "/" + path,
			}
			return u.String()
		}
	} else if strings.HasPrefix(gitURL, "https://") || strings.HasPrefix(gitURL, "http://") {
		// SECURITY: Parse and validate HTTP(S) URLs
		parsedURL, err := url.Parse(gitURL)
		if err != nil {
			// Fallback to simple string manipulation if parsing fails
			return strings.TrimSuffix(gitURL, ".git")
		}
		// Remove .git suffix from path
		parsedURL.Path = strings.TrimSuffix(parsedURL.Path, ".git")
		return parsedURL.String()
	} else if strings.HasPrefix(gitURL, "ssh://") {
		// SECURITY: Properly parse SSH URLs
		parsedURL, err := url.Parse(gitURL)
		if err != nil {
			// Fallback to simple string manipulation
			gitURL = strings.Replace(gitURL, "ssh://git@", "https://", 1)
			gitURL = strings.Replace(gitURL, "ssh://", "https://", 1)
			return strings.TrimSuffix(gitURL, ".git")
		}

		// Convert ssh:// to https://
		parsedURL.Scheme = "https"
		// Remove username (git@) from host
		if parsedURL.User != nil {
			parsedURL.User = nil
		}
		// Remove .git suffix
		parsedURL.Path = strings.TrimSuffix(parsedURL.Path, ".git")
		return parsedURL.String()
	}

	return gitURL
}

// TranslateGit2UrlWithPage converts a Git URL to an HTTP(S) URL with optional page suffix
func TranslateGit2UrlWithPage(gitURL string, pageType PageType) string {
	baseURL := TranslateGit2Url(gitURL)
	platform := DetectPlatform(gitURL)
	pagePath := GetPagePath(platform, pageType)

	// SECURITY: Use proper URL joining instead of string concatenation
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		// Fallback to simple concatenation if parsing fails
		return baseURL + pagePath
	}

	parsedURL.Path = parsedURL.Path + pagePath
	return parsedURL.String()
}
