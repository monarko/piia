package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// OverReadingsIndex gets all OverReadings. This function is mapped to the path
func OverReadingsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("participant", participant)

	overReadings := &models.OverReadings{}
	q := tx.Eager("OverReader").Where("participant_id = ?", c.Param("pid")).PaginateFromParams(c.Params()).Order("created_at ASC")
	// Retrieve all OverReadings from the DB
	if err := q.All(overReadings); err != nil {
		return errors.WithStack(err)
	}
	c.Set("overReadings", overReadings)

	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Participants"] = "/participants/index"
	breadcrumbMap["Over Readings"] = "/participants/" + c.Param("pid") + "/overreadings/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("over_readings/index.html"))
}

// OverReadingsCreateGet renders the form for creating a new OverReading.
func OverReadingsCreateGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager("Screenings").Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	screening := participant.Screenings[0]
	c.Set("participant", participant)
	c.Set("screening", screening)
	c.Set("overReading", &models.OverReading{})

	// images
	leftEye := "https://ivmartel.github.io/dwv-jqmobile/demo/stable/index.html?input=https://upload.wikimedia.org/wikipedia/commons/7/7f/Brain_MRI_112010_rgbca.png"
	rightEye := "https://ivmartel.github.io/dwv-jqmobile/demo/stable/index.html?input=https://raw.githubusercontent.com/ivmartel/dwv/master/tests/data/bbmri-53323851.dcm"

	c.Set("leftEyeLink", leftEye)
	c.Set("rightEyeLink", rightEye)

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Participants"] = "/participants/index"
	breadcrumbMap["Over Readings"] = "/participants/" + c.Param("pid") + "/overreadings/index"
	breadcrumbMap["New Over Reading"] = "/participants/" + c.Param("pid") + "/overreadings/create"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("over_readings/create.html"))
}

// OverReadingsCreatePost renders the form for creating a new OverReading.
func OverReadingsCreatePost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager("Screenings").Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	screening := participant.Screenings[0]
	user := c.Value("current_user").(*models.User)
	overReading := &models.OverReading{}

	if err := c.Bind(overReading); err != nil {
		return errors.WithStack(err)
	}

	overReading.OverReaderID = user.ID
	overReading.ParticipantID = participant.ID

	// images
	leftEye := c.Param("leftEyeLink")
	rightEye := c.Param("rightEyeLink")

	c.Set("leftEyeLink", leftEye)
	c.Set("rightEyeLink", rightEye)

	verrs, err := tx.ValidateAndCreate(overReading)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("participant", participant)
		c.Set("screening", screening)
		c.Set("overReading", overReading)
		c.Set("errors", verrs.Errors)
		breadcrumbMap := make(map[string]interface{})
		breadcrumbMap["Participants"] = "/participants/index"
		breadcrumbMap["Over Readings"] = "/participants/" + c.Param("pid") + "/overreadings/index"
		breadcrumbMap["New Over Reading"] = "/participants/" + c.Param("pid") + "/overreadings/create"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("over_readings/create.html"))
	}

	participant.Status = "111"
	perrs, err := tx.ValidateAndUpdate(participant)
	if err != nil {
		return errors.WithStack(err)
	}
	if perrs.HasAny() {
		c.Set("participant", participant)
		c.Set("screening", screening)
		c.Set("overReading", overReading)
		c.Set("errors", verrs.Errors)
		breadcrumbMap := make(map[string]interface{})
		breadcrumbMap["Participants"] = "/participants/index"
		breadcrumbMap["Over Readings"] = "/participants/" + c.Param("pid") + "/overreadings/index"
		breadcrumbMap["New Over Reading"] = "/participants/" + c.Param("pid") + "/overreadings/create"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("over_readings/create.html"))
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "New over reading added successfully.")

	return c.Redirect(302, "/participants/"+c.Param("pid")+"/overreadings/index")
}
