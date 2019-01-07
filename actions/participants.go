package actions

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/helpers"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// ParticipantsIndex default implementation.
func ParticipantsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participants := &models.Participants{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	var q *pop.Query

	where := make([]string, 0)
	wheres := make([]interface{}, 0)
	if len(c.Param("status")) > 0 {
		where = append(where, "status = ?")
		wheres = append(wheres, c.Param("status"))
	}

	if len(c.Param("search")) > 0 {
		where = append(where, "participant_id LIKE ?")
		wheres = append(wheres, "%"+strings.ToUpper(c.Param("search"))+"%")
	}

	whereStmt := strings.Join(where, " AND ")

	user := c.Value("current_user").(*models.User)
	if user.Admin || user.Permission.StudyCoordinator || user.Permission.Screening {
		if len(whereStmt) > 0 {
			q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where(whereStmt, wheres...).PaginateFromParams(c.Params()).Order("created_at DESC")
		} else {
			q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").PaginateFromParams(c.Params()).Order("created_at DESC")
		}
	} else {
		// If there are no errors set a success message
		c.Flash().Add("danger", "You don't have sufficient permission.")
		InsertLog("error", "User viewed participants error", "Insufficient permission", "", "", user.ID, c)
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
		InsertLog("error", "User viewed participants error", err.Error(), "", "", user.ID, c)
		return c.Redirect(302, "/")
	}
	// Make posts available inside the html template
	c.Set("participants", participants)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["page_participants_title"] = "/participants/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	c.Set("filterStatus", c.Params().Get("status"))
	c.Set("filterSearch", c.Params().Get("search"))
	logErr := InsertLog("view", "User viewed participants", "", "", "", user.ID, c)
	if logErr != nil {
		// return errors.WithStack(logErr)
		errStr := logErr.Error()
		errs := map[string][]string{
			"index_error": {errStr},
		}
		c.Set("errors", errs)
		InsertLog("error", "User viewed participants error", logErr.Error(), "", "", user.ID, c)
		return c.Render(422, r.HTML("participants/index.html"))
	}
	return c.Render(200, r.HTML("participants/index.html"))
}

// ParticipantsCreateGet for the insert form
func ParticipantsCreateGet(c buffalo.Context) error {
	c.Set("participant", &models.Participant{})
	// user := c.Value("current_user").(*models.User)
	// luhnID := helpers.GenerateLuhnIDWithGivenPrefix(user.Site)
	// c.Set("luhnID", luhnID.ID)
	user := c.Value("current_user").(*models.User)
	prefix := "P"
	if len(user.Site) > 0 {
		prefix = prefix + user.Site
	}
	c.Set("participantIDPrefix", prefix)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["page_participants_title"] = "/participants/index"
	breadcrumbMap["breadcrumb_enrol_participant"] = "/participants/create"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("participants/create.html"))
}

