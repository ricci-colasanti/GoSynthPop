package main

import (
	"math"
	"math/rand"
	"time"
)

// Constants defining distance metrics and numerical stability parameters
const (
	// EPSILON is a small value to prevent division by zero and ensure numerical stability
	EPSILON = 1e-10

	// Distance metric types
	KL_DIVERGENCE  = iota // Kullback-Leibler divergence
	CHI_SQUARED           // Chi-squared distance
	EUCLIDEAN             // Standard Euclidean distance
	NORM_EUCLIDEAN        // Normalized Euclidean distance
	MANHATTEN             // Manhattan distance
)

// Distance calculates the distance between two distributions using the specified metric
//
// Parameters:
//   - metric: The distance metric to use (KL_DIVERGENCE, CHI_SQUARED, etc.)
//   - constraints: The target distribution values
//   - testData: The generated distribution values to compare
//
// Returns:
//   - The calculated distance between the distributions
func Distance(metric int, constraints, testData []float64) float64 {
	switch metric {
	case CHI_SQUARED:
		return ChiSquaredDistance(constraints, testData)
	case EUCLIDEAN:
		return EuclideanDistance(constraints, testData)
	case NORM_EUCLIDEAN:
		return NormalizedEuclideanDistance(constraints, testData)
	case MANHATTEN:
		return ManhattanDistance(constraints, testData)
	default: // KL_DIVERGENCE
		return KLDivergence(constraints, testData)
	}
}

// KLDivergence calculates the Kullback-Leibler divergence between two distributions
//
// Parameters:
//   - constraints: The target probability distribution (P)
//   - testData: The approximate distribution (Q)
//
// Returns:
//   - The KL divergence D(P||Q)
//
// Note:
//   - Uses EPSILON to avoid numerical instability
func KLDivergence(constraints, testData []float64) float64 {
	divergence := 0.0
	for i := range constraints {
		p := constraints[i] + EPSILON
		q := testData[i] + EPSILON
		divergence += p * math.Log(p/q)
	}
	return divergence
}

// ChiSquaredDistance calculates the chi-squared distance between observed and expected values
//
// Parameters:
//   - constraints: The expected values
//   - testData: The observed values
//
// Returns:
//   - The chi-squared statistic
func ChiSquaredDistance(constraints, testData []float64) float64 {
	distance := 0.0
	for i := range constraints {
		observed := testData[i] + EPSILON
		expected := constraints[i] + EPSILON
		diff := observed - expected
		distance += (diff * diff) / expected
	}
	return distance
}

// EuclideanDistance calculates the standard Euclidean distance between two vectors
//
// Parameters:
//   - constraints: The first vector
//   - testData: The second vector
//
// Returns:
//   - The Euclidean distance (L2 norm)
func EuclideanDistance(constraints, testData []float64) float64 {
	distance := 0.0
	for i := range constraints {
		diff := testData[i] - constraints[i]
		distance += diff * diff
	}
	return math.Sqrt(distance)
}

// NormalizedEuclideanDistance calculates a normalized version of Euclidean distance
//
// Parameters:
//   - constraints: The reference vector used for normalization
//   - testData: The vector to compare
//
// Returns:
//   - The normalized Euclidean distance
//
// Note:
//   - Applies special handling for zero/very small constraints
//   - Adds large penalty for violating zero constraints
func NormalizedEuclideanDistance(constraints, testData []float64) float64 {
	distance := 0.0
	for i := range constraints {
		norm := constraints[i]
		if math.Abs(norm) < EPSILON {
			if math.Abs(testData[i]) > EPSILON {
				distance += 1000.0 * testData[i] * testData[i]
			}
			continue
		}
		diff := (testData[i] - constraints[i]) / norm
		distance += diff * diff
	}
	return math.Sqrt(distance)
}

