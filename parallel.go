package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// parallelRun executes population synthesis in parallel across multiple workers.
// It takes constraint data, microdata, output file paths, and annealing configuration,
// then distributes the work across CPU cores and writes results to CSV files.
//
// Parameters:
//   - constraints: Slice of ConstraintData defining each geographical area's constraints
//   - microData: Slice of MicroData containing individual population records
//   - outputfile1: Path for output CSV mapping area IDs to synthetic population IDs
//   - outputfile2: Path for output CSV comparing synthetic vs constraint fractions
//   - config: AnnealingConfig with optimization parameters
//
// Returns:
//   - error: Any error encountered during processing
func parallelRun(constraints []ConstraintData, microData []MicroData, outputfile1 string, outputfile2 string, config AnnealingConfig) error {
	// Dynamic worker count - use either CPU count or constraint count, whichever is smaller
	numWorkers := runtime.NumCPU()
	if len(constraints) < numWorkers {
		numWorkers = len(constraints)
	}
	fmt.Printf("🚀 Starting %d workers for %d population areas\n", numWorkers, len(constraints))

	// Setup communication channels:
	// - jobs: feeds constraints to workers
	// - resultsChan: collects processed results from workers
	// - errChan: receives any processing errors (buffered to prevent deadlocks)
	jobs := make(chan ConstraintData, numWorkers*2)
	resultsChan := make(chan results, numWorkers*2)
	errChan := make(chan error, 1)

	// Create output files for:
	// 1. ID mappings (area_id → synthetic population IDs)
	// 2. Fraction comparisons (synthetic vs constraint fractions by variable)
	idsFile, err := os.Create(outputfile1)
	if err != nil {
		return fmt.Errorf("cannot create IDs file: %w", err)
	}
	defer idsFile.Close()

	fractionsFile, err := os.Create(outputfile2)
	if err != nil {
		return fmt.Errorf("cannot create fractions file: %w", err)
	}
	defer fractionsFile.Close()

	// Initialize CSV writers with buffering
	idsWriter := csv.NewWriter(idsFile)
	defer idsWriter.Flush() // Ensure all data is written even if function exits early

	fractionsWriter := csv.NewWriter(fractionsFile)
	defer fractionsWriter.Flush()

	// Write CSV headers for both output files
	if err := idsWriter.Write([]string{"area_id", "microdata_id"}); err != nil {
		return fmt.Errorf("error writing IDs headers: %w", err)
	}
	if err := fractionsWriter.Write([]string{"area_id", "variable", "synthetic_fraction", "constraint_fraction"}); err != nil {
		return fmt.Errorf("error writing fractions headers: %w", err)
	}

	// Progress tracking setup
	var (
		processed      atomic.Int32 // Thread-safe counter for completed jobs
		totalJobs      = len(constraints)
		startTime      = time.Now()                      // Capture start time for ETA calculation
		progressTicker = time.NewTicker(2 * time.Second) // Update progress every 2s
	)
	defer progressTicker.Stop()

	// Progress reporter goroutine - displays real-time statistics
	go func() {
		for range progressTicker.C {
			elapsed := time.Since(startTime).Round(time.Second)
			done := processed.Load()
			remaining := totalJobs - int(done)
			percent := float64(done) / float64(totalJobs) * 100

			// Calculate ETA based on current processing rate
			var eta time.Duration
			if done > 0 {
				perItem := elapsed / time.Duration(done)
				eta = time.Duration(remaining) * perItem
			}

			// Include memory usage in progress report
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			fmt.Printf("\r📊 Progress: %d/%d (%.1f%%) | ⏱️ Elapsed: %v | 🕒 ETA: %v | 🧠 Memory: %vMB",
				done, totalJobs, percent, elapsed, eta.Round(time.Second), m.Alloc/1024/1024)
		}
	}()

	// Writer goroutine - handles all output file writing
	var writerWg sync.WaitGroup
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		for res := range resultsChan {
			areaId := res.area

			// Write ID mappings (area_id → synthetic population IDs)
			for _, id := range res.ids {
				if err := idsWriter.Write([]string{areaId, id}); err != nil {
					// Non-blocking error reporting
					select {
					case errChan <- fmt.Errorf("error writing ID row: %w", err):
					default: // Skip if error channel is full
					}
					return
				}
			}

			// Write fraction comparisons for each variable
			for i := range res.synthpop_totals {
				row := []string{
					areaId,
					fmt.Sprintf("var_%d", i), // Generic variable name
					// Calculate fractions by dividing by total population
					strconv.FormatFloat(res.synthpop_totals[i]/res.population, 'f', -1, 64),
					strconv.FormatFloat(res.constraint_totals[i]/res.population, 'f', -1, 64),
				}
				if err := fractionsWriter.Write(row); err != nil {
					select {
					case errChan <- fmt.Errorf("error writing fraction row: %w", err):
					default:
					}
					return
				}
			}

			processed.Add(1) // Increment progress counter
		}
	}()

	// Worker pool - processes constraints in parallel
	var workerWg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()
			for constraint := range jobs {
				// Generate synthetic population for this constraint area
				res := syntheticPopulation(constraint, microData, config)

				// Send result or abort if error occurred
				select {
				case resultsChan <- res:
				case <-errChan: // Channel closed means error occurred
					return
				}
			}
		}(i)
	}

	// Feed jobs to workers with error checking
	for _, constraint := range constraints {
		select {
		case jobs <- constraint: // Send next job
		case err := <-errChan: // Handle any errors from writers
			close(jobs)        // Signal workers to stop
			workerWg.Wait()    // Wait for workers to finish
			close(resultsChan) // Close results channel
			writerWg.Wait()    // Wait for writer to finish
			return err         // Return the error
		}
	}
	close(jobs) // All jobs sent

	// Wait for completion
	workerWg.Wait()    // All workers finished
	close(resultsChan) // No more results coming
	writerWg.Wait()    // All results written

	// Final performance report
	elapsed := time.Since(startTime).Round(time.Second)
	fmt.Printf("\n✅ Completed %d populations in %v (avg %.2f/sec)\n",
		totalJobs, elapsed, float64(totalJobs)/elapsed.Seconds())

	return nil
}
