package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func Fade(t float64) float64 {
	t = math.Abs(t)
	return t * t * t * (t*(t*6-15) + 10)
}

func GeneratePermutations() []int {
	temp := make([]int, 256)
	for i := 0; i < 256; i++ {
		temp[i] = i
	}
	rand.Shuffle(len(temp), func(i, j int) {
		temp[i], temp[j] = temp[j], temp[i]
	})
	result := make([]int, 512)
	for i := 0; i < 512; i++ {
		result[i] = temp[i%256]
	}
	return result
}

func GenerateGradients() []vec2 {
	grads := make([]vec2, 256)

	for i := 0; i < len(grads); i++ {
		var gradient vec2
		for {
			gradient = vec2{rand.Float64()*2 - 1, rand.Float64()*2 - 1}
			if gradient.LengthSquared() >= 1.0 {
				break
			}
		}
		gradient.Normalize()
		grads[i] = gradient
	}

	return grads
}

func Q(uv vec2) float64 {
	return Fade(uv.x) * Fade(uv.y)
}

func Noise(pos vec2, perms []int, grads []vec2) float64 {
	cell := vec2{math.Floor(pos.x), math.Floor(pos.y)}
	total := 0.0
	corners := [4]vec2{{0, 0}, {0, 1}, {1, 0}, {1, 1}}
	for _, corner := range corners {
		ij := cell.Add(&corner)
		uv := pos.Sub(&ij)
		index := perms[int(ij.x)%len(perms)]
		index = perms[(index+int(ij.y))%len(perms)]
		grad := grads[index%len(grads)]
		total += Q(uv) * grad.Dot(&uv)
	}
	return math.Max(math.Min(total, 1.0), -1.0)
}

func CubicInterpolation(p []float64, x float64) float64 {
	return CubicInterpAux(p[0], p[1], p[2], p[3], x)
}

func CubicInterpAux(v0, v1, v2, v3, x float64) float64 {
	P := (v3 - v2) - (v0 - v1)
	Q := (v0 - v1) - P
	R := v2 - v0
	S := v1
	return P*x*x*x + Q*x*x + R*x + S
}

func StretchedNoise(pos vec2, perms []int, grads []vec2, stretch float64) float64 {
	xf := pos.x / stretch
	yf := pos.y / stretch
	x := int(math.Floor(xf))
	y := int(math.Floor(yf))
	fracX := xf - float64(x)
	fracY := yf - float64(y)
	p := make([]float64, 4)
	for j := 0; j < 4; j++ {
		p2 := make([]float64, 4)
		for i := 0; i < 4; i++ {
			p2[i] = Noise(
				vec2{float64(x + i), float64(y + j)},
				perms,
				grads)
		}
		p[j] = CubicInterpolation(p2, fracX)
	}
	return CubicInterpolation(p, fracY)
}

func GenerateNoiseMap(width int, height int, octave float64, stretch float64, multiplier float64) []float64 {

	perms := GeneratePermutations()
	grads := GenerateGradients()
	data := make([]float64, width*height)

	index := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := vec2{float64(x) * octave, float64(y) * octave}
			data[index] = StretchedNoise(pos, perms, grads, stretch) * multiplier
			data[index] *= 0.5
			index++
		}
	}

	return data
}

func MergeNoiseData(multipliers []float64, redistribution float64, waterHeight float64, layers ...[]float64) []float64 {
	result := make([]float64, len(layers[0]))

	sumMultipliers := 0.0
	for _, m := range multipliers {
		sumMultipliers += m
	}

	for i := range result {
		for _, ns := range layers {
			result[i] += ns[i]
		}
		result[i] /= sumMultipliers
		result[i] = math.Pow(result[i], redistribution)
		if result[i] < waterHeight {
			result[i] = 0.0
		}
	}

	return result
}

func NoiseDataToColor(noiseData []float64) []color {
	colors := make([]color, len(noiseData))
	for i, n := range noiseData {
		colors[i] = color{n, n, n}
	}
	return colors
}

func RenderToImage() {
	const imageWidth = 600
	const imageHeight = 600

	const redistribution = 1.2
	const waterHeight = 0.1

	multipliers := []float64{1.0, 0.5, 0.25}

	noiseData1 := GenerateNoiseMap(imageWidth, imageHeight, 1, imageWidth/10, multipliers[0])
	noiseData2 := GenerateNoiseMap(imageWidth, imageHeight, 2, imageWidth/20, multipliers[1])
	noiseData3 := GenerateNoiseMap(imageWidth, imageHeight, 4, imageWidth/40, multipliers[2])

	mapData := MergeNoiseData(multipliers, redistribution, waterHeight, noiseData1, noiseData2, noiseData3)
	colors := NoiseDataToColor(mapData)

	WriteToPPMFile("output.ppm", imageWidth, imageHeight, colors)
}

func main() {
	start := time.Now()

	RenderToImage()

	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println("Elapsed Time", elapsed.Seconds(), "seconds")
}
