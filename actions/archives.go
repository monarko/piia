package actions

import (
	"encoding/json"
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/monarko/piia/models"
)

// ArchiveIndex gets all Archives. This function is mapped to the path
// GET /archives
func ArchiveIndex(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	archives := &models.Archives{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	// Retrieve all Archives from the DB
	if err := q.Eager().All(archives); err != nil {
		return err
	}

	c.Set("archives", archives)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Archives"] = "/archives/index"
	c.Set("breadcrumbMap", breadcrumbMap)

	return c.Render(200, r.HTML("archives/index.html"))
}

// ArchiveShow gets the data for one Archive. This function is mapped to
// the path GET /archives/{archive_id}
func ArchiveShow(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Allocate an empty Archive
	archive := &models.Archive{}

	// To find the Archive the parameter archive_id is used.
	if err := tx.Find(archive, c.Param("aid")); err != nil {
		return c.Error(404, err)
	}

	c.Set("archive", archive)
	switch archive.ArchiveType {
	case "OverReading":
		o := models.OverReading{}
		json.Unmarshal(archive.Data, &o)
		c.Set("data", o)
	}

	return c.Render(200, r.HTML("archives/detail.html"))
}

// ArchiveCreatePost adds a Archive to the DB. This function is mapped to the
// path POST /archives
func ArchiveCreatePost(c buffalo.Context) error {
	// Allocate an empty Archive
	archive := &models.Archive{}

	// Bind archive to the html form elements
	if err := c.Bind(archive); err != nil {
		return err
	}

	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(archive)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		// Make the errors available inside the html template
		c.Set("errors", verrs)

		// Render again the new.html template that the user can
		// correct the input.
		return c.Render(422, r.Auto(c, archive))
	}

	// If there are no errors set a success message
	c.Flash().Add("success", T.Translate(c, "archive.created.success"))
	// and redirect to the archives index page
	return c.Render(201, r.Auto(c, archive))
}

// ArchiveDestroy deletes a Archive from the DB. This function is mapped
// to the path DELETE /archives/{archive_id}
func ArchiveDestroy(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Allocate an empty Archive
	archive := &models.Archive{}

	// To find the Archive the parameter archive_id is used.
	if err := tx.Find(archive, c.Param("archive_id")); err != nil {
		return c.Error(404, err)
	}

	if err := tx.Destroy(archive); err != nil {
		return err
	}

	// If there are no errors set a flash message
	c.Flash().Add("success", T.Translate(c, "archive.destroyed.success"))
	// Redirect to the archives index page
	return c.Render(200, r.Auto(c, archive))
}

// ArchiveMake functions
func ArchiveMake(c buffalo.Context, userID, modelID uuid.UUID, archiveType string, data interface{}) error {
	archive := &models.Archive{}
	archive.ArchiverID = userID
	archive.ArchiveType = archiveType
	archive.ModelID = modelID
	bt, err := json.Marshal(data)
	if err != nil {
		return err
	}
	archive.Data = bt

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}
	_, err = tx.ValidateAndCreate(archive)
	if err != nil {
		return err
	}

	return nil
}

// ArchiveRestore functions
func ArchiveRestore(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Allocate an empty Archive
	archive := &models.Archive{}

	// To find the Archive the parameter archive_id is used.
	if err := tx.Find(archive, c.Param("aid")); err != nil {
		return c.Error(404, err)
	}

	switch archive.ArchiveType {
	case "OverReading":
		o := &models.OverReading{}
		json.Unmarshal(archive.Data, o)
		cr := o.CreatedAt
		err := tx.Create(o)
		if err != nil {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(302, "/archives/index")
		}
		err = tx.RawQuery("UPDATE over_readings SET created_at = ? WHERE id = ?", cr, o.ID).Exec()
		if err != nil {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(302, "/archives/index")
		}
		prt := &models.Participant{}
		if err := tx.Find(prt, o.ParticipantID); err != nil {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(302, "/archives/index")
		}
		prt.Status = "111"
		perrs, err := tx.ValidateAndUpdate(prt)
		if err != nil {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(302, "/archives/index")
		}
		if perrs.HasAny() {
			c.Set("errors", perrs.Errors)
			return c.Redirect(302, "/archives/index")
		}
	}

	if err := tx.Destroy(archive); err != nil {
		c.Flash().Add("danger", err.Error())
		return c.Redirect(302, "/archives/index")
	}

	c.Flash().Add("success", "Archive restored successfully")

	return c.Redirect(302, "/archives/index")
}
