package main

import (
	"context"
	"flag"
	"fmt"
    	"math/rand"
	"log"
	"time"
)

func main() {
	flag.Parse()

	// go LogStats()

	/*
	go func() {
		for {
			err := MainQueue.Add(CountTask.WithArgs(context.Background()))
			if err != nil {
				log.Fatal(err)
			}
			IncrLocalCounter()
		}
	}()
	*/

	go func() {
		for {
			key :=rand.Intn(1000)
			fmt.Printf("creating a new task with key %v\n", key) 

			err := MainQueue.Add(FailingTask.WithArgs(context.Background(), key))
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(5*time.Second);
		}
	}()

	sig := WaitSignal()
	log.Println(sig.String())
}
