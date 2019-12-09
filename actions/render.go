package actions

import (
	"encoding/json"
	"html/template"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr/v2"
	"github.com/monarko/piia/helpers"
	"github.com/monarko/piia/models"
)

var r *render.Engine
var assetsBox = packr.New("../public", "../public")

// var assetsBox = packr.NewBox("../public")

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
				return helpers.Age(d)
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
				if i := helpers.IndexInSlice(sa, s); i >= 0 {
					return true
				}
				return false
			},
			"lastElement": func(s string, sep string) string {
				ss := strings.Split(s, sep)
				last := ss[len(ss)-1]

				return last
			},
			"currentDate": func(calendar string) string {
				return helpers.CurrentDateInFormat(calendar, "2006-01-02")
			},
			"currentDateInFormat": func(calendar, format string) string {
				return helpers.CurrentDateInFormat(calendar, format)
			},
			"languageDate": func(gregorianDate time.Time, format, calendar string) string {
				return helpers.LanguageDate(gregorianDate, format, calendar)
			},
			"changeDate": func(givenDate time.Time, format, fromCalendar, toCalendar string) string {
				return helpers.ChangeDate(givenDate, format, fromCalendar, toCalendar)
			},
			"arrayStringView": func(s string) string {
				if len(s) > 0 {
					return helpers.SliceStringToCommaSeparatedValue(s)
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
				return helpers.MatchTimes(a, b)
			},
			"getMatchingAudit": func(l models.SystemLog, as []models.Audit) string {
				for _, a := range as {
					rType := strings.ToLower(strings.Replace(l.ResourceType, "_", "", -1))
					mType := strings.ToLower(strings.Replace(a.ModelType, "_", "", -1))
					if helpers.MatchTimes(l.CreatedAt, a.CreatedAt) && (rType == mType) {
						jsonString, err := json.Marshal(a.Changes)
						if err == nil {
							return string(jsonString)
						}
					}
				}
				return ""
			},
		},
	})
}
