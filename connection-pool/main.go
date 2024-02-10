package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func BenchmarkFunction(fn func()) float64 {
	startTime := time.Now()
	fn()
	elapsedTime := time.Since(startTime).Seconds() * 1000 // Convert to milliseconds
	fmt.Printf("Your function took %.2f milliseconds to run.\n", elapsedTime)
	return elapsedTime
}

func nonConnectionPoolExample() {
	numOfConnections := 150
	var wg sync.WaitGroup
	for i := 0; i < numOfConnections; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			db := NewDBConnection()
			_, err := db.Exec("SELECT SLEEP(0.01);")
			if err != nil {
				panic(err)
			}

			db.Close()
		}(i)
	}
	wg.Wait()
	fmt.Println("Number of connections created:", numOfConnections)
}

func staticConnectionPoolExample() {
	numOfConnections := 500
	connectionPool := NewStaticConnectionPool(10)
	var wg sync.WaitGroup
	for i := 0; i < numOfConnections; i++ {
		wg.Add(1)
		go func(id int) {
			db := connectionPool.Get()
			defer func(db *sql.DB) {
				connectionPool.Release(db)
				wg.Done()
			}(db)

			_, err := db.Exec("SELECT SLEEP(0.01);")
			if err != nil {
				panic(err)
			}
		}(i)
	}
	wg.Wait()
	fmt.Println("Number of connections created:", len(connectionPool.conns))
	connectionPool.CleanUp()
}

func dynamicConnectionPoolExample() {
	numOfConnections := 500
	connectionPool := NewDynamicConnectionPool(10)
	var wg sync.WaitGroup
	for i := 0; i < numOfConnections; i++ {
		wg.Add(1)
		go func(id int) {
			db := connectionPool.Get()
			defer func(db *sql.DB) {
				// randomly either release or discard a connection
				if rand.Intn(2) == 0 {
					connectionPool.Release(db)
				} else {
					connectionPool.Discard(db)
				}
				wg.Done()
			}(db)

			_, err := db.Exec("SELECT SLEEP(0.01);")
			if err != nil {
				panic(err)
			}
		}(i)
	}
	wg.Wait()
	fmt.Println("Number of open connections:", connectionPool.numOfOpenConnections, "==", len(connectionPool.conns))
	connectionPool.CleanUp()
}

func main() {
	// BenchmarkFunction(nonConnectionPoolExample)
	// BenchmarkFunction(staticConnectionPoolExample)
	BenchmarkFunction(dynamicConnectionPoolExample)
}
