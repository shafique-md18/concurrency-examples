package main

import (
	"database/sql"
	"sync"
)

type ConnectionPool struct {
	m        sync.Mutex
	notFull  sync.Cond
	notEmpty sync.Cond
	conns    []*sql.DB
	capacity int
	size     int
}

func NewConnectionPool(capacity int) *ConnectionPool {
	return &ConnectionPool{
		notFull:  *sync.NewCond(&sync.Mutex{}),
		notEmpty: *sync.NewCond(&sync.Mutex{}),
		conns:    make([]*sql.DB, 0, capacity),
		capacity: capacity,
		size:     0,
	}
}
