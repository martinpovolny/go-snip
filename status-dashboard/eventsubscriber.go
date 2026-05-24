package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type EventSubscriber struct {
	url      string
	user     string
	password string
	store    *Store
	client   *http.Client
}

func NewEventSubscriber(baseURL, topic, user, password string, store *Store) *EventSubscriber {
	return &EventSubscriber{
		url:      strings.TrimRight(baseURL, "/") + "/" + topic + "/json",
		user:     user,
		password: password,
		store:    store,
		client:   &http.Client{}, // no timeout — this is a long-lived streaming connection
	}
}

func (es *EventSubscriber) Start() {
	go func() {
		for {
			if err := es.run(); err != nil {
				log.Printf("events: %v; reconnecting in 30s", err)
				time.Sleep(30 * time.Second)
			}
		}
	}()
}

func (es *EventSubscriber) run() error {
	since := es.store.LastEventID()
	req, err := http.NewRequest("GET", es.url+"?since="+since, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if es.password != "" {
		req.SetBasicAuth(es.user, es.password)
	}

	resp, err := es.client.Do(req)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	type ntfyMsg struct {
		ID       string   `json:"id"`
		Time     int64    `json:"time"`
		Event    string   `json:"event"`
		Message  string   `json:"message"`
		Title    string   `json:"title"`
		Tags     []string `json:"tags"`
		Priority int      `json:"priority"`
	}

	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var msg ntfyMsg
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if msg.Event != "message" {
			continue
		}
		if err := es.store.RecordEvent(msg.ID, msg.Time, msg.Title, msg.Message, msg.Tags, msg.Priority); err != nil {
			log.Printf("events: record %s: %v", msg.ID, err)
		}
	}
	return sc.Err()
}
