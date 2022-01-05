package main

import (
	"context"
	"flag"
	"log"

	// "github.com/vmihailenco/taskq/example/api_worker"
)

func main() {
	flag.Parse()

	c := context.Background()

	err := QueueFactory.StartConsumers(c)
	if err != nil {
		log.Fatal(err)
	}

	//go LogStats()

	sig := WaitSignal()
	log.Println(sig.String())

	err = QueueFactory.Close()
	if err != nil {
		log.Fatal(err)
	}
}
