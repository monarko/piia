package actions

import (
	"strings"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop/v5"

	"github.com/gobuffalo/buffalo"
	"github.com/monarko/piia/helpers"
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

	site := ""
	if c.Value("current_site") != nil {
		site = strings.TrimSpace(c.Value("current_site").(string))
	}
	whereso := make([]string, 0)
	wheresso := make([]interface{}, 0)
	whereso = append(whereso, "referral->>'referred' = ?")
	wheresso = append(wheresso, "true")
	whereStmtso := strings.Join(whereso, " AND ")

	screenings := &models.Screenings{}
	sq := tx.Eager("Participant").Where(whereStmtso, wheresso...)
	if err := sq.All(screenings); err != nil {
		InsertLog("error", "User viewed referrals error", err.Error(), "", "", user.ID, c)
		return errors.WithStack(err)
	}

	overs := &models.OverReadings{}
	oq := tx.Eager("Participant").Where(whereStmtso, wheresso...)
	if err := oq.All(overs); err != nil {
		InsertLog("error", "User viewed referrals error", err.Error(), "", "", user.ID, c)
		return errors.WithStack(err)
	}

	ids := make([]string, 0)
	for _, s := range *screenings {
		pid := s.Participant.ParticipantID
		if len(site) == 0 || string(pid[1]) == site {
			ids = append(ids, s.ParticipantID.String())
		}
	}

	for _, o := range *overs {
		pid := o.Participant.ParticipantID
		if len(site) == 0 || string(pid[1]) == site {
			ids = append(ids, o.ParticipantID.String())
		}
	}

	refers := &models.ReferredMessages{}
	if err := tx.All(refers); err != nil {
		InsertLog("error", "User viewed referrals error", err.Error(), "", "", user.ID, c)
		return errors.WithStack(err)
	}

	rIds := make([]string, 0)

	for _, s := range *refers {
		t := s.ParticipantID.String()
		for _, o := range ids {
			if o == t {
				rIds = append(rIds, t)
			}
		}
	}

	ids = helpers.SliceStringUnique(ids, true)
	rIds = helpers.SliceStringUnique(rIds, true)

	where := make([]string, 0)
	wheres := make([]interface{}, 0)

	var idsToSearch []string

	if len(c.Param("status")) > 0 {
		if c.Param("status") == "completed" {
			idsToSearch = make([]string, len(rIds))
			copy(idsToSearch, rIds)
		} else if c.Param("status") == "open" {
			idsToSearch = make([]string, 0)
			for _, id := range ids {
				_, found := helpers.SliceContainsString(rIds, id)
				if !found {
					idsToSearch = append(idsToSearch, id)
				}
			}
		} else {
			idsToSearch = make([]string, len(ids))
			copy(idsToSearch, ids)
		}
	} else {
		idsToSearch = make([]string, len(ids))
		copy(idsToSearch, ids)
	}

	if len(c.Param("search")) > 0 {
		where = append(where, "replace(participants.participant_id, '-', '') LIKE ?")
		wheres = append(wheres, "%"+strings.Replace(strings.ToUpper(c.Param("search")), "-", "", -1)+"%")
	}

	if len(site) > 0 {
		where = append(where, "SUBSTRING(participants.participant_id,2,1) = ?")
		wheres = append(wheres, site)
	}

	whereStmt := strings.Join(where, " AND ")

	var q *pop.Query
	if len(idsToSearch) > 0 {
		if len(where) > 0 {
			q = tx.Eager("User", "Screenings", "Screenings.Screener", "OverReadings", "OverReadings.OverReader", "Referrals").Where("id in (?)", idsToSearch).Where("status LIKE ?", "1%").Where(whereStmt, wheres...).PaginateFromParams(c.Params()).Order("created_at DESC")
		} else {
			q = tx.Eager("User", "Screenings", "Screenings.Screener", "OverReadings", "OverReadings.OverReader", "Referrals").Where("id in (?)", idsToSearch).Where("status LIKE ?", "1%").PaginateFromParams(c.Params()).Order("created_at DESC")
		}
	} else {
		q = tx.Eager("User", "Screenings", "Screenings.Screener", "OverReadings", "OverReadings.OverReader", "Referrals").Where("gender = ? ", "abc").Where("status LIKE ?", "1%").PaginateFromParams(c.Params()).Order("created_at DESC")
	}

	// Retrieve all Posts from the DB
	if err := q.All(participants); err != nil {
		InsertLog("error", "User viewed referrals error", err.Error(), "", "", user.ID, c)
		return errors.WithStack(err)
	}
	// Make posts available inside the html template
	c.Set("participants", participants)
	c.Set("finished", rIds)
	c.Set("all_ids", ids)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "Referrals", Path: "/referrals/index"})
	c.Set("breadcrumb", b)

	c.Set("filterStatus", c.Params().Get("status"))
	c.Set("filterSearch", c.Params().Get("search"))
	logErr := InsertLog("view", "User viewed referrals", "", "", "", user.ID, c)
	if logErr != nil {
		InsertLog("error", "User viewed referrals error", logErr.Error(), "", "", user.ID, c)
		return errors.WithStack(logErr)
	}
	return c.Render(200, r.HTML("referrals/index.html"))
}

// ReferralsParticipantsGet returns form
func ReferralsParticipantsGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader", "Referrals").Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("participant", participant)

	hospitals := envy.Get("HOSPITALS", "")
	listHospitals := strings.Split(hospitals, ",")

	c.Set("hospitals", listHospitals)

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "Referrals", Path: "/referrals/index"})
	b = append(b, helpers.Breadcrumb{Title: "Referrals Update", Path: "/referrals/participants/" + participant.ID.String()})
	c.Set("breadcrumb", b)

	return c.Render(200, r.HTML("referrals/create.html"))
}

// ReferralsParticipantsView returns form
func ReferralsParticipantsView(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader", "Referrals").Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("participant", participant)

	hospitals := envy.Get("HOSPITALS", "")
	listHospitals := strings.Split(hospitals, ",")

	c.Set("hospitals", listHospitals)

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "Referrals", Path: "/referrals/index"})
	b = append(b, helpers.Breadcrumb{Title: "Referrals Details", Path: "/referrals/participants/" + participant.ID.String()})
	c.Set("breadcrumb", b)

	return c.Render(200, r.HTML("referrals/details.html"))
}

// ReferralsDestroy function
func ReferralsDestroy(c buffalo.Context) error {
	returnURL := "/referrals/index"
	user := c.Value("current_user").(*models.User)
	if !user.Admin {
		c.Flash().Add("danger", "Access denied")
		return c.Redirect(302, returnURL)
	}

	tx := c.Value("tx").(*pop.Connection)
	referral := &models.ReferredMessage{}
	if err := tx.Eager().Find(referral, c.Param("rid")); err != nil {
		return c.Error(404, err)
	}
	participant := referral.Participant
	if c.Param("pid") != participant.ID.String() {
		c.Flash().Add("danger", "Not Found")
		return c.Redirect(302, returnURL)
	}

	reason := c.Request().FormValue("reason")

	err := ArchiveMake(c, user.ID, referral.ID, "ReferredMessage", referral, reason)
	if err != nil {
		c.Flash().Add("danger", err.Error())
		return c.Redirect(302, returnURL)
	}

	// If there are no errors set a flash message
	c.Flash().Add("success", "Archived successfully")

	return c.Redirect(302, returnURL)
}
