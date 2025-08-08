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

//	func worker(id int, change int, microData []MicroData, config AnnealingConfig, jobs <-chan ConstraintData, resultsChan chan<- results, wg *sync.WaitGroup) {
//		defer wg.Done()
//		for constraint := range jobs {
//			//fmt.Printf("Worker %d processing area %s\n", id, constraint.ID)
//			res := syntheticPopulation(constraint, microData, config)
//			resultsChan <- res
//		}
//	}
func parallelRun(constraints []ConstraintData, microData []MicroData, outputfile1 string, outputfile2 string, config AnnealingConfig) error {
	// Dynamic worker count
	numWorkers := runtime.NumCPU()
	if len(constraints) < numWorkers {
		numWorkers = len(constraints)
	}
	fmt.Printf("🚀 Starting %d workers for %d population areas\n", numWorkers, len(constraints))

	// Setup channels
	jobs := make(chan ConstraintData, numWorkers*2)
	resultsChan := make(chan results, numWorkers*2)
	errChan := make(chan error, 1)

	// Create both output files
	idsFile, err := os.Create(outputfile1) // For ID output
	if err != nil {
		return fmt.Errorf("cannot create IDs file: %w", err)
	}
	defer idsFile.Close()

	fractionsFile, err := os.Create(outputfile2) // For fraction output
	if err != nil {
		return fmt.Errorf("cannot create fractions file: %w", err)
	}
	defer fractionsFile.Close()

	// Create writers for both files
	idsWriter := csv.NewWriter(idsFile)
	defer idsWriter.Flush()

	fractionsWriter := csv.NewWriter(fractionsFile)
	defer fractionsWriter.Flush()

	// Write headers for both files
	if err := idsWriter.Write([]string{"area_id", "microdata_id"}); err != nil {
		return fmt.Errorf("error writing IDs headers: %w", err)
	}
	if err := fractionsWriter.Write([]string{"area_id", "variable", "synthetic_fraction", "constraint_fraction"}); err != nil {
		return fmt.Errorf("error writing fractions headers: %w", err)
	}

	// Progress tracking variables
	var (
		processed      atomic.Int32
		totalJobs      = len(constraints)
		startTime      = time.Now()
		progressTicker = time.NewTicker(2 * time.Second)
	)
	defer progressTicker.Stop()

	// Progress reporter
	go func() {
		for range progressTicker.C {
			elapsed := time.Since(startTime).Round(time.Second)
			done := processed.Load()
			remaining := totalJobs - int(done)
			percent := float64(done) / float64(totalJobs) * 100

			var eta time.Duration
			if done > 0 {
				perItem := elapsed / time.Duration(done)
				eta = time.Duration(remaining) * perItem
			}

			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			fmt.Printf("\r📊 Progress: %d/%d (%.1f%%) | ⏱️ Elapsed: %v | 🕒 ETA: %v | 🧠 Memory: %vMB",
				done, totalJobs, percent, elapsed, eta.Round(time.Second), m.Alloc/1024/1024)
		}
	}()

	// Writer goroutine
	var writerWg sync.WaitGroup
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		for res := range resultsChan {
			areaId := res.area

			// Write to IDs file
			for _, id := range res.ids {
				if err := idsWriter.Write([]string{areaId, id}); err != nil {
					select {
					case errChan <- fmt.Errorf("error writing ID row: %w", err):
					default:
					}
					return
				}
			}

			// Write to fractions file (one row per variable)
			for i := range res.synthpop_totals {
				row := []string{
					areaId,
					fmt.Sprintf("var_%d", i), // or use actual variable names if available
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

			processed.Add(1)
		}
	}()

	// Worker pool
	var workerWg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()
			for constraint := range jobs {
				res := syntheticPopulation(constraint, microData, config)
				select {
				case resultsChan <- res:
				case <-errChan:
					return
				}
			}
		}(i)
	}

	// Feed jobs to workers
	for _, constraint := range constraints {
		select {
		case jobs <- constraint:
		case err := <-errChan:
			close(jobs)
			workerWg.Wait()
			close(resultsChan)
			writerWg.Wait()
			return err
		}
	}
	close(jobs)

	// Wait for completion
	workerWg.Wait()
	close(resultsChan)
	writerWg.Wait()

	// Final report
	elapsed := time.Since(startTime).Round(time.Second)
	fmt.Printf("\n✅ Completed %d populations in %v (avg %.2f/sec)\n",
		totalJobs, elapsed, float64(totalJobs)/elapsed.Seconds())

	return nil
}
