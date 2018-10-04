package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// SystemLogsIndex returns the logs list
func SystemLogsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	logs := &models.SystemLogs{}

	var q *pop.Query

	if len(c.Param("status")) > 0 {
		q = tx.Eager("User").Where("action = ?", c.Param("status")).PaginateFromParams(c.Params()).Order("created_at DESC")
	} else {
		q = tx.Eager("User").PaginateFromParams(c.Params()).Order("created_at DESC")
	}
	// Retrieve all Screenings from the DB
	if err := q.All(logs); err != nil {
		return errors.WithStack(err)
	}
	c.Set("logs", logs)
	c.Set("filterStatus", c.Params().Get("status"))

	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Logs"] = "/logs/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("system_logs/index.html"))
}

// InsertLog inserts a log into db
func InsertLog(action, activity, errorMessage, resourceID, resourceType string, userID uuid.UUID, c buffalo.Context) error {
	log := &models.SystemLog{}
	log.Action = action
	log.Activity = activity
	if len(errorMessage) > 0 {
		log.Error = true
		log.ErrorMessage = errorMessage
	}
	ip := c.Request().RemoteAddr
	log.ClientIP = "not found"
	if len(ip) > 0 {
		log.ClientIP = ip
	}
	log.UserID = userID
	log.ResourceID = resourceID
	log.ResourceType = resourceType

	tx := c.Value("tx").(*pop.Connection)
	_, err := tx.ValidateAndCreate(log)
	if err != nil {
		return err
	}

	return nil
}
