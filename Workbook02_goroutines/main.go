package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// worker is a goroutine function that processes jobs from a channel and sends results to another channel
func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done() // Move WaitGroup management into worker function
	// This  line!  works like a Python generator:
	// - Continuously reads from the 'jobs' channel until it's closed
	// - Automatically handles synchronization and blocking
	// - Each iteration gets the next available job from the channel
	for job := range jobs {
		fmt.Printf("Worker %d processing job %d\n", id, job)
		time.Sleep(time.Second)
		results <- job * 2
	}
	fmt.Printf("Worker %d: All jobs completed, shutting down\n", id)
}

func main() {
	fmt.Printf("CPU Cores: %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))

	jobs := make(chan int, 10)
	results := make(chan int, 10)

	var wg sync.WaitGroup

	// Start 3 worker goroutines
	for w := 1; w <= 3; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg) // Pass WaitGroup to worker
	}

	// Send 16 jobs
	for j := 1; j <= 16; j++ {
		jobs <- j
	}
	close(jobs)

	// Start collecting results IN PARALLEL with worker completion
	go func() {
		// Wait for all workers to finish first
		wg.Wait()
		// Then close results channel
		close(results)
		fmt.Println("All workers finished. Results channel closed.")
	}()

	fmt.Println("All jobs sent. Collecting results...")

	// Collect results - this runs in main goroutine WHILE workers are still working
	resultCount := 0
	for result := range results {
		resultCount++
		fmt.Printf("Main: Received result %d: %d\n", resultCount, result)
	}

	fmt.Printf("All done! Collected %d results. Program exiting cleanly.\n", resultCount)
}
