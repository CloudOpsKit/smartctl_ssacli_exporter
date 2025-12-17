package parser

import (
	"log"
	"strconv"
	"strings"
)

func toINT(s string) int64 {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalln(err)
	}
	return int64(i)
}

func toFLO(s string) float64 {
	i, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatalln(err)
	}
	return float64(i)
}

func trim(s string) string {
	return strings.Trim(s, " \t")
}

func parseSmartRawValue(s string) *float64 {
	// 1. Clean up excess garbage.
	// If input is "0/200164573", take the part before "/".
	// If input is "26", keep "26".
	if idx := strings.Index(s, "/"); idx != -1 {
		s = s[:idx]
	}

	// Just in case, trim spaces (although strings.Fields usually handles this).
	s = strings.TrimSpace(s)

	// 2. Parse float
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// If parsing failed (not a number), return nil.
		return nil
	}

	// 3. Return pointer
	return &val
}
