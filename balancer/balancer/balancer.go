package balancer

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
)

type Client interface {
	// Weight is unit-less number that determines how much processing capacity can a client be allocated
	// when running in parallel with other clients. The higher the weight, the more capacity the client receives.
	Weight() int
	// Workload returns a channel of work chunks that are ment to be processed through the Server.
	// Client's channel is always filled with work chunks.
	Workload(ctx context.Context) chan int
	Id() int
}

// Server defines methods required to process client's work chunks (requests).
type Server interface {
	// Process takes one work chunk (request) and does something with it. The error can be ignored.
	Process(ctx context.Context, workChunk int) error
}

// Balancer makes sure the Server is not smashed with incoming requests (work chunks) by only enabling certain number
// of parallel requests processed by the Server. Imagine there's a SLO defined, and we don't want to make the expensive
// service people angry.
//
// If implementing more advanced balancer, ake sure to correctly assign processing capacity to a client based on other
// clients currently in process.
// To give an example of this, imagine there's a maximum number of work chunks set to 100 and there are two clients
// registered, both with the same priority. When they are both served in parallel, each of them gets to send
// 50 chunks at the same time.
// In the same scenario, if there were two clients with priority 1 and one client with priority 2, the first
// two would be allowed to send 25 requests and the other one would send 50. It's likely that the one sending 50 would
// be served faster, finishing the work early, meaning that it would no longer be necessary that those first two
// clients only send 25 each but can and should use the remaining capacity and send 50 again.
type Balancer struct {
	maxParallel int32
	clients     []*clientWorkload
	server      Server

	mux  sync.Mutex
	once sync.Once
	cond *sync.Cond
}

type clientWorkload struct {
	client   Client
	workload chan int
}

// New creates a new Balancer instance. It needs the server that it's going to balance for and a maximum number of work
// chunks that can the processor process at a time. THIS IS A HARD REQUIREMENT - THE SERVICE CANNOT PROCESS MORE THAN
// <PROVIDED NUMBER> OF WORK CHUNKS IN PARALLEL.
func New(server Server, maxParallel int32) *Balancer {
	b := &Balancer{
		maxParallel: maxParallel,
		server:      server,
		clients:     make([]*clientWorkload, 0),
	}
	b.cond = sync.NewCond(&b.mux)
	return b
}

type Job struct {
	clientId int
	workload int
}

// weightedChoice implements "Weighted random selection" algorithm to randomly select a client to server respecting the weights.
// if client A has weight of 1 and client B has weight of 3 then B is 3x more likely to be picked.
func (b *Balancer) weightedChoice() int {
	totalWeight := 0
	for _, cw := range b.clients {
		totalWeight += cw.client.Weight()
	}

	randValue := rand.Intn(totalWeight)

	for index, cw := range b.clients {
		if randValue < cw.client.Weight() {
			return index
		}
		randValue -= cw.client.Weight()
	}

	return 0 // not reachable
}

// Register a client to the balancer and start processing its work chunks through provided processor (server).
// For the sake of simplicity, assume that the client has no identifier, meaning the same client can register themselves
// multiple times.
func (b *Balancer) Register(ctx context.Context, client Client) {
	b.mux.Lock()
	fmt.Println("added client: ", client.Id())
	b.clients = append(b.clients, &clientWorkload{
		client:   client,
		workload: client.Workload(ctx),
	})
	b.mux.Unlock()
	b.cond.Signal() // signal that the number of clients changed (non-zero now)

	b.once.Do(func() { // start the workers
		// Create a channel for jobs, no more that maxParallel
		jobs := make(chan Job, b.maxParallel)

		// Start workers
		for w := 0; w < int(b.maxParallel); w++ {
			go func(slotId int, jobs <-chan Job) {
				for workItem := range jobs {
					fmt.Printf("in slot %v doing work %v for client %v\n", slotId, workItem.workload, workItem.clientId)
					_ = b.server.Process(ctx, workItem.workload)
				}
			}(w, jobs)
		}

		// Start the scheduling process
		for {
			var index int
			var cw *clientWorkload
			for {
				b.mux.Lock()
				l := len(b.clients)

				if l > 0 {
					index = b.weightedChoice()
					cw = b.clients[index]
					b.mux.Unlock()
					break
				}

				// no more clients --> wait for more clients to come
				b.cond.Wait()
				continue
			}

			workChunk, ok := <-cw.workload
			if ok {
				fmt.Println("picked client: ", cw.client.Id(), " workChunk: ", workChunk)
				jobs <- Job{cw.client.Id(), workChunk}

			} else { // if !ok then channel was closed, work is done --> remove the client
				b.mux.Lock()
				b.clients = append(b.clients[:index], b.clients[index+1:]...)
				b.mux.Unlock()
			}
		}
	})
}
