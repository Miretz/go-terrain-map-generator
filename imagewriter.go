package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
)

type color struct {
	r float64
	g float64
	b float64
}

func Clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func MapRange(x, inMin, inMax, outMin, outMax float64) float64 {
	slope := 1.0 * (outMax - outMin) / (inMax - inMin)
	return outMin + math.Round(slope*(x-inMin))
}

func WriteColor(pixelColor *color) string {
	return fmt.Sprintf("%d %d %d\n",
		int(math.Round(MapRange(pixelColor.r, 0.0, 1.0, 0.0, 255.0))),
		int(math.Round(MapRange(pixelColor.g, 0.0, 1.0, 0.0, 255.0))),
		int(math.Round(MapRange(pixelColor.b, 0.0, 1.0, 0.0, 255.0))))
}

func WriteToPPMFile(outputFile string, imageWidth int, imageHeight int, colorData []color) {
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	writer := bufio.NewWriter(file)
	writer.WriteString(fmt.Sprintf("P3\n%d %d\n255\n", imageWidth, imageHeight))
	for _, c := range colorData {
		_, err := writer.WriteString(WriteColor(&c))
		if err != nil {
			log.Fatalf("Error while writing to file. Err: %s", err.Error())
		}
	}
	writer.Flush()
}
