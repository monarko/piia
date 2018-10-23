package actions

import (
	"sort"

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

	user := c.Value("current_user").(*models.User)
	if user.Admin || user.PermissionStudyCoordinator {
		if len(c.Param("status")) > 0 {
			q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status = ?", c.Param("status")).PaginateFromParams(c.Params()).Order("created_at ASC")
		} else {
			q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").PaginateFromParams(c.Params()).Order("created_at ASC")
		}
	} else if user.PermissionScreening && user.PermissionOverRead {
		q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status != ?", "111").Where("participants.participant_id LIKE '" + user.Site + "%'").PaginateFromParams(c.Params()).Order("created_at ASC")
	} else if user.PermissionScreening {
		q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status = ?", "1").Where("participants.participant_id LIKE '" + user.Site + "%'").PaginateFromParams(c.Params()).Order("created_at ASC")
	} else if user.PermissionOverRead {
		q = tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Where("status = ?", "11").Where("participants.participant_id LIKE '" + user.Site + "%'").PaginateFromParams(c.Params()).Order("created_at ASC")
	} else {
		// If there are no errors set a success message
		c.Flash().Add("danger", "You don't have sufficient permission.")
		// and redirect to the index page
		return c.Redirect(302, "/")
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
	breadcrumbMap["page_participants_title"] = "/participants/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	c.Set("filterStatus", c.Params().Get("status"))
	logErr := InsertLog("view", "User viewed participants", "", "", "", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}
	return c.Render(200, r.HTML("participants/index.html"))
}

// ParticipantsCreateGet for the insert form
func ParticipantsCreateGet(c buffalo.Context) error {
	c.Set("participant", &models.Participant{})
	user := c.Value("current_user").(*models.User)
	luhnID := helpers.GenerateLuhnIDWithGivenPrefix(user.Site)
	c.Set("luhnID", luhnID.ID)
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
	// Bind participant to the html form elements
	if err := c.Bind(participant); err != nil {
		return errors.WithStack(err)
	}
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Validate the data from the html form
	participant.UserID = user.ID
	participant.Status = "1"
	c.Set("luhnID", c.Request().FormValue("ParticipantID"))
	verrs, err := tx.ValidateAndCreate(participant)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("participant", participant)
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("participants/create.html"))
	}
	logErr := InsertLog("create", "User created a participant", "", participant.ID.String(), "participant", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "New participant added successfully.")
	// and redirect to the index page
	return c.Redirect(302, "/participants/index")
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
	if err := tx.Eager("User", "Screenings.Screener", "OverReadings.OverReader").Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("participant", participant)

	userActivities := make(map[string][]map[string]string)
	participantCreatedDate := participant.CreatedAt.Format("2006 Jan 02")
	participantCreatedTime := participant.CreatedAt.Format("3:04")
	participantCreatedPm := participant.CreatedAt.Format("pm")
	participantCreatedMsg := participant.User.Name + " registered the participant"

	screeningCreatedDate := participant.Screenings[0].CreatedAt.Format("2006 Jan 02")
	screeningCreatedTime := participant.Screenings[0].CreatedAt.Format("3:04")
	screeningCreatedPm := participant.Screenings[0].CreatedAt.Format("pm")
	screeningCreatedMsg := participant.Screenings[0].Screener.Name + " screened the participant"

	overReadCreatedDate := participant.OverReadings[0].CreatedAt.Format("2006 Jan 02")
	overReadCreatedTime := participant.OverReadings[0].CreatedAt.Format("3:04")
	overReadCreatedPm := participant.OverReadings[0].CreatedAt.Format("pm")
	overReadCreatedMsg := participant.OverReadings[0].OverReader.Name + " over read the participant"

	userActivities[participantCreatedDate] = append(userActivities[participantCreatedDate], map[string]string{
		"time": participantCreatedTime,
		"ampm": participantCreatedPm,
		"msg":  participantCreatedMsg,
	})

	userActivities[screeningCreatedDate] = append(userActivities[screeningCreatedDate], map[string]string{
		"time": screeningCreatedTime,
		"ampm": screeningCreatedPm,
		"msg":  screeningCreatedMsg,
	})

	userActivities[overReadCreatedDate] = append(userActivities[overReadCreatedDate], map[string]string{
		"time": overReadCreatedTime,
		"ampm": overReadCreatedPm,
		"msg":  overReadCreatedMsg,
	})

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
			logCreatedDate := l.CreatedAt.Format("2006 Jan 02")
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

	var keys []string
	for k := range userActivities {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	c.Set("user_activities", userActivities)
	c.Set("activities_keys", keys)

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["page_participants_title"] = "/participants/index"
	breadcrumbMap["breadcrumb_enrol_participant"] = ""
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("participants/detail.html"))
}
