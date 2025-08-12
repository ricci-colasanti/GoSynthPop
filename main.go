package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"
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
}

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
