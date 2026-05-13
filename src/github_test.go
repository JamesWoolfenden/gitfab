package gitfab

import (
	"testing"
	"time"
)

func TestParseOwnerRepo(t *testing.T) {
	cases := []struct {
		in          string
		wantOwner   string
		wantRepo    string
		wantErr     bool
		description string
	}{
		{"git@github.com:JamesWoolfenden/gitfab.git", "JamesWoolfenden", "gitfab", false, "ssh shorthand"},
		{"https://github.com/JamesWoolfenden/gitfab.git", "JamesWoolfenden", "gitfab", false, "https"},
		{"https://github.com/JamesWoolfenden/gitfab", "JamesWoolfenden", "gitfab", false, "https no .git"},
		{"ssh://git@github.com/JamesWoolfenden/gitfab", "JamesWoolfenden", "gitfab", false, "ssh url"},
		{"https://github.com/", "", "", true, "no path"},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			owner, repo, err := ParseOwnerRepo(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got owner=%q repo=%q", owner, repo)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if owner != tc.wantOwner || repo != tc.wantRepo {
				t.Errorf("got %s/%s, want %s/%s", owner, repo, tc.wantOwner, tc.wantRepo)
			}
		})
	}
}

func TestPlainStatus(t *testing.T) {
	cases := []struct {
		r    WorkflowRun
		want string
	}{
		{WorkflowRun{Status: "in_progress"}, "* in_progress"},
		{WorkflowRun{Status: "queued"}, "* queued"},
		{WorkflowRun{Status: "completed", Conclusion: "success"}, "success"},
		{WorkflowRun{Status: "completed", Conclusion: "failure"}, "failure"},
		{WorkflowRun{Status: "completed", Conclusion: ""}, "- completed"},
	}
	for _, tc := range cases {
		if got := plainStatus(tc.r); got != tc.want {
			t.Errorf("plainStatus(%+v) = %q, want %q", tc.r, got, tc.want)
		}
	}
}

func TestStripLogTimestamp(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"2026-05-13T10:24:33.9087072Z ##[error]boom", "##[error]boom"},
		{"2026-05-13T10:24:33.0000000Z plain", "plain"},
		{"no timestamp here", "no timestamp here"},
		{"", ""},
		{"2026-05-13", "2026-05-13"},
	}
	for _, tc := range cases {
		if got := stripLogTimestamp(tc.in); got != tc.want {
			t.Errorf("stripLogTimestamp(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestHumanAge(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{45 * time.Second, "45s"},
		{3 * time.Minute, "3m"},
		{2 * time.Hour, "2h"},
		{49 * time.Hour, "2d"},
	}
	for _, tc := range cases {
		if got := humanAge(tc.d); got != tc.want {
			t.Errorf("humanAge(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}
