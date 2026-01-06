package utils

import (
	"strconv"
	"strings"
)

func ParseStockFloat(s string) (float64, error) {
	dotIndex := strings.LastIndex(s, ".")

	if dotIndex != -1 {
		fractionalPart := s[dotIndex+1:]

		if len(fractionalPart) == 3 {
			cleanString := strings.ReplaceAll(s, ".", "")
			return strconv.ParseFloat(cleanString, 64)
		}
	}

	return strconv.ParseFloat(s, 64)
}
