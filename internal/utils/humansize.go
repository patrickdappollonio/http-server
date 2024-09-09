package utils

import (
	"fmt"
	"math"
)

var symbols = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

func index(s int64) float64 {
	x := math.Log(float64(s)) / math.Log(1024)
	return math.Floor(x)
}

func countSize(s int64, i float64) float64 {
	return float64(s) / math.Pow(1024, math.Floor(i))
}

func Humansize(s int64) string {
	if s < 10 {
		return fmt.Sprintf("%dB", s)
	}

	i := index(s)
	size := countSize(s, i)
	format := "%.0f"

	if size < 10 {
		format = "%.1f"
	}

	return fmt.Sprintf(format+" %s", size, symbols[int(i)])
}
