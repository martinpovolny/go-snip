package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// parseHosts reads /etc/hosts and returns ip → first hostname.
// Loopback addresses and comment lines are skipped.
func parseHosts(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	names := make(map[string]string)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line, _, _ = strings.Cut(line, "#")
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ip := fields[0]
		if strings.HasPrefix(ip, "127.") || ip == "::1" {
			continue
		}
		if _, exists := names[ip]; !exists {
			names[ip] = fields[1]
		}
	}
	return names, sc.Err()
}
