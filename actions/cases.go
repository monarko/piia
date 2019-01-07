package actions

import (
	"strings"

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
	where := make([]string, 0)
	wheres := make([]interface{}, 0)

	status := "11%"
	modifier := "LIKE"
	if len(c.Param("status")) > 0 {
		if c.Param("status") == "completed" {
			status = "111"
			modifier = "="
		} else if c.Param("status") == "open" {
			status = "11"
			modifier = "="
		}
	}
	where = append(where, "status "+modifier+" ?")
	wheres = append(wheres, status)

	if len(c.Param("search")) > 0 {
		where = append(where, "participant_id LIKE ?")
		wheres = append(wheres, "%"+strings.ToUpper(c.Param("search"))+"%")
	}

	whereStmt := strings.Join(where, " AND ")

	user := c.Value("current_user").(*models.User)
	q := tx.Eager("User", "Screenings", "Screenings.Screener", "OverReadings", "OverReadings.OverReader").Where(whereStmt, wheres...).PaginateFromParams(c.Params()).Order("created_at DESC")

	// Retrieve all Posts from the DB
	if err := q.All(participants); err != nil {
		InsertLog("error", "User viewed cases error", err.Error(), "", "", user.ID, c)
		return errors.WithStack(err)
	}
	// Make posts available inside the html template
	c.Set("participants", participants)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Cases"] = "/cases/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	c.Set("filterStatus", c.Params().Get("status"))
	c.Set("filterSearch", c.Params().Get("search"))
	logErr := InsertLog("view", "User viewed cases", "", "", "", user.ID, c)
	if logErr != nil {
		InsertLog("error", "User viewed cases error", logErr.Error(), "", "", user.ID, c)
		return errors.WithStack(logErr)
	}
	return c.Render(200, r.HTML("cases/index.html"))
}
