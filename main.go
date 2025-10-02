package main

import (
	"encoding/json"
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

	myApp := app.New()
	myWindow := myApp.NewWindow("UK-808")
	myWindow.Resize(fyne.NewSize(600, 100))

	// Create our UI
	statusLabel := widget.NewLabel("Ready to start...")

	// Create channel for UI updates
	uiUpdates := make(chan UIUpdate, 10)

	// Start the UI update handler (runs forever)
	go func() {
		for update := range uiUpdates {
			// Use fyne.Do for thread-safe UI updates
			fyne.Do(func() {
				statusLabel.SetText(update.Text)
			})
		}
	}()
	// Button that starts the worker in a goroutine
	var startButton *widget.Button
	startButton = widget.NewButton("Start", func() {
		if reflect.DeepEqual(constraintHeader, microDataHeader) {
			startButton.Disable()
			start := time.Now()
			// Run parallelRun in a goroutine to avoid blocking UI
			go func() {
				parallelRun(constraints, microData, microDataHeader, config.Output.File, config.Validate.File, annealingConfig, uiUpdates)

				elapsed := time.Since(start)
				// Send completion message
				uiUpdates <- UIUpdate{Text: fmt.Sprintf("✅ Completed in %s", elapsed)}
				// SAFE: fyne.Do() ensures it runs on main UI thread
				fyne.Do(func() {
					startButton.Enable()
				})
			}()
		} else {
			fmt.Printf("Error: The Constraints header and the MicroData header not the same\n")
			os.Exit(1)
		}
	})
	content := container.NewVBox(statusLabel, startButton)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
	close(uiUpdates)
}
