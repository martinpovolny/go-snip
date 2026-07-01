// apitest: smoke-tests API credentials from .env and prints results.
// Only read-only (GET) calls are made — no data is modified.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	edookit "github.com/martinpovolny/go-snip/edookit"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Fprintln(os.Stderr, "warning: no .env file loaded:", err)
	}

	baseURL := os.Getenv("API_URL")
	user := os.Getenv("API_USER")
	pass := os.Getenv("API_PASSWORD")

	if baseURL == "" {
		fmt.Fprintln(os.Stderr, "API_URL is not set (check .env)")
		os.Exit(1)
	}

	c := edookit.New(baseURL, user, pass)
	fmt.Printf("main:   %s\nlegacy: %s\nuser:   %s\n\n", c.BaseURL, c.LegacyBaseURL, user)

	today := time.Now().Format("2006-01-02")

	type check struct {
		name string
		fn   func() (any, error)
	}

	checks := []check{
		// --- Změnový rozvrh (legacy -login domain, known working) ---
		{"schedule changes today", func() (any, error) {
			return c.ListScheduleChanges(edookit.ScheduleChangeOpts{From: today, To: today})
		}},
		{"schedule changes this week", func() (any, error) {
			mon := monday(time.Now())
			fri := mon.AddDate(0, 0, 4)
			return c.ListScheduleChanges(edookit.ScheduleChangeOpts{
				From: mon.Format("2006-01-02"),
				To:   fri.Format("2006-01-02"),
			})
		}},

		// --- Vyhledání osob: criterion is an Edookit ID integer, not a name ---
		{"person by ID 200", func() (any, error) { return c.SearchPerson("200") }},
		{"person by ID 237", func() (any, error) { return c.SearchPerson("237") }},

		// --- Other modules (likely need different creds/enablement) ---
		{"rooms", func() (any, error) { return c.ListRooms() }},
		{"student stats", func() (any, error) { return c.ListStudentStats("") }},
	}

	ok, fail := 0, 0
	for _, ch := range checks {
		result, err := ch.fn()
		if err != nil {
			fmt.Printf("  FAIL  %s\n        %v\n", ch.name, err)
			fail++
			continue
		}
		b, _ := json.Marshal(result)
		preview := string(b)
		if len(preview) > 160 {
			preview = preview[:160] + "..."
		}
		fmt.Printf("  OK    %s\n        %s\n", ch.name, preview)
		ok++
	}

	fmt.Printf("\n%d ok, %d failed\n", ok, fail)
	if fail > 0 {
		os.Exit(1)
	}
}

func monday(t time.Time) time.Time {
	for t.Weekday() != time.Monday {
		t = t.AddDate(0, 0, -1)
	}
	return t
}
