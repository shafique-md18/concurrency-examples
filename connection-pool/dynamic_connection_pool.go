package main

import (
	"database/sql"
	"fmt"
	"sync"
)

// TODO: Fix bug which causes panic when Discard is used.

type DynamicConnectionPool struct {
	m                    sync.Mutex
	notFull              sync.Cond
	notEmpty             sync.Cond
	conns                []*sql.DB
	capacity             int
	numOfOpenConnections int
}

func NewDynamicConnectionPool(capacity int) *DynamicConnectionPool {
	var mu sync.Mutex
	return &DynamicConnectionPool{
		notFull:              *sync.NewCond(&mu),
		notEmpty:             *sync.NewCond(&mu),
		conns:                make([]*sql.DB, 0, capacity),
		capacity:             capacity,
		numOfOpenConnections: 0,
	}
}

// Returns a connection from pool
func (cp *DynamicConnectionPool) Get() *sql.DB {
	cp.notEmpty.L.Lock()
	defer cp.notEmpty.L.Unlock()

	// we cannot create any more connections, capacity exhausted and all connections are being used
	if cp.numOfOpenConnections == cp.capacity && cp.isEmpty() {
		cp.notEmpty.Wait()
	}

	// if capacity is still not exhausted but queue is empty we can create a new connection
	if cp.numOfOpenConnections < cp.capacity && cp.isEmpty() {
		cp.conns = append(cp.conns, NewDBConnection())
		cp.numOfOpenConnections++
	}

	var conn *sql.DB
	if len(cp.conns) == 1 {
		conn = cp.conns[0]
		cp.conns = []*sql.DB{} // Empty the slice
	} else if len(cp.conns) > 1 {
		conn = cp.conns[0]
		cp.conns = cp.conns[1:]
	} else {
		panic("Connection pool is empty unexpectedly.")
	}

	// signals to Release that pool is not full
	cp.notFull.Signal()
	return conn
}

// Adds connection back to pool
func (cp *DynamicConnectionPool) Release(conn *sql.DB) {
	cp.notFull.L.Lock()
	defer cp.notFull.L.Unlock()

	if cp.isFull() {
		fmt.Println(len(cp.conns), cp.capacity)
		panic("Connection pool is full, cannot release the connection.")
	}

	cp.conns = append(cp.conns, conn)
	// signals to Get that pool is not empty
	cp.notEmpty.Signal()
}

// Discards an open connection
func (cp *DynamicConnectionPool) Discard(conn *sql.DB) {
	// acquire lock so that it doesn't run parallel to Get or Release
	cp.m.Lock()
	defer cp.m.Unlock()

	conn.Close()
	cp.numOfOpenConnections--
	// signal Get which can create a new connection if required
	cp.notEmpty.Signal()
	// signal Release which can add any connection back to pool
	cp.notFull.Signal()
}

func (cp *DynamicConnectionPool) CleanUp() {
	cp.m.Lock()
	defer cp.m.Unlock()

	for idx := range cp.conns {
		cp.conns[idx].Close()
	}
	cp.conns = nil
	cp.numOfOpenConnections = 0

	fmt.Println("All connections closed and connection pool set to nil!")
}

func (cp *DynamicConnectionPool) isEmpty() bool {
	return len(cp.conns) == 0
}

func (cp *DynamicConnectionPool) isFull() bool {
	return len(cp.conns) == cp.capacity
}
