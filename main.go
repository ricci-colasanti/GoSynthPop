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
	InitialTemp      float64
	MinTemp          float64
	CoolingRate      float64
	ReheatFactor     float64
	FitnessThreshold float64
	MinImprovement   float64
	MaxIterations    int
	WindowSize       int
	Change           int
}

//Fields and Tags:
//
//Each nested struct has a single field named File, which is a string.
//The json:"file" tag is used to map the JSON key "file" to the File field in the struct.
//The outer tags (json:"constraints", json:"microdata", json:"output")
//are used to map the JSON keys "constraints", "microdata",
//and "output" to the corresponding nested structs.

type Config struct {
	Constraints struct {
		File string `json:"file"`
	} `json:"constraints"`
	Microdata struct {
		File string `json:"file"`
	} `json:"microdata"`
	Output struct {
		File string `json:"file"`
	} `json:"output"`
	Scatter struct {
		File string `json:"file"`
	} `json:"scatter"`
}

func main() {
	configFileName := "config.json"
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./program <config.json>")
	} else {
		configFileName = os.Args[1]
	}

	// Open and read the JSON file
	file, err := os.Open(configFileName)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Decode the JSON data into the Config struct
	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return
	}

	// Access the file names
	constraintsFile := config.Constraints.File
	microdataFile := config.Microdata.File
	outputFile1 := config.Output.File
	outputFile2 := config.Scatter.File
	// Get the file name from the command-line arguments

	constraints, constarintHeader, err := ReadConstraintCSV(constraintsFile)
	if err == nil {
		fmt.Printf("Areas: %v\n", len(constraints))
	} else {
		fmt.Printf("Error %v\n", err)
	}
	microData, microDataHEader, err := ReadMicroDataCSV(microdataFile)
	if err == nil {
		fmt.Printf("Sample population: %v \n", len(microData))
	} else {
		fmt.Printf("Error: %v\n", err)
	}
	if reflect.DeepEqual(constarintHeader, microDataHEader) {
		annealingConfig := AnnealingConfig{
			InitialTemp:      5000.0,
			MinTemp:          0.00001,
			CoolingRate:      0.999,
			ReheatFactor:     0.8,
			FitnessThreshold: 0.001,
			MinImprovement:   0.0001,
			MaxIterations:    5000000,
			WindowSize:       1000,
			Change:           100000,
		}

		start := time.Now()
		parallelRun(constraints, microData, outputFile1, outputFile2, annealingConfig)

		elapsed := time.Since(start) // Calculate duration
		fmt.Printf("slowFunction took %s\n", elapsed)
	} else {
		fmt.Printf("Error: The Constraints header and the MiroData header not the same\n")
	}
}
