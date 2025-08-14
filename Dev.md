# SynthPopGo V0.2

### 08/08/24
- Change config
    - Config -> PopulationConfig
    - Scatter -> Validate


Create a simulated annealing configuration file 

```go
type AnnealingConfig struct {
    InitialTemp      float64 `json:"initialTemp"`
    MinTemp          float64 `json:"minTemp"`
    CoolingRate      float64 `json:"coolingRate"`
    ReheatFactor     float64 `json:"reheatFactor"`
    FitnessThreshold float64 `json:"fitnessThreshold"`
    MinImprovement   float64 `json:"minImprovement"`
    MaxIterations    int     `json:"maxIterations"`
    WindowSize       int     `json:"windowSize"`
    Change           int     `json:"change"`
}


if reflect.DeepEqual(constarintHeader, microDataHEader) {
    // Load annealing config from JSON file
    configFile, err := os.Open("annealing_config.json")
    if err != nil {
        fmt.Printf("Error opening annealing config: %v\n", err)
        return
    }
    defer configFile.Close()

    var annealingConfig AnnealingConfig
    if err := json.NewDecoder(configFile).Decode(&annealingConfig); err != nil {
        fmt.Printf("Error decoding annealing config: %v\n", err)
        return
    }
```



Changed the type AnnealingConfig struct

### 12/08/25

1) Refactored the main go put all of the data loading in seprate functions

```go
// loadConfig loads the population configuration from a JSON file.
func loadConfig(configFileName string) (PopulationConfig, error) {
	var config PopulationConfig
	file, err := os.Open(configFileName)
	if err != nil {
		return config, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return config, fmt.Errorf("error decoding config JSON: %w", err)
	}
	return config, nil
}

// loadAnnealingConfig loads annealing parameters from a JSON file.
func loadAnnealingConfig(annealingFileName string) (AnnealingConfig, error) {
	var annealingConfig AnnealingConfig
	file, err := os.Open(annealingFileName)
	if err != nil {
		return annealingConfig, fmt.Errorf("error opening annealing config: %w", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&annealingConfig); err != nil {
		return annealingConfig, fmt.Errorf("error decoding annealing config: %w", err)
	}
	return annealingConfig, nil
}

// readArgs parses command-line arguments with default fallbacks.
func readArgs() (string, string) {
	configFileName := "config.json"
	annealingFileName := "annealing_config.json"

	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}
	if len(os.Args) > 2 {
		annealingFileName = os.Args[2]
	}

	return configFileName, annealingFileName
}

// loadConstraints loads constraint data from CSV and validates headers.
func loadConstraints(constraintsFile string) ([]ConstraintData, []string, error) {
	constraints, header, err := ReadConstraintCSV(constraintsFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read constraints CSV: %w", err)
	}
	fmt.Printf("Loaded %d constraint areas", len(constraints))
	return constraints, header, nil
}

// loadMicrodata loads microdata from CSV and validates headers.
func loadMicrodata(microdataFile string) ([]MicroData, []string, error) {
	microData, header, err := ReadMicroDataCSV(microdataFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read microdata CSV: %w", err)
	}
	fmt.Printf("Loaded %d microdata records", len(microData))
	return microData, header, nil
}

func main() {
	configFileName, anellingFileName := readArgs()

	config, err := loadConfig(configFileName)
	if err != nil {
		fmt.Printf("Config error: %v", err)
	}

	annealingConfig, err := loadAnnealingConfig(anellingFileName)
	if err != nil {
		fmt.Printf("Annealing config error: %v", err)
	}

	// Load data
	constraints, constraintHeader, err := loadConstraints(config.Constraints.File)
	if err != nil {
		fmt.Printf("Constraint loading error: %v", err)
	}

	microData, microDataHeader, err := loadMicrodata(config.Microdata.File)
	if err != nil {
		fmt.Printf("Microdata loading error: %v", err)
	}

	if reflect.DeepEqual(constraintHeader, microDataHeader) {
		start := time.Now()
		parallelRun(constraints, microData, config.Output.File, config.Validate.File, annealingConfig)

		elapsed := time.Since(start) // Calculate duration
		fmt.Printf("slowFunction took %s\n", elapsed)
	} else {
		fmt.Printf("Error: The Constraints header and the MiroData header not the same\n")
		os.Exit(1)
	}
}

```

2) Output the synthpop survay for each area  


```go

	header := append([]string{"area_id"}, microdataHeader...)
	if err := fractionsWriter.Write(header); err != nil {
		return fmt.Errorf("error writing fractions headers: %w", err)
	}
	fractionsWriter.Flush() // This will write the line to file immediately
	if err := fractionsWriter.Error(); err != nil {
		return fmt.Errorf("error flushing fractions headers: %w", err)
	}




var writerWg sync.WaitGroup
writerWg.Add(1)
go func() {
    defer writerWg.Done()
    for res := range resultsChan {
        areaId := res.area

        // Write ID mappings (using existing CSV writer)
        for _, id := range res.ids {
            if err := idsWriter.Write([]string{areaId, id}); err != nil {
                select {
                case errChan <- fmt.Errorf("error writing ID row: %w", err):
                default:
                }
                return
            }
        }

        // Build the unquoted CSV line
        var buf strings.Builder
        buf.WriteString(areaId)
        for _, val := range res.synthpop_totals {
            buf.WriteByte(',')
            buf.WriteString(strconv.FormatFloat(val, 'f', -1, 64))
        }
        buf.WriteByte('\n')

        // Write raw string directly to file
        if _, err := fractionsFile.WriteString(buf.String()); err != nil {
            select {
            case errChan <- fmt.Errorf("error writing fraction row: %w", err):
            default:
            }
            return
        }

        processed.Add(1)
    }
}()
```

### Key Improvements:

1. **String Building Optimization**:
   - Uses `strings.Builder` for efficient string concatenation
   - Avoids intermediate string allocations

2. **Direct File Writing**:
   - Bypasses `csv.Writer` to prevent automatic quoting
   - Writes the exact format you want: `S00135308,33,0,8,...`

3. **Maintained Concurrency Safety**:
   - Keeps all existing synchronization (`WaitGroup`, channels)
   - Preserves non-blocking error reporting

4. **Memory Efficiency**:
   - Builder reuses buffer space
   - Single write operation per record

### Why This Works Better:

- For your ID mappings (which need proper CSV handling), keeps using `csv.Writer`
- For your synthetic totals (which need unquoted output), writes directly
- Maintains all error handling and progress tracking
- Still properly closes/flushes files via your existing `defer` statements
