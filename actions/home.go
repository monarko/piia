package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/monarko/piia/models"
)

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c buffalo.Context) error {
	_, ok := c.Value("current_user").(*models.User)
	if !ok {
		return c.Render(200, r.HTML("index-non-logged-in.html", "application-non-logged-in.html"))
	}
	return c.Render(200, r.HTML("index.html"))
}
