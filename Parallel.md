# **Synthetic Population Generator: Comprehensive Explanation with Code**

---

## **1. Overview**
This parallelized system generates synthetic populations by distributing geographic area processing across CPU cores. It creates:
1. **ID Mappings**: Links geographic areas to synthetic individuals  
2. **Fraction Comparisons**: Shows statistical alignment between synthetic and constraint data  

Key features:
- **Dynamic worker allocation** (optimized for CPU cores)
- **Real-time progress tracking** (ETA, memory usage)
- **Graceful error handling** with clean shutdowns
- **Efficient I/O** with buffered CSV writing

---

## **2. Initialization Phase**

### **Worker Pool Setup**
The system first determines how many parallel workers to use, balancing between available CPU cores and the number of geographic areas to process.

```go
numWorkers := runtime.NumCPU()  // Default to CPU core count
if len(constraints) < numWorkers {
    numWorkers = len(constraints)  // Use fewer workers if fewer areas exist
}
fmt.Printf("🚀 Starting %d workers for %d population areas\n", numWorkers, len(constraints))
```

**Channels for Coordination**:
- `jobs`: Feeds constraints to workers (buffered for efficiency)  
- `resultsChan`: Collects processed results  
- `errChan`: Reports errors without blocking  

```go
jobs := make(chan ConstraintData, numWorkers*2)      // Buffered job queue
resultsChan := make(chan results, numWorkers*2)     // Buffered results queue
errChan := make(chan error, 1)                     // Non-blocking error reporting
```

---

### **Output File Preparation**
Two CSV files are created:
1. **ID Mappings** (`outputfile1`): `area_id → synthetic_id`  
2. **Fraction Comparisons** (`outputfile2`): Statistical validation  

```go
// Create and configure output files
idsFile, err := os.Create(outputfile1)  // ID mappings
if err != nil {
    return fmt.Errorf("cannot create IDs file: %w", err)
}
defer idsFile.Close()  // Ensures file closure even if errors occur

fractionsFile, err := os.Create(outputfile2)  // Statistical comparisons
if err != nil {
    return fmt.Errorf("cannot create fractions file: %w", err)
}
defer fractionsFile.Close()

// Initialize CSV writers with buffering
idsWriter := csv.NewWriter(idsFile)
defer idsWriter.Flush()  // Flush remaining data on exit

fractionsWriter := csv.NewWriter(fractionsFile)
defer fractionsWriter.Flush()

// Write CSV headers
if err := idsWriter.Write([]string{"area_id", "microdata_id"}); err != nil {
    return fmt.Errorf("error writing IDs headers: %w", err)
}
if err := fractionsWriter.Write([]string{"area_id", "variable", "synthetic_fraction", "constraint_fraction"}); err != nil {
    return fmt.Errorf("error writing fractions headers: %w", err)
}
```

---

## **3. Progress Monitoring System**
A dedicated goroutine provides real-time updates every 2 seconds, including:
- Completion percentage  
- Elapsed time and ETA  
- Memory usage  

```go
var (
    processed      atomic.Int32  // Thread-safe counter
    totalJobs      = len(constraints)
    startTime      = time.Now()
    progressTicker = time.NewTicker(2 * time.Second)  // Update every 2s
)
defer progressTicker.Stop()  // Clean up ticker on exit

// Progress reporter goroutine
go func() {
    for range progressTicker.C {
        elapsed := time.Since(startTime).Round(time.Second)
        done := processed.Load()
        percent := float64(done)/float64(totalJobs)*100
        
        // Calculate ETA (if at least one job completed)
        var eta time.Duration
        if done > 0 {
            perItemTime := elapsed / time.Duration(done)
            eta = time.Duration(totalJobs-int(done)) * perItemTime
        }

        // Memory usage (MB)
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        fmt.Printf("\r📊 Progress: %d/%d (%.1f%%) | ⏱️ %v | 🕒 %v | 🧠 %vMB",
            done, totalJobs, percent, elapsed, eta.Round(time.Second), m.Alloc/1024/1024)
    }
}()
```

