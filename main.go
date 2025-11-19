package main

import (
	"flag"
	"fmt"
	gitfab "gitfab/src"
	"log"
	"os"
	"os/exec"
	"runtime"

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

	openBrowser(remote.URLs[0])
}

func openBrowser(origin string) {
	var err error

	url := gitfab.TranslateGit2Url(origin)

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
