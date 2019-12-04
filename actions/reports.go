package actions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// ReportsIndex default implementation.
func ReportsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	var err error
	user := c.Value("current_user").(*models.User)

	/** STATISTICS **/
	s := 0
	if s, err = tx.Count(&models.Screening{}); err != nil {
		s = 0
	}

	o := 0
	if o, err = tx.Count(&models.OverReading{}); err != nil {
		o = 0
	}

	sReferred := 0
	if sReferred, err = tx.Where("referral->>'referred' = ?", "true").Count(&models.Screening{}); err != nil {
		sReferred = 0
	}

	oReferred := 0
	if oReferred, err = tx.Where("referral->>'referred' = ?", "true").Count(&models.OverReading{}); err != nil {
		oReferred = 0
	}

	stat := make(map[string]interface{})
	stat["total_screenings"] = s
	stat["screening_referred"] = sReferred
	stat["screening_referred_percent"] = 0.0
	if s > 0 {
		stat["screening_referred_percent"] = fmt.Sprintf("%.2f", ((float64(sReferred) / float64(s)) * 100))
	}

	stat["total_overreadings"] = o
	stat["overreading_referred"] = oReferred
	stat["overreading_referred_percent"] = 0.0
	if o > 0 {
		stat["overreading_referred_percent"] = fmt.Sprintf("%.2f", ((float64(oReferred) / float64(o)) * 100))
	}

	dQuery := `SELECT ov.id AS "OID", sc.id AS "SID"
	FROM over_readings ov
	LEFT JOIN screenings sc ON (ov.screening_id = sc.id)
	WHERE ov.referral->'referred' != sc.referral->'referred'`

	type disagreement struct {
		OID string `db:"OID"`
		SID string `db:"SID"`
	}

	dis := 0
	if dis, err = tx.RawQuery(dQuery).Count(&disagreement{}); err != nil {
		dis = 0
	}

	stat["disagreement"] = dis
	if o > 0 {
		stat["disagreement_percent"] = fmt.Sprintf("%.2f", ((float64(dis) / float64(o)) * 100))
	}

	c.Set("stat", stat)

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["page_analytics_title"] = "/analytics/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	// c.Set("filterStatus", c.Params().Get("status"))
	logErr := InsertLog("view", "User viewed analytics", "", "", "", user.ID, c)
	if logErr != nil {
		// return errors.WithStack(logErr)
		errStr := logErr.Error()
		errs := map[string][]string{
			"index_error": {errStr},
		}
		c.Set("errors", errs)
		InsertLog("error", "User viewed analytics error", logErr.Error(), "", "", user.ID, c)
		return c.Render(422, r.HTML("reports/index.html"))
	}
	return c.Render(200, r.HTML("reports/index.html"))
}

// ReportsIndexAPI default implementation.
func ReportsIndexAPI(c buffalo.Context) error {
	var err error
	pr := DatatableRequest{}
	err = json.NewDecoder(c.Request().Body).Decode(&pr)
	if err != nil {
		return c.Render(400, r.JSON(err))
	}
	tx := c.Value("tx").(*pop.Connection)
	responseResult := DatatableResponse{}
	responseResult.Draw = pr.Parameters.Draw
	participants := &models.Participants{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	var q *pop.Query
	page := (pr.Parameters.Start / pr.Parameters.Length) + 1
	orders := []string{
		"participant_id",
		"participant_id",
		"participant_id",
		"created_at",
		"participant_id",
		"participant_id",
		"participant_id",
	}
	order := fmt.Sprintf("%s %s", orders[pr.Parameters.Order[0].Column], strings.ToUpper(pr.Parameters.Order[0].Direction))
	user := c.Value("current_user").(*models.User)
	if user.Admin || user.Permission.StudyCoordinator || user.Permission.StudyTeamMember {
		q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Paginate(page, pr.Parameters.Length).Order(order)
	} else {
		err = errors.New("not enough previleges")
		return c.Render(403, r.JSON(err))
	}

	// Retrieve all Posts from the DB
	if err = q.All(participants); err != nil {
		return c.Render(404, r.JSON(err))
	}

	s := 0
	if s, err = tx.Count(&models.Participant{}); err != nil {
		s = 0
	}

	responseResult.RecordsTotal = s
	responseResult.RecordsFiltered = q.Paginator.TotalEntriesSize

	data := make([][]string, 0)
	for _, p := range *participants {
		scRef, ovRef := "", ""
		if len(p.Screenings) > 0 {
			if p.Screenings[0].Referral.Referred.Valid {
				if p.Screenings[0].Referral.Referred.Bool {
					scRef = "Yes"
				} else {
					scRef = "No"
				}
			}
		}
		if len(p.OverReadings) > 0 {
			if p.OverReadings[0].Referral.Referred.Valid {
				if p.OverReadings[0].Referral.Referred.Bool {
					ovRef = "Yes"
				} else {
					ovRef = "No"
				}
			}
		}
		temp := []string{
			p.ParticipantID[1:2],
			p.ParticipantID,
			p.ID.String(),
			Age(p.DOB.CalculatedDate),
			p.Gender,
			LanguageDate(p.CreatedAt, "2006-01-02", c.Value("current_lang").(string)),
			strconv.Itoa(p.Completeness()) + "%",
			scRef,
			ovRef,
		}

		data = append(data, temp)
	}

	responseResult.Data = data

	return c.Render(200, r.JSON(responseResult))
}

// DatatableResponse object
type DatatableResponse struct {
	Draw            int        `json:"draw"`
	RecordsTotal    int        `json:"recordsTotal"`
	RecordsFiltered int        `json:"recordsFiltered"`
	Data            [][]string `json:"data"`
}

// DatatableRequest object
type DatatableRequest struct {
	Parameters Parameters `json:"parameters"`
}

// Parameters object
type Parameters struct {
	Draw   int     `json:"draw"`
	Length int     `json:"length"`
	Start  int     `json:"start"`
	Search Search  `json:"search"`
	Order  []Order `json:"order"`
}

// Search object
type Search struct {
	Regex bool   `json:"regex"`
	Value string `json:"value"`
}

// Order object
type Order struct {
	Column    int    `json:"column"`
	Direction string `json:"dir"`
}
