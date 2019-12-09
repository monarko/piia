package actions

import (
	"github.com/gobuffalo/buffalo"
)

// ChangeLanguage changes the site language
func ChangeLanguage(c buffalo.Context) error {
	selectedLanguage := c.Param("lang")
	lang := "en"

	if selectedLanguage == "bn" {
		lang = "bn"
	} else if selectedLanguage == "th" {
		lang = "th"
	}

	c.Session().Set("lang", lang)

	referrer := c.Request().Referer()

	return c.Redirect(302, referrer)
}
