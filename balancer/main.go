package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"gitlab.com/kiwicom/search-team/balancer/balancer"
	"gitlab.com/kiwicom/search-team/balancer/client"
	"gitlab.com/kiwicom/search-team/balancer/service"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	//maxParallel := int32(50 + rand.Intn(150))
	maxParallel := int32(50 + rand.Intn(15))
	fmt.Println("maxParallel: ", maxParallel)
	b := balancer.New(&service.TheExpensiveFragileService{}, maxParallel)

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	//nbClients := 1 + rand.Intn(5)
	nbClients := 2 + rand.Intn(5)
	fmt.Println("maxParallel: ", maxParallel, " clients: ", nbClients)
	for i := 0; i < nbClients; i++ {
		go func(j int) {
			//workload := 500 + rand.Intn(1000)
			workload := 50 + rand.Intn(100)
			weight := 1 + rand.Intn(3)

			//time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
			time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
			fmt.Printf("NEW client %v, workload: %v, weight: %v\n", j, workload, weight)
			b.Register(ctx, client.New(workload, weight, j))
		}(i)
	}
	<-ctx.Done()
}
