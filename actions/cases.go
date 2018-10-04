package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// CasesIndex default implementation.
func CasesIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participants := &models.Participants{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	user := c.Value("current_user").(*models.User)
	q := tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status = ?", "11").PaginateFromParams(c.Params()).Order("created_at ASC")
	// Retrieve all Posts from the DB
	if err := q.All(participants); err != nil {
		return errors.WithStack(err)
	}
	// Make posts available inside the html template
	c.Set("participants", participants)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Cases"] = "/cases/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	logErr := InsertLog("view", "User viewed cases", "", "", "", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}
	return c.Render(200, r.HTML("cases/index.html"))
}
