package helpers

import (
	"testing"
)

func TestValid(t *testing.T) {
	test := "1111111116"
	if !Valid(test) {
		t.Errorf("Expexted ”%s” to be valid according to the Luhn algorithm", test)
	}

	test = "1111111111"
	if Valid(test) {
		t.Errorf("Expexted ”%s” to be invalid according to the Luhn algorithm", test)
	}

	test = "G-1253-4"
	if !Valid(test) {
		t.Errorf("Expexted ”%s” to be valid according to the Luhn algorithm", test)
	}

	test = "J75580"
	if !Valid(test) {
		t.Errorf("Expexted ”%s” to be valid according to the Luhn algorithm", test)
	}

	test = "G-1253-3"
	if Valid(test) {
		t.Errorf("Expexted ”%s” to be invalid according to the Luhn algorithm", test)
	}
}

func TestCalculate(t *testing.T) {
	type Test struct {
		input  string
		output string
	}

	st := []Test{
		{"A-9576", "A-9576-3"},
		{"P-9577", "P-9577-0"},
		{"A-9578", "A-9578-9"},
		{"P-9579", "P-9579-6"},
		{"A-9580", "A-9580-5"},
		{"P-9581", "P-9581-2"},
		{"A-9582", "A-9582-1"},
		{"P-9583", "P-9583-8"},
		{"A-9584", "A-9584-7"},
		{"P-9585", "P-9585-3"},
		{"A-9586", "A-9586-2"},
		{"P-9587", "P-9587-9"},
		{"A-9588", "A-9588-8"},
		{"P-9589", "P-9589-5"},
		{"A-9590", "A-9590-4"},
		{"P-9591", "P-9591-1"},
		{"A-9592", "A-9592-0"},
		{"P-9593", "P-9593-7"},
		{"A-9594", "A-9594-6"},
		{"PK-0001", "PK-0001-2"},
		{"PK-0002", "PK-0002-0"},
		{"PK-0003", "PK-0003-8"},
		{"PK-0004", "PK-0004-6"},
		{"PK-0005", "PK-0005-3"},
		{"PK-0006", "PK-0006-1"},
		{"PK-0007", "PK-0007-9"},
		{"PK-0008", "PK-0008-7"},
		{"PK-0009", "PK-0009-5"},
		{"PK-0010", "PK-0010-3"},
	}

	for _, s := range st {
		cd := generateControlDigit(s.input)
		actualOutput := s.input + "-" + integersToString([]int{cd})

		if s.output != actualOutput {
			t.Errorf("For %s, expected output %s, got %s", s.input, s.output, actualOutput)
		}
	}
}
