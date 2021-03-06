package actions

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
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
		where = append(where, "replace(participants.participant_id, '-', '') LIKE ?")
		wheres = append(wheres, "%"+strings.Replace(strings.ToUpper(c.Param("search")), "-", "", -1)+"%")
	}

	user := c.Value("current_user").(*models.User)
	site := ""
	if c.Value("current_site") != nil {
		site = c.Value("current_site").(string)
	}

	if len(site) > 0 {
		where = append(where, "SUBSTRING(participant_id,2,1) = ?")
		wheres = append(wheres, site)
	}

	whereStmt := strings.Join(where, " AND ")

	if user.Admin || user.Permission.StudyCoordinator || user.Permission.Screening || user.Permission.StudyTeamMember {
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

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "page_participants_title", Path: "/participants/index"})
	c.Set("breadcrumb", b)

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
	currentSite := c.Value("current_site").(string)
	prefix := "P"
	if len(currentSite) > 0 {
		prefix = prefix + currentSite
	}
	c.Set("participantIDPrefix", prefix)

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "page_participants_title", Path: "/participants/index"})
	b = append(b, helpers.Breadcrumb{Title: "breadcrumb_enrol_participant", Path: "/participants/create"})
	c.Set("breadcrumb", b)

	return c.Render(200, r.HTML("participants/create.html"))
}

// ParticipantsCreatePost for posting
func ParticipantsCreatePost(c buffalo.Context) error {
	// Allocate an empty Post
	participant := &models.Participant{}
	oldParticipant := participant.Maps()
	user := c.Value("current_user").(*models.User)

	currentSite := c.Value("current_site").(string)
	prefix := "P"
	if len(currentSite) > 0 {
		prefix = prefix + currentSite
	}
	c.Set("participantIDPrefix", prefix)

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "page_participants_title", Path: "/participants/index"})
	b = append(b, helpers.Breadcrumb{Title: "breadcrumb_enrol_participant", Path: "/participants/create"})
	c.Set("breadcrumb", b)

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

	newParticipant := participant.Maps()
	auditErr := MakeAudit("Participant", participant.ID, oldParticipant, newParticipant, user.ID, c)
	if auditErr != nil {
		return errors.WithStack(auditErr)
	}

	logErr := InsertLog("create", "User created a participant", "", participant.ID.String(), "participant", user.ID, c)
	if logErr != nil {
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

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "page_participants_title", Path: "/participants/index"})
	b = append(b, helpers.Breadcrumb{Title: "breadcrumb_enrol_update_participant", Path: "/participants/edit/" + participant.ID.String()})
	c.Set("breadcrumb", b)

	return c.Render(200, r.HTML("participants/edit.html"))
}

