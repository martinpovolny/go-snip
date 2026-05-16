package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const failuresBeforeAlert = 2 // consecutive failures required before firing a down alert

type monitorState struct {
	consecutive int  // consecutive failure count
	alerted     bool // true = down alert already sent; suppresses repeats until recovery
}

type Notifier struct {
	url      string
	username string
	password string
	client   *http.Client
	store    *Store

	mu    sync.Mutex
	state map[string]*monitorState
}

func NewNotifier(baseURL, topic, username, password string, store *Store) *Notifier {
	return &Notifier{
		url:      strings.TrimRight(baseURL, "/") + "/" + topic,
		username: username,
		password: password,
		client:   &http.Client{Timeout: 10 * time.Second},
		store:    store,
		state:    make(map[string]*monitorState),
	}
}

// Notify updates per-monitor state and fires alerts on confirmed outages and recoveries.
func (n *Notifier) Notify(statuses []MonitorStatus) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, s := range statuses {
		st, seen := n.state[s.Name]
		if !seen {
			st = &monitorState{}
			if n.store != nil {
				st.consecutive, st.alerted = n.store.LoadAlertState(s.Name, failuresBeforeAlert)
			}
			n.state[s.Name] = st
			if st.consecutive == 0 && !st.alerted {
				continue // genuinely first observation — establish baseline silently
			}
			// fall through: resume from persisted state
		}

		if s.Up {
			if st.alerted {
				n.send(s.Name, "recovery",
					"✅ "+s.Name+" is back up",
					fmt.Sprintf("Responded %d in %d ms", s.StatusCode, s.ResponseMs),
					"default", "white_check_mark")
			}
			st.consecutive = 0
			st.alerted = false
		} else {
			st.consecutive++
			if st.consecutive >= failuresBeforeAlert && !st.alerted {
				msg := fmt.Sprintf("Status %d", s.StatusCode)
				if s.Error != "" {
					msg = s.Error
				}
				n.send(s.Name, "down", "🔴 "+s.Name+" is DOWN", msg, "urgent", "rotating_light")
				st.alerted = true
			}
		}
		if n.store != nil {
			if err := n.store.SaveAlertState(s.Name, st.consecutive, st.alerted); err != nil {
				log.Printf("ntfy: save alert state %s: %v", s.Name, err)
			}
		}
	}
}

func (n *Notifier) send(monitor, event, title, message, priority, tags string) {
	req, err := http.NewRequest("POST", n.url, strings.NewReader(message))
	if err != nil {
		log.Printf("ntfy: build request: %v", err)
		return
	}
	req.SetBasicAuth(n.username, n.password)
	req.Header.Set("Title", title)
	req.Header.Set("Priority", priority)
	req.Header.Set("Tags", tags)

	resp, err := n.client.Do(req)
	if err != nil {
		log.Printf("ntfy: send: %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode >= 300 {
			log.Printf("ntfy: unexpected status %d for %q", resp.StatusCode, title)
		}
	}

	if n.store != nil {
		if err := n.store.RecordNotification(monitor, event, title, message); err != nil {
			log.Printf("ntfy: record notification: %v", err)
		}
	}
}
