package actions

import (
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr/v2"
)

var r *render.Engine
var assetsBox = packr.NewBox("../public")

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.html",

		// Box containing all of the templates:
		TemplatesBox: packr.New("../templates", "../templates"),
		AssetsBox:    assetsBox,

		// Add template helpers here:
		Helpers: render.Helpers{
			// uncomment for non-Bootstrap form helpers:
			// "form":     plush.FormHelper,
			// "form_for": plush.FormForHelper,
			"csrf": func() template.HTML {
				return template.HTML("<input name=\"authenticity_token\" value=\"<%= authenticity_token %>\" type=\"hidden\">")
			},
			"ageHelper": func(d time.Time) string {
				return Age(d)
			},
			"genderHelper": func(s string) string {
				if s == "M" {
					return "Male"
				} else if s == "F" {
					return "Female"
				} else {
					return "Other"
				}
			},
			"stringInSlice": func(s string, sa []string) bool {
				for _, str := range sa {
					if str == s {
						return true
					}
				}
				return false
			},
			"lastElement": func(s string, sep string) string {
				ss := strings.Split(s, sep)
				last := ss[len(ss)-1]

				return last
			},
			"currentDate": func(calendar string) string {
				currentTime := time.Now()
				if calendar == "thai" {
					currentTime = currentTime.AddDate(543, 0, 0)
				}
				return currentTime.Format("2006-01-02")
			},
			"currentDateInFormat": func(calendar, format string) string {
				currentTime := time.Now()
				if calendar == "thai" {
					currentTime = currentTime.AddDate(543, 0, 0)
				}
				return currentTime.Format(format)
			},
			"languageDate": func(gregorianDate time.Time, format, calendar string) string {
				theDate := gregorianDate
				if calendar == "thai" {
					theDate = theDate.AddDate(543, 0, 0)
				}
				return theDate.Format(format)
			},
			"changeDate": func(givenDate time.Time, format, fromCalendar, toCalendar string) string {
				gregorionDate := givenDate
				if fromCalendar == "thai" {
					gregorionDate = gregorionDate.AddDate(-543, 0, 0)
				}

				theDate := gregorionDate
				if toCalendar == "thai" {
					theDate = gregorionDate.AddDate(543, 0, 0)
				}

				return theDate.Format(format)
			},
			"arrayStringView": func(s string) string {
				if len(s) > 0 {
					return SliceStringToCommaSeparatedValue(s)
				}
				return ""
			},
			"trimText": func(s string) string {
				return strings.TrimSpace(s)
			},
			"appendIfNotFound": func(s, toAdd string) string {
				if strings.Contains(s, toAdd) {
					return s
				}
				return strings.Join([]string{s, toAdd}, " ")
			},
			"getSite": func(participantID string) string {
				return participantID[1:2]
			},
			"tolower": func(s string) string {
				return strings.ToLower(s)
			},
			"matchTimes": func(a time.Time, b time.Time) bool {
				end := b.Add(time.Duration(1) * time.Second)
				start := b.Add(time.Duration(-1) * time.Second)
				return a.After(start) && a.Before(end)
			},
		},
	})
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
	//monthText := ""
	remainingText := ""

	if year > 0 {
		if year > 1 {
			yearText = strconv.Itoa(year) + " years"
		} else {
			yearText = strconv.Itoa(year) + " year"
		}
	}
	/*
		if month > 0 {
			if month > 1 {
				monthText = strconv.Itoa(month) + " months"
			} else {
				monthText = strconv.Itoa(month) + " month"
			}
		}
	*/

	if month == 0 && year == 0 {
		if day > 1 {
			remainingText = strconv.Itoa(day) + " days"
		} else {
			remainingText = strconv.Itoa(day) + " day"
		}
	}

	ageText := strings.Join([]string{yearText, remainingText}, " ")

	return ageText
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
