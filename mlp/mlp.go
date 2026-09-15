package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
)

// MLP structures
type MLPRequest struct {
	Pixels []float64 `json:"pixels"`
}

type MLPResponse struct {
	Class         int       `json:"class"`
	Probabilities []float64 `json:"probabilities"`
	Confidence    float64   `json:"confidence"`
}

// MLP state
var mlpWeights [][][]float64
var mlpBiases [][]float64

func initMLP() {
	data, err := os.ReadFile("aigo/mlp_weights.json")
	if err != nil {
		fmt.Println("Warning: Could not load MLP weights:", err)
		return
	}

	var weightsData struct {
		Weights [][][]float64 `json:"weights"`
		Biases  [][]float64   `json:"biases"`
	}
	if err := json.Unmarshal(data, &weightsData); err != nil {
		fmt.Println("Warning: Could not parse MLP weights:", err)
		return
	}

	mlpWeights = weightsData.Weights
	mlpBiases = weightsData.Biases
	fmt.Println("MLP weights loaded successfully")
}

func handleMLPPredict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MLPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Pixels) != 784 {
		http.Error(w, "Expected 784 pixels", http.StatusBadRequest)
		return
	}

	output := predictMLP(req.Pixels)
	class := argmaxMLP(output)
	confidence := 0.0
	if len(output) >= 2 {
		if output[0] > output[1] {
			confidence = output[0] - output[1]
		} else {
			confidence = output[1] - output[0]
		}
	}

	resp := MLPResponse{
		Class:         class,
		Probabilities: output,
		Confidence:    confidence,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func predictMLP(input []float64) []float64 {
	current := make([]float64, len(input))
	copy(current, input)

	for l := 0; l < len(mlpWeights); l++ {
		next := make([]float64, len(mlpWeights[l]))
		for j := 0; j < len(mlpWeights[l]); j++ {
			sum := mlpBiases[l][j]
			for k := 0; k < len(current); k++ {
				sum += current[k] * mlpWeights[l][j][k]
			}
			if l < len(mlpWeights)-1 {
				next[j] = math.Tanh(sum)
			} else {
				next[j] = sum
			}
		}
		if l == len(mlpWeights)-1 {
			next = softmaxMLP(next)
		}
		current = next
	}
	return current
}

func softmaxMLP(x []float64) []float64 {
	maxVal := x[0]
	for _, v := range x {
		if v > maxVal {
			maxVal = v
		}
	}
	exp := make([]float64, len(x))
	sum := 0.0
	for i, v := range x {
		exp[i] = math.Exp(v - maxVal)
		sum += exp[i]
	}
	for i := range exp {
		exp[i] /= sum
	}
	return exp
}

func argmaxMLP(a []float64) int {
	idx := 0
	for i := range a {
		if a[i] > a[idx] {
			idx = i
		}
	}
	return idx
}

func init() {
	initMLP()
}
