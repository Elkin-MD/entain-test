// Package money handles monetary values as integer minor units (cents) to
// avoid floating-point precision errors.
package money

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

// ErrInvalidAmount is returned when a string is not a valid non-negative
// amount with at most 2 decimal places.
var ErrInvalidAmount = errors.New("invalid amount")

// Parse converts a decimal string such as "10.15" into cents (1015).
func Parse(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrInvalidAmount
	}

	if strings.ContainsAny(s, "+-") {
		return 0, ErrInvalidAmount
	}

	intPart, fracPart, hasDot := strings.Cut(s, ".")
	if hasDot && (len(fracPart) == 0 || len(fracPart) > 2) {
		return 0, ErrInvalidAmount
	}

	if intPart == "" {
		return 0, ErrInvalidAmount
	}

	whole, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return 0, ErrInvalidAmount
	}

	frac := int64(0)

	if hasDot {
		padded := fracPart
		for len(padded) < 2 {
			padded += "0"
		}

		frac, err = strconv.ParseInt(padded, 10, 64)
		if err != nil {
			return 0, ErrInvalidAmount
		}
	}

	if whole > math.MaxInt64/100 {
		return 0, ErrInvalidAmount
	}

	return whole*100 + frac, nil
}

// Format renders cents as a decimal string with exactly 2 places (925 -> "9.25").
func Format(cents int64) string {
	sign := ""

	if cents < 0 {
		sign = "-"
		cents = -cents
	}

	return sign + strconv.FormatInt(cents/100, 10) + "." + pad2(cents%100)
}

func pad2(n int64) string {
	if n < 10 {
		return "0" + strconv.FormatInt(n, 10)
	}

	return strconv.FormatInt(n, 10)
}
