package gitfab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	colReset  = "\033[0m"
	colYellow = "\033[33m"
	colGreen  = "\033[32m"
	colRed    = "\033[31m"
	colDim    = "\033[90m"
	clearScr  = "\033[H\033[J"
)

// ParseOwnerRepo extracts the <owner>/<repo> pair from a git remote URL.
func ParseOwnerRepo(remoteURL string) (owner, repo string, err error) {
	httpsURL := TranslateGit2Url(remoteURL)
	u, err := url.Parse(httpsURL)
	if err != nil {
		return "", "", fmt.Errorf("parse remote url: %w", err)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("could not determine owner/repo from %q", remoteURL)
	}
	return parts[0], parts[1], nil
}

type WorkflowRun struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	HeadBranch string    `json:"head_branch"`
	Event      string    `json:"event"`
	HTMLURL    string    `json:"html_url"`
	CreatedAt  time.Time `json:"created_at"`
}

type workflowRunsResponse struct {
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

func fetchWorkflowRuns(ctx context.Context, owner, repo, token, branch string) ([]WorkflowRun, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs?per_page=20", url.PathEscape(owner), url.PathEscape(repo))
	if branch != "" {
		endpoint += "&branch=" + url.QueryEscape(branch)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "gitfab")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("github api %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var out workflowRunsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return out.WorkflowRuns, nil
}

// ListRuns fetches workflow runs once and writes them to out, then returns.
// With asJSON it emits the raw run objects; otherwise a plain table with no
// ANSI escapes, suitable for piping/grep.
func ListRuns(out io.Writer, owner, repo, token, branch string, asJSON bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	runs, err := fetchWorkflowRuns(ctx, owner, repo, token, branch)
	if err != nil {
		return err
	}
	return writeRuns(out, runs, asJSON)
}

// WaitRuns polls until all runs that were active during the wait have
// completed, prints a final summary, and reports whether they all passed.
// If no active run appears within a short grace window it exits early.
func WaitRuns(out io.Writer, owner, repo, token, branch string, interval time.Duration, asJSON bool) (allPassed bool, err error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	const graceTicks = 6
	watched := map[int64]bool{}
	idle := 0
	var runs []WorkflowRun

	for {
		runs, err = fetchWorkflowRuns(ctx, owner, repo, token, branch)
		if err != nil {
			if ctx.Err() != nil {
				return false, nil
			}
			return false, err
		}

		active := 0
		for _, r := range runs {
			if isActive(r.Status) {
				active++
				watched[r.ID] = true
			}
		}

		if active == 0 {
			if len(watched) > 0 {
				break
			}
			idle++
			if idle >= graceTicks {
				fmt.Fprintln(out, "no active runs appeared")
				return true, writeRuns(out, runs, asJSON)
			}
		}

		select {
		case <-ctx.Done():
			return false, nil
		case <-ticker.C:
		}
	}

	failed := 0
	for _, r := range runs {
		if !watched[r.ID] {
			continue
		}
		switch r.Conclusion {
		case "failure", "timed_out", "cancelled", "startup_failure":
			failed++
		}
	}
	if err := writeRuns(out, runs, asJSON); err != nil {
		return failed == 0, err
	}
	if failed > 0 {
		fmt.Fprintf(out, "\n%d watched run(s) failed\n", failed)
		return false, nil
	}
	fmt.Fprintf(out, "\nall %d watched run(s) passed\n", len(watched))
	return true, nil
}

func writeRuns(out io.Writer, runs []WorkflowRun, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(runs)
	}
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "STATUS\tWORKFLOW\tBRANCH\tEVENT\tAGE\tID")
	for _, r := range runs {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\n",
			plainStatus(r), r.Name, r.HeadBranch, r.Event,
			humanAge(time.Since(r.CreatedAt)), r.ID)
	}
	return tw.Flush()
}

func plainStatus(r WorkflowRun) string {
	if isActive(r.Status) {
		return "* " + r.Status
	}
	if r.Conclusion == "" {
		return "- " + r.Status
	}
	return r.Conclusion
}

// WatchRuns polls GitHub Actions workflow runs and renders them until interrupted.
func WatchRuns(out io.Writer, owner, repo, token, branch string, interval time.Duration) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		runs, err := fetchWorkflowRuns(ctx, owner, repo, token, branch)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Fprintln(out)
				return nil
			}
			return err
		}
		renderRuns(out, owner, repo, token, runs)

		active := 0
		for _, r := range runs {
			if isActive(r.Status) {
				active++
			}
		}
		if active == 0 {
			fmt.Fprintln(out, "\nNo builds running.")
			return nil
		}

		select {
		case <-ctx.Done():
			fmt.Fprintln(out)
			return nil
		case <-ticker.C:
		}
	}
}

func isActive(status string) bool {
	switch status {
	case "queued", "in_progress", "waiting", "requested", "pending":
		return true
	}
	return false
}

func statusCell(r WorkflowRun) string {
	if isActive(r.Status) {
		return colYellow + "● " + r.Status + colReset
	}
	switch r.Conclusion {
	case "success":
		return colGreen + "✓ success" + colReset
	case "failure", "timed_out":
		return colRed + "✗ " + r.Conclusion + colReset
	case "":
		return colDim + "- " + r.Status + colReset
	default:
		return colDim + "- " + r.Conclusion + colReset
	}
}

func humanAge(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func renderRuns(out io.Writer, owner, repo, token string, runs []WorkflowRun) {
	fmt.Fprint(out, clearScr)
	fmt.Fprintf(out, "Watching %s/%s — Ctrl+C to stop\n\n", owner, repo)

	var active, done []WorkflowRun
	for _, r := range runs {
		if isActive(r.Status) {
			active = append(active, r)
		} else {
			done = append(done, r)
		}
	}
	if len(done) > 5 {
		done = done[:5]
	}
	display := append(active, done...)

	if len(display) == 0 {
		fmt.Fprintln(out, colDim+"  no workflow runs found"+colReset)
	} else {
		tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "STATUS\tWORKFLOW\tBRANCH\tEVENT\tAGE\tID")
		for _, r := range display {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\n",
				statusCell(r), r.Name, r.HeadBranch, r.Event,
				humanAge(time.Since(r.CreatedAt)), r.ID)
		}
		tw.Flush()
	}

	fmt.Fprintf(out, "\nUpdated %s", time.Now().Format("15:04:05"))
	if token == "" {
		fmt.Fprint(out, "  "+colDim+"(no GITHUB_TOKEN — public repos only, low rate limit)"+colReset)
	}
	fmt.Fprintln(out)
}
