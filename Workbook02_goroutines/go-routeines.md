## Goroutines: The Basics

Goroutines are lightweight threads managed by the Go runtime. They're incredibly cheap compared to OS threads.

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // Start a goroutine
    go sayHello()
    
    // Give the goroutine time to execute
    time.Sleep(100 * time.Millisecond)
    fmt.Println("Main function")
}

func sayHello() {
    fmt.Println("Hello from goroutine!")
}
```

## Channels: Passing Messages

Channels are the pipes that connect goroutines, allowing them to communicate safely.

### Basic Channel Operations

```go
package main

import "fmt"

func main() {
    // Create an unbuffered channel
    messageChannel := make(chan string)
    
    // Start a goroutine that sends a message
    go func() {
        messageChannel <- "Hello from the goroutine!"
    }()
    
    // Receive the message in main
    message := <-messageChannel
    fmt.Println(message)
}
```

### Buffered Channels

```go
func bufferedExample() {
    // Buffered channel with capacity of 2
    ch := make(chan string, 2)
    
    ch <- "first"
    ch <- "second" // Won't block until buffer is full
    
    fmt.Println(<-ch) // "first"
    fmt.Println(<-ch) // "second"
}
```

## Practical Examples

### 1. Producer-Consumer Pattern

```go
package main

import (
    "fmt"
    "time"
)

func producer(ch chan<- int) {
    for i := 0; i < 5; i++ {
        fmt.Printf("Producing: %d\n", i)
        ch <- i
        time.Sleep(time.Second)
    }
    close(ch) // Important: close channel when done
}

func consumer(ch <-chan int) {
    for num := range ch {
        fmt.Printf("Consuming: %d\n", num)
    }
}

func main() {
    ch := make(chan int)
    
    go producer(ch)
    consumer()
}
```

### 2. Multiple Workers with Results

```go
func worker(id int, jobs <-chan int, results chan<- int) {
    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job)
        time.Sleep(time.Second) // Simulate work
        results <- job * 2
    }
}

func main() {
    jobs := make(chan int, 10)
    results := make(chan int, 10)
    
    // Start 3 workers
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }
    
    // Send 5 jobs
    for j := 1; j <= 5; j++ {
        jobs <- j
    }
    close(jobs)
    
    // Collect results
    for r := 1; r <= 5; r++ {
        fmt.Printf("Result: %d\n", <-results)
    }
}
```


## Key Concepts to Remember

1. **Unbuffered channels** are synchronous - sender blocks until receiver is ready
2. **Buffered channels** allow sending without blocking until buffer is full
3. **Always close channels** from the sender side when done
4. **Use `range` on channels** to receive values until channel is closed


