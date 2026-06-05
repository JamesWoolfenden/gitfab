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

// ParseGitLabProject extracts the host and full project path (e.g.
// "ocpo-it/cybernova/cybernova-poc") from a git remote URL. GitLab project
// paths may be arbitrarily nested, unlike GitHub's owner/repo pair.
func ParseGitLabProject(remoteURL string) (host, projectPath string, err error) {
	httpsURL := TranslateGit2Url(remoteURL)
	u, err := url.Parse(httpsURL)
	if err != nil {
		return "", "", fmt.Errorf("parse remote url: %w", err)
	}
	path := strings.Trim(u.Path, "/")
	if path == "" {
		return "", "", fmt.Errorf("could not determine project path from %q", remoteURL)
	}
	return u.Host, path, nil
}

func glGet(ctx context.Context, host, endpoint, token string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+host+"/api/v4"+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gitfab")
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		resp.Body.Close()
		return nil, fmt.Errorf("gitlab api %s %s: %s", endpoint, resp.Status, strings.TrimSpace(string(body)))
	}
	return resp, nil
}

type Pipeline struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	Ref       string    `json:"ref"`
	SHA       string    `json:"sha"`
	Source    string    `json:"source"`
	WebURL    string    `json:"web_url"`
	CreatedAt time.Time `json:"created_at"`
}

func fetchPipelines(ctx context.Context, host, projectPath, token, branch string) ([]Pipeline, error) {
	endpoint := fmt.Sprintf("/projects/%s/pipelines?per_page=20&order_by=id&sort=desc", url.PathEscape(projectPath))
	if branch != "" {
		endpoint += "&ref=" + url.QueryEscape(branch)
	}
	resp, err := glGet(ctx, host, endpoint, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out []Pipeline
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode pipelines: %w", err)
	}
	return out, nil
}

func glIsActive(status string) bool {
	switch status {
	case "created", "waiting_for_resource", "preparing", "pending", "running", "scheduled":
		return true
	}
	return false
}

func glStatusCell(p Pipeline) string {
	if glIsActive(p.Status) {
		return colYellow + "● " + p.Status + colReset
	}
	switch p.Status {
	case "success":
		return colGreen + "✓ success" + colReset
	case "failed":
		return colRed + "✗ failed" + colReset
	default:
		return colDim + "- " + p.Status + colReset
	}
}

func glPlainStatus(p Pipeline) string {
	if glIsActive(p.Status) {
		return "* " + p.Status
	}
	return p.Status
}

func writePipelines(out io.Writer, pipes []Pipeline, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(pipes)
	}
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "STATUS\tREF\tSOURCE\tSHA\tAGE\tID")
	for _, p := range pipes {
		sha := p.SHA
		if len(sha) > 8 {
			sha = sha[:8]
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\n",
			glPlainStatus(p), p.Ref, p.Source, sha,
			humanAge(time.Since(p.CreatedAt)), p.ID)
	}
	return tw.Flush()
}

// ListPipelines fetches pipelines once and writes them to out.
func ListPipelines(out io.Writer, host, projectPath, token, branch string, asJSON bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pipes, err := fetchPipelines(ctx, host, projectPath, token, branch)
	if err != nil {
		return err
	}
	return writePipelines(out, pipes, asJSON)
}

// WaitPipelines polls until all pipelines that were active during the wait
// have completed, prints a final summary, and reports whether they all passed.
func WaitPipelines(out io.Writer, host, projectPath, token, branch string, interval time.Duration, asJSON bool) (allPassed bool, err error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	const graceTicks = 6
	watched := map[int64]bool{}
	idle := 0
	var pipes []Pipeline

	for {
		pipes, err = fetchPipelines(ctx, host, projectPath, token, branch)
		if err != nil {
			if ctx.Err() != nil {
				return false, nil
			}
			return false, err
		}

		active := 0
		for _, p := range pipes {
			if glIsActive(p.Status) {
				active++
				watched[p.ID] = true
			}
		}

		if active == 0 {
			if len(watched) > 0 {
				break
			}
			idle++
			if idle >= graceTicks {
				fmt.Fprintln(out, "no active pipelines appeared")
				return true, writePipelines(out, pipes, asJSON)
			}
		}

		select {
		case <-ctx.Done():
			return false, nil
		case <-ticker.C:
		}
	}

	failed := 0
	for _, p := range pipes {
		if !watched[p.ID] {
			continue
		}
		if p.Status == "failed" || p.Status == "canceled" {
			failed++
		}
	}
	if err := writePipelines(out, pipes, asJSON); err != nil {
		return failed == 0, err
	}
	if failed > 0 {
		fmt.Fprintf(out, "\n%d watched pipeline(s) failed\n", failed)
		return false, nil
	}
	fmt.Fprintf(out, "\nall %d watched pipeline(s) passed\n", len(watched))
	return true, nil
}

