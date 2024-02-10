package main

import (
	"fmt"
	"testing"
	"time"
)

func TestMPMCBoundedQueue(t *testing.T) {
	q := NewMPMCBoundedQueue(5)

	// Multiple Producers - 3
	for i := 0; i < 3; i++ {
		go func(id int) {
			for j := 0; j < 5; j++ {
				item := fmt.Sprintf("Producer - %v, Item - %v", id, j)
				q.Enqueue(item)
				fmt.Printf("Enqueued: %v\n", item)
				time.Sleep(time.Millisecond * 200)
			}
		}(i)
	}

	// Multiple Consumers - 2
	for i := 0; i < 2; i++ {
		go func(id int) {
			// infinitely consume
			for {
				item := q.Dequeue()
				fmt.Printf("Consumer - %v, Consumed - %v\n", id, item)
				time.Sleep(time.Millisecond * 200)
			}
		}(i)
	}

	// Allow time to complete execution
	time.Sleep(5 * time.Second)
}
