package controllers

import (
	"encoding/json"
	"github.com/eleby/aigo/MLP"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"sync"
)

// aigoNetwork holds the trained MLP network used by the live demo on the Aigo project page.
// It is loaded once at startup from pre-trained weights (produced with the real
// github.com/eleby/aigo training pipeline) and reused for every prediction request.
var (
	aigoNetwork   *MLP.Network
	aigoNetworkMu sync.Once
)

// aigoWeights mirrors the JSON structure exported by the aigo training pipeline.
type aigoWeights struct {
	LayerSizes   []int         `json:"layerSizes"`
	NumFeatures  int           `json:"numFeatures"`
	LearningRate float64       `json:"learningRate"`
	Weights      [][][]float64 `json:"weights"`
	Biases       [][]float64   `json:"biases"`
}

// InitAigoMLP loads the pre-trained MLP weights and builds the network using the
// real github.com/eleby/aigo/MLP package (no custom reimplementation of the algorithm).
func InitAigoMLP() {
	aigoNetworkMu.Do(func() {
		data, err := os.ReadFile("aigo/mlp_weights.json")
		if err != nil {
			return
		}
		var w aigoWeights
		if err := json.Unmarshal(data, &w); err != nil {
			return
		}
		network, err := MLP.CreateNetwork(MLP.Init{
			LearningRate: w.LearningRate,
			NumFeatures:  w.NumFeatures,
			LayerSizes:   w.LayerSizes,
		})
		if err != nil {
			return
		}
		if err := network.SetWeights(w.Weights); err != nil {
			return
		}
		if err := network.SetBiases(w.Biases); err != nil {
			return
		}
		aigoNetwork = network
	})
}

// aigoPredictRequest is the payload sent by the drawing canvas: a flat 16x16
// grid already preprocessed client-side to match the CenterNormalize format
// (1.0 for black/drawn pixels, 0.0 for white/background).
type aigoPredictRequest struct {
	Pixels []float64 `json:"pixels"`
}

// PredictAigoMLP handles POST /api/mlp/predict, running the real MLP.Network.PredictOne
// on the client-provided 16x16 normalized pixel grid.
func PredictAigoMLP() gin.HandlerFunc {
	return func(c *gin.Context) {
		if aigoNetwork == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "model not loaded"})
			return
		}
		var req aigoPredictRequest
		if err := c.ShouldBindJSON(&req); err != nil || len(req.Pixels) != 256 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expected 256 pixels (16x16 grid)"})
			return
		}
		prediction := aigoNetwork.PredictOne(req.Pixels)
		label := "circle"
		if prediction.Class == 1 {
			label = "cross"
		}
		c.JSON(http.StatusOK, gin.H{
			"class":         prediction.Class,
			"label":         label,
			"confidence":    prediction.Confidence,
			"probabilities": prediction.Probabilities,
		})
	}
}
