package algorithms

import (
	"testing"
)

func TestToBoolArray(t *testing.T) {
	var tests = []struct {
		input byte
		want  [8]bool
	}{
		{0x00, [8]bool{false, false, false, false, false, false, false, false}},
		{0x01, [8]bool{false, false, false, false, false, false, false, true}},
		{0x10, [8]bool{false, false, false, true, false, false, false, false}},
		{0xAB, [8]bool{true, false, true, false, true, false, true, true}},
		{0x44, [8]bool{false, true, false, false, false, true, false, false}},
		{0x0F, [8]bool{false, false, false, false, true, true, true, true}},
		{0xF0, [8]bool{true, true, true, true, false, false, false, false}},
		{0xFF, [8]bool{true, true, true, true, true, true, true, true}},
	}

	for _, test := range tests {
		if got := ToBoolArray(test.input); got != test.want {
			t.Errorf("ToBoolArray(%v) = %v; expected %v", test.input, got, test.want)
		}
	}
}
