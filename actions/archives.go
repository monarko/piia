package actions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

var (
	types = map[string]map[string]string{
		"OverReading":     map[string]string{"log": "overReading", "table": "over_readings"},
		"ReferredMessage": map[string]string{"log": "referred_message", "table": "referred_messages"},
		"Notification":    map[string]string{"log": "notification", "table": "notifications"},
		"Screening":       map[string]string{"log": "screening", "table": "screenings"},
		"Participant":     map[string]string{"log": "participant", "table": "participants"},
	}
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
	// q := tx.PaginateFromParams(c.Params())

	// Retrieve all Archives from the DB
	if err := tx.Eager().All(archives); err != nil {
		return err
	}

	c.Set("archives", archives)
	// Add the paginator to the context so it can be used in the template.
	// c.Set("pagination", q.Paginator)

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
	if err := tx.Eager().Find(archive, c.Param("aid")); err != nil {
		return c.Error(404, err)
	}

	c.Set("archive", archive)
	switch archive.ArchiveType {
	case "OverReading":
		o := models.OverReading{}
		json.Unmarshal(archive.Data, &o)
		c.Set("data", o)
	case "Notification":
		o := models.Notification{}
		json.Unmarshal(archive.Data, &o)
		c.Set("data", o)
	case "ReferredMessage":
		o := models.ReferredMessage{}
		json.Unmarshal(archive.Data, &o)
		c.Set("data", o)
	case "Screening":
		o := models.Screening{}
		json.Unmarshal(archive.Data, &o)
		c.Set("data", o)
	case "Participant":
		o := models.Participant{}
		json.Unmarshal(archive.Data, &o)
		c.Set("data", o)
	}

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Archives"] = "/archives/index"
	breadcrumbMap["Archive Detail"] = "#"

	c.Set("breadcrumbMap", breadcrumbMap)

	return c.Render(200, r.HTML("archives/detail.html"))
}

// ArchiveDestroy deletes a Archive from the DB. This function is mapped
// to the path DELETE /archives/{archive_id}
func ArchiveDestroy(c buffalo.Context) error {
	returnURL := "/archives/index"
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}
	user := c.Value("current_user").(*models.User)

	// Allocate an empty Archive
	archive := &models.Archive{}

	// To find the Archive the parameter archive_id is used.
	if err := tx.Find(archive, c.Param("aid")); err != nil {
		return c.Error(404, err)
	}

	reason := c.Request().FormValue("reason")
	aType := archive.ArchiveType
	if err := tx.Destroy(archive); err != nil {
		c.Flash().Add("danger", err.Error())
		return c.Redirect(302, returnURL)
	}

	logErr := InsertLog("delete", "Archive ("+aType+") deleted PERMANENTLY, reason: "+reason, "", c.Param("aid"), "archive", user.ID, c)
	if logErr != nil {
		c.Flash().Add("danger", logErr.Error())
		return c.Redirect(302, returnURL)
	}

	// If there are no errors set a flash message
	c.Flash().Add("success", T.Translate(c, "archive.destroyed.success"))
	// Redirect to the archives index page
	return c.Redirect(302, returnURL)
}

// ArchiveMake functions
func ArchiveMake(c buffalo.Context, userID, modelID uuid.UUID, archiveType string, data interface{}, reason string) error {
	archive := &models.Archive{}
	archive.ArchiverID = userID
	archive.ArchiveType = archiveType
	archive.ModelID = modelID
	archive.Reason = reason
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

	logErr := InsertLog("archive", "Archive created ("+archiveType+")", "", modelID.String(), types[archiveType]["log"], userID, c)
	if logErr != nil {
		return logErr
	}
	id := modelID
	if err := tx.Destroy(data); err != nil {
		return err
	}

	logErr = InsertLog("delete", archiveType+" deleted, reason: "+reason, "", id.String(), types[archiveType]["log"], userID, c)
	if logErr != nil {
		return logErr
	}

	return nil
}

