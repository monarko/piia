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

// SetCurrentLang attempts to find a user based on the current_user_id
// in the session. If one is found it is set on the context.
func SetCurrentLang(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if lang := c.Session().Get("lang"); lang != nil {
			c.Set("current_lang", lang)
		} else {
			c.Set("current_lang", "en")
		}
		return next(c)
	}
}
