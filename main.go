package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type MicroData struct {
	ID     string
	Values []float64
}

type ConstraintData struct {
	ID     string
	Values []float64
	Total  float64
}

type results struct {
	area              string
	population        float64
	synthpop_totals   []float64
	ids               []string
	constraint_totals []float64
	fitness           float64
}

type CombinedConfig struct {
	// ── Annealing section ────────────────────────────────────────
	InitialTemp      float64 `json:"initialTemp"`
	MinTemp          float64 `json:"minTemp"`
	CoolingRate      float64 `json:"coolingRate"`
	ReheatFactor     float64 `json:"reheatFactor"`
	FitnessThreshold float64 `json:"fitnessThreshold"`
	MinImprovement   float64 `json:"minImprovement"`
	MaxIterations    int     `json:"maxIterations"`
	WindowSize       int     `json:"windowSize"`
	Change           int     `json:"change"`
	Distance         string  `json:"distance"`
	UseRandomSeed    string  `json:"useRandomSeed"`
	RandomSeed       *int64  `json:"randomSeed,omitempty"` // optional seed for reproducibility

	// ── Population section ───────────────────────────────────────
	ConstraintsFile string `json:"constraintsFile"` // renamed for clarity
	MicrodataFile   string `json:"microdataFile"`   // renamed for clarity
	OutputFile      string `json:"outputFile"`      // renamed for clarity
	ValidateFile    string `json:"validateFile"`    // renamed for clarity
}

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
	Distance         string  `json:"distance"`
	UseRandomSeed    string  `json:"useRandomSeed"`
	RandomSeed       *int64  `json:"randomSeed,omitempty"` // Optional seed for reproducibility
}

// UIUpdate struct for messages
type UIUpdate struct {
	Text string
}

var ValidMetrics = []string{"CHI_SQUARED", "EUCLIDEAN", "NORM_EUCLIDEAN", "MANHATTEN", "KL_DIVERGENCE", "COSINE", "JSDIVERGENCE"}

type PopulationConfig struct {
	Constraints struct {
		File string `json:"file"`
	} `json:"constraints"`
	Microdata struct {
		File string `json:"file"`
	} `json:"microdata"`
	Output struct {
		File string `json:"file"`
	} `json:"output"`
	Validate struct {
		File string `json:"file"`
	} `json:"validate"`
}

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
	var config AnnealingConfig

	file, err := os.Open(annealingFileName)
	if err != nil {
		return config, fmt.Errorf("error opening config: %w", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return config, fmt.Errorf("invalid config format: %w", err)
	}

	// Validate distance metric
	valid := false
	for _, m := range ValidMetrics {
		if config.Distance == m {
			valid = true
			break
		}
	}

	if !valid {
		return config, fmt.Errorf(
			"invalid distance metric '%s'. Must be one of: %v",
			config.Distance,
			ValidMetrics,
		)
	}

	return config, nil
}

type FileConfig struct {
	File string `json:"file"`
}

type JSONConfig struct {
	Constraints *FileConfig `json:"constraints"`
	Microdata   *FileConfig `json:"microdata"`
	Output      *FileConfig `json:"output"`
	Validate    *FileConfig `json:"validate"`

	InitialTemp      float64 `json:"initialTemp"`
	MinTemp          float64 `json:"minTemp"`
	CoolingRate      float64 `json:"coolingRate"`
	ReheatFactor     float64 `json:"reheatFactor"`
	FitnessThreshold float64 `json:"fitnessThreshold"`
	MinImprovement   float64 `json:"minImprovement"`
	MaxIterations    int     `json:"maxIterations"`
	WindowSize       int     `json:"windowSize"`
	Change           int     `json:"change"`
	Distance         string  `json:"distance"`
	UseRandomSeed    string  `json:"useRandomSeed"`
	RandomSeed       *int64  `json:"randomSeed,omitempty"`
}

// Combined struct for JSON parsing
type RootConfig struct {
	AnnealingConfig  `json:",inline"`
	PopulationConfig `json:",inline"`
}

