package main

import (
	"flag"
	"fmt"
	gitfab "gitfab/src"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/go-git/go-git/v5"
)

const version = "0.1.0"

func main() {
	// Define CLI flags
	var (
		showVersion   = flag.Bool("version", false, "Show version information")
		showHelp      = flag.Bool("help", false, "Show help information")
		remoteName    = flag.String("remote", "origin", "Name of the remote to open")
		openPipeline  = flag.Bool("target", false, "Open pipeline/actions page (auto-detects based on platform)")
		openPipelineT = flag.Bool("t", false, "Shorthand for -target")
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

	repoPath, err := gitfab.FindGitRepository(path)
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

	remoteURL := remote.URLs[0]

	// Determine which page to open
	var page gitfab.PageType
	if *openPipeline || *openPipelineT {
		page = gitfab.PagePipeline
	} else {
		page = gitfab.PageRepo
	}

	openBrowser(remoteURL, page)
}

func openBrowser(origin string, page gitfab.PageType) {
	var err error

	urlStr := gitfab.TranslateGit2UrlWithPage(origin, page)

	// SECURITY: Validate URL before passing to exec to prevent command injection
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Fatalf("Invalid URL: %v", err)
	}

	// Whitelist allowed schemes
	allowedSchemes := map[string]bool{
		"http":  true,
		"https": true,
	}
	if !allowedSchemes[parsedURL.Scheme] {
		log.Fatalf("Unsupported URL scheme: %s (allowed: http, https)", parsedURL.Scheme)
	}

	// Additional validation - check for dangerous characters
	dangerousChars := []string{"`", "$", "(", ")", ";", "&", "|", "<", ">", "\n", "\r"}
	for _, char := range dangerousChars {
		if strings.Contains(urlStr, char) {
			log.Fatalf("URL contains dangerous characters: %s", char)
		}
	}

	// Re-encode the URL for safety
	safeURL := parsedURL.String()

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", safeURL).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", safeURL).Start()
	case "darwin":
		log.Printf("opening %s in browser", safeURL)
		err = exec.Command("open", safeURL).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}
