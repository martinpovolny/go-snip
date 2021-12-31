package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

func waitForNotification(l *pq.Listener, shutdownChan <-chan interface{}) {
	for {
		select {
		case n := <-l.Notify:
			if n == nil {
				// when PG restarts a nil notifications comes here
				fmt.Println("Received nil notification")
			} else {
				fmt.Println("Received data from channel [", n.Channel, "] :")
				// Prepare notification payload for pretty print
				var prettyJSON bytes.Buffer
				err := json.Indent(&prettyJSON, []byte(n.Extra), "", "\t")
				if err != nil {
					fmt.Println("Error processing JSON: ", err)
					return
				}
				fmt.Println(string(prettyJSON.String()))
			}
		case <-time.After(90 * time.Second):
			fmt.Println("Received no events for 90 seconds, checking connection")
			if nil != l.Ping() {
				// I have experimentally verified that there's no need to reconnect.
				fmt.Println("Connection lost.")
			}
		case <-shutdownChan:
			fmt.Println("Shutdown request received")
			return
		}
	}
}

func main() {
	var conninfo string = "dbname=webapp user=webapp password=webapp sslmode=disable"

	_, err := sql.Open("postgres", conninfo)
	if err != nil {
		panic(err)
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	listener := pq.NewListener(conninfo, 10*time.Second, time.Minute, reportProblem)
	err = listener.Listen("events")
	if err != nil {
		panic(err)
	}

	fmt.Println("Start monitoring PostgreSQL...")
	shutdownChan := make(chan interface{})

	go func() {
		waitForNotification(listener, shutdownChan)
	}()

	time.Sleep(5 * time.Minute)
	fmt.Println("Stopping the listener goroutine...")
	close(shutdownChan)

	time.Sleep(10 * time.Second)
	fmt.Println("Exiting main...")
}
