package visualization

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Canvas struct {
	width    float64
	height   float64
	offsetX  float64
	offsetY  float64
	scale    float64
	gridSize float64
	points   []Point
	curves   [][]Point
}

type Point struct {
	X float64
	Y float64
}

func NewCanvas(width, height float64) *Canvas {
	return &Canvas{
		width:    width,
		height:   height,
		offsetX:  width / 2,
		offsetY:  height / 2,
		scale:    3.0,
		gridSize: 30.0,
		points:   make([]Point, 0),
		curves:   make([][]Point, 0),
	}
}

func (c *Canvas) Clear() {
	c.points = make([]Point, 0)
	c.curves = make([][]Point, 0)
}

func (c *Canvas) AddPoint(x, y float64) {
	c.points = append(c.points, Point{X: x, Y: y})
}

func (c *Canvas) AddCurve(points []Point) {
	c.curves = append(c.curves, points)
}

func (c *Canvas) Draw(screen *ebiten.Image) {
	c.drawGrid(screen)
	c.drawAxis(screen)
	c.drawPoints(screen)
	c.drawCurves(screen)
}

func (c *Canvas) drawGrid(screen *ebiten.Image) {
	gridColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}

	for x := c.offsetX; x < c.width; x += c.gridSize {
		vector.StrokeLine(screen, float32(x), 0, float32(x), float32(c.height), 1, gridColor, true)
	}
	for x := c.offsetX; x >= 0; x -= c.gridSize {
		vector.StrokeLine(screen, float32(x), 0, float32(x), float32(c.height), 1, gridColor, true)
	}
	for y := c.offsetY; y < c.height; y += c.gridSize {
		vector.StrokeLine(screen, 0, float32(y), float32(c.width), float32(y), 1, gridColor, true)
	}
	for y := c.offsetY; y >= 0; y -= c.gridSize {
		vector.StrokeLine(screen, 0, float32(y), float32(c.width), float32(y), 1, gridColor, true)
	}
}

func (c *Canvas) drawAxis(screen *ebiten.Image) {
	axisColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}

	vector.StrokeLine(screen, 0, float32(c.offsetY), float32(c.width), float32(c.offsetY), 2, axisColor, true)
	vector.StrokeLine(screen, float32(c.offsetX), 0, float32(c.offsetX), float32(c.height), 2, axisColor, true)
}

func (c *Canvas) drawPoints(screen *ebiten.Image) {
	pointColor := color.RGBA{R: 255, G: 100, B: 100, A: 255}

	for _, p := range c.points {
		screenX := c.offsetX + p.X*c.scale
		screenY := c.offsetY - p.Y*c.scale
		vector.StrokeLine(screen, float32(screenX)-3, float32(screenY), float32(screenX)+3, float32(screenY), 2, pointColor, true)
		vector.StrokeLine(screen, float32(screenX), float32(screenY)-3, float32(screenX), float32(screenY)+3, 2, pointColor, true)
	}
}

func (c *Canvas) drawCurves(screen *ebiten.Image) {
	curveColor := color.RGBA{R: 100, G: 150, B: 255, A: 255}

	for _, curve := range c.curves {
		if len(curve) < 2 {
			continue
		}

		for i := 0; i < len(curve)-1; i++ {
			p1 := curve[i]
			p2 := curve[i+1]

			screenX1 := c.offsetX + p1.X*c.scale
			screenY1 := c.offsetY - p1.Y*c.scale
			screenX2 := c.offsetX + p2.X*c.scale
			screenY2 := c.offsetY - p2.Y*c.scale

			vector.StrokeLine(screen, float32(screenX1), float32(screenY1), float32(screenX2), float32(screenY2), 2, curveColor, true)
		}
	}
}

func (c *Canvas) WorldToScreen(x, y float64) (float64, float64) {
	return c.offsetX + x*c.scale, c.offsetY - y*c.scale
}

func (c *Canvas) ScreenToWorld(screenX, screenY float64) (float64, float64) {
	return (screenX - c.offsetX) / c.scale, (c.offsetY - screenY) / c.scale
}

func GeneratePolynomialCurve(coeffs []int64, prime int64, start, end, step float64) []Point {
	points := make([]Point, 0)

	for x := start; x <= end; x += step {
		y := evaluatePolynomial(coeffs, x, prime)
		points = append(points, Point{X: x, Y: float64(y)})
	}

	return points
}

func evaluatePolynomial(coeffs []int64, x float64, prime int64) int64 {
	result := int64(0)
	power := int64(1)

	for _, coeff := range coeffs {
		result = (result + coeff*int64(power)) % prime
		if result < 0 {
			result += prime
		}
		power = (power * int64(x)) % prime
	}

	return result
}

func (c *Canvas) DrawPolynomial(screen *ebiten.Image, coeffs []int64, prime int64) {
	curve := GeneratePolynomialCurve(coeffs, prime, -10, 10, 0.5)
	c.AddCurve(curve)
	c.Draw(screen)
}

func (c *Canvas) DrawShare(screen *ebiten.Image, x, y int64) {
	hue := float64(x%360) / 360.0
	r := uint8(128 + 127*hue)
	g := uint8(128 + 127*(1-hue))
	b := uint8(200)
	pointColor := color.RGBA{R: r, G: g, B: b, A: 255}

	screenX := c.offsetX + float64(x)*c.scale
	screenY := c.offsetY - float64(y)*c.scale

	vector.StrokeLine(screen, float32(screenX)-4, float32(screenY), float32(screenX)+4, float32(screenY), 2, pointColor, true)
	vector.StrokeLine(screen, float32(screenX), float32(screenY)-4, float32(screenX), float32(screenY)+4, 2, pointColor, true)
}

func (c *Canvas) SetScale(scale float64) {
	c.scale = scale
}

func (c *Canvas) SetOffset(x, y float64) {
	c.offsetX = x
	c.offsetY = y
}
