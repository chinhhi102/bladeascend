// Command artgen generates game art with the minigen /jobs API and post-processes it
// into assets/ following docs/ART_PROMPTS.md.
//
//	go run ./cmd/artgen plan                 # list jobs
//	go run ./cmd/artgen run -c 2 -p 2        # generate missing images up to priority 2
//	go run ./cmd/artgen prep                 # remove backgrounds, slice sheets, build atlases + manifest
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"kiemxoay/server/internal/config"
)

const pollEvery = 6 * time.Second

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: artgen plan|run|prep [flags]")
		os.Exit(2)
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	conc := fs.Int("c", 2, "concurrent minigen jobs")
	maxPri := fs.Int("p", 9, "only jobs with priority <= p")
	only := fs.String("only", "", "only jobs whose name contains this text (comma separated)")
	redo := fs.Bool("redo", false, "regenerate even if the raw image exists")
	limit := fs.Int("n", 60, "max images to generate this run (minigen allows ~60-100 per 5 hours)")
	_ = fs.Parse(os.Args[2:])

	assets := config.FindRepoFile("assets")
	jobs := filterJobs(allJobs(), *maxPri, *only)
	switch os.Args[1] {
	case "plan":
		plan(jobs, assets)
	case "run":
		secrets := loadSecrets(config.FindRepoFile("docs/private/secret"))
		mg, err := newMinigen(secrets)
		if err != nil {
			log.Fatal(err)
		}
		run(mg, jobs, assets, runOpts{conc: *conc, redo: *redo, limit: *limit})
	case "prep":
		prepAll(jobs, assets)
	case "status":
		writeStatus(jobs, assets, filepath.Join(filepath.Dir(assets), "docs", "ASSETS_STATUS.md"))
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
}

func filterJobs(jobs []Job, maxPri int, only string) []Job {
	var out []Job
	for _, j := range jobs {
		if j.Priority > maxPri || (only != "" && !matchesAny(j.Name, only)) {
			continue
		}
		out = append(out, j)
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Priority < out[b].Priority })
	return out
}

func matchesAny(name, list string) bool {
	for _, s := range strings.Split(list, ",") {
		if s != "" && strings.Contains(name, s) {
			return true
		}
	}
	return false
}

func rawPath(assets string, j Job) string { return filepath.Join(assets, "_raw", j.Name+".png") }

// writeStatus lists which art exists and which is still a placeholder (M6 deliverable).
func writeStatus(jobs []Job, assets, out string) {
	var done, missing []string
	for _, j := range jobs {
		line := fmt.Sprintf("| `%s` | p%d | %s |", j.Name, j.Priority, jobOutputs(j))
		if _, err := os.Stat(rawPath(assets, j)); err == nil {
			done = append(done, line)
		} else {
			missing = append(missing, line)
		}
	}
	var b strings.Builder
	b.WriteString("# Tình trạng art\n\nTạo lại file này: `cd server && go run ./cmd/artgen status`.\n")
	b.WriteString("Asset thiếu → game tự vẽ placeholder (hình tròn + chữ). Sinh thêm: `go run ./cmd/artgen run -p 2 -n 30` rồi `go run ./cmd/artgen prep`.\n")
	b.WriteString("Nếu tự tải ảnh từ Gemini: lưu thành `assets/_raw/<tên job>.png` rồi chạy `prep`.\n\n")
	fmt.Fprintf(&b, "## Còn thiếu (%d ảnh)\n\n| Job | Ưu tiên | Ra file |\n|---|---|---|\n%s\n\n", len(missing), strings.Join(missing, "\n"))
	fmt.Fprintf(&b, "## Đã có (%d ảnh)\n\n| Job | Ưu tiên | Ra file |\n|---|---|---|\n%s\n", len(done), strings.Join(done, "\n"))
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d done, %d missing → %s\n", len(done), len(missing), out)
}

func jobOutputs(j Job) string {
	if j.Mode != modeSheet {
		return j.Out
	}
	var parts []string
	for _, s := range j.Segs {
		parts = append(parts, fmt.Sprintf("%s (%s ×%d)", s.Out, s.Anim, s.Frames))
	}
	return strings.Join(parts, ", ")
}

