package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// ScreeningsIndex gets all Screenings. This function is mapped to the path
func ScreeningsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("participant", participant)

	screenings := &models.Screenings{}
	q := tx.Eager("Screener").Where("participant_id = ?", c.Param("pid")).PaginateFromParams(c.Params()).Order("created_at DESC")
	// Retrieve all Screenings from the DB
	if err := q.All(screenings); err != nil {
		return errors.WithStack(err)
	}
	c.Set("screenings", screenings)

	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Participants"] = "/participants/index"
	// breadcrumbMap["Screenings"] = "/participants/" + c.Param("pid") + "/screenings/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("screenings/index.html"))
}

// ScreeningsCreateGet renders the form for creating a new Screening.
func ScreeningsCreateGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager().Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	if len(participant.Screenings) > 0 {
		scr := participant.Screenings[0]
		red := "/participants/" + c.Param("pid") + "/screenings/edit/" + scr.ID.String()
		c.Redirect(302, red)
	}
	c.Set("participant", participant)
	c.Set("screening", &models.Screening{})
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Participants"] = "/participants/index"
	// breadcrumbMap["Screenings"] = "/participants/" + c.Param("pid") + "/screenings/index"
	breadcrumbMap["New Screening"] = "/participants/" + c.Param("pid") + "/screenings/create"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("screenings/create.html"))
}

// ScreeningsCreatePost renders the form for creating a new Screening.
func ScreeningsCreatePost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager().Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	if len(participant.Screenings) > 0 {
		scr := participant.Screenings[0]
		red := "/participants/" + c.Param("pid") + "/screenings/edit/" + scr.ID.String()
		c.Redirect(302, red)
	}
	user := c.Value("current_user").(*models.User)
	screening := &models.Screening{}
	oldScreening := screening.Maps()
	if err := c.Bind(screening); err != nil {
		return errors.WithStack(err)
	}

	screening.ScreenerID = user.ID
	screening.ParticipantID = participant.ID
	referral := c.Request().FormValue("referral")
	if referral == "yes" {
		screening.Referral.Referred = true
	}

	verrs, err := tx.ValidateAndCreate(screening)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("participant", participant)
		c.Set("screening", screening)
		c.Set("errors", verrs.Errors)
		breadcrumbMap := make(map[string]interface{})
		breadcrumbMap["Participants"] = "/participants/index"
		// breadcrumbMap["Screenings"] = "/participants/" + c.Param("pid") + "/screenings/index"
		breadcrumbMap["New Screening"] = "/participants/" + c.Param("pid") + "/screenings/create"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("screenings/create.html"))
	}

	if len(screening.Eyes.RightEye.VisualAcuity.String) > 0 && len(screening.Eyes.RightEye.DRGrading.String) > 0 && len(screening.Eyes.RightEye.DMEAssessment.String) > 0 && len(screening.Eyes.LeftEye.VisualAcuity.String) > 0 && len(screening.Eyes.LeftEye.DRGrading.String) > 0 && len(screening.Eyes.LeftEye.DMEAssessment.String) > 0 {
		participant.Status = "11"
		perrs, err := tx.ValidateAndUpdate(participant)
		if err != nil {
			return errors.WithStack(err)
		}
		if perrs.HasAny() {
			c.Set("participant", participant)
			c.Set("screening", screening)
			c.Set("errors", verrs.Errors)
			breadcrumbMap := make(map[string]interface{})
			breadcrumbMap["Participants"] = "/participants/index"
			// breadcrumbMap["Screenings"] = "/participants/" + c.Param("pid") + "/screenings/index"
			breadcrumbMap["New Screening"] = "/participants/" + c.Param("pid") + "/screenings/create"
			c.Set("breadcrumbMap", breadcrumbMap)
			return c.Render(422, r.HTML("screenings/create.html"))
		}
	}

	newScreening := screening.Maps()
	auditErr := MakeAudit("Screening", screening.ID, oldScreening, newScreening, user.ID, c)
	if auditErr != nil {
		return errors.WithStack(auditErr)
	}

	logErr := InsertLog("create", "User did a screening", "", screening.ID.String(), "screening", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "New screening added successfully.")

	return c.Redirect(302, "/participants/index")
}

// ScreeningsEditGet renders the form for creating a new Screening.
func ScreeningsEditGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("participant", participant)
	screening := &models.Screening{}
	if err := tx.Find(screening, c.Param("sid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("screening", screening)
	// statuses := screening.StatusesMap()
	// c.Set("screeningStatuses", statuses)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Participants"] = "/participants/index"
	breadcrumbMap["Edit Screening"] = "/participants/" + c.Param("pid") + "/screenings/edit"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("screenings/edit.html"))
}

// ScreeningsEditPost renders the form for creating a new Screening.
func ScreeningsEditPost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	user := c.Value("current_user").(*models.User)
	screening := &models.Screening{}
	if err := tx.Find(screening, c.Param("sid")); err != nil {
		return c.Error(404, err)
	}
	oldScreening := screening.Maps()
	if err := c.Bind(screening); err != nil {
		return errors.WithStack(err)
	}
	// screening.ScreenerID = user.ID
	// screening.ParticipantID = participant.ID
	referral := c.Request().FormValue("referral")
	if referral == "yes" {
		screening.Referral.Referred = true
	}

	verrs, err := tx.ValidateAndUpdate(screening)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("participant", participant)
		c.Set("screening", screening)
		c.Set("errors", verrs.Errors)
		breadcrumbMap := make(map[string]interface{})
		breadcrumbMap["Participants"] = "/participants/index"
		breadcrumbMap["Edit Screening"] = "/participants/" + c.Param("pid") + "/screenings/edit"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("screenings/edit.html"))
	}

	if participant.Status == "1" && len(screening.Eyes.RightEye.VisualAcuity.String) > 0 && len(screening.Eyes.RightEye.DRGrading.String) > 0 && len(screening.Eyes.RightEye.DMEAssessment.String) > 0 && len(screening.Eyes.LeftEye.VisualAcuity.String) > 0 && len(screening.Eyes.LeftEye.DRGrading.String) > 0 && len(screening.Eyes.LeftEye.DMEAssessment.String) > 0 {
		participant.Status = "11"
		perrs, err := tx.ValidateAndUpdate(participant)
		if err != nil {
			return errors.WithStack(err)
		}
		if perrs.HasAny() {
			c.Set("participant", participant)
			c.Set("screening", screening)
			c.Set("errors", verrs.Errors)
			breadcrumbMap := make(map[string]interface{})
			breadcrumbMap["Participants"] = "/participants/index"
			breadcrumbMap["Edit Screening"] = "/participants/" + c.Param("pid") + "/screenings/edit"
			c.Set("breadcrumbMap", breadcrumbMap)
			return c.Render(422, r.HTML("screenings/edit.html"))
		}
	}

	newScreening := screening.Maps()
	auditErr := MakeAudit("Screening", screening.ID, oldScreening, newScreening, user.ID, c)
	if auditErr != nil {
		return errors.WithStack(auditErr)
	}

	logErr := InsertLog("update", "User updated a screening", "", screening.ID.String(), "screening", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "Screening updated successfully.")

	return c.Redirect(302, "/participants/index")
}
