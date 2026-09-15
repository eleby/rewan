package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Weights struct {
	W [][][]float64 `json:"weights"`
	B [][]float64   `json:"biases"`
}

func main() {
	dir := "/var/www/html/rewan.dev/aigo/training"
	out := "/var/www/html/rewan.dev/aigo/mlp_weights.json"
	os.Remove(out)
	
	fmt.Println("Chargement des images depuis", dir)
	images, labels := loadImages(dir)
	fmt.Printf("%d images chargees (%d cercles, %d croix)\n", len(images), count(labels, 0), count(labels, 1))
	
	if len(images) == 0 {
		log.Fatal("Aucune image trouvee!")
	}
	
	sizes := []int{256, 128, 64, 2}
	lr := 0.01
	epochs := 100
	
	fmt.Printf("Entrainement MLP %v...\n", sizes)
	start := time.Now()
	w, b := train(images, labels, sizes, lr, epochs)
	fmt.Printf("Termine en %v\n", time.Since(start))
	
	data := Weights{W: w, B: b}
	f, _ := os.Create(out)
	defer f.Close()
	json.NewEncoder(f).Encode(data)
	fmt.Println("Sauvegarde dans", out)
}

func count(l []int, t int) int {
	c := 0
	for _, v := range l {
		if v == t {
			c++
		}
	}
	return c
}

func loadImages(dir string) ([][]float64, []int) {
	var imgs [][]float64
	var lbls []int
	
	entries, _ := os.ReadDir(dir)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".png") {
			continue
		}
		
		path := filepath.Join(dir, e.Name())
		p := preprocess(path)
		
		if strings.HasPrefix(e.Name(), "circle") {
			lbls = append(lbls, 0)
		} else if strings.HasPrefix(e.Name(), "cross") {
			lbls = append(lbls, 1)
		} else {
			continue
		}
		imgs = append(imgs, p)
	}
	
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(imgs) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		imgs[i], imgs[j] = imgs[j], imgs[i]
		lbls[i], lbls[j] = lbls[j], lbls[i]
	}
	return imgs, lbls
}

func preprocess(path string) []float64 {
	f, err := os.Open(path)
	if err != nil {
		return make([]float64, 256)
	}
	defer f.Close()
	
	img, err := png.Decode(f)
	if err != nil {
		return make([]float64, 256)
	}
	
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	
	// Binariser
	bin := make([][]bool, h)
	for y := 0; y < h; y++ {
		bin[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			c := color.GrayModel.Convert(img.At(bounds.Min.X+x, bounds.Min.Y+y))
			gray := c.(color.Gray).Y
			bin[y][x] = gray < 128 // true = encre
		}
	}
	
	// Bounding box
	minX, minY := w, h
	maxX, maxY := 0, 0
	has := false
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if bin[y][x] {
				has = true
				if x < minX { minX = x }
				if y < minY { minY = y }
				if x > maxX { maxX = x }
				if y > maxY { maxY = y }
			}
		}
	}
	
	if !has {
		return make([]float64, 256)
	}
	
	cropW := maxX - minX + 1
	cropH := maxY - minY + 1
	
	// Scale 14x14
	scW, scH := 14, 14
	scaled := make([][]bool, scH)
	for y := 0; y < scH; y++ {
		scaled[y] = make([]bool, scW)
		for x := 0; x < scW; x++ {
			sx := int(float64(x) / float64(scW) * float64(cropW))
			sy := int(float64(y) / float64(scH) * float64(cropH))
			if sx >= cropW { sx = cropW - 1 }
			if sy >= cropH { sy = cropH - 1 }
			scaled[y][x] = bin[minY+sy][minX+sx]
		}
	}
	
	// Center 16x16
	offX := (16 - scW) / 2
	offY := (16 - scH) / 2
	final := make([]bool, 256)
	for y := 0; y < scH; y++ {
		for x := 0; x < scW; x++ {
			final[(offY+y)*16+(offX+x)] = scaled[y][x]
		}
	}
	
	out := make([]float64, 256)
	for i := 0; i < 256; i++ {
		if final[i] {
			out[i] = 1.0
		}
	}
	return out
}

