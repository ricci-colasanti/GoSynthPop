
# GoSynthPop 
**🚧 Work in Progress - Not Ready for Production Use 🚧**  

*This is a research/experimental project for generating UK spatial synthetic populations.  
The code is under active development and not yet intended for public release or real-world use.*  

---

# UK Spatial Synthetic Population Generator

A Go-based tool for generating synthetic populations by combining UK census data with Understanding Society survey data using simulated annealing optimization.  


## Overview

This tool creates spatially detailed synthetic populations that match statistical constraints from census data while preserving individual characteristics from survey microdata. It's particularly designed for UK local authorities and uses parallel processing for efficient large-scale population synthesis.

## Features

- 🚀 **Parallel processing** - Utilizes all CPU cores for fast population generation
- 🔍 **Multiple distance metrics** - KL Divergence, Chi-Squared, Euclidean, and more
- ❄️ **Simulated annealing** - Intelligent optimization algorithm to match constraints
- 📊 **Validation outputs** - Generates comparison files to verify constraint matching
- 🏙️ **UK-focused** - Designed for UK census geography and Understanding Society data

## How It Works

1. **Inputs**:
   - Constraint data (CSV) - Census totals for each area
   - Microdata (CSV) - Understanding Society individual records
   - Configuration (JSON) - Specifies files and annealing parameters

2. **Process**:
   - For each geographical area:
     - Creates initial random population from microdata
     - Uses simulated annealing to optimize population to match constraints
     - Tracks fitness and automatically adjusts optimization parameters

3. **Outputs**:
   - Population IDs mapping area to individuals
   - Fractional comparisons showing constraint matching

## Installation

1. Ensure Go (1.16+) is installed
2. Clone this repository
3. Build the application:
   ```bash
   go build



 V0.22  
