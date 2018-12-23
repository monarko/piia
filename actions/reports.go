package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
)

// ReportsIndex default implementation.
func ReportsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participants := &models.Participants{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	var q *pop.Query

	user := c.Value("current_user").(*models.User)
	if user.Admin || user.Permission.StudyCoordinator {
		if len(c.Param("status")) > 0 {
			q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status = ?", c.Param("status")).PaginateFromParams(c.Params()).Order("created_at DESC")
		} else {
			q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").PaginateFromParams(c.Params()).Order("created_at DESC")
		}
	} else if user.Permission.Screening && user.Permission.OverRead {
		q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status != ?", "111").Where("participants.participant_id LIKE '_" + user.Site + "%'").PaginateFromParams(c.Params()).Order("created_at DESC")
	} else if user.Permission.Screening {
		q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status LIKE ?", "1%").Where("participants.participant_id LIKE '_" + user.Site + "%'").PaginateFromParams(c.Params()).Order("created_at DESC")
	} else if user.Permission.OverRead {
		q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status LIKE ?", "11%").PaginateFromParams(c.Params()).Order("created_at DESC")
	} else {
		// If there are no errors set a success message
		c.Flash().Add("danger", "You don't have sufficient permission.")
		// and redirect to the index page
		return c.Redirect(302, "/")
	}

	// Retrieve all Posts from the DB
	if err := q.All(participants); err != nil {
		// return errors.WithStack(err)
		errStr := err.Error()
		errs := map[string][]string{
			"index_error": {errStr},
		}
		c.Set("errors", errs)
		return c.Redirect(302, "/")
	}
	// Make posts available inside the html template
	c.Set("participants", participants)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["page_participants_title"] = "/reports/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	c.Set("filterStatus", c.Params().Get("status"))
	logErr := InsertLog("view", "User viewed reports", "", "", "", user.ID, c)
	if logErr != nil {
		// return errors.WithStack(logErr)
		errStr := logErr.Error()
		errs := map[string][]string{
			"index_error": {errStr},
		}
		c.Set("errors", errs)
		return c.Render(422, r.HTML("reports/index.html"))
	}
	return c.Render(200, r.HTML("reports/index.html"))
}
