package main

import (
	"database/sql"
	"fmt"
	"sync"
)

type StaticConnectionPool struct {
	m                sync.Mutex
	notFull          sync.Cond
	notEmpty         sync.Cond
	conns            []*sql.DB
	capacity         int
	numOfConnections int
}

func NewStaticConnectionPool(capacity int) *StaticConnectionPool {
	conns := make([]*sql.DB, 0, capacity)
	// statically create all required connections at once
	for i := 0; i < capacity; i++ {
		conns = append(conns, NewDBConnection())
	}
	var mu sync.Mutex
	return &StaticConnectionPool{
		// notFull and notEmpty are conditions with Shared Mutex
		// In the shared mutex, locking one condition implicitly
		// locks the other because they share the same mutex. -> Get and Release don't run in parallel
		notFull:          *sync.NewCond(&mu),
		notEmpty:         *sync.NewCond(&mu),
		conns:            conns,
		capacity:         capacity,
		numOfConnections: capacity,
	}
}

// TODO: We are defining the size of connection pool, but not the max number of connections
// Returns a connection from pool
func (cp *StaticConnectionPool) Get() *sql.DB {
	cp.notEmpty.L.Lock()
	defer cp.notEmpty.L.Unlock()

	if cp.isEmpty() {
		cp.notEmpty.Wait()
	}

	var conn *sql.DB
	if len(cp.conns) == 1 {
		conn = cp.conns[0]
		cp.conns = []*sql.DB{} // Empty the slice
	} else if len(cp.conns) > 1 {
		conn = cp.conns[0]
		cp.conns = cp.conns[1:]
	}
	cp.numOfConnections--
	// signals to Release that pool is not full
	cp.notFull.Signal()
	return conn
}

// Adds connection back to pool
func (cp *StaticConnectionPool) Release(conn *sql.DB) {
	cp.notFull.L.Lock()
	defer cp.notFull.L.Unlock()

	if cp.isFull() {
		cp.notFull.Wait()
	}

	cp.conns = append(cp.conns, conn)
	cp.numOfConnections++
	// signals to Get that pool is not empty
	cp.notEmpty.Signal()
}

func (cp *StaticConnectionPool) CleanUp() {
	connectionPool := cp.conns
	cp.conns = nil
	cp.numOfConnections = 0

	for idx := range connectionPool {
		connectionPool[idx].Close()
	}

	fmt.Println("All connections closed and connection pool set to nil!")
}

func (cp *StaticConnectionPool) isEmpty() bool {
	return len(cp.conns) == 0
}

func (cp *StaticConnectionPool) isFull() bool {
	return len(cp.conns) == cp.capacity
}
