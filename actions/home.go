package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c buffalo.Context) error {
	loggedInUser, ok := c.Value("current_user").(*models.User)
	if !ok {
		return c.Redirect(302, "/users/login")
	}

	tx := c.Value("tx").(*pop.Connection)
	var err error
	site := ""
	if c.Value("current_site") != nil {
		site = c.Value("current_site").(string)
	}

	p := 0
	if len(site) > 0 {
		if p, err = tx.Where("SUBSTRING(participant_id,2,1) = ?", site).Count(&models.Participant{}); err != nil {
			p = 0
		}
	} else {
		if p, err = tx.Count(&models.Participant{}); err != nil {
			p = 0
		}
	}
	c.Set("participants", p)

	s := 0
	if len(site) > 0 {
		allParticipants := tx.Where("SUBSTRING(participants.participant_id,2,1) = ?", site)
		query := allParticipants.LeftJoin("participants", "participants.id=screenings.participant_id")
		sql, args := query.ToSQL(&pop.Model{Value: models.Screening{}}, "screenings.id", "participants.participant_id")
		if s, err = allParticipants.RawQuery(sql, args...).Count(&models.Screening{}); err != nil {
			s = 0
		}
	} else {
		if s, err = tx.Count(&models.Screening{}); err != nil {
			s = 0
		}
	}
	c.Set("screenings", s)

	o := 0
	if len(site) > 0 {
		allParticipants := tx.Where("SUBSTRING(participants.participant_id,2,1) = ?", site)
		query := allParticipants.LeftJoin("participants", "participants.id=over_readings.participant_id")
		sql, args := query.ToSQL(&pop.Model{Value: models.OverReading{}}, "over_readings.id", "participants.participant_id")
		if o, err = allParticipants.RawQuery(sql, args...).Count(&models.OverReading{}); err != nil {
			o = 0
		}
	} else {
		if o, err = tx.Count(&models.OverReading{}); err != nil {
			o = 0
		}
	}
	c.Set("overreadings", o)

	u := 0
	if u, err = tx.Count(&models.User{}); err != nil {
		u = 0
	}
	c.Set("users", u)

	participants := &models.Participants{}
	var qov *pop.Query

	if len(site) > 0 {
		qov = tx.Eager("User", "Screenings", "Screenings.Screener", "OverReadings", "OverReadings.OverReader").Where("status LIKE ?", "11%").Where("SUBSTRING(participants.participant_id,2,1) = ?", site)
	} else {
		qov = tx.Eager("User", "Screenings", "Screenings.Screener", "OverReadings", "OverReadings.OverReader").Where("status LIKE ?", "11%")
	}
	// Retrieve all Posts from the DB
	cn := 0
	if cn, err = qov.Count(participants); err != nil {
		cn = 0
	}

	// c.Set("total_cases", len(*participants))
	openCases := 0
	if cn > o {
		openCases = cn - o
	}
	c.Set("open_cases", openCases)

	// Notifications
	notifications := &models.Notifications{}
	var q *pop.Query
	openNotificationStatuses := []string{"open", "nurse-notified", "patient-contacted", "referral-arranged"}
	to := 5

	if len(site) > 0 {
		q = tx.Eager().Where("site = ?", site).Where("status in (?)", openNotificationStatuses).Paginate(1, to).Order("created_at DESC")
	} else if loggedInUser.Admin || loggedInUser.Permission.StudyCoordinator {
		q = tx.Eager().Where("status in (?)", openNotificationStatuses).Paginate(1, to).Order("created_at DESC")
	} else {
		q = tx.Eager().Where("status != ?", "unknown").Paginate(1, to).Order("created_at DESC")
	}

	// Retrieve all Notifications from the DB
	if err := q.All(notifications); err != nil {
		return errors.WithStack(err)
	}

	// Make posts available inside the html template
	c.Set("notifications", notifications)
	c.Set("total_notifications", q.Paginator.TotalEntriesSize)

	return c.Render(200, r.HTML("index.html"))
}
