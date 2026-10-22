// Command artgen generates game art with the minigen /jobs API and post-processes it
// into assets/ following docs/ART_PROMPTS.md.
//
//	go run ./cmd/artgen plan                 # list jobs
//	go run ./cmd/artgen run -p 2             # send each missing prompt once, one at a time (-gap 30s)
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
	conc := fs.Int("c", 1, "concurrent minigen jobs (keep 1: minigen answers \"too many requests\")")
	gap := fs.Duration("gap", 30*time.Second, "pause between prompts")
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
		run(mg, jobs, assets, runOpts{conc: *conc, gap: *gap, redo: *redo, limit: *limit})
	case "prep":
		prepAll(jobs, assets)
	case "prompts":
		writePrompts(jobs, assets, filepath.Join(filepath.Dir(assets), "docs", "ART_JOBS.md"))
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

// writePrompts exports every missing job's prompt so images can be made with any tool and
// dropped into assets/_raw/<job>.png before running `artgen prep`.
func writePrompts(jobs []Job, assets, out string) {
	var b strings.Builder
	b.WriteString("# Danh sách prompt cần sinh ảnh\n\n")
	b.WriteString("Tạo lại file này: `cd server && go run ./cmd/artgen prompts` (chỉ liệt kê ảnh còn thiếu, xếp theo ưu tiên).\n\n")
	b.WriteString("**Cách dùng**\n")
	b.WriteString("1. Dán prompt vào công cụ sinh ảnh bất kỳ (Gemini, ChatGPT, Midjourney…). Nếu công cụ trả lời bằng chữ, thêm vào đầu: `Generate an image:`.\n")
	b.WriteString("2. Ảnh nên khổ ngang 16:9 (hoặc vuông với ảnh 1 món), độ phân giải càng cao càng tốt.\n")
	b.WriteString("3. Lưu đúng tên ở dòng **Lưu thành** (đuôi `.png`; ảnh JPEG đổi tên thành .png vẫn được).\n")
	b.WriteString("4. Chạy `cd server && go run ./cmd/artgen prep` → tool tự xóa nền magenta/trắng, cắt frame, căn pivot, ghép atlas.\n")
	b.WriteString("5. Sheet lệch nhẹ vẫn dùng được (tool cắt theo vùng pixel liền nhau). Nhân vật sai số frame vẫn chạy; chỉ sinh lại khi hình dính chồng lên nhau.\n\n")
	n := 0
	for _, j := range jobs {
		if _, err := os.Stat(rawPath(assets, j)); err == nil {
			continue
		}
		n++
		fmt.Fprintf(&b, "## %d. `%s` · ưu tiên %d\n\n", n, j.Name, j.Priority)
		fmt.Fprintf(&b, "- **Lưu thành**: `assets/_raw/%s.png`\n", j.Name)
		fmt.Fprintf(&b, "- **Bố cục**: %s\n", jobLayout(j))
		fmt.Fprintf(&b, "- **Ra file game**: %s\n\n", jobOutputs(j))
		fmt.Fprintf(&b, "```text\n%s\n```\n\n", j.Prompt)
	}
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d prompts → %s\n", n, out)
}

func jobLayout(j Job) string {
	switch j.Mode {
	case modeSheet:
		parts := make([]string, len(j.Rows))
		for i, r := range j.Rows {
			parts[i] = fmt.Sprintf("hàng %d: %d hình", i+1, r)
		}
		return "sprite sheet, " + strings.Join(parts, " · ") + ", nền magenta #FF00FF"
	case modeTile:
		return "texture vuông lặp liền mạch, phủ kín ảnh"
	case modeWide:
		return "tranh ngang 16:9, có nền"
	default:
		return "1 hình ở giữa, nền magenta #FF00FF"
	}
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
	gap         time.Duration
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
		if sent > 0 {
			time.Sleep(o.gap)
		}
		queue <- j
		sent++
	}
	close(queue)
	wg.Wait()
	log.Printf("done: %d ok, %d failed", okN, failN)
}

// generate sends the prompt to minigen once and downloads its image. A prompt is sent
// again only when minigen refused it for quota (nothing was drawn): we then wait until
// reset_at. A job that timed out keeps its id so the next run resumes it.
func generate(mg *minigen, st *state, j Job, dest string) error {
	for waits := 0; ; waits++ {
		reset, err := generateOnce(mg, st, j, dest)
		if reset == "" || waits >= 3 {
			return err
		}
		waitQuota(reset)
	}
}

// generateOnce returns the quota reset time when the job was refused for quota.
func generateOnce(mg *minigen, st *state, j Job, dest string) (string, error) {
	id := st.get(j.Name)
	if id == "" {
		out, err := submitOnce(mg, j.Prompt)
		if err != nil {
			return "", err
		}
		id = out.ID
		st.set(j.Name, id)
	}
	out, err := waitJob(mg, id)
	if err != nil {
		if out == nil || out.Status != "failed" {
			return "", err // id kept: resumed next run
		}
		st.set(j.Name, "")
		if strings.Contains(strings.ToLower(out.Error), "quota") {
			return out.ResetAt, err
		}
		return "", err
	}
	if err := mg.download(out.imageURL(), dest); err != nil {
		return "", err // id kept: the next run downloads it again without a new prompt
	}
	st.set(j.Name, "")
	return "", nil
}

func waitQuota(reset string) {
	d := 10 * time.Minute
	if t, err := time.Parse(time.RFC3339, reset); err == nil {
		d = max(time.Until(t)+30*time.Second, time.Minute)
	}
	log.Printf("quota locked until %s, waiting %s", reset, d.Round(time.Second))
	time.Sleep(d)
}

// submitOnce retries only while the service refuses to accept the job.
func submitOnce(mg *minigen, prompt string) (*jobOut, error) {
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		out, err := mg.submit(prompt)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if !rejected(err) {
			return nil, err
		}
		d := backoff(attempt) * 4
		log.Printf("minigen busy (%v), waiting %s before sending again", truncate(err.Error(), 80), d)
		time.Sleep(d)
	}
	return nil, lastErr
}

func rejected(err error) bool {
	m := strings.ToLower(err.Error())
	return strings.Contains(m, "http 429") || strings.Contains(m, "too many") || strings.Contains(m, "quota")
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