// ManhattanDistance calculates the Manhattan distance (L1 norm) between two vectors
//
// Parameters:
//   - constraints: The first vector
//   - testData: The second vector
//
// Returns:
//   - The Manhattan distance
func ManhattanDistance(constraints, testData []float64) float64 {
	distance := 0.0
	for i := range constraints {
		distance += math.Abs(testData[i] - constraints[i])
	}
	return distance
}

// replaceValue copies values from new slice to old slice
//
// Parameters:
//   - old: The destination slice
//   - new: The source slice
//
// Note:
//   - Modifies the old slice in-place
func replaceValue(old []float64, new []float64) {
	for i := range old {
		old[i] = new[i]
	}
}

// removeUnordered removes an element from a slice in O(1) time (does not preserve order)
//
// Parameters:
//   - slice: The slice to modify
//   - index: The index of the element to remove
//
// Returns:
//   - The modified slice
func removeUnordered(slice []int, index int) []int {
	slice[index] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

// isValidMicrodata checks if microdata values satisfy all constraints
//
// Parameters:
//   - mdValues: The microdata values to check
//   - constraints: The constraints to validate against
//
// Returns:
//   - true if all zero constraints are satisfied, false otherwise
func isValidMicrodata(mdValues, constraints []float64) bool {
	for i, constraintVal := range constraints {
		if constraintVal == 0 && mdValues[i] != 0 {
			return false
		}
	}
	return true
}

// replace performs a replacement operation in the synthetic population using simulated annealing
//
// Parameters:
//   - microdata: The source microdata records
//   - constraint: The area constraints
//   - synthPopTotals: Current aggregate statistics
//   - synthPopMicrodataIndexess: Current population indices
//   - fitness: Current fitness score
//   - temp: Current temperature
//   - rng: Random number generator
//
// Returns:
//   - newFitness: The fitness after replacement
//   - flag: True if replacement was accepted, false if reverted
func replace(microdata []MicroData, constraint ConstraintData, synthPopTotals []float64,
	synthPopMicrodataIndexess []int, fitness float64, temp float64, rng *rand.Rand) (float64, bool) {

	flag := true

	var randomReplacmentIndex int
	var newValues []float64
	validFound := false
	maxAttempts := 100

	// Find valid replacement candidate
	for attempts := 0; attempts < maxAttempts; attempts++ {
		randomReplacmentIndex = rng.Intn(len(microdata))
		newValues = microdata[randomReplacmentIndex].Values
		if isValidMicrodata(newValues, constraint.Values) {
			validFound = true
			break
		}
	}

	if !validFound {
		return fitness, false
	}

	// Perform replacement
	randomReplceIndex := rng.Intn(len(synthPopMicrodataIndexess))
	replacementIndex := synthPopMicrodataIndexess[randomReplceIndex]
	oldValues := microdata[replacementIndex].Values

	// Update aggregates
	for i := 0; i < len(synthPopTotals); i++ {
		synthPopTotals[i] = synthPopTotals[i] - oldValues[i] + newValues[i]
	}

	newFitness := Distance(EUCLIDEAN, constraint.Values, synthPopTotals)

	// Metropolis acceptance criterion
	if newFitness >= fitness || math.Exp((fitness-newFitness)/temp) < rng.Float64() {
		// Revert changes
		for i := 0; i < len(synthPopTotals); i++ {
			synthPopTotals[i] = synthPopTotals[i] - newValues[i] + oldValues[i]
		}
		newFitness = fitness
		flag = false
	} else {
		// Accept changes
		synthPopMicrodataIndexess[randomReplceIndex] = randomReplacmentIndex
	}

	return newFitness, flag
}

// initPopulation creates an initial synthetic population for an area
//
// Parameters:
//   - constraint: The area constraints
//   - microdata: The source microdata
//
// Returns:
//   - synthPopTotals: Initial aggregate statistics
//   - synthPopMicrodataIndexs: Indices of selected microdata records
func initPopulation(constraint ConstraintData, microdata []MicroData) ([]float64, []int) {
	synthPopTotals := make([]float64, len(constraint.Values))
	synthPopMicrodataIndexs := make([]int, 0, int(constraint.Total))

	// Pre-filter valid microdata
	var validIndices []int
	for i, md := range microdata {
		if isValidMicrodata(md.Values, constraint.Values) {
			validIndices = append(validIndices, i)
		}
	}

	if len(validIndices) == 0 {
		panic("No valid microdata records match constraints")
	}

	// Create initial population
	for i := 0; i < int(constraint.Total); i++ {
		randomIndex := validIndices[rand.Intn(len(validIndices))]
		randomElement := microdata[randomIndex]

		synthPopMicrodataIndexs = append(synthPopMicrodataIndexs, randomIndex)
		for j := 0; j < len(synthPopTotals); j++ {
			synthPopTotals[j] += randomElement.Values[j]
		}
	}

	return synthPopTotals, synthPopMicrodataIndexs
}

// syntheticPopulation generates a synthetic population for one area using simulated annealing
//
// Parameters:
//   - constraint: The area constraints
//   - microdata: The source microdata
//   - config: Annealing configuration parameters
//
// Returns:
//   - results: The best solution found
func syntheticPopulation(constraint ConstraintData, microdata []MicroData, config AnnealingConfig) results {
	var synthPopResults results

	// Initialize population and fitness
	synthPopTotals, synthPopIDs := initPopulation(constraint, microdata)
	fitness := KLDivergence(constraint.Values, synthPopTotals)

	// Setup annealing parameters
	changes := config.Change
	temp := config.InitialTemp
	improvementWindow := make([]float64, config.WindowSize)
	windowIndex := 0
	bestFitness := fitness
	improvementWindow[windowIndex] = fitness
	windowIndex++

	// Track best solution
	bestSynthPopTotals := make([]float64, len(synthPopTotals))
	copy(bestSynthPopTotals, synthPopTotals)
	bestSynthPopIDs := make([]int, len(synthPopIDs))
	copy(bestSynthPopIDs, synthPopIDs)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Main optimization loop
	for iteration := 0; iteration < config.MaxIterations && changes > 0 && temp > config.MinTemp; iteration++ {
		flag := true
		fitness, flag = replace(microdata, constraint, synthPopTotals, synthPopIDs, fitness, temp, rng)

		// Update best solution
		if fitness < bestFitness {
			bestFitness = fitness
			copy(bestSynthPopTotals, synthPopTotals)
			copy(bestSynthPopIDs, synthPopIDs)

			if bestFitness <= config.FitnessThreshold {
				break
			}
		}

		// Track improvements
		improvementWindow[windowIndex] = fitness
		windowIndex = (windowIndex + 1) % config.WindowSize

		// Check for stagnation
		if iteration >= config.WindowSize {
			windowBest, windowWorst := improvementWindow[0], improvementWindow[0]
			for _, val := range improvementWindow {
				if val < windowBest {
					windowBest = val
				}
				if val > windowWorst {
					windowWorst = val
				}
			}

			relativeImprovement := (windowWorst - windowBest) / windowWorst
			if relativeImprovement < config.MinImprovement {
				temp = math.Max(temp*(1+config.ReheatFactor), config.InitialTemp*0.1)
				if relativeImprovement < config.MinImprovement/10 {
					break
				}
			}
		}

		temp *= config.CoolingRate

		if !flag {
			changes--
		}
	}

	// Prepare results
	synthPopResults.area = constraint.ID
	synthPopResults.synthpop_totals = bestSynthPopTotals
	synthPopResults.ids = make([]string, len(bestSynthPopIDs))
	for i, id := range bestSynthPopIDs {
		synthPopResults.ids[i] = microdata[id].ID
	}
	synthPopResults.constraint_totals = constraint.Values
	synthPopResults.fitness = bestFitness
	synthPopResults.population = constraint.Total

	return synthPopResults
}
