package actions

import (
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// ReferralsIndex default implementation.
func ReferralsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participants := &models.Participants{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".

	user := c.Value("current_user").(*models.User)
	screenings := &models.Screenings{}
	sq := tx.Where("referral->>'referred' = ?", "true")
	if err := sq.All(screenings); err != nil {
		return errors.WithStack(err)
	}

	overs := &models.OverReadings{}
	oq := tx.Where("referral->>'referred' = ?", "true")
	if err := oq.All(overs); err != nil {
		return errors.WithStack(err)
	}

	ids := make([]string, 0)
	for _, s := range *screenings {
		ids = append(ids, s.ParticipantID.String())
	}

	for _, o := range *overs {
		ids = append(ids, o.ParticipantID.String())
	}

	var q *pop.Query
	if len(c.Param("search")) > 0 {
		q = tx.Eager("User", "Screenings", "Screenings.Screener", "OverReadings", "OverReadings.OverReader").Where("id in (?)", ids).Where("referral_appointment != ?", true).Where("participant_id = ?", strings.ToUpper(c.Param("search"))).PaginateFromParams(c.Params()).Order("created_at ASC")
		c.Set("search", c.Param("search"))
	} else {
		q = tx.Eager("User", "Screenings", "Screenings.Screener", "OverReadings", "OverReadings.OverReader").Where("id in (?)", ids).Where("referral_appointment != ?", true).PaginateFromParams(c.Params()).Order("created_at ASC")
		c.Set("search", "")
	}

	// Retrieve all Posts from the DB
	if err := q.All(participants); err != nil {
		return errors.WithStack(err)
	}
	// Make posts available inside the html template
	c.Set("participants", participants)

	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Referrals"] = "/referrals/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	logErr := InsertLog("view", "User viewed referrals", "", "", "", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}
	return c.Render(200, r.HTML("referrals/index.html"))
}
