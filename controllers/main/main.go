package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"rewan/controllers"
	"rewan/controllers/routines"
	"rewan/models"
	"rewan/models/memorygame"
	"rewan/models/persistence"
	"strconv"
)

var funcMap = template.FuncMap{
	"msg":         controllers.Msg,
	"timeMachine": controllers.TimeMachine,
	"year":        controllers.GetCurrentYear,
	"pieceColor":  memorygame.GetColorPiece,
	"cornerClass": memorygame.GetCornerClass,
}

var mlpWeights [][][]float64
var mlpBiases [][]float64

type MLPRequest struct {
	Pixels []float64 `json:"pixels"`
}

type MLPResponse struct {
	Class         int       `json:"class"`
	Probabilities []float64 `json:"probabilities"`
	Confidence    float64   `json:"confidence"`
}

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

	if len(req.Pixels) != 256 {
		http.Error(w, "Expected 256 pixels", http.StatusBadRequest)
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

func main() {
	releaseMode := os.Args[2] != "dev"
	if releaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	f, _ := os.OpenFile(models.Conf.LogDir+models.Conf.LogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	ferr, _ := os.OpenFile(models.Conf.LogDir+models.Conf.ErrLogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	multiWriter := io.MultiWriter(f, os.Stdout)
	errWriter := io.MultiWriter(ferr, os.Stdout)
	gin.DefaultWriter = multiWriter
	gin.DefaultErrorWriter = errWriter
	log.SetPrefix("[LOG] ")
	log.SetOutput(multiWriter)

	r := gin.New()
	controllers.InitBundle()
	r.Static("/resources", "./resources")
	r.SetFuncMap(funcMap)
	r.LoadHTMLGlob("views/**/*")

	r.Use(func(c *gin.Context) {
		lang := c.Request.Header.Get("Accept-Language")
		c.Set("lang", lang)
		c.Next()
	})

	r.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		errorCode := persistence.GenerateUID()
		var logger *log.Logger
		logger = log.New(gin.DefaultErrorWriter, "[ERR] ", log.LstdFlags)
		errStr := []byte(fmt.Sprintf("%v", err))
		logger.Print("Error '" + string(errStr) + "' logged with " + errorCode)
		lang := controllers.GetLang(c)
		title := controllers.MsgStr("500.title", lang, string(errStr))
		content := controllers.MsgStr("500.content", lang, errorCode)
		c.HTML(http.StatusInternalServerError, "error.html", controllers.Render(c, gin.H{"title": "500",
			"titleError": template.HTML(title), "contentError": template.HTML(content)}))
	}))

	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/resources/*"}}))

	// Initialize MLP with real images
	initMLP()

	// Routes
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", controllers.Render(c, gin.H{"title": "index"}))
	})
	r.GET("/career", func(c *gin.Context) {
		c.HTML(http.StatusOK, "career.html", controllers.Render(c, gin.H{"title": "career"}))
	})
	r.GET("/projects", func(c *gin.Context) {
		c.HTML(http.StatusOK, "projects.html", controllers.Render(c, gin.H{"title": "projects"}))
	})
	r.GET("/projects/aigo", func(c *gin.Context) {
		c.HTML(http.StatusOK, "aigo.html", controllers.Render(c, gin.H{"title": "aigo"}))
	})
	r.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "contact.html", controllers.Render(c, gin.H{"title": "contact"}))
	})
	r.GET("/pages", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pages.html", controllers.Render(c, gin.H{"title": "pages"}))
	})

	// MLP API route
	r.POST("/api/mlp/predict", func(c *gin.Context) {
		var req MLPRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(req.Pixels) != 256 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expected 256 pixels (16x16)"})
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
		c.JSON(http.StatusOK, MLPResponse{
			Class:         class,
			Probabilities: output,
			Confidence:    confidence,
		})
	})

	r.GET("/about/bootstrap", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		titleContent := controllers.MsgStr("bootstrap.main.title", lang, nil)
		textContent := controllers.MsgStr("bootstrap.main.content", lang, nil)
		alt := controllers.MsgStr("alt.wiki.bootstrap", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "bootstrap", "image": "bootstrap.svg", "alt": alt,
				"credits": "", "titleContent": titleContent, "textContent": textContent,
				"website": "https://getbootstrap.com/", "wiki": "https://en.wikipedia.org/wiki/Bootstrap_(framework)"}))
	})
	r.GET("/about/play", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		titleContent := controllers.MsgStr("play.main.title", lang, nil)
		textContent := controllers.MsgStr("play.main.content", lang, nil)
		alt := controllers.MsgStr("alt.wiki.play", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "play", "image": "play.svg", "alt": alt,
				"credits": "", "titleContent": titleContent, "textContent": textContent,
				"website": "https://www.playframework.com/", "wiki": "https://en.wikipedia.org/wiki/Play_Framework"}))
	})
	r.GET("/about/letsencrypt", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		titleContent := controllers.MsgStr("letsencrypt.main.title", lang, nil)
		textContent := controllers.MsgStr("letsencrypt.main.content", lang, nil)
		alt := controllers.MsgStr("alt.wiki.letsencrypt", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "letsencrypt", "image": "letsencrypt.svg", "alt": alt,
				"credits": "", "titleContent": titleContent, "textContent": textContent,
				"website": "https://letsencrypt.org/", "wiki": "https://en.wikipedia.org/wiki/Let%27s_Encrypt"}))
	})
	r.GET("/about/gandi", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		titleContent := controllers.MsgStr("gandi.main.title", lang, nil)
		textContent := controllers.MsgStr("gandi.main.content", lang, nil)
		alt := controllers.MsgStr("alt.wiki.gandi", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "gandi", "image": "gandi.svg", "alt": alt,
				"credits": "", "titleContent": titleContent, "textContent": textContent,
				"website": "https://www.gandi.net/en", "wiki": "https://en.wikipedia.org/wiki/Gandi"}))
	})
	r.GET("/about/go", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		titleContent := controllers.MsgStr("go.main.title", lang, nil)
		textContent := controllers.MsgStr("go.main.content", lang, nil)
		alt := controllers.MsgStr("alt.wiki.go", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "go", "image": "go.svg", "alt": alt,
				"credits": "", "titleContent": titleContent, "textContent": textContent,
				"website": "https://golang.org/", "wiki": "https://fr.wikipedia.org/wiki/Go_(langage)"}))
	})
	r.GET("/about/gin", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		titleContent := controllers.MsgStr("gin.main.title", lang, nil)
		textContent := controllers.MsgStr("gin.main.content", lang, nil)
		alt := controllers.MsgStr("alt.wiki.gin", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "gin", "image": "gin.png", "alt": alt,
				"credits": "", "titleContent": titleContent, "textContent": template.HTML(textContent),
				"website": "https://gin-gonic.com/"}))
	})
	r.GET("/about/rouen", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		credits := controllers.MsgStr("rouen.credits", lang, nil)
		titleContent := controllers.MsgStr("rouen.main.title", lang, nil)
		textContent := controllers.MsgStr("rouen.main.content", lang, nil)
		alt := controllers.MsgStr("alt.wiki.rouen", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "rouen", "image": "rouen.jpg", "alt": alt,
				"credits": credits, "titleContent": titleContent, "textContent": textContent,
				"website": "https://rouen.fr/", "wiki": "https://fr.wikipedia.org/wiki/Rouen"}))
	})
	r.GET("/about/2.0", func(c *gin.Context) {
		c.HTML(http.StatusOK, "version.html", controllers.Render(c, gin.H{"title": "version"}))
	})
	r.GET("/tools/memory", func(c *gin.Context) {
		c.HTML(http.StatusOK, "memory.html", controllers.Render(c, gin.H{"title": "memory"}))
	})
	r.GET("/tools/memory/game", func(c *gin.Context) {
		game := memorygame.GetSetMemoryGame(c)
		c.HTML(http.StatusOK, "memorygame.html", controllers.Render(c, gin.H{"title": "memory", "board": game, "pieceType": memorygame.PieceType, "gameState": memorygame.GameState}))
	})
	r.GET("/tools/memory/game/reset", memorygame.Reset())
	r.GET("/tools/memory/game/finish", memorygame.Finish())
	r.GET("/tools/time", func(c *gin.Context) {
		c.HTML(http.StatusOK, "time.html", controllers.Render(c, gin.H{"title": "time"}))
	})
	r.GET("/tools/memory/game/row/:rowId/piece/:pieceId/type", memorygame.GetPieceType())
	r.POST("/contact/mail", controllers.SendMail())
	r.GET("/tools/time/get/:query/:year/:month/:day/:hour/:minute", controllers.GetTimeMachine())
	r.GET("/sitemap.xml", controllers.RenderSitemap())
	r.GET("/robots.txt", renderRobotsTxtFile)

	r.NoRoute(func(c *gin.Context) {
		lang := controllers.GetLang(c)
		title := controllers.MsgStr("404.title", lang, c.Request.URL)
		content := controllers.MsgStr("404.content", lang, nil)
		c.HTML(http.StatusInternalServerError, "error.html", controllers.Render(c, gin.H{"title": "404",
			"titleError": template.HTML(title), "contentError": template.HTML(content)}))
	})

	controllers.GenerateSitemap(r)
	routines.ManageLogsRoutine()
	_ = r.Run(":" + strconv.Itoa(models.Conf.Port))
}

func renderRobotsTxtFile(c *gin.Context) {
	c.File("robots.txt")
}
