package money_test

import (
	"testing"

	"entaintest/internal/money"

	"github.com/stretchr/testify/require"
)

func Test_Parse_ReturnsCents_InCaseAmountIsValid(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int64
	}{
		{name: "with_two_decimals", in: "10.15", want: 1015},
		{name: "with_one_decimal", in: "10.1", want: 1010},
		{name: "without_decimals", in: "10", want: 1000},
		{name: "less_than_one", in: "0.99", want: 99},
		{name: "zero", in: "0", want: 0},
		{name: "zero_with_decimals", in: "0.00", want: 0},
		{name: "with_surrounding_spaces", in: " 10.15 ", want: 1015},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got, err := money.Parse(tt.in)

			// Assert
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_Parse_ReturnsError_InCaseAmountIsInvalid(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "empty", in: ""},
		{name: "not_a_number", in: "abc"},
		{name: "too_many_decimals", in: "10.123"},
		{name: "trailing_dot", in: "10."},
		{name: "negative", in: "-1.00"},
		{name: "explicit_sign", in: "+1.00"},
		{name: "multiple_dots", in: "1.2.3"},
		{name: "missing_integer_part", in: ".5"},
		{name: "wrong_separator", in: "10,15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, err := money.Parse(tt.in)

			// Assert
			require.ErrorIs(t, err, money.ErrInvalidAmount)
		})
	}
}

func Test_Format_ReturnsDecimalString_InCaseGivenCents(t *testing.T) {
	tests := []struct {
		name string
		in   int64
		want string
	}{
		{name: "two_decimals", in: 1015, want: "10.15"},
		{name: "rounded_value", in: 925, want: "9.25"},
		{name: "whole_number", in: 1000, want: "10.00"},
		{name: "single_cent", in: 5, want: "0.05"},
		{name: "zero", in: 0, want: "0.00"},
		{name: "less_than_one", in: 99, want: "0.99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := money.Format(tt.in)

			// Assert
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_Format_ReturnsOriginalString_InCaseAppliedAfterParse(t *testing.T) {
	inputs := []string{"0.00", "9.25", "10.15", "1234.56"}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			// Arrange
			cents, err := money.Parse(input)
			require.NoError(t, err)

			// Act
			got := money.Format(cents)

			// Assert
			require.Equal(t, input, got)
		})
	}
}
