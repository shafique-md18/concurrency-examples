package main

import (
	"fmt"
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

func main() {
	BenchmarkFunction(nonConnectionPoolExample)
}
