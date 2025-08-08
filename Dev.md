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