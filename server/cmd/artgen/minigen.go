package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// minigen is a tiny client for the async /jobs API of the minigen service.
type minigen struct {
	base, key string
	http      *http.Client
}

type jobOut struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Error   string `json:"error"`
	ResetAt string `json:"reset_at"`
	Result  *struct {
		Status string `json:"status"`
		Media  []struct {
			Kind      string `json:"kind"`
			SourceURL string `json:"source_url"`
			StreamURL string `json:"stream_url"`
		} `json:"media"`
	} `json:"result"`
}

// loadSecrets reads KEY=VALUE lines (docs/private/secret); environment variables win.
func loadSecrets(path string) map[string]string {
	out := map[string]string{}
	if f, err := os.Open(path); err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if k, v, ok := strings.Cut(strings.TrimSpace(sc.Text()), "="); ok {
				out[k] = v
			}
		}
	}
	for _, k := range []string{"MINIGEN_URL", "MINIGEN_SERVICE_API_KEY"} {
		if v := os.Getenv(k); v != "" {
			out[k] = v
		}
	}
	return out
}

func newMinigen(secrets map[string]string) (*minigen, error) {
	m := &minigen{base: strings.TrimRight(secrets["MINIGEN_URL"], "/"), key: secrets["MINIGEN_SERVICE_API_KEY"],
		http: &http.Client{Timeout: 120 * time.Second}}
	if m.base == "" || m.key == "" {
		return nil, errors.New("missing MINIGEN_URL / MINIGEN_SERVICE_API_KEY (docs/private/secret)")
	}
	return m, nil
}

func (m *minigen) do(method, path string, body any) ([]byte, error) {
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rd = bytes.NewReader(b)
	}
	url := path
	if strings.HasPrefix(path, "/") {
		url = m.base + path
	}
	req, err := http.NewRequest(method, url, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", m.key)
	req.Header.Set("content-type", "application/json")
	resp, err := m.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s: HTTP %d %s", method, path, resp.StatusCode, truncate(string(data), 200))
	}
	return data, err
}

func (m *minigen) submit(prompt string) (*jobOut, error) {
	data, err := m.do("POST", "/jobs", map[string]any{
		"prompt": prompt, "res-type": "image", "media": "url", "remove_watermark": true,
	})
	if err != nil {
		return nil, err
	}
	var j jobOut
	return &j, json.Unmarshal(data, &j)
}

func (m *minigen) get(id string) (*jobOut, error) {
	data, err := m.do("GET", "/jobs/"+id, nil)
	if err != nil {
		return nil, err
	}
	var j jobOut
	return &j, json.Unmarshal(data, &j)
}

// imageURL prefers the service's stream URL (watermark-cleaned bytes).
func (j *jobOut) imageURL() string {
	if j.Result == nil {
		return ""
	}
	for _, md := range j.Result.Media {
		if md.Kind != "image" {
			continue
		}
		if md.StreamURL != "" {
			return md.StreamURL
		}
		return md.SourceURL
	}
	return ""
}

func (m *minigen) download(url, dest string) error {
	// The service may report an http:// stream URL; reuse our base scheme/host.
	if i := strings.Index(url, "/media/"); i >= 0 {
		url = m.base + url[i:]
	}
	data, err := m.do("GET", url, nil)
	if err != nil {
		return err
	}
	if len(data) < 1000 {
		return fmt.Errorf("download too small (%d bytes)", len(data))
	}
	tmp := dest + ".part"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dest)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
