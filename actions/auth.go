package actions

import "github.com/gobuffalo/buffalo"

// AuthHandler is a default handler to serve up
// a home page.
func AuthHandler(c buffalo.Context) error {
	return c.Render(200, r.HTML("login.html"))
}
