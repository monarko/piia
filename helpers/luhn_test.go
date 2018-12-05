package helpers

import (
	"testing"
)

func TestValid(t *testing.T) {
	test := "1111111116"
	if !Valid(test, false) {
		t.Errorf("Expexted ”%s” to be valid according to the Luhn algorithm", test)
	}

	test = "1111111111"
	if Valid(test, false) {
		t.Errorf("Expexted ”%s” to be invalid according to the Luhn algorithm", test)
	}

	test = "G-1253-4"
	if !Valid(test, false) {
		t.Errorf("Expexted ”%s” to be valid according to the Luhn algorithm", test)
	}

	test = "J75580"
	if !Valid(test, false) {
		t.Errorf("Expexted ”%s” to be valid according to the Luhn algorithm", test)
	}

	test = "G-1253-3"
	if Valid(test, false) {
		t.Errorf("Expexted ”%s” to be invalid according to the Luhn algorithm", test)
	}
}

func TestCalculate(t *testing.T) {
	type Test struct {
		input  string
		output string
		alpha  bool
	}

	st := []Test{
		{"A-9578", "A-9578-9", true},
		{"P-9577", "P-9577-0", true},
		{"P-9579", "P-9579-6", true},
		{"A-9580", "A-9580-5", true},
		{"P-9581", "P-9581-2", true},
		{"A-9582", "A-9582-1", true},
		{"P-9583", "P-9583-8", true},
		{"A-9584", "A-9584-7", true},
		{"P-9585", "P-9585-3", true},
		{"A-9586", "A-9586-2", true},
		{"P-9587", "P-9587-9", true},
		{"A-9588", "A-9588-8", true},
		{"P-9589", "P-9589-5", true},
		{"A-9590", "A-9590-4", true},
		{"P-9591", "P-9591-1", true},
		{"A-9592", "A-9592-0", true},
		{"P-9593", "P-9593-7", true},
		{"A-9576", "A-9576-3", true},
		{"A-9594", "A-9594-6", true},
		{"PK-0001", "PK-0001-2", true},
		{"PK-0002", "PK-0002-0", true},
		{"PK-0003", "PK-0003-8", true},
		{"PK-0004", "PK-0004-6", true},
		{"PK-0005", "PK-0005-3", true},
		{"PK-0006", "PK-0006-1", true},
		{"PK-0007", "PK-0007-9", true},
		{"PK-0008", "PK-0008-7", true},
		{"PK-0009", "PK-0009-5", true},
		{"PK-0010", "PK-0010-3", true},

		{"PK-0001", "PK-0001-8", false},
		{"PK-0002", "PK-0002-6", false},
		{"PK-0003", "PK-0003-4", false},
		{"PK-0004", "PK-0004-2", false},
		{"PK-0005", "PK-0005-9", false},
		{"PK-0006", "PK-0006-7", false},
		{"PK-0007", "PK-0007-5", false},
		{"PK-0008", "PK-0008-3", false},
		{"PK-0009", "PK-0009-1", false},
		{"PK-0010", "PK-0010-9", false},
		{"PK-0011", "PK-0011-7", false},
		{"PK-0012", "PK-0012-5", false},
		{"PK-0013", "PK-0013-3", false},
		{"PK-0014", "PK-0014-1", false},
		{"PK-0015", "PK-0015-8", false},
		{"PK-0016", "PK-0016-6", false},
		{"PK-0017", "PK-0017-4", false},
		{"PK-0018", "PK-0018-2", false},
		{"PK-0019", "PK-0019-0", false},
		{"PK-0020", "PK-0020-8", false},
		{"PK-0021", "PK-0021-6", false},
		{"PK-0022", "PK-0022-4", false},
		{"PK-0023", "PK-0023-2", false},
		{"PK-0024", "PK-0024-0", false},
		{"PK-0025", "PK-0025-7", false},
		{"PK-0026", "PK-0026-5", false},
		{"PK-0027", "PK-0027-3", false},
		{"PK-0028", "PK-0028-1", false},
		{"PK-0029", "PK-0029-9", false},
		{"PK-0030", "PK-0030-7", false},
		{"PK-0031", "PK-0031-5", false},
		{"PK-0032", "PK-0032-3", false},
		{"PK-0033", "PK-0033-1", false},
		{"PK-0034", "PK-0034-9", false},
		{"PK-0035", "PK-0035-6", false},
		{"PK-0036", "PK-0036-4", false},
		{"PK-0037", "PK-0037-2", false},
		{"PK-0038", "PK-0038-0", false},
		{"PK-0039", "PK-0039-8", false},
		{"PK-0040", "PK-0040-6", false},
		{"PK-0041", "PK-0041-4", false},
		{"PK-0042", "PK-0042-2", false},
		{"PK-0043", "PK-0043-0", false},
		{"PK-0044", "PK-0044-8", false},
		{"PK-0045", "PK-0045-5", false},
		{"PK-0046", "PK-0046-3", false},
		{"PK-0047", "PK-0047-1", false},
		{"PK-0048", "PK-0048-9", false},
		{"PK-0049", "PK-0049-7", false},
		{"PK-0050", "PK-0050-5", false},
		{"PK-0051", "PK-0051-3", false},
		{"PK-0052", "PK-0052-1", false},
		{"PK-0053", "PK-0053-9", false},
		{"PK-0054", "PK-0054-7", false},
		{"PK-0055", "PK-0055-4", false},
		{"PK-0056", "PK-0056-2", false},
		{"PK-0057", "PK-0057-0", false},
		{"PK-0058", "PK-0058-8", false},
		{"PK-0059", "PK-0059-6", false},
		{"PK-0060", "PK-0060-4", false},
		{"PK-0061", "PK-0061-2", false},
		{"PK-0062", "PK-0062-0", false},
		{"PK-0063", "PK-0063-8", false},
		{"PK-0064", "PK-0064-6", false},
		{"PK-0065", "PK-0065-3", false},
		{"PK-0066", "PK-0066-1", false},
		{"PK-0067", "PK-0067-9", false},
		{"PK-0068", "PK-0068-7", false},
		{"PK-0069", "PK-0069-5", false},
		{"PK-0070", "PK-0070-3", false},
		{"PK-0071", "PK-0071-1", false},
		{"PK-0072", "PK-0072-9", false},
		{"PK-0073", "PK-0073-7", false},
		{"PK-0074", "PK-0074-5", false},
		{"PK-0075", "PK-0075-2", false},
		{"PK-0076", "PK-0076-0", false},
		{"PK-0077", "PK-0077-8", false},
		{"PK-0078", "PK-0078-6", false},
		{"PK-0079", "PK-0079-4", false},
		{"PK-0080", "PK-0080-2", false},
		{"PK-0081", "PK-0081-0", false},
		{"PK-0082", "PK-0082-8", false},
		{"PK-0083", "PK-0083-6", false},
		{"PK-0084", "PK-0084-4", false},
		{"PK-0085", "PK-0085-1", false},
		{"PK-0086", "PK-0086-9", false},
		{"PK-0087", "PK-0087-7", false},
		{"PK-0088", "PK-0088-5", false},
		{"PK-0089", "PK-0089-3", false},
		{"PK-0090", "PK-0090-1", false},
		{"PK-0091", "PK-0091-9", false},
		{"PK-0092", "PK-0092-7", false},
		{"PK-0093", "PK-0093-5", false},
		{"PK-0094", "PK-0094-3", false},
		{"PK-0095", "PK-0095-0", false},
		{"PK-0096", "PK-0096-8", false},
		{"PK-0097", "PK-0097-6", false},
		{"PK-0098", "PK-0098-4", false},
		{"PK-0099", "PK-0099-2", false},
	}

	for _, s := range st {
		cd := generateControlDigit(s.input, s.alpha)
		actualOutput := s.input + "-" + integersToString([]int{cd})

		if s.output != actualOutput {
			t.Errorf("For %s, expected output %s, got %s", s.input, s.output, actualOutput)
		}
	}
}