// ParticipantsCreatePost for posting
func ParticipantsCreatePost(c buffalo.Context) error {
	// Allocate an empty Post
	participant := &models.Participant{}
	user := c.Value("current_user").(*models.User)

	prefix := "P"
	if len(user.Site) > 0 {
		prefix = prefix + user.Site
	}
	c.Set("participantIDPrefix", prefix)

	// Bind participant to the html form elements
	if err := c.Bind(participant); err != nil {
		// return errors.WithStack(err)
		errStr := err.Error()
		errs := map[string][]string{
			"create_error": {errStr},
		}
		c.Set("participant", participant)
		c.Set("errors", errs)
		return c.Render(422, r.HTML("participants/create.html"))
	}

	currentDate := time.Now()
	birthYear, err := strconv.Atoi(c.Request().FormValue("BirthYear"))
	maxYear := currentDate.Year() - 10
	minYear := currentDate.Year() - 100
	currentCalendar := "Gregorian"
	if participant.DOB.Calendar == "thai" {
		maxYear += 543
		minYear += 543
		currentCalendar = "Thai"
	}
	if err == nil && birthYear >= minYear && birthYear <= maxYear {
		today := time.Now().Year()
		diff := birthYear - today
		participant.DOB.GivenDate = time.Now().AddDate(diff, 0, 0)
	} else {
		errStr := fmt.Sprintf("Invalid birth year given, please re-check your input. For %s calendar, valid year of birth is between %d and %d.", currentCalendar, minYear, maxYear)
		errs := map[string][]string{
			"checksum_error": {errStr},
		}
		c.Set("participant", participant)
		c.Set("errors", errs)
		return c.Render(422, r.HTML("participants/create.html"))
	}

	if len(participant.ParticipantID) != 9 || !helpers.Valid(participant.ParticipantID, false) || strings.Contains(participant.ParticipantID, "_") {
		errStr := "Invalid Participant ID, please check your input again for valid checksum."
		errs := map[string][]string{
			"checksum_error": {errStr},
		}
		c.Set("participant", participant)
		c.Set("errors", errs)
		return c.Render(422, r.HTML("participants/create.html"))
	}

	if len(participant.ParticipantID) == 9 && !strings.HasPrefix(participant.ParticipantID, prefix) {
		errStr := "Participant ID should start with letter \"" + prefix + "\"."
		errs := map[string][]string{
			"checksum_error": {errStr},
		}
		c.Set("participant", participant)
		c.Set("errors", errs)
		return c.Render(422, r.HTML("participants/create.html"))
	}

	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Validate the data from the html form
	participant.UserID = user.ID
	participant.Status = "1"

	// c.Set("luhnID", c.Request().FormValue("ParticipantID"))
	verrs, err := tx.ValidateAndCreate(participant)
	if err != nil {
		// return errors.WithStack(err)
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate") && strings.Contains(errStr, "participants_participant_id_idx") {
			errStr = "Participant ID already in use. Please check again for correct Participant ID."
		}
		errs := map[string][]string{
			"create_error": {errStr},
		}
		c.Set("participant", participant)
		c.Set("errors", errs)
		return c.Render(422, r.HTML("participants/create.html"))
	}
	if verrs.HasAny() {
		c.Set("participant", participant)
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("participants/create.html"))
	}
	logErr := InsertLog("create", "User created a participant", "", participant.ID.String(), "participant", user.ID, c)
	if logErr != nil {
		// return errors.WithStack(logErr)
		errStr := logErr.Error()
		errs := map[string][]string{
			"create_error": {errStr},
		}
		c.Set("participant", participant)
		c.Set("errors", errs)
		return c.Render(422, r.HTML("participants/create.html"))
	}

	btnAction := c.Request().FormValue("submitBtn")
	redirectPath := "/participants/index"
	message := "New participant registered into study with Participant ID " + participant.ParticipantID
	if btnAction == "enrollAddAnother" {
		redirectPath = "/participants/create"
		message += ". You can add another one now."
	} else if btnAction == "enrollGoToScreening" {
		redirectPath = "/participants/" + participant.ID.String() + "/screenings/create"
		message += ". You can add screening for this participant here."
	}

	// If there are no errors set a success message
	c.Flash().Add("success", message)
	// and redirect to the index page
	return c.Redirect(302, redirectPath)
}

// ParticipantsEditGet displays a form to edit the post.
func ParticipantsEditGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("participant", participant)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["page_participants_title"] = "/participants/index"
	breadcrumbMap["breadcrumb_enrol_participant"] = "/participants/edit"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("participants/edit.html"))
}

// ParticipantsEditPost updates a post.
func ParticipantsEditPost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	participant.Consented = false
	if err := c.Bind(participant); err != nil {
		return errors.WithStack(err)
	}
	verrs, err := tx.ValidateAndUpdate(participant)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("participant", participant)
		c.Set("errors", verrs.Errors)
		breadcrumbMap := make(map[string]interface{})
		breadcrumbMap["page_participants_title"] = "/participants/index"
		breadcrumbMap["breadcrumb_enrol_participant"] = "/participants/edit"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("participants/edit.html"))
	}
	user := c.Value("current_user").(*models.User)
	logErr := InsertLog("update", "User updated a participant", "", participant.ID.String(), "participant", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}
	c.Flash().Add("success", "Participant was updated successfully.")
	return c.Redirect(302, "/participants/index")
}

