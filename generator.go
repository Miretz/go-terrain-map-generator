package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Fade function as defined by Ken Perlin. This will smooth out the result.
func Fade(t float64) float64 {
	t = math.Abs(t)
	return t * t * t * (t*(t*6-15) + 10)
}

// Generates a randomly arranged array of 512 values ranging between 0-255 inclusive
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

// Generates 2d gradients which are uses for composing the final noise pattern
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

// Applies Fade on a 2d vector and multiplies the it's values
func Q(uv vec2) float64 {
	return Fade(uv.x) * Fade(uv.y)
}

// Generate a perlin noise point from predefined permutations and gradients
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

// Cubic interpolation of the given values float values
func CubicInterpolation(p []float64, x float64) float64 {
	return cubicInterpAux(p[0], p[1], p[2], p[3], x)
}

func cubicInterpAux(v0, v1, v2, v3, x float64) float64 {
	P := (v3 - v2) - (v0 - v1)
	Q := (v0 - v1) - P
	R := v2 - v0
	S := v1
	return P*x*x*x + Q*x*x + R*x + S
}

// Generates a perlin noise value where each point is stretched over multiple points and smoothened
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

// Generates a unique noise pattern with the given parameters
func GenerateNoiseMap(width int, height int, octave float64, stretch float64, multiplier float64) []float64 {

	rand.Seed(time.Now().UnixNano())

	perms := GeneratePermutations()
	grads := GenerateGradients()
	data := make([]float64, width*height)

	index := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := vec2{float64(x) * octave, float64(y) * octave}
			res := StretchedNoise(pos, perms, grads, stretch) * multiplier
			res = Clamp(res, 0.0, 1.0)
			res *= 0.5
			data[index] = res
			index++
		}
	}

	return data
}

// Merges multiple noise patterns and applies redistribution and water level cutoff
func MergeNoiseData(multipliers []float64, redistribution float64, waterHeight float64, layers ...[]float64) []float64 {
	result := make([]float64, len(layers[0]))

	sumMultipliers := 0.0
	for _, m := range multipliers {
		sumMultipliers += m
	}

	for i := range result {
		res := 0.0
		for _, ns := range layers {
			res += ns[i]
		}
		res /= sumMultipliers
		res = math.Pow(res, redistribution)
		if res < waterHeight {
			res = waterHeight
		}
		result[i] = Clamp(res, 0.0, 1.0)
	}

	return result
}

// Converts a float64 array into an array of color values using the same value per each channel
func NoiseDataToColor(noiseData []float64) []color {
	colors := make([]color, len(noiseData))
	for i, n := range noiseData {
		colors[i] = color{n, n, n}
	}
	return colors
}

// Generates a terrain heightmap using perlin noise and store the result in a PPM file
func RenderToImage() {
	const imageWidth = 600
	const imageHeight = 600

	const redistribution = 0.72
	const waterHeight = 0.1

	multipliers := []float64{1.0, 0.5, 0.25}

	noiseData1 := GenerateNoiseMap(imageWidth, imageHeight, 1, 100, multipliers[0])
	noiseData2 := GenerateNoiseMap(imageWidth, imageHeight, 2, 100, multipliers[1])
	noiseData3 := GenerateNoiseMap(imageWidth, imageHeight, 4, 100, multipliers[2])

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