func plan(jobs []Job, assets string) {
	done := 0
	for _, j := range jobs {
		mark := " "
		if _, err := os.Stat(rawPath(assets, j)); err == nil {
			mark, done = "x", done+1
		}
		fmt.Printf("[%s] p%d %-32s → %s\n", mark, j.Priority, j.Name, j.Out)
	}
	fmt.Printf("%d jobs, %d generated\n", len(jobs), done)
}

// state remembers submitted job ids so a restarted run resumes polling instead of resubmitting.
type state struct {
	mu   sync.Mutex
	path string
	IDs  map[string]string `json:"ids"`
}

func loadState(path string) *state {
	s := &state{path: path, IDs: map[string]string{}}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, s)
	}
	return s
}

func (s *state) set(name, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id == "" {
		delete(s.IDs, name)
	} else {
		s.IDs[name] = id
	}
	b, _ := json.MarshalIndent(s, "", " ")
	_ = os.WriteFile(s.path, b, 0o644)
}

func (s *state) get(name string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.IDs[name]
}

type runOpts struct {
	conc, limit int
	redo        bool
}

func run(mg *minigen, jobs []Job, assets string, o runOpts) {
	if err := os.MkdirAll(filepath.Join(assets, "_raw"), 0o755); err != nil {
		log.Fatal(err)
	}
	st := loadState(filepath.Join(assets, "_raw", "state.json"))
	queue := make(chan Job)
	var wg sync.WaitGroup
	var okN, failN int
	var mu sync.Mutex
	for w := 0; w < o.conc; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range queue {
				err := generate(mg, st, j, rawPath(assets, j))
				mu.Lock()
				if err != nil {
					failN++
					log.Printf("FAIL %s: %v", j.Name, err)
				} else {
					okN++
					log.Printf("ok   %s (%d ok, %d fail)", j.Name, okN, failN)
				}
				mu.Unlock()
			}
		}()
	}
	sent := 0
	for _, j := range jobs {
		if _, err := os.Stat(rawPath(assets, j)); err == nil && !o.redo {
			continue
		}
		if sent >= o.limit {
			log.Printf("reached -n %d; run again later for the rest", o.limit)
			break
		}
		queue <- j
		sent++
	}
	close(queue)
	wg.Wait()
	log.Printf("done: %d ok, %d failed", okN, failN)
}

// generate submits (or resumes) one job and downloads its image, retrying on quota/account errors.
func generate(mg *minigen, st *state, j Job, dest string) error {
	var lastErr error
	for attempt := 0; attempt < 6; attempt++ {
		id := st.get(j.Name)
		if id == "" {
			out, err := mg.submit(j.Prompt)
			if err != nil {
				lastErr = err
				time.Sleep(backoff(attempt))
				continue
			}
			id = out.ID
			st.set(j.Name, id)
		}
		out, err := waitJob(mg, id)
		if err == nil {
			err = mg.download(out.imageURL(), dest)
		}
		if err == nil {
			st.set(j.Name, "")
			return nil
		}
		lastErr = err
		st.set(j.Name, "") // resubmit next attempt
		d, quota := retryDelay(out, attempt)
		if quota {
			attempt-- // waiting for quota is not a failed attempt
		}
		time.Sleep(d)
	}
	return lastErr
}

func waitJob(mg *minigen, id string) (*jobOut, error) {
	deadline := time.Now().Add(10 * time.Minute)
	for time.Now().Before(deadline) {
		out, err := mg.get(id)
		if err != nil {
			return nil, err
		}
		switch out.Status {
		case "completed":
			if out.imageURL() == "" {
				return out, fmt.Errorf("completed without image")
			}
			return out, nil
		case "failed":
			return out, fmt.Errorf("job failed: %s", out.Error)
		default:
			time.Sleep(pollEvery)
		}
	}
	return nil, fmt.Errorf("job %s timed out", id)
}

func backoff(attempt int) time.Duration { return time.Duration(15*(attempt+1)) * time.Second }

// retryDelay waits until the account quota unlocks when the service tells us when.
func retryDelay(out *jobOut, attempt int) (time.Duration, bool) {
	if out != nil && out.ResetAt != "" {
		if t, err := time.Parse(time.RFC3339, out.ResetAt); err == nil {
			if d := time.Until(t) + 30*time.Second; d > 0 && d < 6*time.Hour {
				log.Printf("quota locked until %s, waiting %s", out.ResetAt, d.Round(time.Second))
				return d, true
			}
		}
	}
	return backoff(attempt), false
}
