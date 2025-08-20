
### **Technical Documentation: Spatial Population Synthesis via Simulated Annealing**

#### **1. Overview and Conceptual Foundation**

This code implements a core component of a **spatial microsimulation** or **population synthesis** tool. Its primary purpose is to generate a realistic, synthetic population for a given geographical area (e.g., a UK Local Authority) that adheres to known statistical aggregates (e.g., census tables) while being built from individual-level records (microdata) from a sample survey.

**The Core Problem:** Census data provides aggregated counts (e.g., "100 people aged 25-34, 50 households with 2 cars") but lacks individual records due to privacy. Survey data (e.g., the Understanding Society survey) provides rich individual-level data but is only a sample and is not geographically representative. This tool bridges that gap by creating a full list of synthetic individuals that, when aggregated, match the census constraints for a specific area.

**The Chosen Method: Simulated Annealing**
The algorithm is based on **Simulated Annealing (SA)**, a probabilistic optimization technique inspired by the process of annealing in metallurgy. SA is well-suited for this problem because it can efficiently navigate large, complex solution spaces (the vast number of possible population combinations) to find a near-optimal solution without getting trapped in poor local minima.

The algorithm starts with a random initial population, then iteratively makes small changes (swapping one individual for another), accepting changes that improve the solution but also occasionally accepting worse changes based on a "temperature" parameter. This controlled randomness allows the algorithm to explore the solution space broadly before gradually "cooling down" and converging on a final solution.

---

#### **2. Detailed Code Explanation**

The code is written in Go, a statically typed, compiled language known for its efficiency and built-in support for concurrency, making it ideal for processing large datasets in parallel.

##### **2.1. Constants and Configuration (Lines 1-15)**

```go
const (
	EPSILON = 1e-10
	KL_DIVERGENCE  = iota
	CHI_SQUARED
	EUCLIDEAN
	NORM_EUCLIDEAN
	MANHATTEN
)
```
*   **`EPSILON`**: A very small constant (`0.0000000001`) used for **numerical stability**. It prevents division by zero and the logarithm of zero in distance calculations, which would otherwise cause the program to crash or produce invalid results (`NaN` or `Inf`).
*   **Distance Metric Constants**: These are integer identifiers (`iota` auto-generates incrementing numbers) for the various statistical methods available to measure the difference between the target census constraints and the current synthetic population aggregates. The choice of metric significantly influences the optimization process.

##### **2.2. Distance Metric Functions**

A **function type** is defined, specifying the signature for all distance functions.
```go
type DistanceFunc func([]float64, []float64) float64
```
This means any function that takes two slices of `float64` and returns a single `float64` can be treated as a `DistanceFunc`. This allows the code to be modular and pass different comparison functions around easily.

**Factory Function: `distanceFunc` (Lines 18-40)**
This function acts as a dispatcher. Based on the string value provided in the configuration (`config.Distance`), it returns the appropriate function for calculating the distance (error) between two distributions.
```go
func distanceFunc(config AnnealingConfig) DistanceFunc {
	switch config.Distance {
	case "CHI_SQUARED":
		return ChiSquaredDistance
	case "EUCLIDEAN":
		return EuclideanDistance
	// ... other cases ...
	default:
		return KLDivergence
	}
}
```
This is a common Go pattern for creating flexible and configurable behavior.

**Specific Distance Functions:**

*   **`KLDivergence` (Kullback-Leibler Divergence)**: Measures how one probability distribution (the synthetic population, `Q`) diverges from a second, expected distribution (the census constraints, `P`). It is not symmetric. A value of 0 means the distributions are identical. It heavily penalizes cases where `P` is large but `Q` is small. `EPSILON` is added to both `P` and `Q` to avoid `math.Log(0)`.

*   **`ChiSquaredDistance`**: A standard statistical measure of the discrepancy between observed (`testData`) and expected (`constraints`) values. It is particularly sensitive to relative differences, especially where the expected value is small.

*   **`EuclideanDistance`**: The straight-line "ordinary" distance between two points in multi-dimensional space. It treats all dimensions (census categories) equally.

*   **`NormalizedEuclideanDistance`**: A crucial enhancement for this domain. It normalizes the difference in each category by the expected value (`constraints[i]`). This means a difference of 10 people is more significant for a constraint of 100 than for a constraint of 10,000. It also includes a special case: if a constraint is zero (e.g., "no households with 5+ cars in this area"), it applies a massive penalty (`1000.0 * testData[i] * testData[i]`) if the synthetic population has any value in that category, enforcing hard constraints.

*   **`ManhattanDistance` (L1 Norm)**: The sum of the absolute differences along each dimension. It is less sensitive to large outliers in a single dimension than Euclidean distance.

*   **`JSdivergence` (Jensen-Shannon Divergence)**: A symmetric and smoothed version of the KL divergence. It is often considered a more robust metric.

*   **`Cosine`**: Measures the cosine of the angle between two vectors. A value of 1 means the vectors are pointing in the same direction (perfect match), 0 means they are orthogonal. It is useful for comparing the overall "shape" of the distributions rather than their absolute magnitudes.

##### **2.3. Core Algorithm Functions**