// ParticipantsDetail default implementation.
func ParticipantsDetail(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader", "Referrals").Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("participant", participant)

	userActivities := make(map[string][]map[string]string)
	participantCreatedDate := participant.CreatedAt.Format("2006-01-02")
	participantCreatedTime := participant.CreatedAt.Format("3:04")
	participantCreatedPm := participant.CreatedAt.Format("pm")
	participantCreatedMsg := participant.User.Name + " registered the participant"

	userActivities[participantCreatedDate] = append(userActivities[participantCreatedDate], map[string]string{
		"time": participantCreatedTime,
		"ampm": participantCreatedPm,
		"msg":  participantCreatedMsg,
	})

	if len(participant.Screenings) > 0 {
		screeningCreatedDate := participant.Screenings[0].CreatedAt.Format("2006-01-02")
		screeningCreatedTime := participant.Screenings[0].CreatedAt.Format("3:04")
		screeningCreatedPm := participant.Screenings[0].CreatedAt.Format("pm")
		screeningCreatedMsg := participant.Screenings[0].Screener.Name + " screened the participant"

		userActivities[screeningCreatedDate] = append(userActivities[screeningCreatedDate], map[string]string{
			"time": screeningCreatedTime,
			"ampm": screeningCreatedPm,
			"msg":  screeningCreatedMsg,
		})
	}

	if len(participant.OverReadings) > 0 {
		overReadCreatedDate := participant.OverReadings[0].CreatedAt.Format("2006-01-02")
		overReadCreatedTime := participant.OverReadings[0].CreatedAt.Format("3:04")
		overReadCreatedPm := participant.OverReadings[0].CreatedAt.Format("pm")
		overReadCreatedMsg := participant.OverReadings[0].OverReader.Name + " over read the participant"

		userActivities[overReadCreatedDate] = append(userActivities[overReadCreatedDate], map[string]string{
			"time": overReadCreatedTime,
			"ampm": overReadCreatedPm,
			"msg":  overReadCreatedMsg,
		})
	}

	openNotifications := &models.Notifications{}
	if err := tx.Eager().Where("participant_id = ?", participant.ID).Where("status != ?", "closed").All(openNotifications); err != nil {
		return c.Error(404, err)
	}
	c.Set("open_notifications", openNotifications)

	notifications := &models.Notifications{}
	if err := tx.Eager().Where("participant_id = ?", participant.ID).All(notifications); err != nil {
		return c.Error(404, err)
	}

	for _, n := range *notifications {
		logs := &models.SystemLogs{}
		if err := tx.Eager().Where("resource_id = ?", n.ID).Where("resource_type = ?", "notification").All(logs); err != nil {
			return c.Error(404, err)
		}

		for _, l := range *logs {
			logCreatedDate := l.CreatedAt.Format("2006-01-02")
			logCreatedTime := l.CreatedAt.Format("3:04")
			logCreatedPm := l.CreatedAt.Format("pm")
			logCreatedMsg := l.User.Name + " performed activity on notification: " + l.Activity

			userActivities[logCreatedDate] = append(userActivities[logCreatedDate], map[string]string{
				"time": logCreatedTime,
				"ampm": logCreatedPm,
				"msg":  logCreatedMsg,
			})
		}
	}

	slogs := &models.SystemLogs{}
	if err := tx.Eager().Where("resource_id = ?", participant.ID).Where("resource_type = ?", "participant").Where("action = ?", "update").All(slogs); err != nil {
		return c.Error(404, err)
	}

	for _, l := range *slogs {
		logCreatedDate := l.CreatedAt.Format("2006-01-02")
		logCreatedTime := l.CreatedAt.Format("3:04")
		logCreatedPm := l.CreatedAt.Format("pm")
		logCreatedMsg := l.User.Name + " performed activity on participant: " + l.Activity

		userActivities[logCreatedDate] = append(userActivities[logCreatedDate], map[string]string{
			"time": logCreatedTime,
			"ampm": logCreatedPm,
			"msg":  logCreatedMsg,
		})
	}

	var keys []string
	for k := range userActivities {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	audits := &models.Audits{}
	if len(participant.Screenings) > 0 {
		if err := tx.Eager().Where("model_type = ?", "Screening").Where("model_id = ?", participant.Screenings[0].ID).All(audits); err != nil {
			return c.Error(404, err)
		}
	}

	allLogs := &models.SystemLogs{}
	if len(participant.Screenings) > 0 {
		if err := tx.Eager().Where("resource_id = ?", participant.Screenings[0].ID).Where("resource_type = ?", "screening").All(allLogs); err != nil {
			return c.Error(404, err)
		}
	}

	c.Set("user_activities", userActivities)
	c.Set("activities_keys", keys)
	c.Set("audits", audits)
	c.Set("logs", allLogs)

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["page_participants_title"] = "/participants/index"
	breadcrumbMap["breadcrumb_enrol_participant_short"] = "/participants/" + participant.ID.String()
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("participants/detail.html"))
}

// ParticipantsReferralAppointmentDone implementation
func ParticipantsReferralAppointmentDone(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	participant.ReferralAppointment = true
	verrs, err := tx.ValidateAndUpdate(participant)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("errors", verrs.Errors)
		return c.Redirect(302, "/referrals/index")
	}
	user := c.Value("current_user").(*models.User)
	logErr := InsertLog("update", "User marked a participant for appointment completed", "", participant.ID.String(), "participant", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}
	c.Flash().Add("success", "Participant is marked successfully.")
	return c.Redirect(302, "/referrals/index")
}
