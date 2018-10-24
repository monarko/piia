package actions

import (
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c buffalo.Context) error {
	loggedInUser, ok := c.Value("current_user").(*models.User)
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

	notifications := &models.Notifications{}
	var q *pop.Query

	if len(strings.TrimSpace(loggedInUser.Site)) > 0 {
		q = tx.Eager().Where("site = ?", loggedInUser.Site).Where("status != ?", "closed").PaginateFromParams(c.Params()).Order("created_at DESC")
	} else if loggedInUser.Admin || loggedInUser.PermissionStudyCoordinator {
		q = tx.Eager().Where("status != ?", "closed").PaginateFromParams(c.Params()).Order("created_at DESC")
	} else {
		q = tx.Eager().Where("status != ?", "unknown").PaginateFromParams(c.Params()).Order("created_at DESC")
	}

	// Retrieve all Notifications from the DB
	if err := q.All(notifications); err != nil {
		return errors.WithStack(err)
	}

	// Make posts available inside the html template
	c.Set("notifications", notifications)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)

	return c.Render(200, r.HTML("index.html"))
}
