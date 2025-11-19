package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTranslateGit2Url(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SSH URL with git@",
			input:    "git@github.com:JamesWoolfenden/gitfab.git",
			expected: "https://github.com/JamesWoolfenden/gitfab",
		},
		{
			name:     "SSH URL without .git",
			input:    "git@github.com:user/repo",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "HTTPS URL with .git",
			input:    "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "HTTPS URL without .git",
			input:    "https://github.com/user/repo",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "HTTP URL with .git",
			input:    "http://example.com/user/repo.git",
			expected: "http://example.com/user/repo",
		},
		{
			name:     "SSH protocol URL with git@",
			input:    "ssh://git@github.com/user/repo.git",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "SSH protocol URL without git@",
			input:    "ssh://github.com/user/repo.git",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "GitLab SSH URL",
			input:    "git@gitlab.com:group/subgroup/project.git",
			expected: "https://gitlab.com/group/subgroup/project",
		},
		{
			name:     "Bitbucket SSH URL",
			input:    "git@bitbucket.org:team/repo.git",
			expected: "https://bitbucket.org/team/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translateGit2Url(tt.input)
			if result != tt.expected {
				t.Errorf("translateGit2Url(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFindGitRepository(t *testing.T) {
	// Create a temporary directory structure with a git repo
	tempDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tempDir, ".git")
	err := os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir", "nested")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	t.Run("Find from root", func(t *testing.T) {
		result, err := findGitRepository(tempDir)
		if err != nil {
			t.Errorf("findGitRepository(%q) returned error: %v", tempDir, err)
		}
		if result != tempDir {
			t.Errorf("findGitRepository(%q) = %q, want %q", tempDir, result, tempDir)
		}
	})

	t.Run("Find from subdirectory", func(t *testing.T) {
		result, err := findGitRepository(subDir)
		if err != nil {
			t.Errorf("findGitRepository(%q) returned error: %v", subDir, err)
		}
		if result != tempDir {
			t.Errorf("findGitRepository(%q) = %q, want %q", subDir, result, tempDir)
		}
	})

	t.Run("Not a git repository", func(t *testing.T) {
		nonGitDir := filepath.Join(t.TempDir(), "notgit")
		err := os.Mkdir(nonGitDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		_, err = findGitRepository(nonGitDir)
		if err == nil {
			t.Error("findGitRepository should return error for non-git directory")
		}
	})
}