func train(images [][]float64, labels []int, sizes []int, lr float64, epochs int) ([][][]float64, [][]float64) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	w := make([][][]float64, len(sizes)-1)
	b := make([][]float64, len(sizes)-1)
	for l := 0; l < len(sizes)-1; l++ {
		limit := math.Sqrt(6.0 / float64(sizes[l]+sizes[l+1]))
		w[l] = make([][]float64, sizes[l+1])
		for j := 0; j < sizes[l+1]; j++ {
			w[l][j] = make([]float64, sizes[l])
			for k := 0; k < sizes[l]; k++ {
				w[l][j][k] = (r.Float64()*2 - 1) * limit
			}
		}
		b[l] = make([]float64, sizes[l+1])
	}
	
	for epoch := 0; epoch < epochs; epoch++ {
		correct := 0
		for i := range images {
			out := fwd(w, b, images[i])
			if argmax(out) == labels[i] {
				correct++
			}
			bwd(w, b, images[i], labels[i], lr)
		}
		if epoch%10 == 0 || epoch == epochs-1 {
			fmt.Printf("Epoch %3d: %.1f%%\n", epoch+1, float64(correct)/float64(len(images))*100)
		}
	}
	return w, b
}

func fwd(w [][][]float64, b [][]float64, input []float64) []float64 {
	cur := input
	for l := 0; l < len(w); l++ {
		next := make([]float64, len(w[l]))
		for j := 0; j < len(w[l]); j++ {
			s := b[l][j]
			for k := 0; k < len(cur); k++ {
				s += cur[k] * w[l][j][k]
			}
			if l < len(w)-1 {
				next[j] = math.Tanh(s)
			} else {
				next[j] = s
			}
		}
		if l == len(w)-1 {
			next = soft(next)
		}
		cur = next
	}
	return cur
}

func soft(x []float64) []float64 {
	m := x[0]
	for _, v := range x {
		if v > m {
			m = v
		}
	}
	e := make([]float64, len(x))
	sum := 0.0
	for i, v := range x {
		e[i] = math.Exp(v - m)
		sum += e[i]
	}
	for i := range e {
		e[i] /= sum
	}
	return e
}

func bwd(w [][][]float64, b [][]float64, input []float64, label int, lr float64) {
	acts := [][]float64{input}
	cur := input
	for l := 0; l < len(w); l++ {
		next := make([]float64, len(w[l]))
		for j := 0; j < len(w[l]); j++ {
			s := b[l][j]
			for k := 0; k < len(cur); k++ {
				s += cur[k] * w[l][j][k]
			}
			if l < len(w)-1 {
				next[j] = math.Tanh(s)
			} else {
				next[j] = s
			}
		}
		if l == len(w)-1 {
			next = soft(next)
		}
		acts = append(acts, next)
		cur = next
	}
	
	target := make([]float64, len(acts[len(acts)-1]))
	target[label] = 1
	
	deltas := make([][]float64, len(w))
	ol := len(w) - 1
	deltas[ol] = make([]float64, len(acts[len(acts)-1]))
	for j := range deltas[ol] {
		deltas[ol][j] = acts[len(acts)-1][j] - target[j]
	}
	for l := ol - 1; l >= 0; l-- {
		deltas[l] = make([]float64, len(acts[l+1]))
		for j := range deltas[l] {
			err := 0.0
			for k := range deltas[l+1] {
				err += deltas[l+1][k] * w[l+1][k][j]
			}
			deltas[l][j] = err * (1 - math.Pow(acts[l+1][j], 2))
		}
	}
	
	for l := range w {
		for j := range w[l] {
			b[l][j] -= lr * deltas[l][j]
			for k := range w[l][j] {
				w[l][j][k] -= lr * deltas[l][j] * acts[l][k]
			}
		}
	}
}

func argmax(a []float64) int {
	idx := 0
	for i := 1; i < len(a); i++ {
		if a[i] > a[idx] {
			idx = i
		}
	}
	return idx
}
