package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	// Open the Git repository
	//
	path, err := os.Getwd()

	if err != nil {
		fmt.Println(err)
		return
	}

	repo, err := git.PlainOpen(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Read the Git configuration
	config, err := repo.Config()
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(config.Remotes) >= 1 {
		if origin, ok := config.Remotes["origin"]; ok {
			openBrowser(origin.URLs[0])
		} else {
			log.Fatalf("No remotes found")
		}

	} else {
		log.Fatalf("No remotes found")
	}

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

func translateGit2Url(url string) string {
	if strings.Contains(url, "git@") {
		url = strings.Replace(url, ":", "/", 1)
		url = strings.Replace(url, "git@", "https://", 1)
		url = strings.Replace(url, ".git", "", 1)
	}

	return url
}
