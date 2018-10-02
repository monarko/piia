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
	q := tx.Eager("Screener").Where("participant_id = ?", c.Param("pid")).PaginateFromParams(c.Params()).Order("created_at ASC")
	// Retrieve all Screenings from the DB
	if err := q.All(screenings); err != nil {
		return errors.WithStack(err)
	}
	c.Set("screenings", screenings)

	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Participants"] = "/participants/index"
	breadcrumbMap["Screenings"] = "/participants/" + c.Param("pid") + "/screenings/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("screenings/index.html"))
}

// ScreeningsCreateGet renders the form for creating a new Screening.
func ScreeningsCreateGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("participant", participant)
	c.Set("screening", &models.Screening{})
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Participants"] = "/participants/index"
	breadcrumbMap["Screenings"] = "/participants/" + c.Param("pid") + "/screenings/index"
	breadcrumbMap["New Screening"] = "/participants/" + c.Param("pid") + "/screenings/create"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("screenings/create.html"))
}

// ScreeningsCreatePost renders the form for creating a new Screening.
func ScreeningsCreatePost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	user := c.Value("current_user").(*models.User)
	screening := &models.Screening{}

	if err := c.Bind(screening); err != nil {
		return errors.WithStack(err)
	}

	screening.ScreenerID = user.ID
	screening.ParticipantID = participant.ID

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
		breadcrumbMap["Screenings"] = "/participants/" + c.Param("pid") + "/screenings/index"
		breadcrumbMap["New Screening"] = "/participants/" + c.Param("pid") + "/screenings/create"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("screenings/create.html"))
	}

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
		breadcrumbMap["Screenings"] = "/participants/" + c.Param("pid") + "/screenings/index"
		breadcrumbMap["New Screening"] = "/participants/" + c.Param("pid") + "/screenings/create"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("screenings/create.html"))
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "New screening added successfully.")

	return c.Redirect(302, "/participants/"+c.Param("pid")+"/screenings/index")
}

// Create adds a Screening to the DB. This function is mapped to the
// path POST /screenings
func Create(c buffalo.Context) error {
	// Allocate an empty Screening
	screening := &models.Screening{}

	// Bind screening to the html form elements
	if err := c.Bind(screening); err != nil {
		return errors.WithStack(err)
	}

	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(screening)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		// Make the errors available inside the html template
		c.Set("errors", verrs)

		// Render again the new.html template that the user can
		// correct the input.
		return c.Render(422, r.Auto(c, screening))
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "Screening was created successfully")

	// and redirect to the screenings index page
	return c.Render(201, r.Auto(c, screening))
}

// Edit renders a edit form for a Screening. This function is
// mapped to the path GET /screenings/{screening_id}/edit
func Edit(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	// Allocate an empty Screening
	screening := &models.Screening{}

	if err := tx.Find(screening, c.Param("screening_id")); err != nil {
		return c.Error(404, err)
	}

	return c.Render(200, r.Auto(c, screening))
}

// Update changes a Screening in the DB. This function is mapped to
// the path PUT /screenings/{screening_id}
func Update(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	// Allocate an empty Screening
	screening := &models.Screening{}

	if err := tx.Find(screening, c.Param("screening_id")); err != nil {
		return c.Error(404, err)
	}

	// Bind Screening to the html form elements
	if err := c.Bind(screening); err != nil {
		return errors.WithStack(err)
	}

	verrs, err := tx.ValidateAndUpdate(screening)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		// Make the errors available inside the html template
		c.Set("errors", verrs)

		// Render again the edit.html template that the user can
		// correct the input.
		return c.Render(422, r.Auto(c, screening))
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "Screening was updated successfully")

	// and redirect to the screenings index page
	return c.Render(200, r.Auto(c, screening))
}

// Destroy deletes a Screening from the DB. This function is mapped
// to the path DELETE /screenings/{screening_id}
func Destroy(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	// Allocate an empty Screening
	screening := &models.Screening{}

	// To find the Screening the parameter screening_id is used.
	if err := tx.Find(screening, c.Param("screening_id")); err != nil {
		return c.Error(404, err)
	}

	if err := tx.Destroy(screening); err != nil {
		return errors.WithStack(err)
	}

	// If there are no errors set a flash message
	c.Flash().Add("success", "Screening was destroyed successfully")

	// Redirect to the screenings index page
	return c.Render(200, r.Auto(c, screening))
}
