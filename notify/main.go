package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lib/pq"
)

type WsClient struct {
	conn *websocket.Conn
	send chan string
	hub  *Hub
}

type Hub struct {
	register   chan *WsClient
	unregister chan *WsClient
	clients    map[*WsClient]bool
}

func newHub() *Hub {
	return &Hub{
		register:   make(chan *WsClient),
		unregister: make(chan *WsClient),
		clients:    make(map[*WsClient]bool),
	}
}

func (h *Hub) waitForNotification(l *pq.Listener, shutdownChan <-chan interface{}) {
	for {
		select {
		case client := <-h.register: // A new client has connected.
			h.clients[client] = true
			fmt.Println("length of channels: ", len(h.clients))

		case client := <-h.unregister: // A client has disconnected.
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case n := <-l.Notify: // A PG Notify is received.
			if n == nil {
				// when PG restarts a nil notifications comes here
				fmt.Println("Received nil notification")
			} else {
				fmt.Println("Received data from PG channel [", n.Channel, "] :")
				// Prepare notification payload for pretty print
				var prettyJSON bytes.Buffer
				err := json.Indent(&prettyJSON, []byte(n.Extra), "", "\t")
				if err != nil {
					fmt.Println("Error processing JSON: ", err)
					return
				}
				fmt.Println(string(prettyJSON.String()))

				// Send notification to all clients
				fmt.Println("Length of channels: ", len(h.clients))
				for c := range h.clients {
					fmt.Println("\tSending data to channel [", c.send, "]")
					c.send <- string(n.Extra)
				}
				fmt.Println("Done with sending data to channel")
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

func (c *WsClient) writeLoop() {
	defer func() {
		c.conn.Close()
		c.hub.unregister <- c
	}()

	// Send message to client when data is available in the channel.
	for {
		select {
		case msg := <-c.send:
			fmt.Println("\tWriting to websocket")
			message := []byte(msg)
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Println("Websocket write failed:", err)
				return
			}
		case <-time.After(30 * time.Second):
			fmt.Println("tick")
			message := []byte("tick")
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Println("Websocket write failed:", err)
				return
			}
		}
		fmt.Println("\tDone with writing to websocket")
	}
}

func serveHttp(hub *Hub) {
	var upgrader = websocket.Upgrader{}

	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Print("HTTP upgrade failed: ", err)
			return
		}

		fmt.Println("New websocket connection")
		client := &WsClient{conn: conn, send: make(chan string), hub: hub}
		client.hub.register <- client
		go client.writeLoop()
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "client.html")
	})

	fmt.Println("Starting HTTP server at localhost:8080")
	http.ListenAndServe(":8080", nil)
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

	hub := newHub()

	go func() {
		hub.waitForNotification(listener, shutdownChan)
	}()

	go func() {
		serveHttp(hub)
	}()

	time.Sleep(5 * time.Minute)
	fmt.Println("Stopping the listener goroutine...")
	close(shutdownChan)

	time.Sleep(10 * time.Second)
	fmt.Println("Exiting main...")
}
