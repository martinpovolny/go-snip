package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Monitor struct {
	Name     string
	URL      string
	Username string
	Password string
}

type MonitorStatus struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	Up         bool   `json:"up"`
	StatusCode int    `json:"status_code,omitempty"`
	ResponseMs int64  `json:"response_ms,omitempty"`
	Error      string `json:"error,omitempty"`
	CheckedAt  int64  `json:"checked_at"`
}

func loadMonitors(path string) ([]Monitor, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	var monitors []Monitor
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "http")
		if idx < 0 {
			continue
		}
		name := strings.TrimSpace(line[:idx])
		rest := strings.Fields(line[idx:])
		url := rest[0]
		var user, pass string
		if len(rest) > 1 {
			user, pass, _ = strings.Cut(rest[1], ":")
		}
		monitors = append(monitors, Monitor{Name: name, URL: url, Username: user, Password: pass})
	}
	return monitors, sc.Err()
}

type MonitorCollector struct {
	mu        sync.Mutex
	monitors  []Monitor
	ttl       time.Duration
	cached    []MonitorStatus
	fetchedAt time.Time
	client    *http.Client
}

func newMonitorCollector(monitors []Monitor, ttl time.Duration) *MonitorCollector {
	return &MonitorCollector{
		monitors: monitors,
		ttl:      ttl,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (mc *MonitorCollector) Get() []MonitorStatus {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.cached != nil && time.Since(mc.fetchedAt) < mc.ttl {
		return mc.cached
	}
	mc.cached = mc.checkAll()
	mc.fetchedAt = time.Now()
	return mc.cached
}

func (mc *MonitorCollector) checkAll() []MonitorStatus {
	type result struct {
		i      int
		status MonitorStatus
	}
	ch := make(chan result, len(mc.monitors))
	for i, m := range mc.monitors {
		go func(i int, m Monitor) {
			ch <- result{i, mc.check(m)}
		}(i, m)
	}
	statuses := make([]MonitorStatus, len(mc.monitors))
	for range mc.monitors {
		r := <-ch
		statuses[r.i] = r.status
	}
	return statuses
}

func (mc *MonitorCollector) check(m Monitor) MonitorStatus {
	req, _ := http.NewRequest("GET", m.URL, nil)
	if m.Username != "" {
		req.SetBasicAuth(m.Username, m.Password)
	}
	start := time.Now()
	resp, err := mc.client.Do(req)
	elapsed := time.Since(start).Milliseconds()
	s := MonitorStatus{Name: m.Name, URL: m.URL, CheckedAt: time.Now().Unix()}
	if err != nil {
		s.Error = err.Error()
		return s
	}
	resp.Body.Close()
	s.StatusCode = resp.StatusCode
	s.ResponseMs = elapsed
	s.Up = resp.StatusCode < 400
	return s
}
