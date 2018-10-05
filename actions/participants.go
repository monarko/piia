package actions

import (
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
	if user.Admin {
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
	breadcrumbMap["Participants"] = "/participants/index"
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
	breadcrumbMap["Participants"] = "/participants/index"
	breadcrumbMap["Enrol Participants"] = "/participants/create"
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
	breadcrumbMap["Participants"] = "/participants/index"
	breadcrumbMap["Update Participants"] = "/participants/edit"
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
		breadcrumbMap["Participants"] = "/participants/index"
		breadcrumbMap["Update Participants"] = "/participants/edit"
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
	return c.Render(200, r.HTML("participants/detail.html"))
}