// Function to load both configs from single file
func LoadCombinedConfigs(filename string) (AnnealingConfig, PopulationConfig, error) {
	data, err := os.ReadFile(filename)
	var root RootConfig
	if err != nil {
		return root.AnnealingConfig, root.PopulationConfig, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &root); err != nil {
		return root.AnnealingConfig, root.PopulationConfig, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return root.AnnealingConfig, root.PopulationConfig, nil
}

func readArgs() (string, bool) {
	// Define flags
	cliMode := flag.Bool("c", false, "Run in command-line mode without GUI")
	configFile := flag.String("f", "config.json", "Config file path (default: combine.json)")

	// Custom usage function
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		fmt.Println("Options:")
		fmt.Println("  -c    Run in command-line mode without GUI")
		fmt.Println("  -f string")
		fmt.Println("        Config file path (default: combine.json)")
		fmt.Println("  -h    Show this help message")
	}

	flag.Parse()

	return *configFile, *cliMode
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
	configFile, cliMode := readArgs()

	annealingConfig, config, err := LoadCombinedConfigs(configFile)
	if err != nil {
		fmt.Printf("Config error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Using config file: %s\n", configFile)
	fmt.Printf("Constraints file: %s\n", config.Constraints.File)
	fmt.Printf("Microdata file: %s\n", config.Microdata.File)
	fmt.Printf("Output file: %s\n", config.Output.File)

	// Load data
	constraints, constraintHeader, err := loadConstraints(config.Constraints.File)
	if err != nil {
		fmt.Printf("Constraint loading error: %v\n", err)
		os.Exit(1)
	}

	microData, microDataHeader, err := loadMicrodata(config.Microdata.File)
	if err != nil {
		fmt.Printf("Microdata loading error: %v\n", err)
		os.Exit(1)
	}

	// Check if headers match
	if !reflect.DeepEqual(constraintHeader, microDataHeader) {
		fmt.Printf("Error: The Constraints header and the MicroData header are not the same\n")
		os.Exit(1)
	}

	// CLI mode (-c flag)
	if cliMode {
		fmt.Println("🚀 Running in command-line mode...")
		start := time.Now()

		// Create a channel for updates
		uiUpdates := make(chan UIUpdate, 10)

		// Start a goroutine to print CLI updates
		go func() {
			for update := range uiUpdates {
				fmt.Println("📢", update.Text)
			}
		}()

		// Run the main process
		parallelRun(constraints, microData, microDataHeader, config.Output.File, config.Validate.File, annealingConfig, uiUpdates)

		elapsed := time.Since(start)
		fmt.Printf("✅ Completed in %s\n", elapsed)
		close(uiUpdates)
		return
	}

	// GUI mode (default)
	fmt.Println("🎨 Launching GUI mode...")
	myApp := app.New()
	myWindow := myApp.NewWindow("UK-808")
	myWindow.Resize(fyne.NewSize(600, 100))

	// Create our UI
	statusLabel := widget.NewLabel("Ready to start...")

	// Create channel for UI updates
	uiUpdates := make(chan UIUpdate, 10)

	// Start the UI update handler
	go func() {
		for update := range uiUpdates {
			if cliMode {
				// Carriage return to beginning of line
				fmt.Print("\r")
				// Print with padding to clear previous content
				fmt.Printf("%-80s", update.Text) // Adjust 80 to your terminal width
			} else {
				fyne.Do(func() {
					statusLabel.SetText(update.Text)
				})
			}
		}
	}()

	// Button that starts the worker in a goroutine
	var startButton *widget.Button
	startButton = widget.NewButton("Start", func() {
		startButton.Disable()
		start := time.Now()

		// Run parallelRun in a goroutine to avoid blocking UI
		go func() {
			parallelRun(constraints, microData, microDataHeader, config.Output.File, config.Validate.File, annealingConfig, uiUpdates)

			elapsed := time.Since(start)
			// Send completion message
			uiUpdates <- UIUpdate{Text: fmt.Sprintf("✅ Completed in %s", elapsed)}

			fyne.Do(func() {
				startButton.Enable()
			})
		}()
	})

	content := container.NewVBox(statusLabel, startButton)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
	close(uiUpdates)
}
