package actions

import (
	"bytes"
	"encoding/csv"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
)

// AnalyticsIndex default implementation.
func AnalyticsIndex(c buffalo.Context) error {
	download := false
	dlValue := c.Params().Get("download")
	if dlValue == "Download CSV" {
		download = true
	}

	tx := c.Value("tx").(*pop.Connection)
	user := c.Value("current_user").(*models.User)
	analyticsScreenings := &models.AnalyticsScreenings{}

	query := `SELECT 
	s.created_at AS "CreatedDate",
	p.participant_id AS "ParticipantID",
	date_part('year', age(((p.dob->>'calculated_date'::text)::date)::timestamp with time zone)) AS "Age",
        CASE
            WHEN ((p.gender)::text = 'M'::text) THEN 'Male'::text
            ELSE 'Female'::text
        END AS "Gender",
	((s.eye->'right'::text)->>'visual_acuity'::text) AS "VAOD",
	((s.eye->'right'::text)->>'last_visual_acuity'::text) AS "VAPreviousOD",
	((s.eye->'right'::text)->>'dr'::text) AS "DRGradeOD",
	((s.eye->'right'::text)->>'dme'::text) AS "DMEOD",
	((s.eye->'left'::text)->>'visual_acuity'::text) AS "VAOS",
	((s.eye->'left'::text)->>'last_visual_acuity'::text) AS "VAPreviousOS",
	((s.eye->'left'::text)->>'dr'::text) AS "DRGradeOS",
	((s.eye->'left'::text)->>'dme'::text) AS "DMEOS",
        CASE
            WHEN ((s.referral->>'referred'::text) = 'true'::text) THEN 'Yes'::text
            ELSE 'No'::text
        END AS "DrReferral",
	(s.referral->>'additional_notes'::text) AS "ReferralNotes"
FROM (
	participants p
	LEFT JOIN screenings s ON ((p.id = s.participant_id))
)
WHERE (
	s.id IS NOT NULL
)
ORDER BY s.created_at DESC`

	//  AND right(left(p.participant_id, 2), 1) = ?

	var q *pop.Query

	if download {
		q = tx.RawQuery(query)
	} else {
		q = tx.RawQuery(query).PaginateFromParams(c.Params())
	}
	// participants := &models.Participants{}
	// // Paginate results. Params "page" and "per_page" control pagination.
	// // Default values are "page=1" and "per_page=20".

	//
	// var err error

	// if user.Admin || user.Permission.StudyCoordinator {
	// 	if len(c.Param("status")) > 0 {
	// 		q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status = ?", c.Param("status")).PaginateFromParams(c.Params()).Order("created_at DESC")
	// 	} else {
	// 		q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").PaginateFromParams(c.Params()).Order("created_at DESC")
	// 	}
	// } else if user.Permission.Screening && user.Permission.OverRead {
	// 	q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status != ?", "111").Where("participants.participant_id LIKE '_" + user.Site + "%'").PaginateFromParams(c.Params()).Order("created_at DESC")
	// } else if user.Permission.Screening {
	// 	q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status LIKE ?", "1%").Where("participants.participant_id LIKE '_" + user.Site + "%'").PaginateFromParams(c.Params()).Order("created_at DESC")
	// } else if user.Permission.OverRead {
	// 	q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status LIKE ?", "11%").PaginateFromParams(c.Params()).Order("created_at DESC")
	// } else {
	// 	// If there are no errors set a success message
	// 	c.Flash().Add("danger", "You don't have sufficient permission.")
	// 	InsertLog("error", "User viewed analytics error", "Insufficient permission", "", "", user.ID, c)
	// 	// and redirect to the index page
	// 	return c.Redirect(302, "/")
	// }

	// Retrieve all Posts from the DB
	if err := q.All(analyticsScreenings); err != nil {
		// return errors.WithStack(err)
		errStr := err.Error()
		errs := map[string][]string{
			"index_error": {errStr},
		}
		c.Set("errors", errs)
		if download {
			InsertLog("error", "User download analytics error", err.Error(), "", "", user.ID, c)
		} else {
			InsertLog("error", "User viewed analytics error", err.Error(), "", "", user.ID, c)
		}
		return c.Redirect(302, "/")
	}

	if download {
		b, err := downloadAnalytics(*analyticsScreenings)
		if err != nil {
			InsertLog("error", "User download analytics error", err.Error(), "", "", user.ID, c)
			return c.Redirect(302, "/")
		}

		InsertLog("error", "User download analytics", "", "", "", user.ID, c)
		appHost := envy.Get("APP_HOST", "http://127.0.0.1")
		hosts := strings.Split(strings.TrimSpace(strings.Replace(appHost, "/", "", -1)), ":")
		filename := hosts[1] + "-analytics-" + time.Now().Format("2006-01-02T15-04-05-0700") + ".csv"

		return c.Render(200, r.Download(c, filename, b))
	}

	// Make posts available inside the html template
	c.Set("participants", analyticsScreenings)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["page_analytics_title"] = "/analytics/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	c.Set("filterStatus", c.Params().Get("status"))
	logErr := InsertLog("view", "User viewed analytics", "", "", "", user.ID, c)
	if logErr != nil {
		// return errors.WithStack(logErr)
		errStr := logErr.Error()
		errs := map[string][]string{
			"index_error": {errStr},
		}
		c.Set("errors", errs)
		InsertLog("error", "User viewed analytics error", logErr.Error(), "", "", user.ID, c)
		return c.Render(422, r.HTML("analytics/index.html"))
	}
	return c.Render(200, r.HTML("analytics/index.html"))
}

func downloadAnalytics(analytics []models.AnalyticsScreening) (*bytes.Buffer, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	headers := []string{
		"Created Date",
		"Participant ID",
		"Age (Year)",
		"Gender",
		"VA (OD)",
		"VAPrevious (OD)",
		"DRGrade (OD)",
		"DME (OD)",
		"VA (OS)",
		"VAPrevious (OS)",
		"DRGrade (OS)",
		"DME (OS)",
		"DrReferral",
		"ReferralNotes",
	}

	if err := w.Write(headers); err != nil {
		return nil, err
	}

	for _, a := range analytics {
		var record []string
		record = append(record, a.CreatedDate.Format(time.RFC3339))
		record = append(record, a.ParticipantID)
		record = append(record, strconv.FormatInt(int64(a.Age), 10))
		record = append(record, a.Gender)
		record = append(record, a.VAOD.String)
		record = append(record, a.VAPreviousOD.String)
		record = append(record, a.DRGradeOD.String)
		record = append(record, a.DMEOD.String)
		record = append(record, a.VAOS.String)
		record = append(record, a.VAPreviousOS.String)
		record = append(record, a.DRGradeOS.String)
		record = append(record, a.DMEOS.String)
		record = append(record, a.DrReferral.String)
		record = append(record, a.ReferralNotes.String)

		if err := w.Write(record); err != nil {
			return nil, err
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return nil, err
	}

	return b, nil
}
