package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/monarko/piia/models"
)

// ChangeSite changes the site
func ChangeSite(c buffalo.Context) error {
	selectedSite := c.Param("site")
	site := ""

	if selectedSite == "N" {
		site = "N"
	} else if selectedSite == "K" {
		site = "K"
	} else if selectedSite == "L" {
		site = "L"
	} else if selectedSite == "T" {
		site = "T"
	} else if selectedSite == "R" {
		site = "R"
	} else if selectedSite == "O" {
		site = "O"
	} else if selectedSite == "S" {
		site = "S"
	} else {
		site = ""
	}

	c.Session().Set("site", site)

	referrer := c.Request().Referer()

	return c.Redirect(302, referrer)
}

// SetCurrentSite attempts to find a user based on the current_user_id
// in the session. If one is found it is set on the context.
func SetCurrentSite(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if site := c.Session().Get("site"); site != nil {
			c.Set("current_site", site)
		} else {
			user, ok := c.Value("current_user").(*models.User)
			if ok {
				if user.Admin || user.Permission.StudyCoordinator {
					c.Set("current_site", "")
				} else {
					c.Set("current_site", user.Site)
				}
			} else {
				c.Set("current_site", "")
			}
		}

		return next(c)
	}
}