---

## **4. Parallel Processing Pipeline**

### **Writer Goroutine**
Handles all file I/O, writing two types of records:
1. **ID Mappings** (area → synthetic individuals)  
2. **Fraction Comparisons** (synthetic vs. constraint data)  

```go
var writerWg sync.WaitGroup
writerWg.Add(1)
go func() {
    defer writerWg.Done()
    for res := range resultsChan {
        // Write ID mappings (area → synthetic IDs)
        for _, id := range res.ids {
            if err := idsWriter.Write([]string{res.area, id}); err != nil {
                select {
                case errChan <- fmt.Errorf("ID write error: %w", err):  // Report error
                default:  // Non-blocking if channel full
                }
                return
            }
        }
        
        // Write fraction comparisons (per variable)
        for i := range res.synthpop_totals {
            row := []string{
                res.area,
                fmt.Sprintf("var_%d", i),  // Variable label
                strconv.FormatFloat(res.synthpop_totals[i]/res.population, 'f', -1, 64),  // Synth fraction
                strconv.FormatFloat(res.constraint_totals[i]/res.population, 'f', -1, 64), // Constraint fraction
            }
            if err := fractionsWriter.Write(row); err != nil {
                select {
                case errChan <- fmt.Errorf("fraction write error: %w", err):
                default:
                }
                return
            }
        }
        processed.Add(1)  // Update progress counter
    }
}()
```

---

### **Worker Pool**
Each worker:
1. Processes constraints from the `jobs` channel  
2. Generates synthetic populations using `syntheticPopulation()`  
3. Sends results to the writer via `resultsChan`  

```go
var workerWg sync.WaitGroup
for i := 0; i < numWorkers; i++ {
    workerWg.Add(1)
    go func(workerID int) {
        defer workerWg.Done()
        for constraint := range jobs {
            res := syntheticPopulation(constraint, microData, config)  // Core processing
            select {
            case resultsChan <- res:  // Send result
            case <-errChan:  // Abort if error occurred
                return
            }
        }
    }(i)
}
```

---

### **Job Distribution & Error Handling**
The main loop feeds jobs to workers while monitoring for errors:
```go
for _, constraint := range constraints {
    select {
    case jobs <- constraint:  // Send next job
    case err := <-errChan:    // Handle errors from writers/workers
        close(jobs)           // Stop workers
        workerWg.Wait()       // Wait for current jobs to finish
        close(resultsChan)    // Signal writer to exit
        writerWg.Wait()       // Wait for final writes
        return err            // Propagate error
    }
}
close(jobs)  // Signal workers that no more jobs remain
```

---

## **5. Completion Phase**
After all jobs finish:
1. Workers complete  
2. Results channel closes  
3. Writer finishes processing  
4. Final performance stats are logged  

```go
workerWg.Wait()      // Wait for all workers
close(resultsChan)   // Signal writer to exit
writerWg.Wait()      // Ensure all data is written

// Final report
elapsed := time.Since(startTime).Round(time.Second)
fmt.Printf("\n✅ Completed %d populations in %v (avg %.2f/sec)\n",
    totalJobs, elapsed, float64(totalJobs)/elapsed.Seconds())
```

---

## **Key Takeaways**
1. **Efficient Parallelism**:  
   - Dynamically scales workers to CPU cores  
   - Buffered channels prevent blocking  
2. **Robust Error Handling**:  
   - Non-blocking error reporting  
   - Clean shutdowns on failure  
3. **Real-Time Monitoring**:  
   - Progress %, ETA, memory usage  
4. **Reliable Output**:  
   - Atomic counters for thread-safe progress  
   - Deferred flushes ensure data integrity  

This design ensures **high throughput** for large datasets while maintaining **clear visibility** into system performance.