// ParticipantsEditPost updates a post.
func ParticipantsEditPost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	user := c.Value("current_user").(*models.User)

	currentSite := c.Value("current_site").(string)
	prefix := "P"
	if len(currentSite) > 0 {
		prefix = prefix + currentSite
	}
	c.Set("participantIDPrefix", prefix)

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "page_participants_title", Path: "/participants/index"})
	b = append(b, helpers.Breadcrumb{Title: "breadcrumb_enrol_update_participant", Path: "/participants/edit/" + participant.ID.String()})
	c.Set("breadcrumb", b)

	if err := tx.Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	oldParticipant := participant.Maps()

	if err := c.Bind(participant); err != nil {
		return errors.WithStack(err)
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
		return c.Render(422, r.HTML("participants/edit.html"))
	}

	if len(participant.ParticipantID) != 9 || !helpers.Valid(participant.ParticipantID, false) || strings.Contains(participant.ParticipantID, "_") {
		errStr := "Invalid Participant ID, please check your input again for valid checksum."
		errs := map[string][]string{
			"checksum_error": {errStr},
		}
		c.Set("participant", participant)
		c.Set("errors", errs)
		return c.Render(422, r.HTML("participants/edit.html"))
	}

	if len(participant.ParticipantID) == 9 && !strings.HasPrefix(participant.ParticipantID, prefix) {
		errStr := "Participant ID should start with letter \"" + prefix + "\"."
		errs := map[string][]string{
			"checksum_error": {errStr},
		}
		c.Set("participant", participant)
		c.Set("errors", errs)
		return c.Render(422, r.HTML("participants/edit.html"))
	}

	verrs, err := tx.ValidateAndUpdate(participant)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("participant", participant)
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("participants/edit.html"))
	}

	newParticipant := participant.Maps()
	auditErr := MakeAudit("Participant", participant.ID, oldParticipant, newParticipant, user.ID, c)
	if auditErr != nil {
		return errors.WithStack(auditErr)
	}

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

	openNotificationStatuses := []string{"open", "nurse-notified", "patient-contacted", "referral-arranged"}
	openNotifications := &models.Notifications{}
	if err := tx.Eager().Where("participant_id = ?", participant.ID).Where("status in (?)", openNotificationStatuses).All(openNotifications); err != nil {
		return c.Error(404, err)
	}
	c.Set("open_notifications", openNotifications)

	notifications := &models.Notifications{}
	if err := tx.Eager().Where("participant_id = ?", participant.ID).All(notifications); err != nil {
		return c.Error(404, err)
	}

	notificationIDs := make([]string, 0)
	for _, n := range *notifications {
		notificationIDs = append(notificationIDs, n.ID.String())
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

	audits := []models.Audit{}
	screeningAudits := &models.Audits{}
	participantAudits := &models.Audits{}
	overReadingAudits := &models.Audits{}
	notificationAudits := &models.Audits{}
	referralAudits := &models.Audits{}
	if err := tx.Eager().Where("model_type = ?", "Participant").Where("model_id = ?", participant.ID).All(participantAudits); err != nil {
		return c.Error(404, err)
	}
	if len(participant.Screenings) > 0 {
		if err := tx.Eager().Where("model_type = ?", "Screening").Where("model_id = ?", participant.Screenings[0].ID).All(screeningAudits); err != nil {
			return c.Error(404, err)
		}
	}
	if len(participant.OverReadings) > 0 {
		if err := tx.Eager().Where("model_type = ?", "OverReading").Where("model_id = ?", participant.OverReadings[0].ID).All(overReadingAudits); err != nil {
			return c.Error(404, err)
		}
	}
	if len(notificationIDs) > 0 {
		if err := tx.Eager().Where("model_type = ?", "Notification").Where("model_id in (?)", notificationIDs).All(notificationAudits); err != nil {
			return c.Error(404, err)
		}
	}
	if len(participant.Referrals) > 0 {
		if err := tx.Eager().Where("model_type = ?", "ReferredMessage").Where("model_id = ?", participant.Referrals[0].ID).All(referralAudits); err != nil {
			return c.Error(404, err)
		}
	}
	audits = append(audits, *participantAudits...)
	audits = append(audits, *screeningAudits...)
	audits = append(audits, *overReadingAudits...)
	audits = append(audits, *notificationAudits...)
	audits = append(audits, *referralAudits...)

	allLogs := []models.SystemLog{}
	participantLogs := &models.SystemLogs{}
	screeningLogs := &models.SystemLogs{}
	overReadingLogs := &models.SystemLogs{}
	notificationLogs := &models.SystemLogs{}
	referralLogs := &models.SystemLogs{}
	if err := tx.Eager().Where("resource_id = ?", participant.ID).Where("resource_type = ?", "participant").All(participantLogs); err != nil {
		return c.Error(404, err)
	}
	if len(participant.Screenings) > 0 {
		if err := tx.Eager().Where("resource_id = ?", participant.Screenings[0].ID).Where("resource_type = ?", "screening").All(screeningLogs); err != nil {
			return c.Error(404, err)
		}
	}
	if len(participant.OverReadings) > 0 {
		if err := tx.Eager().Where("resource_id = ?", participant.OverReadings[0].ID).Where("resource_type = ?", "overReading").All(overReadingLogs); err != nil {
			return c.Error(404, err)
		}
	}
	if len(notificationIDs) > 0 {
		if err := tx.Eager().Where("resource_id in (?)", notificationIDs).Where("resource_type = ?", "notification").All(notificationLogs); err != nil {
			return c.Error(404, err)
		}
	}
	if len(participant.Referrals) > 0 {
		if err := tx.Eager().Where("resource_id = ?", participant.Referrals[0].ID).Where("resource_type = ?", "referred_message").All(referralLogs); err != nil {
			return c.Error(404, err)
		}
	}
	allLogs = append(allLogs, *participantLogs...)
	allLogs = append(allLogs, *screeningLogs...)
	allLogs = append(allLogs, *overReadingLogs...)
	allLogs = append(allLogs, *notificationLogs...)
	allLogs = append(allLogs, *referralLogs...)

	right, left, err := getImage(participant.ParticipantID)
	if err != nil {
		left = ""
		right = ""
	}

	c.Set("leftEyeLink", left)
	c.Set("rightEyeLink", right)

	drReferral := false
	ovReferral := false

	if len(participant.Screenings) > 0 {
		drReferral = participant.Screenings[0].Referral.Referred.Valid && participant.Screenings[0].Referral.Referred.Bool
	}
	if len(participant.OverReadings) > 0 {
		ovReferral = participant.OverReadings[0].Referral.Referred.Valid && participant.OverReadings[0].Referral.Referred.Bool
	}

	c.Set("drReferral", drReferral)
	c.Set("ovReferral", ovReferral)

	c.Set("user_activities", userActivities)
	c.Set("activities_keys", keys)
	c.Set("audits", audits)
	c.Set("logs", allLogs)

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "page_participants_title", Path: "/participants/index"})
	b = append(b, helpers.Breadcrumb{Title: "breadcrumb_enrol_participant_short", Path: "/participants/" + participant.ID.String()})
	c.Set("breadcrumb", b)

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

// ParticipantsDestroy for the insert form
func ParticipantsDestroy(c buffalo.Context) error {
	returnURL := "/participants/index"
	user := c.Value("current_user").(*models.User)
	if !user.Admin {
		c.Flash().Add("danger", "Access denied")
		return c.Redirect(302, returnURL)
	}

	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager().Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}

	reason := c.Request().FormValue("reason")

	for _, o := range participant.OverReadings {
		err := ArchiveMake(c, user.ID, o.ID, "OverReading", &o, reason)
		if err != nil {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(302, returnURL)
		}
	}

	for _, o := range participant.Notifications {
		err := ArchiveMake(c, user.ID, o.ID, "Notification", &o, reason)
		if err != nil {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(302, returnURL)
		}
	}

	for _, o := range participant.Referrals {
		err := ArchiveMake(c, user.ID, o.ID, "ReferredMessage", &o, reason)
		if err != nil {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(302, returnURL)
		}
	}

	for _, o := range participant.Screenings {
		err := ArchiveMake(c, user.ID, o.ID, "Screening", &o, reason)
		if err != nil {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(302, returnURL)
		}
	}

	err := ArchiveMake(c, user.ID, participant.ID, "Participant", participant, reason)
	if err != nil {
		c.Flash().Add("danger", err.Error())
		return c.Redirect(302, returnURL)
	}

	c.Flash().Add("success", "Archived successfully")

	return c.Redirect(302, returnURL)
}
