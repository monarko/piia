package helpers

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// Luhn object
type Luhn struct {
	ID string
}

// GenerateLuhnID generates a new luhn id
func GenerateLuhnID() Luhn {
	rand.Seed(time.Now().UnixNano())
	// sitePrefix := chars[rand.Intn(len(chars))]
	sitePrefix := chars[1]
	luhnID := GenerateWithPrefix(6, string(sitePrefix))
	return Luhn{ID: luhnID}
}

// GenerateLuhnIDWithGivenPrefix generates a new luhn id
func GenerateLuhnIDWithGivenPrefix(prefix string) Luhn {
	sitePrefix := chars[1]
	if len(prefix) > 0 {
		sitePrefix = prefix[0]
	}
	luhnID := GenerateWithPrefix(6, string(sitePrefix))
	return Luhn{ID: luhnID}
}

// Valid returns a boolean indicating if the argument was valid according to the Luhn algorithm.
func Valid(luhnString string) bool {
	checksumMod := calculateChecksum(luhnString, false) % 10

	return checksumMod == 0
}

// Generate creates and returns a string of the length of the argument targetSize.
// The returned string is valid according to the Luhn algorithm.
func Generate(size int) string {
	random := randomString(size - 1)
	controlDigit := strconv.Itoa(generateControlDigit(random))

	return random + controlDigit
}

// GenerateWithPrefix creates and returns a string of the length of the argument targetSize
// but prefixed with the second argument.
// The returned string is valid according to the Luhn algorithm.
func GenerateWithPrefix(size int, prefix string) string {
	size = size - 1 - len(prefix)

	random := prefix + "-" + randomString(size)
	controlDigit := strconv.Itoa(generateControlDigit(random))

	return random + "-" + controlDigit
}

func randomString(size int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	source := make([]int, size)

	for i := 0; i < size; i++ {
		source[i] = rand.Intn(9)
	}

	return integersToString(source)
}

func generateControlDigit(luhnString string) int {
	controlDigit := calculateChecksum(luhnString, true) % 10

	if controlDigit != 0 {
		controlDigit = 10 - controlDigit
	}

	return controlDigit
}

func calculateChecksum(luhnString string, double bool) int {
	prefixInt := fmt.Sprintf("%d", luhnString[0])
	suffix := luhnString[1:]
	theString := string(prefixInt) + suffix
	source := strings.Split(theString, "")
	checksum := 0

	for i := len(source) - 1; i > -1; i-- {
		if source[i] == "-" {
			continue
		}

		t, _ := strconv.ParseInt(source[i], 10, 8)
		n := int(t)

		if double {
			n = n * 2
		}
		double = !double

		if n >= 10 {
			n = n - 9
		}

		checksum += n
	}

	return checksum
}

func integersToString(integers []int) string {
	result := make([]string, len(integers))

	for i, number := range integers {
		result[i] = strconv.Itoa(number)
	}

	return strings.Join(result, "")
}
