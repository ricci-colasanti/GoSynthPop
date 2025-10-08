# Go Goroutines and Channels Example

A comprehensive demonstration of Go's concurrency model using goroutines and channels, showcasing how to implement a producer-consumer pattern with multiple workers with **proper synchronization**.

## 📋 Overview

This example demonstrates:
- **Goroutines**: Lightweight concurrent functions
- **Channels**: Thread-safe communication between goroutines  
- **Worker Pool Pattern**: Multiple workers processing jobs from a shared queue
- **WaitGroup Synchronization**: Proper coordination between goroutines
- **Deadlock Prevention**: Avoiding common concurrency pitfalls
- **Channel Ranging**: Python generator-like iteration over channels

## 🚀 Code Explanation

### Core Concepts

#### Worker Function with WaitGroup
The `worker` function runs as a goroutine and processes jobs from a channel with proper synchronization:

```go
func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
    defer wg.Done() // Crucial: ensures WaitGroup is decremented even if worker panics
    
    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job)
        time.Sleep(time.Second) // Simulate work
        results <- job * 2      // Send result
    }
    fmt.Printf("Worker %d: All jobs completed, shutting down\n", id)
}
```

**Key Points:**
- `defer wg.Done()` ensures the WaitGroup is always decremented
- **`for job := range jobs` acts like a Python generator, continuously reading until channel closure**
- Each worker competes for available jobs automatically
- Workers cleanly exit when jobs channel is closed

#### Main Execution Flow with Proper Synchronization

```go
func main() {
    // System info
    fmt.Printf("CPU Cores: %d\n", runtime.NumCPU())
    fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))

    // Create buffered channels
    jobs := make(chan int, 10)
    results := make(chan int, 10)

    var wg sync.WaitGroup

    // Start worker goroutines with WaitGroup
    for w := 1; w <= 3; w++ {
        wg.Add(1) // Increment counter BEFORE starting goroutine
        go worker(w, jobs, results, &wg)
    }

    // Send jobs
    for j := 1; j <= 16; j++ {
        jobs <- j
    }
    close(jobs) // Signal no more jobs

    // Critical: Start results collector in separate goroutine
    go func() {
        wg.Wait()        // Wait for ALL workers to finish
        close(results)   // THEN safely close results channel
        fmt.Println("All workers finished. Results channel closed.")
    }()

    // Collect results in main goroutine
    resultCount := 0
    for result := range results {
        resultCount++
        fmt.Printf("Main: Received result %d: %d\n", resultCount, result)
    }

    fmt.Printf("All done! Collected %d results.\n", resultCount)
}
```

## 🔧 How It Works

### 1. Channel Creation
```go
jobs := make(chan int, 10)    // Buffered channel (capacity 10)
results := make(chan int, 10) // Buffered channel (capacity 10)
```
**Why buffered?**
- Allows sending multiple items without immediate blocking
- Main can queue up to 10 jobs without waiting for workers
- Improves throughput by reducing synchronization overhead

### 2. Goroutine Launch with WaitGroup
```go
for w := 1; w <= 3; w++ {
    wg.Add(1) // Must be called in main goroutine, not inside the worker
    go worker(w, jobs, results, &wg)
}
```
- Creates 3 concurrent worker goroutines
- `wg.Add(1)` increments counter for each worker
- All workers share the same channels but have individual WaitGroup tracking

### 3. Job Distribution
```go
for j := 1; j <= 16; j++ {
    jobs <- j  // Send job to channel
}
close(jobs)    // Signal completion to workers
```
**Channel Behavior:**
- First 10 jobs fill the buffer immediately
- Jobs 11-16 block until workers free up buffer space
- `close(jobs)` tells workers to stop waiting for more work

### 4. Result Collection with Deadlock Prevention
```go
// Critical: Handle channel closing in separate goroutine
go func() {
    wg.Wait()      // Wait for ALL workers to call wg.Done()
    close(results) // Safe to close - no more sends will happen
}()

// Main collects results concurrently
for result := range results {
    // Process results as they arrive
}
```

**Why this pattern prevents deadlocks:**
- Workers can send results while main is receiving them
- Results channel only closes after ALL workers are done
- No goroutines get stuck waiting indefinitely

## 🎯 Key Concurrency Patterns

### The Magic: `for job := range jobs`
This line implements several important behaviors:
- **Automatic Synchronization**: Blocks gracefully when no jobs available
- **Clean Termination**: Exits loop automatically when channel closes
- **Fair Distribution**: Workers compete naturally for available work
- **Load Balancing**: Faster workers automatically process more jobs

### WaitGroup Synchronization Pattern
```go
var wg sync.WaitGroup

// For each worker:
wg.Add(1)        // BEFORE starting goroutine
go func() {
    defer wg.Done()  // DEFER the Done() call
    // ... work ...
}()

// To wait for completion:
wg.Wait()  // Blocks until all Done() calls are made
```

### Channel Direction Types
```go
jobs <-chan int    // Read-only channel
results chan<- int // Write-only channel
```
**Benefits:**
- Compile-time safety prevents misuse
- Clear communication intent
- Prevents accidental channel operations

## ⚠️ Common Pitfalls & Solutions

### Deadlock Scenario (DON'T DO THIS):
```go
// This causes deadlock!
wg.Wait()       // Main waits for workers
close(results)  // Then closes channel
for result := range results {  // But workers are blocked trying to send!
    // ...
}
```

### Correct Pattern :
```go
// This prevents deadlock!
go func() {
    wg.Wait()       // Wait for workers in background
    close(results)  // Then close channel safely
}()
for result := range results {  // Main receives while workers work
    // ...
}
```

## 📊 Expected Output

```
CPU Cores: 8
GOMAXPROCS: 8
All jobs sent. Collecting results...
Worker 1 processing job 1
Worker 2 processing job 2
Worker 3 processing job 3
Main: Received result 1: 2
Worker 1 processing job 4
Main: Received result 2: 4
Worker 2 processing job 5
Main: Received result 3: 6
Worker 3 processing job 6
Main: Received result 4: 8
Worker 1 processing job 7
Main: Received result 5: 10
... (continues with interleaved output) ...
Worker 3: All jobs completed, shutting down
Worker 1: All jobs completed, shutting down
Worker 2: All jobs completed, shutting down
All workers finished. Results channel closed.
Main: Received result 16: 32
All done! Collected 16 results. Program exiting cleanly.
```
