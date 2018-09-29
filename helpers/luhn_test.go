package helpers

import "testing"

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
