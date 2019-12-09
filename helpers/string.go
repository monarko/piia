package helpers

import (
	"strconv"
	"strings"
	"time"
)

// IndexInSlice returns the index of string `b` on slice of string `a` and returns -1 if not found.
func IndexInSlice(a []string, b string) int {
	for k, v := range a {
		if b == v {
			return k
		}
	}
	return -1
}

// LanguageDate returns localized time
func LanguageDate(gregorianDate time.Time, format, calendar string) string {
	theDate := gregorianDate
	if calendar == "thai" {
		theDate = theDate.AddDate(543, 0, 0)
	}
	return theDate.Format(format)
}

// Age calculates the participant's age
func Age(a time.Time) string {
	if a.IsZero() {
		return "Not given"
	}

	b := time.Now()
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}

	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year := int(y2 - y1)
	month := int(M2 - M1)
	day := int(d2 - d1)
	hour := int(h2 - h1)
	min := int(m2 - m1)
	sec := int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	yearText := ""
	remainingText := ""

	if year > 0 {
		if year > 1 {
			yearText = strconv.Itoa(year) + " years"
		} else {
			yearText = strconv.Itoa(year) + " year"
		}
	}

	if month == 0 && year == 0 {
		if day > 1 {
			remainingText = strconv.Itoa(day) + " days"
		} else {
			remainingText = strconv.Itoa(day) + " day"
		}
	}

	ageText := strings.Join([]string{yearText, remainingText}, " ")

	return strings.TrimSpace(ageText)
}

// SliceStringToCommaSeparatedValue function
func SliceStringToCommaSeparatedValue(s string) string {
	temp := strings.Replace(s, "[", "", -1)
	temp = strings.Replace(temp, "]", "", -1)
	temp = strings.Replace(temp, "\"", "", -1)
	sl := strings.Split(temp, ",")
	for i, slt := range sl {
		sl[i] = strings.TrimSpace(slt)
	}

	return strings.Join(sl, ", ")
}

// MatchTimes match any time within a timerange
func MatchTimes(a time.Time, b time.Time) bool {
	end := b.Add(time.Duration(1) * time.Second)
	start := b.Add(time.Duration(-1) * time.Second)
	return a.After(start) && a.Before(end)
}

// ChangeDate converts a date from a given calendar to another calendar
func ChangeDate(givenDate time.Time, format, fromCalendar, toCalendar string) string {
	gregorionDate := givenDate
	if fromCalendar == "thai" {
		gregorionDate = gregorionDate.AddDate(-543, 0, 0)
	}
	theDate := gregorionDate
	if toCalendar == "thai" {
		theDate = gregorionDate.AddDate(543, 0, 0)
	}
	return theDate.Format(format)
}

// CurrentDateInFormat returns current date in given calendar format
func CurrentDateInFormat(calendar, format string) string {
	currentTime := time.Now()
	if calendar == "thai" {
		currentTime = currentTime.AddDate(543, 0, 0)
	}
	return currentTime.Format(format)
}

// IntersectTwoStringSlices returns an intersection of two slices
func IntersectTwoStringSlices(a []string, b []string) (inter []string) {
	// interacting on the smallest list first can potentailly be faster...but not by much, worse case is the same
	low, high := a, b
	if len(a) > len(b) {
		low = b
		high = a
	}

	done := false
	for i, l := range low {
		for j, h := range high {
			// get future index values
			f1 := i + 1
			f2 := j + 1
			if l == h {
				inter = append(inter, h)
				if f1 < len(low) && f2 < len(high) {
					// if the future values aren't the same then that's the end of the intersection
					if low[f1] != high[f2] {
						done = true
					}
				}
				// we don't want to interate on the entire list everytime, so remove the parts we already looped on will make it faster each pass
				high = high[:j+copy(high[j:], high[j+1:])]
				break
			}
		}
		// nothing in the future so we are done
		if done {
			break
		}
	}
	return
}

// SliceStringUnique returns a slice of unique strings by discarding duplicates from the original.
func SliceStringUnique(original []string, caseSensitive bool) []string {
	if original == nil {
		return nil
	}

	unique := make([]string, 0)
	keys := make(map[string]struct{})
	for _, val := range original {
		keyToCheck := val
		if !caseSensitive {
			keyToCheck = strings.ToLower(val)
		}

		if _, ok := keys[keyToCheck]; !ok {
			keys[keyToCheck] = struct{}{}
			unique = append(unique, val)
		}
	}

	return unique
}

// SliceContainsString returns idx and true if a found in s. Otherwise -1 and false.
func SliceContainsString(s []string, a string) (int, bool) {
	for i, b := range s {
		if b == a {
			return i, true
		}
	}

	return -1, false
}