**`initPopulation` (Lines 189-218)**
This function creates the initial, random synthetic population for an area.
1.  It first pre-filters the entire `microdata` pool to find all individual records (`MicroData`) that are **valid** for the given area's constraints (using `isValidMicrodata`). A record is invalid if it has a non-zero value for a category where the constraint is zero (e.g., a survey respondent has 3 cars, but the area constraint for "3+ car households" is 0).
2.  It then randomly samples from these valid records until the synthetic population reaches the required total number of individuals (`constraint.Total`).
3.  It simultaneously builds two key data structures:
    *   `synthPopTotals []float64`: A running aggregate total for every census category, built by summing the values of all selected individuals.
    *   `synthPopMicrodataIndexs []int`: A list of indices pointing back to the original `microdata` slice, representing *which* individuals were chosen.

**`replace` (Lines 122-172) - The Heart of Simulated Annealing**
This function performs a single iteration of the algorithm: a proposed change and a decision to accept or reject it.
1.  **Propose a Change**: It randomly selects a new candidate individual from the microdata (ensuring it's valid) and a random individual currently in the synthetic population to be replaced.
2.  **Calculate New Fitness**: It temporarily updates the `synthPopTotals` aggregates by subtracting the old individual's values and adding the new candidate's values. It then calculates the new fitness (error) using the selected `DistanceFunc`.
3.  **Metropolis Criterion**: This is the core logic of SA. It decides whether to accept the new solution.
    *   **Always accept** if the new fitness is *better* (lower) than the old fitness (`newFitness < fitness`).
    *   **Probabilistically accept** a *worse* solution with a probability of `math.Exp((fitness - newFitness)/temp)`. This probability is high when the temperature `temp` is high (early in the process, encouraging exploration) and decreases as the system cools down. This prevents the algorithm from getting stuck in a local optimum early on.
4.  **Commit or Revert**: If the change is accepted, the list of individual indices (`synthPopMicrodataIndexess`) is updated. If it's rejected, the changes to the `synthPopTotals` aggregates are reverted.

**`syntheticPopulation` (Lines 221-292) - The Main Optimization Loop**
This function orchestrates the entire simulated annealing process for one geographic area.
1.  **Initialization**: It calls `initPopulation` and sets the initial `fitness` (error). It retrieves the correct `DistanceFunc` using the factory function.
2.  **Annealing Parameters Setup**: It configures the starting temperature (`config.InitialTemp`), cooling rate (`config.CoolingRate`), and an `improvementWindow` to track progress over recent iterations.
3.  **Main Loop**: The loop continues for a maximum number of iterations (`config.MaxIterations`) or until the system has cooled (`temp > config.MinTemp`) and no more changes are being accepted (`changes > 0`).
    *   It calls `replace` to attempt a change.
    *   It keeps track of the **best solution ever found** (`bestSynthPopTotals`, `bestSynthPopIDs`).
    *   **Stagnation Detection & Reheating**: Using the `improvementWindow`, it checks if the solution has stopped improving significantly (`relativeImprovement < config.MinImprovement`). If so, it artificially **reheats** the system (`temp = math.Max(temp*(1+config.ReheatFactor), ...`), raising the temperature to help the algorithm jump out of a local trough and explore new areas of the solution space. This is an advanced feature that improves robustness.
    *   **Cooling**: The temperature is reduced every iteration by multiplying it by the `config.CoolingRate` (e.g., `0.999`), gradually reducing the probability of accepting worse solutions.
4.  **Result Preparation**: After the loop finishes, it packages the best-found solution into a `results` struct, which includes the area ID, the final aggregated totals, the list of microdata IDs that make up the population, the original constraints, and the final fitness score. This `results` struct is the final output for the area.

---

#### **3. Key Go Language Features Utilized**

*   **Slices (`[]float64`, `[]int`)**: These are dynamic, flexible arrays and are the primary data structure for handling sequences of data (constraint vectors, population indices).
*   **Functions as First-Class Citizens**: The use of the `DistanceFunc` type and the `distanceFunc` factory function is idiomatic Go. It allows for clean, decoupled code where the optimization logic is separate from the specific choice of error metric.
*   **Pass-by-Reference for Slices**: In Go, slices are passed by reference to functions. This is why functions like `replace` and `replaceValue` can modify the contents of the original `synthPopTotals` slice directly, which is efficient for large datasets.
*   **`rand.Rand` as a Parameter**: Passing a random number generator (`rng *rand.Rand`) as a parameter is a best practice. It makes the code deterministic and testable if a seeded RNG is used, and it is essential for safe concurrent execution (each Goroutine can have its own RNG instance).

#### **4. Application Context: UK Local Authorities**

This code is designed to run for **many areas independently**. A typical use case would involve:
1.  Loading a national microdataset and constraint data for all ~400 UK Local Authorities.
2.  Using Go's powerful **concurrency features** (like Goroutines and channels, not shown in this snippet) to run the `syntheticPopulation` function for dozens of areas **in parallel** on a multi-core machine.
3.  Collecting all the `results` structs and writing them to a database or file.

The output is a spatially detailed synthetic population where each individual is a realistic entity with a full set of attributes (age, income, household structure, etc.), and the collective attributes of all individuals in any given area accurately match the published census statistics for that area. This synthetic population is invaluable for policy analysis, disaster modeling, transport planning, and social research.