// ArchiveRestore functions
func ArchiveRestore(c buffalo.Context) error {
	returnURL := "/archives/index"
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}
	user := c.Value("current_user").(*models.User)

	// Allocate an empty Archive
	archive := &models.Archive{}
	// To find the Archive the parameter archive_id is used.
	if err := tx.Find(archive, c.Param("aid")); err != nil {
		return c.Error(404, err)
	}

	all, res := checkForDependency(archive.Dependency, tx)
	if !all {
		message := "Root elements not found: "
		t := make([]string, 0)
		for k, v := range res {
			if !v {
				t = append(t, k)
			}
		}
		message += strings.Join(t, ", ")
		c.Flash().Add("danger", message)
		return c.Redirect(302, returnURL)
	}

	err := restoreModel(tx, archive, user.ID, c)
	if err != nil {
		c.Flash().Add("danger", err.Error())
		return c.Redirect(302, returnURL)
	}

	c.Flash().Add("success", "Archive restored successfully")

	return c.Redirect(302, returnURL)
}

func restoreModel(tx *pop.Connection, archive *models.Archive, userID uuid.UUID, c buffalo.Context) error {
	switch archive.ArchiveType {
	case "OverReading":
		o := &models.OverReading{}
		json.Unmarshal(archive.Data, o)
		err := tx.Create(o)
		if err != nil {
			return err
		}
		prt := &models.Participant{}
		if err := tx.Find(prt, o.ParticipantID); err != nil {
			return err
		}
		prt.Status = "111"
		perrs, err := tx.ValidateAndUpdate(prt)
		if err != nil {
			return err
		}
		if perrs.HasAny() {
			msg := ""
			for k, v := range perrs.Errors {
				msg += k + ": "
				msg += strings.Join(v, ", ")
			}
			return errors.New(msg)
		}
	case "ReferredMessage":
		o := &models.ReferredMessage{}
		json.Unmarshal(archive.Data, o)
		err := tx.Create(o)
		if err != nil {
			return err
		}
	case "Notification":
		o := &models.Notification{}
		json.Unmarshal(archive.Data, o)
		err := tx.Create(o)
		if err != nil {
			return err
		}
	case "Screening":
		o := &models.Screening{}
		json.Unmarshal(archive.Data, o)
		err := tx.Create(o)
		if err != nil {
			return err
		}
		prt := &models.Participant{}
		if err := tx.Find(prt, o.ParticipantID); err != nil {
			return err
		}
		prt.Status = "11"
		perrs, err := tx.ValidateAndUpdate(prt)
		if err != nil {
			return err
		}
		if perrs.HasAny() {
			msg := ""
			for k, v := range perrs.Errors {
				msg += k + ": "
				msg += strings.Join(v, ", ")
			}
			return errors.New(msg)
		}
	case "Participant":
		o := &models.Participant{}
		json.Unmarshal(archive.Data, o)
		err := tx.Create(o)
		if err != nil {
			return err
		}
	}

	if err := updateCreatedTime(tx, *archive); err != nil {
		return err
	}

	id := archive.ModelID
	idType := types[archive.ArchiveType]

	if err := tx.Destroy(archive); err != nil {
		return err
	}

	logErr := InsertLog("restore", "Archive restored ("+archive.ArchiveType+")", "", id.String(), idType["log"], userID, c)
	if logErr != nil {
		return logErr
	}

	return nil
}

func checkForDependency(d models.Mapping, tx *pop.Connection) (bool, map[string]bool) {
	res := make(map[string]bool)
	all := true

	for k, v := range d {
		res[k] = false
		temp := v.(map[string]interface{})
		current := false
		for _, id := range temp {
			switch k {
			case "Participant":
				prt := &models.Participant{}
				if err := tx.Find(prt, id); err == nil {
					res[k] = true
					current = true
				}
			case "Screening":
				prt := &models.Screening{}
				if err := tx.Find(prt, id); err == nil {
					res[k] = true
					current = true
				}
			case "User":
				prt := &models.User{}
				if err := tx.Find(prt, id); err == nil {
					res[k] = true
					current = true
				}
			}
			all = all && current
		}
	}

	return all, res
}

func updateCreatedTime(tx *pop.Connection, a models.Archive) error {
	table := types[a.ArchiveType]["table"]
	id := a.ModelID
	data := make(map[string]interface{})
	json.Unmarshal(a.Data, &data)
	err := tx.RawQuery("UPDATE "+table+" SET created_at = ? WHERE id = ?", data["created_at"], id).Exec()
	if err != nil {
		return err
	}
	return nil
}
