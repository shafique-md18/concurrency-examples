package main

import (
	"sync"
)

// We are using two conditions so that Enqueue and Dequeue are blocked independently of each other

type MPMCBoundedQueue struct {
	m        sync.Mutex
	data     []interface{}
	notEmpty sync.Cond
	notFull  sync.Cond
	capacity int
}

func NewMPMCBoundedQueue(capacity int) *MPMCBoundedQueue {
	var mu sync.Mutex
	return &MPMCBoundedQueue{
		data:     make([]interface{}, 0, capacity),
		notEmpty: *sync.NewCond(&mu),
		notFull:  *sync.NewCond(&mu),
		capacity: capacity,
	}
}

// Enqueue acquires lock for notFull (Enqueue operations will not run parallely)
// and after enqueing sends signals to notEmpty for dequeuing
func (q *MPMCBoundedQueue) Enqueue(item interface{}) {
	q.notFull.L.Lock()
	defer q.notFull.L.Unlock()

	if len(q.data) == q.capacity {
		q.notFull.Wait()
	}

	q.data = append(q.data, item)
	q.notEmpty.Signal()
}

// Dequeue acquires lock for notEmpty (Dequeue operations will not run parallely)
// and after dequeing sends signal to notFull for enqueing
func (q *MPMCBoundedQueue) Dequeue() interface{} {
	q.notEmpty.L.Lock()
	defer q.notEmpty.L.Unlock()

	if len(q.data) == 0 {
		q.notEmpty.Wait()
	}

	item := q.data[0]
	q.data = q.data[1:len(q.data)]
	q.notFull.Signal()
	return item
}
