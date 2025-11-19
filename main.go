package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-git/go-git/v5"
)

const version = "0.1.0"

func main() {
	// Define CLI flags
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help information")
		remoteName  = flag.String("remote", "origin", "Name of the remote to open")
	)

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("gitfab version %s\n", version)
		os.Exit(0)
	}

	// Handle help flag
	if *showHelp {
		fmt.Println("gitfab - Opens a git repository in a browser")
		fmt.Println("\nUsage:")
		fmt.Println("  gitfab [flags]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Find the Git repository starting from current directory
	path, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	repoPath, err := findGitRepository(path)
	if err != nil {
		log.Fatalf("Failed to find git repository: %v", err)
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Fatalf("Failed to open git repository: %v", err)
	}

	// Read the Git configuration
	config, err := repo.Config()
	if err != nil {
		log.Fatalf("Failed to read git config: %v", err)
	}

	// Get the specified remote
	remote, ok := config.Remotes[*remoteName]
	if !ok {
		log.Fatalf("Remote '%s' not found", *remoteName)
	}

	if len(remote.URLs) == 0 {
		log.Fatalf("Remote '%s' has no URLs configured", *remoteName)
	}

	openBrowser(remote.URLs[0])
}

func openBrowser(origin string) {
	var err error

	url := translateGit2Url(origin)

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		log.Printf("opening %s in browser", url)
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func findGitRepository(startPath string) (string, error) {
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

func translateGit2Url(url string) string {
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
