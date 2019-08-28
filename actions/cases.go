package actions

import (
	"math"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// CasesIndex default implementation.
func CasesIndex(c buffalo.Context) error {
	var err error
	tx := c.Value("tx").(*pop.Connection)
	participants := &models.Participants{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	user := c.Value("current_user").(*models.User)
	site := ""
	if c.Value("current_site") != nil {
		site = c.Value("current_site").(string)
	}

	type scr struct {
		SID            string    `db:"sid"`
		ParticipantID  string    `db:"partid"`
		OverReadingID  *string   `db:"oid"`
		PID            string    `db:"pid"`
		CreatedAt      time.Time `db:"created"`
		AssessmentDate *string   `db:"assessment"`
	}
	screenings := make([]scr, 0)
	fresh := make([]string, 0)
	in7Days := make([]string, 0)
	before7Days := make([]string, 0)
	finished := make([]string, 0)
	allParticipants := tx.Where("participants.status LIKE ?", "11%")
	if len(site) > 0 {
		allParticipants = allParticipants.Where("SUBSTRING(participants.participant_id,2,1) = ?", site)
	}
	query := allParticipants.LeftJoin("participants", "participants.id=screenings.participant_id").LeftJoin("over_readings", "participants.id=over_readings.participant_id")
	columns := []string{
		"over_readings.id AS oid",
		"participants.id AS pid",
		"participants.participant_id AS partid",
		"screenings.created_at AS created",
		"screenings.id AS sid",
		"screenings.eye->'assessment_date'->>'calculated_date' AS assessment",
	}
	sql, args := query.ToSQL(&pop.Model{Value: models.Screening{}}, columns...)
	if err = allParticipants.RawQuery(sql, args...).All(&screenings); err != nil {
		InsertLog("error", "User viewed cases error", err.Error(), "", "", user.ID, c)
		return errors.WithStack(err)
	}

	for _, s := range screenings {
		if s.OverReadingID != nil {
			finished = append(finished, s.PID)
			continue
		}
		date := s.CreatedAt
		if s.AssessmentDate != nil {
			aDate, err := time.Parse(time.RFC3339, *s.AssessmentDate)
			if err == nil {
				date = aDate
			}
		}
		ago := daysAgo(date)
		if ago < 4 {
			fresh = append(fresh, s.PID)
		} else if ago < 8 {
			in7Days = append(in7Days, s.PID)
		} else {
			before7Days = append(before7Days, s.PID)
		}
	}

	where := make([]string, 0)
	wheres := make([]interface{}, 0)

	inIds := make([]string, 0)

	status := "11%"
	modifier := "LIKE"
	if len(c.Param("status")) > 0 {
		if c.Param("status") == "completed" {
			status = "111"
			modifier = "="
			inIds = append(inIds, finished...)
		} else if c.Param("status") == "open" {
			status = "11"
			modifier = "="
			inIds = append(inIds, fresh...)
			inIds = append(inIds, in7Days...)
			inIds = append(inIds, before7Days...)
		} else if c.Param("status") == "open_in_7_days" {
			status = "11"
			modifier = "="
			inIds = append(inIds, in7Days...)
		} else if c.Param("status") == "open_before_7_days" {
			status = "11"
			modifier = "="
			inIds = append(inIds, before7Days...)
		} else {
			inIds = append(inIds, finished...)
			inIds = append(inIds, fresh...)
			inIds = append(inIds, in7Days...)
			inIds = append(inIds, before7Days...)
		}
	} else {
		inIds = append(inIds, finished...)
		inIds = append(inIds, fresh...)
		inIds = append(inIds, in7Days...)
		inIds = append(inIds, before7Days...)
	}
	where = append(where, "status "+modifier+" ?")
	wheres = append(wheres, status)

	if len(c.Param("search")) > 0 {
		where = append(where, "replace(participants.participant_id, '-', '') LIKE ?")
		wheres = append(wheres, "%"+strings.Replace(strings.ToUpper(c.Param("search")), "-", "", -1)+"%")
	}

	if len(site) > 0 {
		where = append(where, "SUBSTRING(participant_id,2,1) = ?")
		wheres = append(wheres, site)
	}

	whereStmt := strings.Join(where, " AND ")
	var q *pop.Query
	q = tx.Eager("User", "Screenings", "Screenings.Screener", "OverReadings", "OverReadings.OverReader").Where(whereStmt, wheres...)
	if len(inIds) > 0 {
		q = q.Where("participants.id in (?)", inIds)
	} else {
		q = q.Where("participants.participant_id = ?", "invalid")
	}
	q = q.PaginateFromParams(c.Params()).Order("created_at DESC")
	// Retrieve all Posts from the DB
	if err = q.All(participants); err != nil {
		InsertLog("error", "User viewed cases error", err.Error(), "", "", user.ID, c)
		return errors.WithStack(err)
	}

	// Make posts available inside the html template
	c.Set("participants", participants)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	c.Set("finished", len(finished))
	c.Set("fresh", len(fresh))
	c.Set("in7Days", len(in7Days))
	c.Set("before7Days", len(before7Days))

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

func daysAgo(c time.Time) int {
	days := int(math.Floor(time.Now().Sub(c).Hours() / 24))

	return days
}