// WatchPipelines polls GitLab pipelines and renders them until interrupted or
// all active pipelines finish.
func WatchPipelines(out io.Writer, host, projectPath, token, branch string, interval time.Duration) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	const grace = 10 * time.Second
	graceUntil := time.Now().Add(grace)
	seenActive := false

	for {
		pipes, err := fetchPipelines(ctx, host, projectPath, token, branch)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Fprintln(out)
				return nil
			}
			return err
		}
		renderPipelines(out, host, projectPath, token, pipes)

		active := 0
		for _, p := range pipes {
			if glIsActive(p.Status) {
				active++
			}
		}
		if active > 0 {
			seenActive = true
		} else if seenActive {
			fmt.Fprintln(out, "\nAll pipelines finished.")
			return nil
		} else if time.Now().After(graceUntil) {
			fmt.Fprintln(out, "\nNo pipelines running.")
			return nil
		} else {
			fmt.Fprintln(out, colDim+"\n  waiting for pipelines to start..."+colReset)
		}

		select {
		case <-ctx.Done():
			fmt.Fprintln(out)
			return nil
		case <-ticker.C:
		}
	}
}

func renderPipelines(out io.Writer, host, projectPath, token string, pipes []Pipeline) {
	fmt.Fprint(out, clearScr)
	fmt.Fprintf(out, "Watching %s/%s — Ctrl+C to stop\n\n", host, projectPath)

	var active, done []Pipeline
	for _, p := range pipes {
		if glIsActive(p.Status) {
			active = append(active, p)
		} else {
			done = append(done, p)
		}
	}
	if len(done) > 5 {
		done = done[:5]
	}
	display := append(active, done...)

	if len(display) == 0 {
		fmt.Fprintln(out, colDim+"  no pipelines found"+colReset)
	} else {
		tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "STATUS\tREF\tSOURCE\tSHA\tAGE\tID")
		for _, p := range display {
			sha := p.SHA
			if len(sha) > 8 {
				sha = sha[:8]
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\n",
				glStatusCell(p), p.Ref, p.Source, sha,
				humanAge(time.Since(p.CreatedAt)), p.ID)
		}
		tw.Flush()
	}

	fmt.Fprintf(out, "\nUpdated %s", time.Now().Format("15:04:05"))
	if token == "" {
		fmt.Fprint(out, "  "+colDim+"(no GITLAB_TOKEN — public projects only)"+colReset)
	}
	fmt.Fprintln(out)
}

type glJob struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Stage  string `json:"stage"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

// ExplainPipelineFailure finds the most recent failed pipeline (optionally on
// a branch) and prints each failing job with the tail of its trace.
func ExplainPipelineFailure(out io.Writer, host, projectPath, token, branch string, tail int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pipes, err := fetchPipelines(ctx, host, projectPath, token, branch)
	if err != nil {
		return err
	}
	var failed *Pipeline
	for i, p := range pipes {
		if p.Status == "failed" || p.Status == "canceled" {
			failed = &pipes[i]
			break
		}
	}
	if failed == nil {
		fmt.Fprintln(out, "No failed pipelines found in the last 20.")
		return nil
	}

	fmt.Fprintf(out, "%s✗ pipeline #%d%s  %s  (%s, %s ago)\n", colRed, failed.ID, colReset, failed.Ref, failed.Status, humanAge(time.Since(failed.CreatedAt)))
	fmt.Fprintf(out, "  %s%s%s\n\n", colDim, failed.WebURL, colReset)

	endpoint := fmt.Sprintf("/projects/%s/pipelines/%d/jobs?per_page=100&include_retried=false", url.PathEscape(projectPath), failed.ID)
	resp, err := glGet(ctx, host, endpoint, token)
	if err != nil {
		return err
	}
	var jobs []glJob
	err = json.NewDecoder(resp.Body).Decode(&jobs)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("decode jobs: %w", err)
	}

	shown := 0
	for _, j := range jobs {
		if j.Status == "success" || j.Status == "skipped" || j.Status == "manual" || j.Status == "created" {
			continue
		}
		shown++
		fmt.Fprintf(out, "%sJob:%s [%s] %s  →  %s%s%s\n", colYellow, colReset, j.Stage, j.Name, colDim, j.WebURL, colReset)
		if err := printGLJobTrace(ctx, out, host, projectPath, token, j.ID, tail); err != nil {
			fmt.Fprintf(out, "  %scould not fetch trace: %v%s\n", colDim, err, colReset)
		}
		fmt.Fprintln(out)
	}
	if shown == 0 {
		fmt.Fprintln(out, "  (no failed jobs reported — pipeline may have been cancelled before jobs started)")
	}
	return nil
}

func printGLJobTrace(ctx context.Context, out io.Writer, host, projectPath, token string, jobID int64, tail int) error {
	endpoint := fmt.Sprintf("/projects/%s/jobs/%d/trace", url.PathEscape(projectPath), jobID)
	resp, err := glGet(ctx, host, endpoint, token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimRight(string(body), "\n"), "\n")
	if tail > 0 && len(lines) > tail {
		lines = lines[len(lines)-tail:]
		fmt.Fprintf(out, "  %s... (last %d lines)%s\n", colDim, tail, colReset)
	}
	for _, l := range lines {
		fmt.Fprintf(out, "  %s\n", l)
	}
	return nil
}
