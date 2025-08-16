package bottools
import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)
var (
	GoldenRatioConjugate = (3.0 - math.Sqrt(5.0)) / 2.0
)
type ColorGenerator interface {
	NextColor() string
}
type goldenRatioGenerator struct {
	hue        float64
	saturation float64
	lightness  float64
	mu         sync.Mutex
}
func NewGoldenRatioGenerator(saturation, lightness float64) ColorGenerator {
	return &goldenRatioGenerator{
		hue:        rand.New(rand.NewSource(time.Now().UnixNano())).Float64(),
		saturation: saturation,
		lightness:  lightness,
	}
}
func (g *goldenRatioGenerator) NextColor() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.hue = math.Mod(g.hue+GoldenRatioConjugate, 1.0)
	r, gr, b := hlsToRgb(g.hue, g.lightness, g.saturation)
	return fmt.Sprintf("#%02X%02X%02X", 
		int(r*255), int(gr*255), int(b*255))
}
func hlsToRgb(h, l, s float64) (r, g, b float64) {
	if s == 0 {
		r, g, b = l, l, l  
		return
	}
	var c2 float64
	if l < 0.5 {
		c2 = l * (1 + s)
	} else {
		c2 = l + s - l*s
	}
	c1 := 2*l - c2
	r = hueToRgb(c1, c2, h+1.0/3.0)
	g = hueToRgb(c1, c2, h)
	b = hueToRgb(c1, c2, h-1.0/3.0)
	return
}
func hueToRgb(c1, c2, hue float64) float64 {
	if hue < 0 {
		hue += 1
	}
	if hue > 1 {
		hue -= 1
	}
	if hue < 1.0/6.0 {
		return c1 + (c2-c1)*6*hue
	}
	if hue < 1.0/2.0 {
		return c2
	}
	if hue < 2.0/3.0 {
		return c1 + (c2-c1)*(2.0/3.0-hue)*6
	}
	return c1
}
type ColorMap struct {
	colorMap  map[string]string
	generator ColorGenerator
	mu        sync.RWMutex
}
func NewColorMap() *ColorMap {
	return &ColorMap{
		colorMap:  make(map[string]string),
		generator: NewGoldenRatioGenerator(0.8, 0.75),
	}
}
func (cm *ColorMap) Get(key string) string {
	cm.mu.RLock()
	if color, exists := cm.colorMap[key]; exists {
		cm.mu.RUnlock()
		return color
	}
	cm.mu.RUnlock()
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if color, exists := cm.colorMap[key]; exists {
		return color
	}
	color := cm.generator.NextColor()
	cm.colorMap[key] = color
	return color
}
func (cm *ColorMap) GetAll() map[string]string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	result := make(map[string]string, len(cm.colorMap))
	for k, v := range cm.colorMap {
		result[k] = v
	}
	return result
}
func (cm *ColorMap) Reset() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.colorMap = make(map[string]string)
	cm.generator = NewGoldenRatioGenerator(0.8, 0.75)
}
func (cm *ColorMap) SetColor(key, color string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.colorMap[key] = color
}
func (cm *ColorMap) Remove(key string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.colorMap, key)
}
func (cm *ColorMap) Size() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.colorMap)
}