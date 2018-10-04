package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
)

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c buffalo.Context) error {
	_, ok := c.Value("current_user").(*models.User)
	if !ok {
		return c.Redirect(302, "/users/login")
		// return c.Render(200, r.HTML("index-non-logged-in.html", "application-non-logged-in.html"))
	}
	tx := c.Value("tx").(*pop.Connection)
	var err error
	p := 0
	if p, err = tx.Count(&models.Participant{}); err != nil {
		p = 0
	}
	c.Set("participants", p)

	s := 0
	if s, err = tx.Count(&models.Screening{}); err != nil {
		s = 0
	}
	c.Set("screenings", s)

	o := 0
	if o, err = tx.Count(&models.OverReading{}); err != nil {
		o = 0
	}
	c.Set("overreadings", o)

	u := 0
	if u, err = tx.Count(&models.User{}); err != nil {
		u = 0
	}
	c.Set("users", u)

	return c.Render(200, r.HTML("index.html"))
}
