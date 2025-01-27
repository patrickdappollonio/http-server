package utils

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

var symbols = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

// index returns the index of the symbol to use for the given size.
func index(s int64) float64 {
	x := math.Log(float64(s)) / math.Log(1024)
	return math.Floor(x)
}

// countSize returns the size in the given index.
func countSize(s int64, i float64) float64 {
	return float64(s) / math.Pow(1024, math.Floor(i))
}

// Humansize converts a size in bytes to a human-readable string.
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

// ParseSize converts a size string (e.g., "2M") to its equivalent in bytes.
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty size string")
	}

	// Find the index where the numeric part ends and the suffix begins
	var i int
	for i = 0; i < len(s); i++ {
		if !unicode.IsDigit(rune(s[i])) {
			break
		}
	}

	if i == 0 {
		return 0, errors.New("no numeric value found")
	}

	numStr := s[:i]
	suffix := strings.ToUpper(strings.TrimSpace(s[i:]))

	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %w", err)
	}

	var multiplier int64

	switch suffix {
	case "", "B":
		multiplier = 1
	case "K":
		multiplier = 1024
	case "M":
		multiplier = 1024 * 1024
	case "G":
		multiplier = 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("invalid size suffix: %s", suffix)
	}

	return num * multiplier, nil
}
